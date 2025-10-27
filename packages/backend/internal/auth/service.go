package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"banca-en-linea/backend/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired      = errors.New("token expired")
	ErrInvalidToken      = errors.New("invalid token")
)

// Claims representa los claims del JWT
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

// Service maneja la autenticación y autorización
type Service struct {
	jwtSecret []byte
}

// NewService crea una nueva instancia del servicio de autenticación
func NewService() *Service {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// En desarrollo, usar un secreto por defecto (NO hacer esto en producción)
		secret = "your-super-secret-jwt-key-change-this-in-production"
	}
	
	return &Service{
		jwtSecret: []byte(secret),
	}
}

// HashPassword hashea una contraseña usando bcrypt
func (s *Service) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// VerifyPassword verifica si una contraseña coincide con su hash
func (s *Service) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GenerateToken genera un JWT token para un usuario
func (s *Service) GenerateToken(user *models.User) (string, error) {
	// Token expira en 24 horas
	expirationTime := time.Now().Add(24 * time.Hour)
	
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "banca-en-linea",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken valida un JWT token y retorna los claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verificar que el método de firma sea el esperado
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// AuthenticateUser autentica un usuario con email y contraseña
func (s *Service) AuthenticateUser(email, password, hashedPassword string) error {
	if err := s.VerifyPassword(hashedPassword, password); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}