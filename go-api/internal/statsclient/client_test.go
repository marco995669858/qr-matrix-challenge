package statsclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestComputeStatsSendsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/stats" {
			t.Errorf("petición inesperada: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer abc" {
			t.Errorf("token no reenviado: %q", r.Header.Get("Authorization"))
		}
		var body struct {
			Matrices []NamedMatrix `json:"matrices"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Matrices) != 1 {
			t.Errorf("cuerpo inesperado: %v %+v", err, body)
		}
		w.Write([]byte(`{"max": 4}`))
	}))
	defer server.Close()

	got, err := New(server.URL+"/").ComputeStats(context.Background(), "abc", []NamedMatrix{{Name: "Q", Data: [][]float64{{4}}}})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"max": 4}` {
		t.Fatalf("respuesta inesperada: %s", got)
	}
}

func TestComputeStatsPropagatesErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"token inválido"}`, http.StatusUnauthorized)
	}))
	defer server.Close()

	if _, err := New(server.URL).ComputeStats(context.Background(), "x", nil); err == nil {
		t.Fatal("se esperaba un error con status 401")
	}
}
