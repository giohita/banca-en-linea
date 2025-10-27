package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"banca-en-linea/backend/models"
)

// UserRepository define la interfaz para operaciones de usuario en la base de datos
type UserRepository interface {
	Create(user *models.CreateUserRequest) (*models.User, error)
	GetByID(id uuid.UUID) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(id uuid.UUID, updates *models.UpdateUserRequest) (*models.User, error)
	Delete(id uuid.UUID) error
	List(limit, offset int) ([]*models.User, error)
	UpdateTigerBeetleAccountID(userID uuid.UUID, accountID int64) error
	VerifyPassword(hashedPassword, password string) error
}

// userRepository implementa UserRepository
type userRepository struct {
	db *sql.DB
}

// NewUserRepository crea una nueva instancia del repositorio de usuarios
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// Create crea un nuevo usuario en la base de datos
func (r *userRepository) Create(req *models.CreateUserRequest) (*models.User, error) {
	// Hash de la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	user := &models.User{
		ID:            uuid.New(),
		Email:         req.Email,
		PasswordHash:  string(hashedPassword),
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		IsActive:      true,
		EmailVerified: false,
	}

	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, created_at, updated_at, is_active, email_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, email, first_name, last_name, phone, date_of_birth, tigerbeetle_account_id, created_at, updated_at, is_active, email_verified`

	err = r.db.QueryRow(
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.CreatedAt,
		user.UpdatedAt,
		user.IsActive,
		user.EmailVerified,
	).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.DateOfBirth,
		&user.TigerBeetleAccountID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
		&user.EmailVerified,
	)

	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return user, nil
}

// GetByID obtiene un usuario por su ID
func (r *userRepository) GetByID(id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, date_of_birth, 
		       tigerbeetle_account_id, created_at, updated_at, is_active, email_verified
		FROM users 
		WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.DateOfBirth,
		&user.TigerBeetleAccountID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
		&user.EmailVerified,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return user, nil
}

// GetByEmail obtiene un usuario por su email
func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, date_of_birth,
		       tigerbeetle_account_id, created_at, updated_at, is_active, email_verified
		FROM users 
		WHERE email = $1`

	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.DateOfBirth,
		&user.TigerBeetleAccountID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
		&user.EmailVerified,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return user, nil
}

// Update actualiza un usuario existente
func (r *userRepository) Update(id uuid.UUID, updates *models.UpdateUserRequest) (*models.User, error) {
	// Construir la consulta dinámicamente basada en los campos a actualizar
	setParts := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIndex := 1

	if updates.Email != nil {
		setParts = append(setParts, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, *updates.Email)
		argIndex++
	}

	if updates.FirstName != nil {
		setParts = append(setParts, fmt.Sprintf("first_name = $%d", argIndex))
		args = append(args, *updates.FirstName)
		argIndex++
	}

	if updates.LastName != nil {
		setParts = append(setParts, fmt.Sprintf("last_name = $%d", argIndex))
		args = append(args, *updates.LastName)
		argIndex++
	}

	// Agregar el ID al final de los argumentos
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE users 
		SET %s
		WHERE id = $%d
		RETURNING id, email, first_name, last_name, phone, date_of_birth, tigerbeetle_account_id, created_at, updated_at, is_active, email_verified`,
		strings.Join(setParts, ", "),
		argIndex,
	)

	user := &models.User{}
	err := r.db.QueryRow(query, args...).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.DateOfBirth,
		&user.TigerBeetleAccountID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
		&user.EmailVerified,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return user, nil
}

// Delete realiza un soft delete del usuario
func (r *userRepository) Delete(id uuid.UUID) error {
	query := `
		UPDATE users 
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// List obtiene una lista paginada de usuarios
func (r *userRepository) List(limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, email, first_name, last_name, phone, date_of_birth, tigerbeetle_account_id, created_at, updated_at, is_active, email_verified
		FROM users 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Phone,
			&user.DateOfBirth,
			&user.TigerBeetleAccountID,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.IsActive,
			&user.EmailVerified,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// UpdateTigerBeetleAccountID actualiza el ID de cuenta de TigerBeetle para un usuario
func (r *userRepository) UpdateTigerBeetleAccountID(userID uuid.UUID, accountID int64) error {
	query := `
		UPDATE users 
		SET tigerbeetle_account_id = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.Exec(query, accountID, userID)
	if err != nil {
		return fmt.Errorf("error updating tigerbeetle account id: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// VerifyPassword verifica si una contraseña coincide con el hash almacenado
func (r *userRepository) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
