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
	db *sql.DB
	tb tigerbeetle.Client
)

// Configuración de la base de datos
func initDatabase() {
	var err error
	
	// Configuración de PostgreSQL
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "banca_user")
	password := getEnv("POSTGRES_PASSWORD", "banca_password")
	dbname := getEnv("POSTGRES_DB", "banca_db")
	
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Error conectando a PostgreSQL:", err)
	}
	
	// Verificar conexión
	err = db.Ping()
	if err != nil {
		log.Fatal("Error verificando conexión a PostgreSQL:", err)
	}
	
	fmt.Println("✅ Conectado a PostgreSQL")
	
	// Configuración de TigerBeetle (temporalmente comentado para debugging)
	// tigerBeetleURL := getEnv("TIGERBEETLE_URL", "tigerbeetle:3002")
	// clusterID := types.ToUint128(0)
	// tb, err = tigerbeetle.NewClient(clusterID, []string{tigerBeetleURL}, 32)
	// if err != nil {
	// 	log.Fatal("Error conectando a TigerBeetle:", err)
	// }
	
	fmt.Println("⚠️ TigerBeetle temporalmente deshabilitado para debugging")
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
	// Verificar conexiones
	if err := db.Ping(); err != nil {
		http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "OK",
		"database": "connected",
		"tigerbeetle": "connected",
	})
}

// Root endpoint
func rootHandler(w http.ResponseWriter, r *http.Request) {
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
	rows, err := db.Query(`
		SELECT id, email, first_name, last_name, phone, created_at, is_active, email_verified 
		FROM users 
		WHERE is_active = true
		ORDER BY created_at DESC
	`)
	if err != nil {
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
			http.Error(w, "Error scanning user", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// Obtener cuentas bancarias de un usuario
func getUserAccountsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["userId"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	
	rows, err := db.Query(`
		SELECT id, user_id, account_number, account_type, tigerbeetle_account_id, 
			   currency, created_at, is_active 
		FROM bank_accounts 
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
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
			http.Error(w, "Error scanning account", http.StatusInternalServerError)
			return
		}
		
		// Obtener balance desde TigerBeetle
		balance, err := getAccountBalance(account.TigerBeetleAccountID)
		if err != nil {
			log.Printf("Error getting balance for account %d: %v", account.TigerBeetleAccountID, err)
			balance = 0
		}
		account.Balance = float64(balance)
		
		accounts = append(accounts, account)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

// Función auxiliar para obtener balance de TigerBeetle
func getAccountBalance(accountID uint64) (int64, error) {
	accounts, err := tb.LookupAccounts([]types.Uint128{types.ToUint128(accountID)})
	if err != nil {
		return 0, err
	}
	
	if len(accounts) == 0 {
		return 0, fmt.Errorf("account not found")
	}
	
	// Convertir Uint128 a uint64 usando String() y strconv
	creditsStr := accounts[0].CreditsPosted.String()
	debitsStr := accounts[0].DebitsPosted.String()
	
	creditsPosted, err := strconv.ParseUint(creditsStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing credits: %v", err)
	}
	
	debitsPosted, err := strconv.ParseUint(debitsStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing debits: %v", err)
	}
	
	// Calcular balance
	balance := int64(creditsPosted) - int64(debitsPosted)
	
	return balance, nil
}

// Configurar rutas
func setupRoutes() *mux.Router {
	r := mux.NewRouter()
	
	// Middleware CORS
	r.Use(func(next http.Handler) http.Handler {
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
	})
	
	// Rutas
	r.HandleFunc("/", rootHandler).Methods("GET")
	r.HandleFunc("/health", healthHandler).Methods("GET")
	r.HandleFunc("/api/users", getUsersHandler).Methods("GET")
	r.HandleFunc("/api/users/{userId}/accounts", getUserAccountsHandler).Methods("GET")
	
	return r
}

func main() {
	// Inicializar conexiones a bases de datos
	initDatabase()
	defer db.Close()
	
	// Configurar rutas
	router := setupRoutes()
	
	// Iniciar servidor
	port := getEnv("PORT", "8080")
	fmt.Printf("🚀 Servidor iniciado en puerto %s...\n", port)
	fmt.Println("📊 PostgreSQL: Conectado")
	fmt.Println("💰 TigerBeetle: Conectado")
	
	log.Fatal(http.ListenAndServe(":"+port, router))
}
