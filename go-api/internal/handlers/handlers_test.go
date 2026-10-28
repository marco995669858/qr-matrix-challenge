package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"matrix-qr-api/internal/auth"
	"matrix-qr-api/internal/config"
	"matrix-qr-api/internal/statsclient"
)

var testCfg = config.Config{
	JWTSecret:    "test-secret",
	TokenTTL:     time.Minute,
	AuthUser:     "admin",
	AuthPassword: "secret",
}

// fakeStats simula la API de Node y guarda lo que recibió.
type fakeStats struct {
	gotToken    string
	gotMatrices []statsclient.NamedMatrix
	err         error
}

func (f *fakeStats) ComputeStats(_ context.Context, token string, m []statsclient.NamedMatrix) (json.RawMessage, error) {
	f.gotToken, f.gotMatrices = token, m
	if f.err != nil {
		return nil, f.err
	}
	return json.RawMessage(`{"max": 1}`), nil
}

func doRequest(t *testing.T, stats StatsComputer, method, path, token string, body any) (*http.Response, map[string]any) {
	t.Helper()
	app := NewApp(testCfg, stats)

	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&decoded)
	return resp, decoded
}

func validToken(t *testing.T) string {
	token, err := auth.GenerateToken(testCfg.JWTSecret, "admin", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestLogin(t *testing.T) {
	resp, body := doRequest(t, &fakeStats{}, "POST", "/auth/login", "", LoginRequest{"admin", "secret"})
	if resp.StatusCode != 200 || body["token"] == "" {
		t.Fatalf("status %d, body %v", resp.StatusCode, body)
	}

	resp, _ = doRequest(t, &fakeStats{}, "POST", "/auth/login", "", LoginRequest{"admin", "mala"})
	if resp.StatusCode != 401 {
		t.Fatalf("se esperaba 401, llegó %d", resp.StatusCode)
	}
}

func TestQRRequiresToken(t *testing.T) {
	resp, _ := doRequest(t, &fakeStats{}, "POST", "/api/qr", "", qrBody([][]float64{{1}}))
	if resp.StatusCode != 401 {
		t.Fatalf("se esperaba 401, llegó %d", resp.StatusCode)
	}

	resp, _ = doRequest(t, &fakeStats{}, "POST", "/api/qr", "token-falso", qrBody([][]float64{{1}}))
	if resp.StatusCode != 401 {
		t.Fatalf("se esperaba 401, llegó %d", resp.StatusCode)
	}
}

func TestQRSuccessForwardsMatricesAndToken(t *testing.T) {
	stats := &fakeStats{}
	token := validToken(t)

	resp, body := doRequest(t, stats, "POST", "/api/qr", token, qrBody([][]float64{{1, 2}, {3, 4}}))
	if resp.StatusCode != 200 {
		t.Fatalf("status %d, body %v", resp.StatusCode, body)
	}
	for _, key := range []string{"q", "r", "stats"} {
		if body[key] == nil {
			t.Fatalf("falta %q en la respuesta: %v", key, body)
		}
	}
	if stats.gotToken != token {
		t.Fatal("el token no se reenvió a la API de estadísticas")
	}
	if len(stats.gotMatrices) != 2 || stats.gotMatrices[0].Name != "Q" || stats.gotMatrices[1].Name != "R" {
		t.Fatalf("matrices enviadas inesperadas: %+v", stats.gotMatrices)
	}
}

func TestQRInvalidMatrix(t *testing.T) {
	resp, body := doRequest(t, &fakeStats{}, "POST", "/api/qr", validToken(t), qrBody([][]float64{{1, 2}, {3}}))
	if resp.StatusCode != 400 {
		t.Fatalf("se esperaba 400, llegó %d (%v)", resp.StatusCode, body)
	}

	resp, _ = doRequest(t, &fakeStats{}, "POST", "/api/qr", validToken(t), map[string]any{"matrix": "hola"})
	if resp.StatusCode != 400 {
		t.Fatalf("se esperaba 400 para tipo inválido, llegó %d", resp.StatusCode)
	}
}

func TestQRStatsUnavailable(t *testing.T) {
	stats := &fakeStats{err: errors.New("caída")}
	resp, _ := doRequest(t, stats, "POST", "/api/qr", validToken(t), qrBody([][]float64{{1}}))
	if resp.StatusCode != 502 {
		t.Fatalf("se esperaba 502, llegó %d", resp.StatusCode)
	}
}

// qrBody arma el cuerpo de POST /api/qr.
func qrBody(m [][]float64) map[string]any {
	return map[string]any{"matrix": m}
}

// Un null no debe tratarse como 0 (encoding/json lo haría con float64).
func TestQRRejectsNull(t *testing.T) {
	resp, body := doRequest(t, &fakeStats{}, "POST", "/api/qr", validToken(t), map[string]any{"matrix": []any{[]any{1, nil}}})
	if resp.StatusCode != 400 {
		t.Fatalf("se esperaba 400, llegó %d (%v)", resp.StatusCode, body)
	}
}
