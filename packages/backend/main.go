package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "strconv"
    "time"

    "github.com/google/uuid"
    "github.com/gorilla/mux"
    _ "github.com/lib/pq"
    tigerbeetle "github.com/tigerbeetle/tigerbeetle-go"
    "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// Estructuras de datos
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Phone        string    `json:"phone"`
	CreatedAt    time.Time `json:"created_at"`
	IsActive     bool      `json:"is_active"`
	EmailVerified bool     `json:"email_verified"`
}

type BankAccount struct {
	ID                   uuid.UUID `json:"id"`
	UserID               uuid.UUID `json:"user_id"`
	AccountNumber        string    `json:"account_number"`
	AccountType          string    `json:"account_type"`
	TigerBeetleAccountID uint64    `json:"tigerbeetle_account_id"`
	Currency             string    `json:"currency"`
	CreatedAt            time.Time `json:"created_at"`
	IsActive             bool      `json:"is_active"`
	Balance              float64   `json:"balance"` // Balance desde TigerBeetle
}

type Transaction struct {
	ID                     uuid.UUID `json:"id"`
	TigerBeetleTransferID  uint64    `json:"tigerbeetle_transfer_id"`
	FromAccountID          *uuid.UUID `json:"from_account_id"`
	ToAccountID            *uuid.UUID `json:"to_account_id"`
	Amount                 int64     `json:"amount"`
	Currency               string    `json:"currency"`
	Description            string    `json:"description"`
	TransactionType        string    `json:"transaction_type"`
	Status                 string    `json:"status"`
	CreatedAt              time.Time `json:"created_at"`
}

// Variables globales para las conexiones
var (
	db     *sql.DB
	tb     tigerbeetle.Client
	logger *zap.Logger
)

// Configuración del logger
func initLogger() {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"
	
	var err error
	logger, err = config.Build()
	if err != nil {
		log.Fatal("Error inicializando logger:", err)
	}
}

// Configuración de la base de datos
func initDatabase() {
    var err error
	
	// Configuración de PostgreSQL
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "postgres")
	password := getEnv("POSTGRES_PASSWORD", "postgres")
	dbname := getEnv("POSTGRES_DB", "banca_db")
	
	logger.Info("Iniciando conexión a PostgreSQL",
		zap.String("host", host),
		zap.String("port", port),
		zap.String("user", user),
		zap.String("database", dbname),
	)
	
	// Construir string de conexión
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	
	// Conectar a PostgreSQL
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		logger.Fatal("Error conectando a PostgreSQL", zap.Error(err))
	}
	
	// Verificar conexión
	err = db.Ping()
	if err != nil {
		logger.Fatal("Error verificando conexión a PostgreSQL", zap.Error(err))
	}
	
	logger.Info("✅ Conectado exitosamente a PostgreSQL")
	
    // Configuración de TigerBeetle (simplificada)
    tigerBeetleHost := getEnv("TIGERBEETLE_HOST", "tigerbeetle")
    tigerBeetlePort := getEnv("TIGERBEETLE_PORT", "3002")
    tigerBeetleAddrEnv := os.Getenv("TIGERBEETLE_ADDR")
	
	logger.Info("Iniciando conexión a TigerBeetle",
		zap.String("host", tigerBeetleHost),
		zap.String("port", tigerBeetlePort),
	)
	
    // Resolver IP IPv4 para evitar "Invalid client cluster address" cuando se usa hostname
    // Dirección final: usar override si existe, de lo contrario hostname directo.
    tigerBeetleAddress := tigerBeetleHost + ":" + tigerBeetlePort
    if tigerBeetleAddrEnv != "" {
        tigerBeetleAddress = tigerBeetleAddrEnv
        logger.Info("Usando TIGERBEETLE_ADDR explícito", zap.String("address", tigerBeetleAddress))
    }
	clusterID := types.ToUint128(0)
	
    logger.Info("Intentando crear cliente TigerBeetle",
        zap.String("cluster_id", "0"),
        zap.String("address", tigerBeetleAddress),
    )
	
	// Crear cliente TigerBeetle con configuración simplificada
    tb, err = tigerbeetle.NewClient(clusterID, []string{tigerBeetleAddress}, 1)
    if err != nil {
        logger.Error("Error conectando a TigerBeetle", 
            zap.Error(err),
            zap.String("address", tigerBeetleAddress),
            zap.String("cluster_id", "0"),
        )
        logger.Warn("TigerBeetle no estará disponible - puedes establecer TIGERBEETLE_ADDR=IP:PUERTO para forzar la dirección")
        tb = nil
    } else {
        logger.Info("✅ Conectado exitosamente a TigerBeetle")
    }

}

