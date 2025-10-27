//go:build ci

package main

import (
	"log"

	"go.uber.org/zap"
)

// Variables globales para el stub
var (
	tb     interface{} // Stub de TigerBeetle client
	logger *zap.Logger
)

// init inicializa el logger para el stub
func init() {
	var err error
	logger, err = zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Error inicializando logger en stub: %v", err)
	}
}

// initTigerBeetle inicializa un stub de TigerBeetle para builds de CI
func initTigerBeetle() {
	logger.Info("🔧 Usando TigerBeetle stub para CI")
	tb = nil
}

// getAccountBalance versión stub para CI
func getAccountBalance(accountID uint64) (int64, error) {
	logger.Debug("Usando stub de TigerBeetle para balance", zap.Uint64("account_id", accountID))
	// Retornar un balance simulado para CI
	return 1000, nil
}
