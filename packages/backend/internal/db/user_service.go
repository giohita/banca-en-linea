package db

import (
	"fmt"
	"log"

	"github.com/google/uuid"

	// "banca-en-linea/backend/internal/tigerbeetle" // Comentado temporalmente
	"banca-en-linea/backend/models"
)

// UserService maneja la lógica de negocio para usuarios
type UserService struct {
	userRepo           UserRepository
	// tigerBeetleService *tigerbeetle.Service // Comentado temporalmente
}

// NewUserService crea una nueva instancia del servicio de usuarios
func NewUserService(userRepo UserRepository, tbService interface{}) *UserService {
	return &UserService{
		userRepo:           userRepo,
		// tigerBeetleService: tbService, // Comentado temporalmente
	}
}

// CreateUserWithAccount crea un usuario (sin TigerBeetle temporalmente)
func (s *UserService) CreateUserWithAccount(req *models.CreateUserRequest) (*models.User, error) {
	// 1. Crear el usuario en PostgreSQL
	user, err := s.userRepo.Create(req)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	// Temporalmente sin TigerBeetle - solo crear el usuario
	log.Printf("Successfully created user %s (TigerBeetle disabled)", user.Email)
	return user, nil
}

// GetUserWithBalance obtiene un usuario (sin balance temporalmente)
func (s *UserService) GetUserWithBalance(userID uuid.UUID) (*models.User, uint64, error) {
	// 1. Obtener el usuario de PostgreSQL
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("error getting user: %w", err)
	}

	// Temporalmente sin TigerBeetle - devolver balance 0
	return user, 0, nil
}

// DepositToUser realiza un depósito a la cuenta de un usuario (temporalmente sin TigerBeetle)
func (s *UserService) DepositToUser(userID uuid.UUID, amount uint64) error {
	// 1. Obtener el usuario
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	// Temporalmente sin TigerBeetle - solo registrar la operación
	log.Printf("TigerBeetle disabled - would deposit %d to user %s", amount, user.Email)
	return nil
}

// WithdrawFromUser realiza un retiro de la cuenta de un usuario (temporalmente sin TigerBeetle)
func (s *UserService) WithdrawFromUser(userID uuid.UUID, amount uint64) error {
	// 1. Obtener el usuario
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	// Temporalmente sin TigerBeetle - solo registrar la operación
	log.Printf("TigerBeetle disabled - would withdraw %d from user %s", amount, user.Email)
	return nil
}

// TransferBetweenUsers realiza una transferencia entre dos usuarios (temporalmente sin TigerBeetle)
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

	// Temporalmente sin TigerBeetle - solo registrar la operación
	log.Printf("TigerBeetle disabled - would transfer %d from user %s to user %s", amount, fromUser.Email, toUser.Email)
	return nil
}

// AssociateTigerBeetleAccount asocia una cuenta TigerBeetle existente a un usuario (temporalmente sin TigerBeetle)
func (s *UserService) AssociateTigerBeetleAccount(userID uuid.UUID) error {
	// 1. Obtener el usuario
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	// Temporalmente sin TigerBeetle - solo registrar la operación
	log.Printf("TigerBeetle disabled - would associate account to user %s", user.Email)
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

// GetUserByEmail obtiene un usuario por su email
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("error getting user by email: %w", err)
	}
	return user, nil
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