// Función auxiliar para obtener variables de entorno
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Handlers HTTP

// Health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Health check solicitado", zap.String("remote_addr", r.RemoteAddr))
	
	// Verificar conexión a PostgreSQL
	dbStatus := "connected"
	if err := db.Ping(); err != nil {
		logger.Error("Error en health check - PostgreSQL", zap.Error(err))
		dbStatus = "disconnected"
		http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
		return
	}
	
	// Verificar conexión a TigerBeetle
	tbStatus := "not_initialized"
	if tb != nil {
		tbStatus = "connected"
	}
	
	response := map[string]string{
		"status":      "OK",
		"database":    dbStatus,
		"tigerbeetle": tbStatus,
	}
	
	logger.Info("Health check completado",
		zap.String("database_status", dbStatus),
		zap.String("tigerbeetle_status", tbStatus),
	)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Root endpoint
func rootHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Root endpoint accedido", zap.String("remote_addr", r.RemoteAddr))
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "¡Banca en Línea API!",
		"version": "1.0.0",
		"status": "running",
	})
}

// Obtener usuarios
func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Solicitando lista de usuarios", zap.String("remote_addr", r.RemoteAddr))
	
	rows, err := db.Query(`
		SELECT id, email, first_name, last_name, phone, created_at, is_active, email_verified 
		FROM users 
		WHERE is_active = true
		ORDER BY created_at DESC
	`)
	if err != nil {
		logger.Error("Error consultando usuarios", zap.Error(err))
		http.Error(w, "Error querying users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, 
			&user.Phone, &user.CreatedAt, &user.IsActive, &user.EmailVerified)
		if err != nil {
			logger.Error("Error escaneando usuario", zap.Error(err))
			http.Error(w, "Error scanning user", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}
	
	logger.Info("Usuarios obtenidos exitosamente", zap.Int("count", len(users)))
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// Obtener cuentas bancarias de un usuario
func getUserAccountsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["userId"])
	if err != nil {
		logger.Error("ID de usuario inválido", zap.String("user_id", vars["userId"]), zap.Error(err))
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	
	logger.Info("Solicitando cuentas de usuario", 
		zap.String("user_id", userID.String()),
		zap.String("remote_addr", r.RemoteAddr),
	)
	
	rows, err := db.Query(`
		SELECT id, user_id, account_number, account_type, tigerbeetle_account_id, 
			   currency, created_at, is_active 
		FROM bank_accounts 
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		logger.Error("Error consultando cuentas", zap.String("user_id", userID.String()), zap.Error(err))
		http.Error(w, "Error querying accounts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	var accounts []BankAccount
	for rows.Next() {
		var account BankAccount
		err := rows.Scan(&account.ID, &account.UserID, &account.AccountNumber, 
			&account.AccountType, &account.TigerBeetleAccountID, &account.Currency,
			&account.CreatedAt, &account.IsActive)
		if err != nil {
			logger.Error("Error escaneando cuenta", zap.Error(err))
			http.Error(w, "Error scanning account", http.StatusInternalServerError)
			return
		}
		
		// Obtener balance desde TigerBeetle
		balance, err := getAccountBalance(account.TigerBeetleAccountID)
		if err != nil {
			logger.Warn("Error obteniendo balance de TigerBeetle", 
				zap.Uint64("tigerbeetle_account_id", account.TigerBeetleAccountID),
				zap.Error(err),
			)
			balance = 0
		}
		account.Balance = float64(balance)
		
		accounts = append(accounts, account)
	}
	
	logger.Info("Cuentas obtenidas exitosamente", 
		zap.String("user_id", userID.String()),
		zap.Int("count", len(accounts)),
	)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

// Función auxiliar para obtener balance de TigerBeetle
func getAccountBalance(accountID uint64) (int64, error) {
	logger.Debug("Consultando balance en TigerBeetle", zap.Uint64("account_id", accountID))
	
	accounts, err := tb.LookupAccounts([]types.Uint128{types.ToUint128(accountID)})
	if err != nil {
		logger.Error("Error consultando cuenta en TigerBeetle", 
			zap.Uint64("account_id", accountID),
			zap.Error(err),
		)
		return 0, err
	}
	
	if len(accounts) == 0 {
		logger.Warn("Cuenta no encontrada en TigerBeetle", zap.Uint64("account_id", accountID))
		return 0, fmt.Errorf("account not found")
	}
	
	// Convertir Uint128 a uint64 usando String() y strconv
	creditsStr := accounts[0].CreditsPosted.String()
	debitsStr := accounts[0].DebitsPosted.String()
	
	creditsPosted, err := strconv.ParseUint(creditsStr, 10, 64)
	if err != nil {
		logger.Error("Error parseando créditos", 
			zap.Uint64("account_id", accountID),
			zap.String("credits_str", creditsStr),
			zap.Error(err),
		)
		return 0, fmt.Errorf("error parsing credits: %v", err)
	}
	
	debitsPosted, err := strconv.ParseUint(debitsStr, 10, 64)
	if err != nil {
		logger.Error("Error parseando débitos", 
			zap.Uint64("account_id", accountID),
			zap.String("debits_str", debitsStr),
			zap.Error(err),
		)
		return 0, fmt.Errorf("error parsing debits: %v", err)
	}
	
	// Calcular balance
	balance := int64(creditsPosted) - int64(debitsPosted)
	
	logger.Debug("Balance calculado exitosamente", 
		zap.Uint64("account_id", accountID),
		zap.Int64("balance", balance),
		zap.Uint64("credits", creditsPosted),
		zap.Uint64("debits", debitsPosted),
	)
	
	return balance, nil
}

// Configurar rutas
func setupRoutes() *mux.Router {
	r := mux.NewRouter()
	
	// Middleware CORS y Logging
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log de request
			logger.Info("Request recibido",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			)
			
			// Headers CORS
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	})
	
	// Rutas
	r.HandleFunc("/", rootHandler).Methods("GET")
	r.HandleFunc("/health", healthHandler).Methods("GET")
	r.HandleFunc("/api/users", getUsersHandler).Methods("GET")
	r.HandleFunc("/api/users/{userId}/accounts", getUserAccountsHandler).Methods("GET")
	
	return r
}

func main() {
	// Inicializar logger
	initLogger()
	defer logger.Sync()
	
	logger.Info("🚀 Iniciando aplicación bancaria")
	
	// Inicializar conexiones a bases de datos
	initDatabase()
	defer db.Close()
	defer func() {
		if tb != nil {
			tb.Close()
		}
	}()
	
	// Configurar rutas
	router := setupRoutes()
	
	// Iniciar servidor
	port := getEnv("PORT", "8080")
	logger.Info("🚀 Servidor iniciado",
		zap.String("port", port),
		zap.String("status", "listening"),
	)
	
	logger.Fatal("Error en servidor HTTP", zap.Error(http.ListenAndServe(":"+port, router)))
}
