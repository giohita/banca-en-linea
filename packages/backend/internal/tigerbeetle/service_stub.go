//go:build ci || docker

package tigerbeetle

import (
	"fmt"
	"log"
)

// AccountType define el tipo de cuenta
type AccountType uint16

const (
	// Cuentas maestras del sistema
	MasterDebitAccount  AccountType = 1 // Cuenta maestra de débito
	MasterCreditAccount AccountType = 2 // Cuenta maestra de crédito

	// Cuentas de usuario
	UserAccount AccountType = 100 // Cuenta de usuario individual
)

// Account representa una cuenta simplificada para el stub
type Account struct {
	ID            uint64
	Ledger        uint32
	Code          uint16
	Flags         uint16
	DebitsPosted  uint64
	CreditsPosted uint64
}

// Implementación de AccountInterface para Account
func (a *Account) GetID() uint64            { return a.ID }
func (a *Account) GetLedger() uint32        { return a.Ledger }
func (a *Account) GetCode() uint16          { return a.Code }
func (a *Account) GetFlags() uint16         { return a.Flags }
func (a *Account) GetDebitsPosted() uint64  { return a.DebitsPosted }
func (a *Account) GetCreditsPosted() uint64 { return a.CreditsPosted }

// Service maneja las operaciones de TigerBeetle (stub para CI)
type Service struct {
	accounts       map[uint64]*Account
	nextTransferID uint64
}

// NewService crea una nueva instancia del servicio TigerBeetle (stub)
func NewService(clusterID interface{}, addresses []string) (*Service, error) {
	log.Println("Using TigerBeetle stub service for CI/testing")

	service := &Service{
		accounts:       make(map[uint64]*Account),
		nextTransferID: 1,
	}

	// Inicializar cuentas maestras
	if err := service.initializeMasterAccounts(); err != nil {
		return nil, fmt.Errorf("failed to initialize master accounts: %w", err)
	}

	return service, nil
}

// NewServiceStub crea una nueva instancia del servicio TigerBeetle stub sin parámetros
func NewServiceStub() *Service {
	log.Println("Using TigerBeetle stub service for development")

	service := &Service{
		accounts:       make(map[uint64]*Account),
		nextTransferID: 1,
	}

	return service
}

// Close cierra la conexión (stub)
func (s *Service) Close() {
	log.Println("Closing TigerBeetle stub service")
}

// InitializeMasterAccounts crea las cuentas maestras del sistema (stub) - método público
func (s *Service) InitializeMasterAccounts() error {
	return s.initializeMasterAccounts()
}

// initializeMasterAccounts crea las cuentas maestras del sistema (stub)
func (s *Service) initializeMasterAccounts() error {
	// Crear cuentas maestras en memoria
	s.accounts[1] = &Account{
		ID:            1,
		Ledger:        1,
		Code:          uint16(MasterDebitAccount),
		Flags:         0,
		DebitsPosted:  0,
		CreditsPosted: 0,
	}

	s.accounts[2] = &Account{
		ID:            2,
		Ledger:        1,
		Code:          uint16(MasterCreditAccount),
		Flags:         0,
		DebitsPosted:  0,
		CreditsPosted: 1000000000, // 10,000,000.00 en centavos como balance inicial
	}

	log.Println("Master accounts initialized (stub)")
	return nil
}

// CreateUserAccount crea una nueva cuenta de usuario (stub)
func (s *Service) CreateUserAccount(userID uint64) (AccountInterface, error) {
	account := &Account{
		ID:            userID,
		Ledger:        1,
		Code:          uint16(UserAccount),
		Flags:         0,
		DebitsPosted:  0,
		CreditsPosted: 0,
	}

	s.accounts[userID] = account
	log.Printf("Created user account %d (stub)", userID)
	return account, nil
}

// GetAccount obtiene una cuenta por ID (stub)
func (s *Service) GetAccount(accountID uint64) (AccountInterface, error) {
	account, exists := s.accounts[accountID]
	if !exists {
		return nil, fmt.Errorf("account %d not found", accountID)
	}
	return account, nil
}

// GetAccountBalance obtiene el balance de una cuenta (stub)
func (s *Service) GetAccountBalance(accountID uint64) (uint64, uint64, error) {
	account, exists := s.accounts[accountID]
	if !exists {
		return 0, 0, fmt.Errorf("account %d not found", accountID)
	}
	return account.DebitsPosted, account.CreditsPosted, nil
}

// Transfer realiza una transferencia entre cuentas (stub)
func (s *Service) Transfer(fromAccountID, toAccountID, amount uint64, transferID uint64) error {
	fromAccount, exists := s.accounts[fromAccountID]
	if !exists {
		return fmt.Errorf("from account %d not found", fromAccountID)
	}

	toAccount, exists := s.accounts[toAccountID]
	if !exists {
		return fmt.Errorf("to account %d not found", toAccountID)
	}

	// Simular transferencia
	fromAccount.DebitsPosted += amount
	toAccount.CreditsPosted += amount

	log.Printf("Transfer %d: %d -> %d, amount: %d (stub)", transferID, fromAccountID, toAccountID, amount)
	return nil
}

// Deposit realiza un depósito a una cuenta de usuario (stub)
func (s *Service) Deposit(userAccountID, amount, transferID uint64) error {
	return s.Transfer(2, userAccountID, amount, transferID) // Desde cuenta maestra de crédito
}

// Withdraw realiza un retiro de una cuenta de usuario (stub)
func (s *Service) Withdraw(userAccountID, amount, transferID uint64) error {
	return s.Transfer(userAccountID, 1, amount, transferID) // Hacia cuenta maestra de débito
}