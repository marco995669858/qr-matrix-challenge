// Package handlers define las rutas HTTP de la API de Go.
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"matrix-qr-api/internal/auth"
	"matrix-qr-api/internal/config"
	"matrix-qr-api/internal/qr"
	"matrix-qr-api/internal/statsclient"
)

// StatsComputer abstrae la llamada a la API de Node. Gracias a esta
// interfaz, los tests pueden usar un servidor falso.
type StatsComputer interface {
	ComputeStats(ctx context.Context, token string, matrices []statsclient.NamedMatrix) (json.RawMessage, error)
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expiresIn"` // segundos
}

// QRRequest usa *float64 para distinguir un null de un 0: Go deja un null
// como 0 en un float64, y así un valor inválido pasaría sin error.
type QRRequest struct {
	Matrix [][]*float64 `json:"matrix"`
}

// toMatrix convierte la entrada a [][]float64 rechazando valores null.
func (r QRRequest) toMatrix() ([][]float64, error) {
	out := make([][]float64, len(r.Matrix))
	for i, row := range r.Matrix {
		out[i] = make([]float64, len(row))
		for j, v := range row {
			if v == nil {
				return nil, fmt.Errorf("el elemento [%d][%d] es null; todos deben ser números", i, j)
			}
			out[i][j] = *v
		}
	}
	return out, nil
}

type QRResponse struct {
	Q     [][]float64     `json:"q"`
	R     [][]float64     `json:"r"`
	Stats json.RawMessage `json:"stats"`
}

// NewApp arma la aplicación Fiber con middlewares y rutas.
func NewApp(cfg config.Config, stats StatsComputer) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "matrix-qr-api",
		ErrorHandler: jsonErrorHandler,
		BodyLimit:    1 << 20, // 1 MB
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	// Frontend estático, servido desde el mismo origen.
	app.Static("/", "./web")

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "matrix-qr-api"})
	})

	// Límite de intentos de login por IP para frenar fuerza bruta.
	app.Post("/auth/login", limiter.New(limiter.Config{Max: 10, Expiration: time.Minute}), login(cfg))

	api := app.Group("/api", auth.Middleware(cfg.JWTSecret))
	api.Post("/qr", factorize(stats))

	return app
}

// login: POST /auth/login → devuelve un JWT si las credenciales son correctas.
func login(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido: se espera {\"username\", \"password\"}")
		}
		if err := auth.CheckCredentials(req.Username, req.Password, cfg.AuthUser, cfg.AuthPassword); err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}
		token, err := auth.GenerateToken(cfg.JWTSecret, req.Username, cfg.TokenTTL)
		if err != nil {
			return err
		}
		return c.JSON(LoginResponse{Token: token, ExpiresIn: int(cfg.TokenTTL.Seconds())})
	}
}

// factorize: POST /api/qr → factoriza la matriz, pide las estadísticas a la
// API de Node y devuelve todo en una sola respuesta.
func factorize(stats StatsComputer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req QRRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido: se espera {\"matrix\": [[números]]}")
		}

		matrix, err := req.toMatrix()
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		result, err := qr.Factorize(matrix)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		token, _ := c.Locals(auth.TokenLocalsKey).(string)
		statsJSON, err := stats.ComputeStats(c.UserContext(), token, []statsclient.NamedMatrix{
			{Name: "Q", Data: result.Q},
			{Name: "R", Data: result.R},
		})
		if err != nil {
			log.Printf("error en la API de estadísticas: %v", err)
			return fiber.NewError(fiber.StatusBadGateway, "la API de estadísticas no está disponible")
		}

		return c.JSON(QRResponse{Q: result.Q, R: result.R, Stats: statsJSON})
	}
}

// jsonErrorHandler responde todos los errores como {"error": "..."}.
func jsonErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "error interno"

	var fe *fiber.Error
	if errors.As(err, &fe) {
		code = fe.Code
		message = fe.Message
	} else {
		log.Printf("error no controlado: %v", err)
	}
	return c.Status(code).JSON(fiber.Map{"error": message})
}
