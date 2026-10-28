// Package config lee la configuración desde variables de entorno
package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	Port         string        // puerto HTTP
	JWTSecret    string        // secreto HS256 compartido con la API de Node
	TokenTTL     time.Duration // duración de los tokens emitidos
	StatsAPIURL  string        // URL base de la API de Node, p. ej. http://node-api:3001
	AuthUser     string        // usuario para POST /auth/login
	AuthPassword string        // contraseña para POST /auth/login
}

func Load() (Config, error) {
	cfg := Config{
		Port:         getEnv("PORT", "3000"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		TokenTTL:     time.Hour,
		StatsAPIURL:  getEnv("STATS_API_URL", "http://localhost:3001"),
		AuthUser:     os.Getenv("AUTH_USER"),
		AuthPassword: os.Getenv("AUTH_PASSWORD"),
	}
	if cfg.JWTSecret == "" || cfg.AuthUser == "" || cfg.AuthPassword == "" {
		return cfg, errors.New("faltan variables de entorno: JWT_SECRET, AUTH_USER y AUTH_PASSWORD son obligatorias")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
