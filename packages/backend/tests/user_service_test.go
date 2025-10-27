package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"

	"banca-en-linea/backend/internal/db"
	"banca-en-linea/backend/internal/tigerbeetle"
	"banca-en-linea/backend/models"
)

// MockUserRepository es un mock del UserRepository para testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.CreateUserRequest) (*models.User, error) {
	args := m.Called(user)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(id uuid.UUID, updates *models.UpdateUserRequest) (*models.User, error) {
	args := m.Called(id, updates)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) List(limit, offset int) ([]*models.User, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateTigerBeetleAccountID(userID uuid.UUID, accountID uint64) error {
	args := m.Called(userID, accountID)
	return args.Error(0)
}

func (m *MockUserRepository) VerifyPassword(hashedPassword, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

// MockTigerBeetleService es un mock del TigerBeetle Service para testing
type MockTigerBeetleService struct {
	mock.Mock
}

func (m *MockTigerBeetleService) CreateUserAccount(userID uint64) (*types.Account, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Account), args.Error(1)
}

func (m *MockTigerBeetleService) GetAccount(accountID uint64) (*types.Account, error) {
	args := m.Called(accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Account), args.Error(1)
}

func (m *MockTigerBeetleService) GetAccountBalance(accountID uint64) (uint64, uint64, error) {
	args := m.Called(accountID)
	return args.Get(0).(uint64), args.Get(1).(uint64), args.Error(2)
}

func (m *MockTigerBeetleService) Transfer(fromAccountID, toAccountID, amount uint64, transferID uint64) error {
	args := m.Called(fromAccountID, toAccountID, amount, transferID)
	return args.Error(0)
}

func (m *MockTigerBeetleService) Deposit(userAccountID, amount, transferID uint64) error {
	args := m.Called(userAccountID, amount, transferID)
	return args.Error(0)
}

func (m *MockTigerBeetleService) Withdraw(userAccountID, amount, transferID uint64) error {
	args := m.Called(userAccountID, amount, transferID)
	return args.Error(0)
}

func (m *MockTigerBeetleService) Close() {
	m.Called()
}

func TestUserService_CreateUserWithAccount_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTB := new(MockTigerBeetleService)

	service := db.NewUserService(mockRepo, mockTB)

	req := &models.CreateUserRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	userID := uuid.New()
	createdUser := &models.User{
		ID:        userID,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	accountID := uint64(12345)
	account := &types.Account{
		ID: accountID,
	}

	// Setup mocks
	mockRepo.On("Create", req).Return(createdUser, nil)
	mockTB.On("CreateUserAccount", mock.AnythingOfType("uint64")).Return(account, nil)
	mockRepo.On("UpdateTigerBeetleAccountID", userID, mock.AnythingOfType("uint64")).Return(nil)

	// Execute
	result, err := service.CreateUserWithAccount(req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.Email, result.Email)
	assert.NotNil(t, result.TigerBeetleAccountID)

	mockRepo.AssertExpectations(t)
	mockTB.AssertExpectations(t)
}

func TestUserService_CreateUserWithAccount_TigerBeetleFailure(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTB := new(MockTigerBeetleService)

	service := db.NewUserService(mockRepo, mockTB)

	req := &models.CreateUserRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	userID := uuid.New()
	createdUser := &models.User{
		ID:        userID,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	// Setup mocks
	mockRepo.On("Create", req).Return(createdUser, nil)
	mockTB.On("CreateUserAccount", mock.AnythingOfType("uint64")).Return(nil, assert.AnError)
	mockRepo.On("Delete", userID).Return(nil) // Rollback

	// Execute
	result, err := service.CreateUserWithAccount(req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error creating TigerBeetle account")

	mockRepo.AssertExpectations(t)
	mockTB.AssertExpectations(t)
}

func TestUserService_GetUserWithBalance_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTB := new(MockTigerBeetleService)

	service := db.NewUserService(mockRepo, mockTB)

	userID := uuid.New()
	accountID := uint64(12345)
	user := &models.User{
		ID:                   userID,
		Email:                "test@example.com",
		TigerBeetleAccountID: &accountID,
	}

	debits := uint64(1000)
	credits := uint64(5000)
	expectedBalance := credits - debits // 4000

	// Setup mocks
	mockRepo.On("GetByID", userID).Return(user, nil)
	mockTB.On("GetAccountBalance", accountID).Return(debits, credits, nil)

	// Execute
	resultUser, balance, err := service.GetUserWithBalance(userID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resultUser)
	assert.Equal(t, user.Email, resultUser.Email)
	assert.Equal(t, expectedBalance, balance)

	mockRepo.AssertExpectations(t)
	mockTB.AssertExpectations(t)
}

func TestUserService_GetUserWithBalance_NoTigerBeetleAccount(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTB := new(MockTigerBeetleService)

	service := db.NewUserService(mockRepo, mockTB)

	userID := uuid.New()
	user := &models.User{
		ID:                   userID,
		Email:                "test@example.com",
		TigerBeetleAccountID: nil, // No TigerBeetle account
	}

	// Setup mocks
	mockRepo.On("GetByID", userID).Return(user, nil)
	// No TigerBeetle calls expected

	// Execute
	resultUser, balance, err := service.GetUserWithBalance(userID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resultUser)
	assert.Equal(t, user.Email, resultUser.Email)
	assert.Equal(t, uint64(0), balance)

	mockRepo.AssertExpectations(t)
	mockTB.AssertExpectations(t)
}

func TestUserService_DepositToUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTB := new(MockTigerBeetleService)

	service := db.NewUserService(mockRepo, mockTB)

	userID := uuid.New()
	accountID := uint64(12345)
	amount := uint64(10000) // 100.00 HNL

	user := &models.User{
		ID:                   userID,
		Email:                "test@example.com",
		TigerBeetleAccountID: &accountID,
	}

	// Setup mocks
	mockRepo.On("GetByID", userID).Return(user, nil)
	mockTB.On("Deposit", accountID, amount, mock.AnythingOfType("uint64")).Return(nil)

	// Execute
	err := service.DepositToUser(userID, amount)

	// Assert
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockTB.AssertExpectations(t)
}

func TestUserService_DepositToUser_NoTigerBeetleAccount(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTB := new(MockTigerBeetleService)

	service := db.NewUserService(mockRepo, mockTB)

	userID := uuid.New()
	amount := uint64(10000)

	user := &models.User{
		ID:                   userID,
		Email:                "test@example.com",
		TigerBeetleAccountID: nil, // No TigerBeetle account
	}

	// Setup mocks
	mockRepo.On("GetByID", userID).Return(user, nil)

	// Execute
	err := service.DepositToUser(userID, amount)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not have a TigerBeetle account")

	mockRepo.AssertExpectations(t)
	mockTB.AssertExpectations(t)
}

func TestUserService_WithdrawFromUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTB := new(MockTigerBeetleService)

	service := db.NewUserService(mockRepo, mockTB)

	userID := uuid.New()
	accountID := uint64(12345)
	amount := uint64(5000) // 50.00 HNL

	user := &models.User{
		ID:                   userID,
		Email:                "test@example.com",
		TigerBeetleAccountID: &accountID,
	}

	// Balance suficiente: credits 10000, debits 0 = balance 10000
	debits := uint64(0)
	credits := uint64(10000)

	// Setup mocks
	mockRepo.On("GetByID", userID).Return(user, nil)
	mockTB.On("GetAccountBalance", accountID).Return(debits, credits, nil)
	mockTB.On("Withdraw", accountID, amount, mock.AnythingOfType("uint64")).Return(nil)

	// Execute
	err := service.WithdrawFromUser(userID, amount)

	// Assert
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockTB.AssertExpectations(t)
}

