// Package statsclient es el cliente HTTP hacia la API de estadísticas (Node.js).
package statsclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// NamedMatrix es el formato que espera la API de Node.
type NamedMatrix struct {
	Name string      `json:"name"`
	Data [][]float64 `json:"data"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		// Timeout amplio: en planes gratuitos de la nube el servicio puede
		// estar "dormido" y tardar en despertar.
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// ComputeStats envía las matrices a POST /api/stats reenviando el JWT del
// usuario. Devuelve la respuesta tal cual (json.RawMessage) para no
// duplicar en Go la estructura que define la API de Node.
func (c *Client) ComputeStats(ctx context.Context, token string, matrices []NamedMatrix) (json.RawMessage, error) {
	body, err := json.Marshal(map[string]any{"matrices": matrices})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/stats", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("no se pudo contactar a la API de estadísticas: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("la API de estadísticas respondió %d: %s", resp.StatusCode, respBody)
	}
	return respBody, nil
}
