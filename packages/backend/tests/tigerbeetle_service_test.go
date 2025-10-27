package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"banca-en-linea/backend/internal/tigerbeetle"
)

func TestTigerBeetleService_InitializeMasterAccounts(t *testing.T) {
	service := tigerbeetle.NewServiceStub()
	defer service.Close()

	err := service.InitializeMasterAccounts()
	require.NoError(t, err)

	// Verificar que las cuentas maestras fueron creadas
	debitAccount, err := service.GetAccount(tigerbeetle.MasterDebitAccount)
	assert.NoError(t, err)
	assert.NotNil(t, debitAccount)
	assert.Equal(t, tigerbeetle.MasterDebitAccount, debitAccount.ID)

	creditAccount, err := service.GetAccount(tigerbeetle.MasterCreditAccount)
	assert.NoError(t, err)
	assert.NotNil(t, creditAccount)
	assert.Equal(t, tigerbeetle.MasterCreditAccount, creditAccount.ID)
}

func TestTigerBeetleService_CreateUserAccount(t *testing.T) {
	service := tigerbeetle.NewServiceStub()
	defer service.Close()

	// Inicializar cuentas maestras primero
	err := service.InitializeMasterAccounts()
	require.NoError(t, err)

	userID := uint64(12345)

	account, err := service.CreateUserAccount(userID)
	assert.NoError(t, err)
	assert.NotNil(t, account)
	assert.Equal(t, userID, account.ID)

	// Verificar que la cuenta se puede obtener después
	retrievedAccount, err := service.GetAccount(userID)
	assert.NoError(t, err)
	assert.Equal(t, userID, retrievedAccount.ID)
}

func TestTigerBeetleService_CreateUserAccount_Duplicate(t *testing.T) {
	service := tigerbeetle.NewServiceStub()
	defer service.Close()

	// Inicializar cuentas maestras primero
	err := service.InitializeMasterAccounts()
	require.NoError(t, err)

	userID := uint64(12345)

	// Crear cuenta por primera vez
	_, err = service.CreateUserAccount(userID)
	assert.NoError(t, err)

	// Intentar crear la misma cuenta otra vez
	_, err = service.CreateUserAccount(userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "account already exists")
}

func TestTigerBeetleService_GetAccount_NotFound(t *testing.T) {
	service := tigerbeetle.NewServiceStub()
	defer service.Close()

	nonExistentID := uint64(99999)

	account, err := service.GetAccount(nonExistentID)
	assert.Error(t, err)
	assert.Nil(t, account)
	assert.Contains(t, err.Error(), "account not found")
}

func TestTigerBeetleService_GetAccountBalance_NewAccount(t *testing.T) {
	service := tigerbeetle.NewServiceStub()
	defer service.Close()

	// Inicializar cuentas maestras primero
	err := service.InitializeMasterAccounts()
	require.NoError(t, err)

	userID := uint64(12345)

	// Crear cuenta de usuario
	_, err = service.CreateUserAccount(userID)
	require.NoError(t, err)

	// Verificar balance inicial (debe ser 0)
	debits, credits, err := service.GetAccountBalance(userID)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), debits)
	assert.Equal(t, uint64(0), credits)
}

func TestTigerBeetleService_Deposit(t *testing.T) {
	service := tigerbeetle.NewServiceStub()
	defer service.Close()

	// Inicializar cuentas maestras primero
	err := service.InitializeMasterAccounts()
	require.NoError(t, err)

	userID := uint64(12345)
	amount := uint64(10000) // 100.00 HNL
	transferID := uint64(1)

	// Crear cuenta de usuario
	_, err = service.CreateUserAccount(userID)
	require.NoError(t, err)

	// Realizar depósito
	err = service.Deposit(userID, amount, transferID)
	assert.NoError(t, err)

	// Verificar balance después del depósito
	debits, credits, err := service.GetAccountBalance(userID)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), debits)
	assert.Equal(t, amount, credits)
}

func TestTigerBeetleService_Withdraw(t *testing.T) {
	service := tigerbeetle.NewServiceStub()
	defer service.Close()

	// Inicializar cuentas maestras primero
	err := service.InitializeMasterAccounts()
	require.NoError(t, err)

	userID := uint64(12345)
	depositAmount := uint64(10000) // 100.00 HNL
	withdrawAmount := uint64(3000) // 30.00 HNL

	// Crear cuenta de usuario
	_, err = service.CreateUserAccount(userID)
	require.NoError(t, err)

	// Realizar depósito inicial
	err = service.Deposit(userID, depositAmount, 1)
	require.NoError(t, err)

	// Realizar retiro
	err = service.Withdraw(userID, withdrawAmount, 2)
	assert.NoError(t, err)

	// Verificar balance después del retiro
	debits, credits, err := service.GetAccountBalance(userID)
	assert.NoError(t, err)
	assert.Equal(t, withdrawAmount, debits)
	assert.Equal(t, depositAmount, credits)
}

