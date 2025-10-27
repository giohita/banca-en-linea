package db

import (
	"fmt"
	"log"

	"github.com/google/uuid"

	"banca-en-linea/backend/internal/tigerbeetle"
	"banca-en-linea/backend/models"
)

// UserService maneja la lógica de negocio para usuarios
type UserService struct {
	userRepo           UserRepository
	tigerBeetleService *tigerbeetle.Service
}

// NewUserService crea una nueva instancia del servicio de usuarios
func NewUserService(userRepo UserRepository, tbService *tigerbeetle.Service) *UserService {
	return &UserService{
		userRepo:           userRepo,
		tigerBeetleService: tbService,
	}
}

// CreateUserWithAccount crea un usuario y su cuenta asociada en TigerBeetle
func (s *UserService) CreateUserWithAccount(req *models.CreateUserRequest) (*models.User, error) {
	// 1. Crear el usuario en PostgreSQL
	user, err := s.userRepo.Create(req)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	// 2. Generar un ID único para la cuenta TigerBeetle
	// Usamos una combinación del timestamp y el UUID del usuario
	accountID := generateTigerBeetleAccountID(user.ID)

	// 3. Crear la cuenta en TigerBeetle
	_, err = s.tigerBeetleService.CreateUserAccount(accountID)
	if err != nil {
		// Si falla la creación de la cuenta TigerBeetle, eliminar el usuario de PostgreSQL
		if deleteErr := s.userRepo.Delete(user.ID); deleteErr != nil {
			log.Printf("Error rolling back user creation: %v", deleteErr)
		}
		return nil, fmt.Errorf("error creating TigerBeetle account: %w", err)
	}

	// 4. Actualizar el usuario con el ID de cuenta TigerBeetle
	signedAccountID := int64(accountID)
	err = s.userRepo.UpdateTigerBeetleAccountID(user.ID, signedAccountID)
	if err != nil {
		log.Printf("Error updating user with TigerBeetle account ID: %v", err)
		// La cuenta ya está creada, pero no está asociada al usuario
		return user, fmt.Errorf("user created but TigerBeetle account association failed: %w", err)
	}

	// 5. Actualizar el objeto user con el ID de cuenta
	user.TigerBeetleAccountID = &signedAccountID

	log.Printf("Successfully created user %s with TigerBeetle account %d", user.Email, accountID)
	return user, nil
}

// GetUserWithBalance obtiene un usuario junto con su balance de TigerBeetle
func (s *UserService) GetUserWithBalance(userID uuid.UUID) (*models.User, uint64, error) {
	// 1. Obtener el usuario de PostgreSQL
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("error getting user: %w", err)
	}

	// 2. Si el usuario no tiene cuenta TigerBeetle, devolver balance 0
	if user.TigerBeetleAccountID == nil {
		return user, 0, nil
	}

	// 3. Obtener el balance de TigerBeetle
	accountID := uint64(*user.TigerBeetleAccountID)
	debits, credits, err := s.tigerBeetleService.GetAccountBalance(accountID)
	if err != nil {
		log.Printf("Error getting TigerBeetle balance for user %s: %v", user.ID, err)
		// Devolver el usuario sin balance en caso de error
		return user, 0, nil
	}

	// El balance es credits - debits
	balance := credits - debits

	return user, balance, nil
}

// DepositToUser realiza un depósito a la cuenta de un usuario
func (s *UserService) DepositToUser(userID uuid.UUID, amount uint64) error {
	// 1. Obtener el usuario
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	// 2. Verificar que el usuario tiene cuenta TigerBeetle
	if user.TigerBeetleAccountID == nil {
		return fmt.Errorf("user does not have a TigerBeetle account")
	}

	// 3. Generar ID único para la transferencia
	transferID := generateTransferID()

	// 4. Realizar el depósito
	accountID := uint64(*user.TigerBeetleAccountID)
	err = s.tigerBeetleService.Deposit(accountID, amount, transferID)
	if err != nil {
		return fmt.Errorf("error depositing to user account: %w", err)
	}

	log.Printf("Successfully deposited %d to user %s (account %d)", amount, user.Email, accountID)
	return nil
}

// WithdrawFromUser realiza un retiro de la cuenta de un usuario
func (s *UserService) WithdrawFromUser(userID uuid.UUID, amount uint64) error {
	// 1. Obtener el usuario
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	// 2. Verificar que el usuario tiene cuenta TigerBeetle
	if user.TigerBeetleAccountID == nil {
		return fmt.Errorf("user does not have a TigerBeetle account")
	}

	// 3. Verificar balance suficiente
	accountID := uint64(*user.TigerBeetleAccountID)
	debits, credits, err := s.tigerBeetleService.GetAccountBalance(accountID)
	if err != nil {
		return fmt.Errorf("error getting account balance: %w", err)
	}

	balance := credits - debits
	if balance < amount {
		return fmt.Errorf("insufficient funds: balance %d, requested %d", balance, amount)
	}

	// 4. Generar ID único para la transferencia
	transferID := generateTransferID()

	// 5. Realizar el retiro
	err = s.tigerBeetleService.Withdraw(accountID, amount, transferID)
	if err != nil {
		return fmt.Errorf("error withdrawing from user account: %w", err)
	}

	log.Printf("Successfully withdrew %d from user %s (account %d)", amount, user.Email, accountID)
	return nil
}

