//go:build scripts

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// UserData representa la estructura de los datos JSON de entrada
type UserData struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FullName  string `json:"full_name"`
	CreatedAt string `json:"created_at"`
}

// CreateUserRequest representa la estructura para crear un nuevo usuario
type CreateUserRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// ImportData representa la estructura del archivo JSON
type ImportData struct {
	Users []UserData `json:"users"`
}

func main() {
	// Leer el archivo JSON
	jsonFile, err := os.Open("../../../datos-prueba-HNL (1).json")
	if err != nil {
		fmt.Printf("Error abriendo archivo JSON: %v\n", err)
		return
	}
	defer jsonFile.Close()

	// Leer el contenido del archivo
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Printf("Error leyendo archivo JSON: %v\n", err)
		return
	}

	// Parsear el JSON
	var importData ImportData
	err = json.Unmarshal(byteValue, &importData)
	if err != nil {
		fmt.Printf("Error parseando JSON: %v\n", err)
		return
	}

	fmt.Printf("Importando %d usuarios...\n", len(importData.Users))

	// URL del API
	apiURL := "http://localhost:8081/api/v1/users"

	successCount := 0
	errorCount := 0

	// Procesar cada usuario
	for i, userData := range importData.Users {
		// Separar el nombre completo en first_name y last_name
		nameParts := strings.Fields(userData.FullName)
		var firstName, lastName string

		if len(nameParts) >= 2 {
			firstName = nameParts[0]
			lastName = strings.Join(nameParts[1:], " ")
		} else if len(nameParts) == 1 {
			firstName = nameParts[0]
			lastName = "Usuario"
		} else {
			firstName = "Usuario"
			lastName = "Desconocido"
		}

		// Crear la estructura de request
		createRequest := CreateUserRequest{
			Email:     userData.Email,
			Password:  userData.Password,
			FirstName: firstName,
			LastName:  lastName,
		}

		// Convertir a JSON
		jsonData, err := json.Marshal(createRequest)
		if err != nil {
			fmt.Printf("Error creando JSON para usuario %s: %v\n", userData.Email, err)
			errorCount++
			continue
		}

		// Hacer la petición HTTP
		resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Error creando usuario %s: %v\n", userData.Email, err)
			errorCount++
			continue
		}

		if resp.StatusCode == 201 {
			fmt.Printf("✅ Usuario %d/%d creado: %s (%s %s)\n", i+1, len(importData.Users), userData.Email, firstName, lastName)
			successCount++
		} else {
			// Leer el cuerpo de la respuesta para obtener más detalles del error
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("❌ Error creando usuario %s (Status: %d): %s\n", userData.Email, resp.StatusCode, string(body))
			errorCount++
		}

		resp.Body.Close()

		// Pequeña pausa para no sobrecargar el servidor
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("\n=== Resumen de importación ===\n")
	fmt.Printf("✅ Usuarios creados exitosamente: %d\n", successCount)
	fmt.Printf("❌ Errores: %d\n", errorCount)
	fmt.Printf("📊 Total procesados: %d\n", successCount+errorCount)
}
