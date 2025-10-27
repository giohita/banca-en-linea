package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"banca-en-linea/backend/database"
	"banca-en-linea/backend/internal/auth"
	"banca-en-linea/backend/internal/db"
	"banca-en-linea/backend/internal/handlers"
	"banca-en-linea/backend/internal/middleware"
	// "banca-en-linea/backend/internal/tigerbeetle" // Comentado temporalmente
	"banca-en-linea/backend/models"
)

type Server struct {
	userService *db.UserService
	// tigerBeetleClient *tigerbeetle.Client // Comentado temporalmente
	authService *auth.Service
	authHandler *handlers.AuthHandler
}

func main() {
	// Configurar logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Iniciando servidor backend...")

	// Obtener configuración de la base de datos
	config := database.GetConfigFromEnv()
	log.Printf("Conectando a la base de datos: %s@%s:%s/%s",
		config.User, config.Host, config.Port, config.DBName)

	// Conectar a la base de datos
	dbConn, err := database.Connect(config)
	if err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}
	defer dbConn.Close()

	// Ejecutar migraciones
	log.Println("Ejecutando migraciones...")
	if err := database.RunMigrations(dbConn, "./migrations"); err != nil {
		log.Fatalf("Error ejecutando migraciones: %v", err)
	}

	// Inicializar TigerBeetle service (usando stub para desarrollo)
	// log.Println("Inicializando servicio TigerBeetle...")
	// tbService := tigerbeetle.NewServiceStub()
	// defer tbService.Close()

	// Inicializar cuentas maestras de TigerBeetle
	// if err := tbService.InitializeMasterAccounts(); err != nil {
	//	log.Fatalf("Error inicializando cuentas maestras TigerBeetle: %v", err)
	// }

	// Crear repositorio y servicio de usuarios
	userRepo := db.NewUserRepository(dbConn)
	userService := db.NewUserService(userRepo, nil) // Pasar nil temporalmente

	// Crear servicio de autenticación
	authService := auth.NewService()

	// Crear handler de autenticación
	authHandler := handlers.NewAuthHandler(userService, authService)

	// Crear servidor
	server := &Server{
		userService: userService,
		// tigerBeetleClient: tbService, // Comentado temporalmente
		authService: authService,
		authHandler: authHandler,
	}

	// Verificar si se debe inicializar con datos de prueba
	if shouldSeedData() {
		log.Println("Inicializando datos de prueba...")
		if err := database.SeedDatabase(userService, "./datos-prueba-HNL (1).json"); err != nil {
			log.Printf("Advertencia: Error inicializando datos de prueba: %v", err)
		} else {
			log.Println("Datos de prueba inicializados exitosamente")
		}
	}

	// Configurar rutas
	router := server.setupRoutes()

	// Obtener puerto del servidor
	port := getServerPort()
	log.Printf("Servidor iniciado en puerto %s", port)
	log.Printf("API disponible en: http://localhost:%s", port)

	// Iniciar servidor
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}

func (s *Server) setupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Middleware para logging
	router.Use(loggingMiddleware)
	router.Use(corsMiddleware)

	// Crear rate limiter para autenticación
	authRateLimiter := middleware.CreateAuthRateLimiter()

	// Rutas de la API
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(corsMiddleware) // Aplicar CORS también al subrouter de API

	// Rutas de autenticación (públicas con rate limiting)
	authRoutes := api.PathPrefix("/auth").Subrouter()
	authRoutes.Use(corsMiddleware) // Aplicar CORS también al subrouter de auth
	authRoutes.Use(authRateLimiter.Middleware)
	authRoutes.HandleFunc("/register", s.authHandler.Register).Methods("POST")
	authRoutes.HandleFunc("/register", s.handleOptions).Methods("OPTIONS")
	authRoutes.HandleFunc("/login", s.authHandler.Login).Methods("POST")
	authRoutes.HandleFunc("/login", s.handleOptions).Methods("OPTIONS")
	authRoutes.HandleFunc("/logout", s.authHandler.Logout).Methods("POST")
	authRoutes.HandleFunc("/logout", s.handleOptions).Methods("OPTIONS")

	// Rutas protegidas
	protectedRoutes := api.PathPrefix("").Subrouter()
	protectedRoutes.Use(middleware.AuthMiddleware(s.authService))

	// Rutas de usuarios (protegidas)
	protectedRoutes.HandleFunc("/users", s.createUser).Methods("POST")
	protectedRoutes.HandleFunc("/users/{id}", s.getUser).Methods("GET")
	protectedRoutes.HandleFunc("/users/{id}/balance", s.getUserBalance).Methods("GET")
	protectedRoutes.HandleFunc("/users", s.listUsers).Methods("GET")

	// Rutas de transacciones (protegidas)
	protectedRoutes.HandleFunc("/users/{id}/deposit", s.depositToUser).Methods("POST")
	protectedRoutes.HandleFunc("/users/{id}/withdraw", s.withdrawFromUser).Methods("POST")
	protectedRoutes.HandleFunc("/transfer", s.transferBetweenUsers).Methods("POST")

	// Ruta para obtener información del usuario autenticado
	protectedRoutes.HandleFunc("/auth/me", s.authHandler.Me).Methods("GET")

	// Ruta de salud (pública)
	api.HandleFunc("/health", s.healthCheck).Methods("GET")

	// Ruta de salud adicional sin prefijo para facilidad de acceso
	router.HandleFunc("/health", s.healthCheck).Methods("GET")

	return router
}