// TransferBetweenUsers realiza una transferencia entre dos usuarios
func (s *UserService) TransferBetweenUsers(fromUserID, toUserID uuid.UUID, amount uint64) error {
	// 1. Obtener ambos usuarios
	fromUser, err := s.userRepo.GetByID(fromUserID)
	if err != nil {
		return fmt.Errorf("error getting source user: %w", err)
	}

	toUser, err := s.userRepo.GetByID(toUserID)
	if err != nil {
		return fmt.Errorf("error getting destination user: %w", err)
	}

	// 2. Verificar que ambos usuarios tienen cuentas TigerBeetle
	if fromUser.TigerBeetleAccountID == nil {
		return fmt.Errorf("source user does not have a TigerBeetle account")
	}

	if toUser.TigerBeetleAccountID == nil {
		return fmt.Errorf("destination user does not have a TigerBeetle account")
	}

	// 3. Verificar balance suficiente
	fromAccountID := uint64(*fromUser.TigerBeetleAccountID)
	debits, credits, err := s.tigerBeetleService.GetAccountBalance(fromAccountID)
	if err != nil {
		return fmt.Errorf("error getting source account balance: %w", err)
	}

	balance := credits - debits
	if balance < amount {
		return fmt.Errorf("insufficient funds: balance %d, requested %d", balance, amount)
	}

	// 4. Generar ID único para la transferencia
	transferID := generateTransferID()

	// 5. Realizar la transferencia
	toAccountID := uint64(*toUser.TigerBeetleAccountID)
	err = s.tigerBeetleService.Transfer(fromAccountID, toAccountID, amount, transferID)
	if err != nil {
		return fmt.Errorf("error transferring between users: %w", err)
	}

	log.Printf("Successfully transferred %d from user %s to user %s", amount, fromUser.Email, toUser.Email)
	return nil
}

// AssociateTigerBeetleAccount asocia una cuenta TigerBeetle existente a un usuario
func (s *UserService) AssociateTigerBeetleAccount(userID uuid.UUID) error {
	// 1. Obtener el usuario
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	// 2. Verificar que el usuario no tiene ya una cuenta asociada
	if user.TigerBeetleAccountID != nil {
		return fmt.Errorf("user already has a TigerBeetle account associated")
	}

	// 3. Generar un ID único para la cuenta TigerBeetle
	accountID := generateTigerBeetleAccountID(user.ID)

	// 4. Crear la cuenta en TigerBeetle
	_, err = s.tigerBeetleService.CreateUserAccount(accountID)
	if err != nil {
		return fmt.Errorf("error creating TigerBeetle account: %w", err)
	}

	// 5. Actualizar el usuario con el ID de cuenta TigerBeetle
	signedAccountID := int64(accountID)
	err = s.userRepo.UpdateTigerBeetleAccountID(user.ID, signedAccountID)
	if err != nil {
		return fmt.Errorf("error updating user with TigerBeetle account ID: %w", err)
	}

	log.Printf("Successfully associated TigerBeetle account %d to user %s", accountID, user.Email)
	return nil
}

// GetUser obtiene un usuario por su ID
func (s *UserService) GetUser(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	return user, nil
}

// ListUsers obtiene una lista paginada de usuarios
func (s *UserService) ListUsers(limit, offset int) ([]*models.User, error) {
	users, err := s.userRepo.List(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing users: %w", err)
	}
	return users, nil
}

// generateTigerBeetleAccountID genera un ID único para una cuenta TigerBeetle basado en el UUID del usuario
func generateTigerBeetleAccountID(userID uuid.UUID) uint64 {
	// Convertir los primeros 8 bytes del UUID a uint64
	bytes := userID[:]
	var id uint64
	for i := 0; i < 8; i++ {
		id = (id << 8) | uint64(bytes[i])
	}

	// Asegurar que el ID no sea 1 o 2 (reservados para cuentas maestras)
	if id <= 2 {
		id += 1000
	}

	return id
}

// generateTransferID genera un ID único para una transferencia
func generateTransferID() uint64 {
	// En una implementación real, esto debería ser un ID único global
	// Por simplicidad, usamos un UUID convertido a uint64
	id := uuid.New()
	bytes := id[:]
	var transferID uint64
	for i := 0; i < 8; i++ {
		transferID = (transferID << 8) | uint64(bytes[i])
	}
	return transferID
}