package middleware

import (
	"context"
	"net/http"
	"strings"

	"banca-en-linea/backend/internal/auth"
)

// ContextKey es el tipo para las claves del contexto
type ContextKey string

const (
	// UserContextKey es la clave para almacenar información del usuario en el contexto
	UserContextKey ContextKey = "user"
)

// AuthMiddleware crea un middleware de autenticación
func AuthMiddleware(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Obtener el token del header Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Verificar que el header tenga el formato "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token := parts[1]

			// Validar el token
			claims, err := authService.ValidateToken(token)
			if err != nil {
				switch err {
				case auth.ErrTokenExpired:
					http.Error(w, "Token expired", http.StatusUnauthorized)
				case auth.ErrInvalidToken:
					http.Error(w, "Invalid token", http.StatusUnauthorized)
				default:
					http.Error(w, "Token validation failed", http.StatusUnauthorized)
				}
				return
			}

			// Agregar la información del usuario al contexto
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			r = r.WithContext(ctx)

			// Continuar con el siguiente handler
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext extrae la información del usuario del contexto
func GetUserFromContext(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*auth.Claims)
	return claims, ok
}

// OptionalAuthMiddleware es un middleware que permite tanto requests autenticados como no autenticados
func OptionalAuthMiddleware(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			
			// Si no hay header de autorización, continuar sin autenticación
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Si hay header, intentar validar el token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				token := parts[1]
				if claims, err := authService.ValidateToken(token); err == nil {
					// Token válido, agregar al contexto
					ctx := context.WithValue(r.Context(), UserContextKey, claims)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}