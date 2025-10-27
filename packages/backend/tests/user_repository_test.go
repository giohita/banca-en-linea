package tests

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"banca-en-linea/backend/internal/db"
	"banca-en-linea/backend/models"
)

// setupTestDB configura una base de datos de prueba en memoria
func setupTestDB(t *testing.T) *sql.DB {
	// Para pruebas reales, necesitarías una base de datos PostgreSQL de prueba
	// Por simplicidad, este ejemplo asume que tienes una DB de prueba configurada

	// Configuración de prueba - ajustar según tu entorno
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=banca_en_linea_test sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	// Verificar conexión
	err = db.Ping()
	if err != nil {
		t.Skip("PostgreSQL test database not available")
	}

	// Limpiar la tabla antes de cada prueba
	_, err = db.Exec("DELETE FROM users")
	require.NoError(t, err)

	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := db.NewUserRepository(db)

	req := &models.CreateUserRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	user, err := repo.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, req.Email, user.Email)
	assert.Equal(t, req.FirstName, user.FirstName)
	assert.Equal(t, req.LastName, user.LastName)
	assert.NotEmpty(t, user.ID)
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotEqual(t, req.Password, user.PasswordHash) // Password should be hashed
	assert.WithinDuration(t, time.Now(), user.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), user.UpdatedAt, time.Second)
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := db.NewUserRepository(db)

	req := &models.CreateUserRequest{
		Email:     "duplicate@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	// Crear el primer usuario
	_, err := repo.Create(req)
	assert.NoError(t, err)

	// Intentar crear un usuario con el mismo email
	_, err = repo.Create(req)
	assert.Error(t, err) // Debería fallar por email duplicado
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := db.NewUserRepository(db)

	// Crear un usuario primero
	req := &models.CreateUserRequest{
		Email:     "getbyid@example.com",
		Password:  "password123",
		FirstName: "Get By ID",
		LastName:  "User",
	}

	createdUser, err := repo.Create(req)
	require.NoError(t, err)

	// Obtener el usuario por ID
	foundUser, err := repo.GetByID(createdUser.ID)

	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, createdUser.ID, foundUser.ID)
	assert.Equal(t, createdUser.Email, foundUser.Email)
	assert.Equal(t, createdUser.FirstName, foundUser.FirstName)
	assert.Equal(t, createdUser.LastName, foundUser.LastName)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := db.NewUserRepository(db)

	// Intentar obtener un usuario que no existe
	nonExistentID := uuid.New()
	_, err := repo.GetByID(nonExistentID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := db.NewUserRepository(db)

	// Crear un usuario primero
	req := &models.CreateUserRequest{
		Email:     "getbyemail@example.com",
		Password:  "password123",
		FirstName: "Get By Email",
		LastName:  "User",
	}

	createdUser, err := repo.Create(req)
	require.NoError(t, err)

	// Obtener el usuario por email
	foundUser, err := repo.GetByEmail(createdUser.Email)

	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, createdUser.ID, foundUser.ID)
	assert.Equal(t, createdUser.Email, foundUser.Email)
	assert.Equal(t, createdUser.FirstName, foundUser.FirstName)
	assert.Equal(t, createdUser.LastName, foundUser.LastName)
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := db.NewUserRepository(db)

	// Crear un usuario primero
	req := &models.CreateUserRequest{
		Email:     "update@example.com",
		Password:  "password123",
		FirstName: "Original",
		LastName:  "Name",
	}

	createdUser, err := repo.Create(req)
	require.NoError(t, err)

	// Actualizar el usuario
	newEmail := "updated@example.com"
	newFirstName := "Updated"
	newLastName := "Name"
	updateReq := &models.UpdateUserRequest{
		Email:     &newEmail,
		FirstName: &newFirstName,
		LastName:  &newLastName,
	}

	updatedUser, err := repo.Update(createdUser.ID, updateReq)

	assert.NoError(t, err)
	assert.NotNil(t, updatedUser)
	assert.Equal(t, newEmail, updatedUser.Email)
	assert.Equal(t, newFirstName, updatedUser.FirstName)
	assert.Equal(t, newLastName, updatedUser.LastName)
	assert.True(t, updatedUser.UpdatedAt.After(createdUser.UpdatedAt))
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := db.NewUserRepository(db)

	// Crear un usuario primero
	req := &models.CreateUserRequest{
		Email:     "delete@example.com",
		Password:  "password123",
		FirstName: "Delete",
		LastName:  "User",
	}

	createdUser, err := repo.Create(req)
	require.NoError(t, err)

	// Eliminar el usuario
	err = repo.Delete(createdUser.ID)
	assert.NoError(t, err)

	// Verificar que el usuario ya no se puede encontrar
	_, err = repo.GetByID(createdUser.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := db.NewUserRepository(db)

	// Crear varios usuarios
	for i := 0; i < 5; i++ {
		req := &models.CreateUserRequest{
			Email:     fmt.Sprintf("list%d@example.com", i),
			Password:  "password123",
			FirstName: fmt.Sprintf("List User %d", i),
			LastName:  "Test",
		}
		_, err := repo.Create(req)
		require.NoError(t, err)
	}

	// Obtener lista de usuarios
	users, err := repo.List(3, 0) // Limit 3, offset 0

	assert.NoError(t, err)
	assert.Len(t, users, 3)

	// Verificar paginación
	moreUsers, err := repo.List(3, 3) // Limit 3, offset 3
	assert.NoError(t, err)
	assert.Len(t, moreUsers, 2) // Deberían quedar 2 usuarios
}

func TestUserRepository_UpdateTigerBeetleAccountID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := db.NewUserRepository(db)

	// Crear un usuario primero
	req := &models.CreateUserRequest{
		Email:     "tigerbeetle@example.com",
		Password:  "password123",
		FirstName: "TigerBeetle",
		LastName:  "User",
	}

	createdUser, err := repo.Create(req)
	require.NoError(t, err)

	// Actualizar el TigerBeetle Account ID
	accountID := uint64(12345)
	err = repo.UpdateTigerBeetleAccountID(createdUser.ID, accountID)
	assert.NoError(t, err)

	// Verificar que se actualizó correctamente
	updatedUser, err := repo.GetByID(createdUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedUser.TigerBeetleAccountID)
	assert.Equal(t, accountID, *updatedUser.TigerBeetleAccountID)
}

func TestUserRepository_VerifyPassword(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := db.NewUserRepository(db)

	password := "testpassword123"
	req := &models.CreateUserRequest{
		Email:     "password@example.com",
		Password:  password,
		FirstName: "Password",
		LastName:  "User",
	}

	createdUser, err := repo.Create(req)
	require.NoError(t, err)

	// Verificar contraseña correcta
	err = repo.VerifyPassword(createdUser.PasswordHash, password)
	assert.NoError(t, err)

	// Verificar contraseña incorrecta
	err = repo.VerifyPassword(createdUser.PasswordHash, "wrongpassword")
	assert.Error(t, err)
}
