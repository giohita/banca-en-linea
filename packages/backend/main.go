package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"banca-en-linea/backend/database"
	"banca-en-linea/backend/internal/db"
	"banca-en-linea/backend/internal/tigerbeetle"
	"banca-en-linea/backend/models"
)

type Server struct {
	userService *db.UserService
	db          *sql.DB
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
	log.Println("Inicializando servicio TigerBeetle...")
	tbService := tigerbeetle.NewServiceStub()
	defer tbService.Close()

	// Inicializar cuentas maestras de TigerBeetle
	if err := tbService.InitializeMasterAccounts(); err != nil {
		log.Fatalf("Error inicializando cuentas maestras TigerBeetle: %v", err)
	}

	// Crear repositorio y servicio de usuarios
	userRepo := db.NewUserRepository(dbConn)
	userService := db.NewUserService(userRepo, tbService)

	// Crear servidor
	server := &Server{
		userService: userService,
		db:          dbConn,
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

	// Rutas de la API
	api := router.PathPrefix("/api/v1").Subrouter()

	// Rutas de usuarios
	api.HandleFunc("/users", s.createUser).Methods("POST")
	api.HandleFunc("/users/{id}", s.getUser).Methods("GET")
	api.HandleFunc("/users/{id}/balance", s.getUserBalance).Methods("GET")
	api.HandleFunc("/users", s.listUsers).Methods("GET")

	// Rutas de transacciones
	api.HandleFunc("/users/{id}/deposit", s.depositToUser).Methods("POST")
	api.HandleFunc("/users/{id}/withdraw", s.withdrawFromUser).Methods("POST")
	api.HandleFunc("/transfer", s.transferBetweenUsers).Methods("POST")

	// Ruta de salud
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
	// Verificar conexión a la base de datos
	if err := s.db.Ping(); err != nil {
		http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
		return
	}

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
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Funciones auxiliares

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
