//go:build !ci && !docker

package tigerbeetle

import (
	"fmt"
	"log"

	"github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// AccountType define los tipos de cuenta en TigerBeetle
type AccountType uint16

const (
	// Cuentas maestras del sistema
	MasterDebitAccount  AccountType = 1 // Cuenta maestra de débito
	MasterCreditAccount AccountType = 2 // Cuenta maestra de crédito

	// Cuentas de usuario
	UserAccount AccountType = 100 // Cuenta de usuario individual
)

// Service maneja las operaciones de TigerBeetle
type Service struct {
	client tigerbeetle.Client
}

// NewService crea una nueva instancia del servicio TigerBeetle
func NewService(clusterID uint128, addresses []string) (*Service, error) {
	client, err := tigerbeetle.NewClient(clusterID, addresses)
	if err != nil {
		return nil, fmt.Errorf("failed to create TigerBeetle client: %w", err)
	}

	service := &Service{client: client}

	// Inicializar cuentas maestras si no existen
	if err := service.initializeMasterAccounts(); err != nil {
		return nil, fmt.Errorf("failed to initialize master accounts: %w", err)
	}

	return service, nil
}

// Close cierra la conexión con TigerBeetle
func (s *Service) Close() {
	s.client.Close()
}

// initializeMasterAccounts crea las cuentas maestras del sistema
func (s *Service) initializeMasterAccounts() error {
	// Definir las cuentas maestras
	masterAccounts := []types.Account{
		{
			ID:     1, // ID fijo para cuenta maestra de débito
			Ledger: 1,
			Code:   uint16(MasterDebitAccount),
			Flags:  types.AccountFlags{}.ToUint16(),
		},
		{
			ID:     2, // ID fijo para cuenta maestra de crédito
			Ledger: 1,
			Code:   uint16(MasterCreditAccount),
			Flags:  types.AccountFlags{}.ToUint16(),
		},
	}

	// Intentar crear las cuentas maestras
	results, err := s.client.CreateAccounts(masterAccounts)
	if err != nil {
		return fmt.Errorf("error creating master accounts: %w", err)
	}

	// Verificar si hubo errores (ignorar si las cuentas ya existen)
	for _, result := range results {
		if result.Result != types.AccountExistsError && result.Result != types.AccountOK {
			log.Printf("Warning: Master account creation result: %v", result.Result)
		}
	}

	log.Println("Master accounts initialized successfully")
	return nil
}

// CreateUserAccount crea una nueva cuenta para un usuario
func (s *Service) CreateUserAccount(userID uint64) (AccountInterface, error) {
	account := types.Account{
		ID:     userID, // Usar el ID del usuario como ID de cuenta
		Ledger: 1,
		Code:   uint16(UserAccount),
		Flags:  types.AccountFlags{}.ToUint16(),
	}

	accounts := []types.Account{account}
	results, err := s.client.CreateAccounts(accounts)
	if err != nil {
		return nil, fmt.Errorf("error creating user account: %w", err)
	}

	// Verificar el resultado
	if len(results) > 0 && results[0].Result != types.AccountOK {
		if results[0].Result == types.AccountExistsError {
			return &AccountWrapper{&account}, fmt.Errorf("account already exists for user %d", userID)
		}
		return nil, fmt.Errorf("failed to create account: %v", results[0].Result)
	}

	log.Printf("Created TigerBeetle account %d for user", userID)
	return &AccountWrapper{&account}, nil
}

// GetAccount obtiene información de una cuenta
func (s *Service) GetAccount(accountID uint64) (AccountInterface, error) {
	accounts, err := s.client.LookupAccounts([]uint64{accountID})
	if err != nil {
		return nil, fmt.Errorf("error looking up account: %w", err)
	}

	if len(accounts) == 0 {
		return nil, fmt.Errorf("account %d not found", accountID)
	}

	return &AccountWrapper{&accounts[0]}, nil
}

// GetAccountBalance obtiene el balance de una cuenta
func (s *Service) GetAccountBalance(accountID uint64) (uint64, uint64, error) {
	account, err := s.GetAccount(accountID)
	if err != nil {
		return 0, 0, err
	}

	return account.GetDebitsPosted(), account.GetCreditsPosted(), nil
}

// Transfer realiza una transferencia entre cuentas
func (s *Service) Transfer(fromAccountID, toAccountID, amount uint64, transferID uint64) error {
	transfer := types.Transfer{
		ID:              transferID,
		DebitAccountID:  fromAccountID,
		CreditAccountID: toAccountID,
		Amount:          amount,
		Ledger:          1,
		Code:            1, // Código de transferencia estándar
		Flags:           types.TransferFlags{}.ToUint16(),
	}

	transfers := []types.Transfer{transfer}
	results, err := s.client.CreateTransfers(transfers)
	if err != nil {
		return fmt.Errorf("error creating transfer: %w", err)
	}

	// Verificar el resultado
	if len(results) > 0 && results[0].Result != types.TransferOK {
		return fmt.Errorf("transfer failed: %v", results[0].Result)
	}

	log.Printf("Transfer completed: %d from account %d to account %d", amount, fromAccountID, toAccountID)
	return nil
}

// Deposit realiza un depósito a una cuenta de usuario desde la cuenta maestra de crédito
func (s *Service) Deposit(userAccountID, amount, transferID uint64) error {
	return s.Transfer(2, userAccountID, amount, transferID) // 2 = MasterCreditAccount
}

// Withdraw realiza un retiro de una cuenta de usuario hacia la cuenta maestra de débito
func (s *Service) Withdraw(userAccountID, amount, transferID uint64) error {
	return s.Transfer(userAccountID, 1, amount, transferID) // 1 = MasterDebitAccount
}

// uint128 helper para crear un uint128 desde dos uint64
func uint128(high, low uint64) types.Uint128 {
	return types.Uint128{High: high, Low: low}
}