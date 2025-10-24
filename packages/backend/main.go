package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Root endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "¡Hola, mundo! - Banca en Línea API"}`))
	})

	fmt.Println("Servidor iniciado en puerto 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
