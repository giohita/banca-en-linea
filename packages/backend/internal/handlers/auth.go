package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"banca-en-linea/backend/internal/auth"
	"banca-en-linea/backend/internal/db"
	"banca-en-linea/backend/models"
)

// AuthHandler maneja las operaciones de autenticación
type AuthHandler struct {
	userService *db.UserService
	authService *auth.Service
}

// NewAuthHandler crea una nueva instancia del handler de autenticación
func NewAuthHandler(userService *db.UserService, authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		authService: authService,
	}
}

// RegisterResponse representa la respuesta del registro
type RegisterResponse struct {
	User  models.UserResponse `json:"user"`
	Token string              `json:"token"`
}

// LoginResponse representa la respuesta del login
type LoginResponse struct {
	User  models.UserResponse `json:"user"`
	Token string              `json:"token"`
}

// Register maneja el registro de nuevos usuarios
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validar que los campos requeridos estén presentes
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		http.Error(w, "Email, password, first name, and last name are required", http.StatusBadRequest)
		return
	}

	// Validar longitud mínima de contraseña
	if len(req.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	// Crear usuario con cuenta TigerBeetle
	user, err := h.userService.CreateUserWithAccount(&req)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		if err.Error() == "user already exists" {
			http.Error(w, "User with this email already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Generar token JWT
	token, err := h.authService.GenerateToken(user)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		http.Error(w, "Error generating authentication token", http.StatusInternalServerError)
		return
	}

	// Responder con el usuario y token
	response := RegisterResponse{
		User:  user.ToResponse(),
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Login maneja el inicio de sesión de usuarios
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validar que los campos requeridos estén presentes
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Obtener usuario por email
	user, err := h.userService.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("Error getting user by email: %v", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Verificar que el usuario esté activo
	if !user.IsActive {
		http.Error(w, "Account is deactivated", http.StatusUnauthorized)
		return
	}

	// Verificar contraseña
	if err := h.authService.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		log.Printf("Invalid password for user %s", req.Email)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generar token JWT
	token, err := h.authService.GenerateToken(user)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		http.Error(w, "Error generating authentication token", http.StatusInternalServerError)
		return
	}

	// Responder con el usuario y token
	response := LoginResponse{
		User:  user.ToResponse(),
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Logout maneja el cierre de sesión (en este caso, simplemente confirma el logout)
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// En una implementación JWT stateless, el logout se maneja en el cliente
	// eliminando el token. Aquí simplemente confirmamos el logout.

	// En el futuro, se podría implementar una blacklist de tokens
	// o usar refresh tokens para un control más granular.

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

// Me retorna la información del usuario autenticado
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	// Obtener claims del contexto (agregado por el middleware de auth)
	claims, ok := r.Context().Value("user").(*auth.Claims)
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	// Obtener información actualizada del usuario
	user, err := h.userService.GetUser(claims.UserID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		http.Error(w, "Error getting user information", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user.ToResponse())
}