func TestUserService_WithdrawFromUser_InsufficientFunds(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTB := new(MockTigerBeetleService)

	service := db.NewUserService(mockRepo, mockTB)

	userID := uuid.New()
	accountID := uint64(12345)
	amount := uint64(15000) // 150.00 HNL (más que el balance)

	user := &models.User{
		ID:                   userID,
		Email:                "test@example.com",
		TigerBeetleAccountID: &accountID,
	}

	// Balance insuficiente: credits 10000, debits 0 = balance 10000
	debits := uint64(0)
	credits := uint64(10000)

	// Setup mocks
	mockRepo.On("GetByID", userID).Return(user, nil)
	mockTB.On("GetAccountBalance", accountID).Return(debits, credits, nil)

	// Execute
	err := service.WithdrawFromUser(userID, amount)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient funds")

	mockRepo.AssertExpectations(t)
	mockTB.AssertExpectations(t)
}

func TestUserService_TransferBetweenUsers_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTB := new(MockTigerBeetleService)

	service := db.NewUserService(mockRepo, mockTB)

	fromUserID := uuid.New()
	toUserID := uuid.New()
	fromAccountID := uint64(12345)
	toAccountID := uint64(67890)
	amount := uint64(5000) // 50.00 HNL

	fromUser := &models.User{
		ID:                   fromUserID,
		Email:                "from@example.com",
		TigerBeetleAccountID: &fromAccountID,
	}

	toUser := &models.User{
		ID:                   toUserID,
		Email:                "to@example.com",
		TigerBeetleAccountID: &toAccountID,
	}

	// Balance suficiente en cuenta origen
	debits := uint64(0)
	credits := uint64(10000)

	// Setup mocks
	mockRepo.On("GetByID", fromUserID).Return(fromUser, nil)
	mockRepo.On("GetByID", toUserID).Return(toUser, nil)
	mockTB.On("GetAccountBalance", fromAccountID).Return(debits, credits, nil)
	mockTB.On("Transfer", fromAccountID, toAccountID, amount, mock.AnythingOfType("uint64")).Return(nil)

	// Execute
	err := service.TransferBetweenUsers(fromUserID, toUserID, amount)

	// Assert
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockTB.AssertExpectations(t)
}
