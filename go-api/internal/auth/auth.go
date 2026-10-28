// Package auth emite y valida JWT (HS256).
//
// La API de Go es la "puerta de entrada": emite el token en /auth/login y
// lo reenvía a la API de Node, que lo valida con el mismo secreto.
package auth

import (
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// TokenLocalsKey es la clave donde el middleware guarda el token crudo,
// para reenviarlo a la API de estadísticas.
const TokenLocalsKey = "token"

// GenerateToken crea un JWT firmado para el usuario indicado.
func GenerateToken(secret, subject string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// ValidateToken verifica firma, algoritmo y expiración.
func ValidateToken(secret, token string) error {
	_, err := jwt.Parse(token,
		func(*jwt.Token) (any, error) { return []byte(secret), nil },
		// Fijar el algoritmo evita ataques de confusión ("alg": "none", etc.).
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	return err
}

// Middleware exige "Authorization: Bearer <token>" válido.
func Middleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, found := strings.CutPrefix(c.Get(fiber.HeaderAuthorization), "Bearer ")
		if !found || token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "falta el token (Authorization: Bearer TOKEN)")
		}
		if err := ValidateToken(secret, token); err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "token inválido o expirado")
		}
		c.Locals(TokenLocalsKey, token)
		return c.Next()
	}
}

// CheckCredentials compara en tiempo constante para no filtrar
// información por diferencias de tiempo de respuesta.
func CheckCredentials(user, password, wantUser, wantPassword string) error {
	userOK := subtle.ConstantTimeCompare([]byte(user), []byte(wantUser)) == 1
	passOK := subtle.ConstantTimeCompare([]byte(password), []byte(wantPassword)) == 1
	if !userOK || !passOK {
		return errors.New("credenciales inválidas")
	}
	return nil
}