// Handlers HTTP

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user, err := s.userService.CreateUserWithAccount(&req)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user.ToResponse())
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := s.userService.GetUser(userID)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error getting user: %v", err)
		http.Error(w, "Error getting user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user.ToResponse())
}

func (s *Server) getUserBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, balance, err := s.userService.GetUserWithBalance(userID)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error getting user balance: %v", err)
		http.Error(w, "Error getting user balance", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"user":    user.ToResponse(),
		"balance": balance,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	// Obtener parámetros de paginación
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // default
	offset := 0 // default

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	users, err := s.userService.ListUsers(limit, offset)
	if err != nil {
		log.Printf("Error listing users: %v", err)
		http.Error(w, "Error listing users", http.StatusInternalServerError)
		return
	}

	// Convertir a responses
	responses := make([]models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

func (s *Server) depositToUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Amount uint64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Amount == 0 {
		http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
		return
	}

	if err := s.userService.DepositToUser(userID, req.Amount); err != nil {
		log.Printf("Error depositing to user: %v", err)
		http.Error(w, "Error processing deposit", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (s *Server) withdrawFromUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Amount uint64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Amount == 0 {
		http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
		return
	}

	if err := s.userService.WithdrawFromUser(userID, req.Amount); err != nil {
		if err.Error() == "insufficient funds" {
			http.Error(w, "Insufficient funds", http.StatusBadRequest)
			return
		}
		log.Printf("Error withdrawing from user: %v", err)
		http.Error(w, "Error processing withdrawal", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (s *Server) transferBetweenUsers(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FromUserID uuid.UUID `json:"from_user_id"`
		ToUserID   uuid.UUID `json:"to_user_id"`
		Amount     uint64    `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Amount == 0 {
		http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
		return
	}

	if req.FromUserID == req.ToUserID {
		http.Error(w, "Cannot transfer to the same user", http.StatusBadRequest)
		return
	}

	if err := s.userService.TransferBetweenUsers(req.FromUserID, req.ToUserID, req.Amount); err != nil {
		if err.Error() == "insufficient funds" {
			http.Error(w, "Insufficient funds", http.StatusBadRequest)
			return
		}
		log.Printf("Error transferring between users: %v", err)
		http.Error(w, "Error processing transfer", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "banca-en-linea-backend",
	})
}

// Middleware

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Configuración estricta de CORS
		origin := r.Header.Get("Origin")
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://localhost:5174",
			"http://localhost:8082",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
			"http://127.0.0.1:5174",
			"http://127.0.0.1:8082",
		}

		// Verificar si el origen está permitido
		isAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 horas

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Funciones auxiliares

func (s *Server) handleOptions(w http.ResponseWriter, r *http.Request) {
	// Esta función maneja las solicitudes OPTIONS para CORS
	// Los headers CORS ya se establecen en el middleware corsMiddleware
	w.WriteHeader(http.StatusOK)
}

func shouldSeedData() bool {
	seedEnv := os.Getenv("SEED_DATA")
	return seedEnv == "true" || seedEnv == "1"
}

func getServerPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}
