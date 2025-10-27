package tigerbeetle

// TigerBeetleService define la interfaz común para el servicio TigerBeetle
type TigerBeetleService interface {
	Close()
	CreateUserAccount(userID uint64) (AccountInterface, error)
	GetAccount(accountID uint64) (AccountInterface, error)
	GetAccountBalance(accountID uint64) (uint64, uint64, error)
	Transfer(fromAccountID, toAccountID, amount uint64, transferID uint64) error
	Deposit(userAccountID, amount, transferID uint64) error
	Withdraw(userAccountID, amount, transferID uint64) error
}

// AccountInterface define la interfaz común para las cuentas
type AccountInterface interface {
	GetID() uint64
	GetLedger() uint32
	GetCode() uint16
	GetFlags() uint16
	GetDebitsPosted() uint64
	GetCreditsPosted() uint64
}