func TestTigerBeetleService_Withdraw_InsufficientFunds(t *testing.T) {
	service := tigerbeetle.NewServiceStub()
	defer service.Close()

	// Inicializar cuentas maestras primero
	err := service.InitializeMasterAccounts()
	require.NoError(t, err)

	userID := uint64(12345)
	withdrawAmount := uint64(10000) // 100.00 HNL (sin fondos)

	// Crear cuenta de usuario (sin depósito inicial)
	_, err = service.CreateUserAccount(userID)
	require.NoError(t, err)

	// Intentar retiro sin fondos
	err = service.Withdraw(userID, withdrawAmount, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient funds")
}

func TestTigerBeetleService_Transfer(t *testing.T) {
	service := tigerbeetle.NewServiceStub()
	defer service.Close()

	// Inicializar cuentas maestras primero
	err := service.InitializeMasterAccounts()
	require.NoError(t, err)

	fromUserID := uint64(12345)
	toUserID := uint64(67890)
	depositAmount := uint64(10000) // 100.00 HNL
	transferAmount := uint64(3000) // 30.00 HNL

	// Crear cuentas de usuario
	_, err = service.CreateUserAccount(fromUserID)
	require.NoError(t, err)
	_, err = service.CreateUserAccount(toUserID)
	require.NoError(t, err)

	// Realizar depósito inicial en cuenta origen
	err = service.Deposit(fromUserID, depositAmount, 1)
	require.NoError(t, err)

	// Realizar transferencia
	err = service.Transfer(fromUserID, toUserID, transferAmount, 2)
	assert.NoError(t, err)

	// Verificar balance de cuenta origen
	fromDebits, fromCredits, err := service.GetAccountBalance(fromUserID)
	assert.NoError(t, err)
	assert.Equal(t, transferAmount, fromDebits)
	assert.Equal(t, depositAmount, fromCredits)

	// Verificar balance de cuenta destino
	toDebits, toCredits, err := service.GetAccountBalance(toUserID)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), toDebits)
	assert.Equal(t, transferAmount, toCredits)
}

func TestTigerBeetleService_Transfer_InsufficientFunds(t *testing.T) {
	service := tigerbeetle.NewServiceStub()
	defer service.Close()

	// Inicializar cuentas maestras primero
	err := service.InitializeMasterAccounts()
	require.NoError(t, err)

	fromUserID := uint64(12345)
	toUserID := uint64(67890)
	transferAmount := uint64(10000) // 100.00 HNL (sin fondos)

	// Crear cuentas de usuario
	_, err = service.CreateUserAccount(fromUserID)
	require.NoError(t, err)
	_, err = service.CreateUserAccount(toUserID)
	require.NoError(t, err)

	// Intentar transferencia sin fondos
	err = service.Transfer(fromUserID, toUserID, transferAmount, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient funds")
}

func TestTigerBeetleService_Transfer_AccountNotFound(t *testing.T) {
	service := tigerbeetle.NewServiceStub()
	defer service.Close()

	// Inicializar cuentas maestras primero
	err := service.InitializeMasterAccounts()
	require.NoError(t, err)

	fromUserID := uint64(12345)
	nonExistentUserID := uint64(99999)
	transferAmount := uint64(1000)

	// Crear solo cuenta origen
	_, err = service.CreateUserAccount(fromUserID)
	require.NoError(t, err)

	// Realizar depósito inicial
	err = service.Deposit(fromUserID, 10000, 1)
	require.NoError(t, err)

	// Intentar transferencia a cuenta inexistente
	err = service.Transfer(fromUserID, nonExistentUserID, transferAmount, 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "account not found")
}

func TestTigerBeetleService_DuplicateTransferID(t *testing.T) {
	service := tigerbeetle.NewServiceStub()
	defer service.Close()

	// Inicializar cuentas maestras primero
	err := service.InitializeMasterAccounts()
	require.NoError(t, err)

	userID := uint64(12345)
	amount := uint64(10000)
	transferID := uint64(1)

	// Crear cuenta de usuario
	_, err = service.CreateUserAccount(userID)
	require.NoError(t, err)

	// Realizar primer depósito
	err = service.Deposit(userID, amount, transferID)
	assert.NoError(t, err)

	// Intentar segundo depósito con mismo transferID
	err = service.Deposit(userID, amount, transferID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transfer ID already exists")
}
