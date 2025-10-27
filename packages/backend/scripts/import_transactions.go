package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type JSONUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FullName  string `json:"full_name"`
	CreatedAt string `json:"created_at"`
}

type JSONAccount struct {
	AccountNumber  string  `json:"account_number"`
	UserID         string  `json:"user_id"`
	InitialBalance float64 `json:"initial_balance"`
	Currency       string  `json:"currency"`
	AccountType    string  `json:"account_type"`
}

type JSONTransaction struct {
	FromAccount string  `json:"from_account"`
	ToAccount   string  `json:"to_account"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Timestamp   string  `json:"timestamp"`
	Status      string  `json:"status"`
}

type JSONData struct {
	Users        []JSONUser        `json:"users"`
	Accounts     []JSONAccount     `json:"accounts"`
	Transactions []JSONTransaction `json:"transactions"`
}

type APIUser struct {
	ID                   string `json:"id"`
	Email                string `json:"email"`
	FirstName            string `json:"first_name"`
	LastName             string `json:"last_name"`
	TigerBeetleAccountID int64  `json:"tigerbeetle_account_id"`
}

type DepositRequest struct {
	Amount uint64 `json:"amount"`
}

type WithdrawRequest struct {
	Amount uint64 `json:"amount"`
}

type TransferRequest struct {
	ToUserID string `json:"to_user_id"`
	Amount   uint64 `json:"amount"`
}

func main() {
	// Leer el archivo JSON
	fmt.Println("Leyendo archivo de datos de prueba...")
	data, err := os.ReadFile("../../../datos-prueba-HNL (1).json")
	if err != nil {
		fmt.Printf("Error leyendo archivo: %v\n", err)
		return
	}

	var jsonData JSONData
	if err := json.Unmarshal(data, &jsonData); err != nil {
		fmt.Printf("Error parseando JSON: %v\n", err)
		return
	}

	fmt.Printf("Datos cargados: %d usuarios, %d cuentas, %d transacciones\n", 
		len(jsonData.Users), len(jsonData.Accounts), len(jsonData.Transactions))

	// Obtener usuarios existentes del API
	fmt.Println("Obteniendo usuarios existentes del API...")
	existingUsers, err := getExistingUsers()
	if err != nil {
		fmt.Printf("Error obteniendo usuarios: %v\n", err)
		return
	}

	// Crear mapeo de email a usuario API
	emailToUser := make(map[string]APIUser)
	for _, user := range existingUsers {
		emailToUser[user.Email] = user
	}

	// Crear mapeo de account_number a user_id del JSON
	accountToJSONUserID := make(map[string]string)
	for _, account := range jsonData.Accounts {
		accountToJSONUserID[account.AccountNumber] = account.UserID
	}

	// Crear mapeo de JSON user_id a email
	jsonUserIDToEmail := make(map[string]string)
	for _, user := range jsonData.Users {
		jsonUserIDToEmail[user.ID] = user.Email
	}

	// Crear mapeo final: account_number -> API user
	accountToAPIUser := make(map[string]APIUser)
	for accountNumber, jsonUserID := range accountToJSONUserID {
		if email, exists := jsonUserIDToEmail[jsonUserID]; exists {
			if apiUser, exists := emailToUser[email]; exists {
				accountToAPIUser[accountNumber] = apiUser
			}
		}
	}

	fmt.Printf("Mapeo creado: %d cuentas mapeadas a usuarios API\n", len(accountToAPIUser))

	// Procesar transacciones
	fmt.Println("Procesando transacciones...")
	successCount := 0
	errorCount := 0

	for i, transaction := range jsonData.Transactions {
		if i%100 == 0 {
			fmt.Printf("Procesando transacción %d/%d...\n", i+1, len(jsonData.Transactions))
		}

		// Convertir amount a uint64 (centavos)
		amountCents := uint64(transaction.Amount * 100)

		switch transaction.Type {
		case "deposit":
			if user, exists := accountToAPIUser[transaction.ToAccount]; exists {
				if err := makeDeposit(user.ID, amountCents); err != nil {
					fmt.Printf("Error en depósito para usuario %s: %v\n", user.Email, err)
					errorCount++
				} else {
					successCount++
				}
			} else {
				fmt.Printf("Usuario no encontrado para cuenta %s\n", transaction.ToAccount)
				errorCount++
			}

		case "withdrawal":
			if user, exists := accountToAPIUser[transaction.FromAccount]; exists {
				if err := makeWithdrawal(user.ID, amountCents); err != nil {
					fmt.Printf("Error en retiro para usuario %s: %v\n", user.Email, err)
					errorCount++
				} else {
					successCount++
				}
			} else {
				fmt.Printf("Usuario no encontrado para cuenta %s\n", transaction.FromAccount)
				errorCount++
			}

		case "transfer":
			fromUser, fromExists := accountToAPIUser[transaction.FromAccount]
			toUser, toExists := accountToAPIUser[transaction.ToAccount]
			
			if fromExists && toExists {
				if err := makeTransfer(fromUser.ID, toUser.ID, amountCents); err != nil {
					fmt.Printf("Error en transferencia de %s a %s: %v\n", fromUser.Email, toUser.Email, err)
					errorCount++
				} else {
					successCount++
				}
			} else {
				if !fromExists {
					fmt.Printf("Usuario origen no encontrado para cuenta %s\n", transaction.FromAccount)
				}
				if !toExists {
					fmt.Printf("Usuario destino no encontrado para cuenta %s\n", transaction.ToAccount)
				}
				errorCount++
			}
		}

		// Pequeña pausa para no sobrecargar el servidor
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Printf("\nImportación completada:\n")
	fmt.Printf("✅ Transacciones exitosas: %d\n", successCount)
	fmt.Printf("❌ Errores: %d\n", errorCount)
	fmt.Printf("📊 Total procesadas: %d\n", successCount+errorCount)
}

func getExistingUsers() ([]APIUser, error) {
	resp, err := http.Get("http://localhost:8081/api/v1/users")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error HTTP: %d", resp.StatusCode)
	}

	var users []APIUser
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}

	return users, nil
}

func makeDeposit(userID string, amount uint64) error {
	depositReq := DepositRequest{Amount: amount}
	jsonData, _ := json.Marshal(depositReq)

	resp, err := http.Post(
		fmt.Sprintf("http://localhost:8081/api/v1/users/%s/deposit", userID),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func makeWithdrawal(userID string, amount uint64) error {
	withdrawReq := WithdrawRequest{Amount: amount}
	jsonData, _ := json.Marshal(withdrawReq)

	resp, err := http.Post(
		fmt.Sprintf("http://localhost:8081/api/v1/users/%s/withdraw", userID),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func makeTransfer(fromUserID, toUserID string, amount uint64) error {
	transferReq := TransferRequest{
		ToUserID: toUserID,
		Amount:   amount,
	}
	jsonData, _ := json.Marshal(transferReq)

	resp, err := http.Post(
		fmt.Sprintf("http://localhost:8081/api/v1/users/%s/transfer", fromUserID),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}