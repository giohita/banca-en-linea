//go:build !ci

package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"go.uber.org/zap"
)

// initTigerBeetle inicializa la conexión a TigerBeetle
func initTigerBeetle() {
	logger.Info("🔧 Inicializando conexión a TigerBeetle")

	// Obtener dirección de TigerBeetle
	tigerBeetleAddress := getEnv("TIGERBEETLE_ADDRESS", "localhost:3000")
	logger.Info("Configurando TigerBeetle", zap.String("address", tigerBeetleAddress))

	// Resolver dirección IP si es necesario
	if host, port, err := net.SplitHostPort(tigerBeetleAddress); err == nil {
		if ips, err := net.LookupIP(host); err == nil && len(ips) > 0 {
			tigerBeetleAddress = net.JoinHostPort(ips[0].String(), port)
			logger.Info("Dirección IP resuelta", zap.String("resolved_address", tigerBeetleAddress))
		}
	}

	// Configurar variables de entorno para TigerBeetle
	if os.Getenv("TIGERBEETLE_IO_MODE") == "" {
		os.Setenv("TIGERBEETLE_IO_MODE", "io_uring")
		logger.Debug("Configurando TIGERBEETLE_IO_MODE", zap.String("mode", "io_uring"))
	}

	if os.Getenv("TIGERBEETLE_DISABLE_IO_URING") == "" {
		os.Setenv("TIGERBEETLE_DISABLE_IO_URING", "false")
		logger.Debug("Configurando TIGERBEETLE_DISABLE_IO_URING", zap.String("disabled", "false"))
	}

	// Configurar cluster ID
	clusterID := uint128FromString("0")
	logger.Debug("Configurando cluster ID", zap.String("cluster_id", "0"))

	// Crear cliente TigerBeetle con configuración simplificada
	var err error
	client, err := tigerbeetle.NewClient(clusterID, []string{tigerBeetleAddress})
	tb = client
	if err != nil {
		logger.Error("❌ Error conectando a TigerBeetle", 
			zap.String("address", tigerBeetleAddress),
			zap.Error(err),
		)
		logger.Warn("⚠️ Continuando sin TigerBeetle - funcionalidad limitada")
		tb = nil
	} else {
		logger.Info("✅ Conectado exitosamente a TigerBeetle")
	}
}

// getAccountBalance obtiene el balance de una cuenta desde TigerBeetle
func getAccountBalance(accountID uint64) (int64, error) {
	logger.Debug("Consultando balance en TigerBeetle", zap.Uint64("account_id", accountID))
	
	if tb == nil {
		logger.Warn("TigerBeetle no está disponible")
		return 0, fmt.Errorf("TigerBeetle not available")
	}
	
	// Cast tb a tigerbeetle.Client para acceder a métodos específicos
	client, ok := tb.(tigerbeetle.Client)
	if !ok {
		logger.Error("Error: tb no es un cliente TigerBeetle válido")
		return 0, fmt.Errorf("invalid TigerBeetle client")
	}
	
	accounts, err := client.LookupAccounts([]types.Uint128{types.ToUint128(accountID)})
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

// uint128FromString convierte un string a Uint128
func uint128FromString(s string) types.Uint128 {
	val, _ := strconv.ParseUint(s, 10, 64)
	return types.ToUint128(val)
}