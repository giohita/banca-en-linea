package database

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"banca-en-linea/backend/internal/db"
	"banca-en-linea/backend/models"
)

// TestUser representa un usuario de prueba del archivo JSON
type TestUser struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	CreatedAt string `json:"created_at"`
}

// SeedDatabase carga los datos de prueba desde el archivo JSON
func SeedDatabase(userService *db.UserService, jsonFilePath string) error {
	log.Println("Starting database seeding...")

	// 1. Leer el archivo JSON
	data, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		return fmt.Errorf("error reading JSON file: %w", err)
	}

	// 2. Parsear el JSON
	var testUsers []TestUser
	if err := json.Unmarshal(data, &testUsers); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	log.Printf("Found %d test users in JSON file", len(testUsers))

	// 3. Crear usuarios en el sistema
	successCount := 0
	errorCount := 0

	for _, testUser := range testUsers {
		// Crear request para el usuario
		createReq := &models.CreateUserRequest{
			Email:     testUser.Email,
			Password:  testUser.Password, // En producción, esto debería ser más seguro
			FirstName: testUser.FirstName,
			LastName:  testUser.LastName,
		}

		// Crear usuario con cuenta TigerBeetle
		user, err := userService.CreateUserWithAccount(createReq)
		if err != nil {
			log.Printf("Error creating user %s: %v", testUser.Email, err)
			errorCount++
			continue
		}

		// Realizar un depósito inicial de prueba (1000.00 HNL = 100000 centavos)
		depositAmount := uint64(100000) // 1000.00 HNL en centavos
		if err := userService.DepositToUser(user.ID, depositAmount); err != nil {
			log.Printf("Warning: Could not deposit initial amount for user %s: %v", user.Email, err)
		} else {
			log.Printf("Deposited initial amount of 1000.00 HNL to user %s", user.Email)
		}

		successCount++
		log.Printf("Successfully created user: %s (ID: %s, TigerBeetle Account: %d)",
			user.Email, user.ID, *user.TigerBeetleAccountID)
	}

	log.Printf("Database seeding completed. Success: %d, Errors: %d", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("seeding completed with %d errors", errorCount)
	}

	return nil
}

// SeedDatabaseFromDefaultPath carga los datos usando la ruta por defecto
func SeedDatabaseFromDefaultPath(userService *db.UserService) error {
	// Ruta por defecto al archivo de datos de prueba
	defaultPath := "c:\\Users\\Giohan Melo\\OneDrive\\Desktop\\proyectos-programacion\\banca-en-linea\\datos-prueba-HNL (1).json"

	log.Printf("Loading test data from: %s", defaultPath)
	return SeedDatabase(userService, defaultPath)
}

// CreateSampleTransactions crea algunas transacciones de ejemplo entre usuarios
func CreateSampleTransactions(userService *db.UserService, userRepo db.UserRepository) error {
	log.Println("Creating sample transactions...")

	// Obtener algunos usuarios para crear transacciones
	users, err := userRepo.List(5, 0) // Obtener los primeros 5 usuarios
	if err != nil {
		return fmt.Errorf("error getting users for sample transactions: %w", err)
	}

	if len(users) < 2 {
		log.Println("Not enough users to create sample transactions")
		return nil
	}

	// Crear algunas transferencias de ejemplo
	transactions := []struct {
		fromIndex   int
		toIndex     int
		amount      uint64 // en centavos
		description string
	}{
		{0, 1, 5000, "Transfer 50.00 HNL"}, // 50.00 HNL
		{1, 2, 2500, "Transfer 25.00 HNL"}, // 25.00 HNL
		{2, 0, 1000, "Transfer 10.00 HNL"}, // 10.00 HNL
	}

	for _, tx := range transactions {
		if tx.fromIndex >= len(users) || tx.toIndex >= len(users) {
			continue
		}

		fromUser := users[tx.fromIndex]
		toUser := users[tx.toIndex]

		err := userService.TransferBetweenUsers(fromUser.ID, toUser.ID, tx.amount)
		if err != nil {
			log.Printf("Error creating sample transaction from %s to %s: %v",
				fromUser.Email, toUser.Email, err)
			continue
		}

		log.Printf("Created sample transaction: %s from %s to %s",
			tx.description, fromUser.Email, toUser.Email)
	}

	log.Println("Sample transactions created successfully")
	return nil
}

// PrintUserBalances imprime los balances de todos los usuarios para verificación
func PrintUserBalances(userService *db.UserService, userRepo db.UserRepository) error {
	log.Println("=== User Balances ===")

	users, err := userRepo.List(100, 0) // Obtener hasta 100 usuarios
	if err != nil {
		return fmt.Errorf("error getting users: %w", err)
	}

	for _, user := range users {
		_, balance, err := userService.GetUserWithBalance(user.ID)
		if err != nil {
			log.Printf("Error getting balance for user %s: %v", user.Email, err)
			continue
		}

		// Convertir centavos a HNL (dividir por 100)
		balanceHNL := float64(balance) / 100.0
		log.Printf("User: %s | Balance: %.2f HNL | TigerBeetle Account: %v",
			user.Email, balanceHNL, user.TigerBeetleAccountID)
	}

	log.Println("=== End User Balances ===")
	return nil
}
