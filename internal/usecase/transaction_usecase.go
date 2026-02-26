package usecase

import (
	"context"
	"time"

	"golang-clean-architecture/internal/entity"
	"golang-clean-architecture/internal/model"
	"golang-clean-architecture/internal/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type TransactionUseCase struct {
	DB                    *gorm.DB
	Log                   *logrus.Logger
	TransactionRepository *repository.TransactionRepository
	AccountRepository     *repository.AccountRepository
}

func NewTransactionUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	transactionRepo *repository.TransactionRepository,
	accountRepo *repository.AccountRepository,
) *TransactionUseCase {
	return &TransactionUseCase{
		DB:                    db,
		Log:                   log,
		TransactionRepository: transactionRepo,
		AccountRepository:     accountRepo,
	}
}

// GetStatus returns the status of a transaction
func (u *TransactionUseCase) GetStatus(ctx context.Context, transactionID string) (*model.TransactionStatusResponse, error) {
	tx := u.DB.WithContext(ctx)

	transaction := new(entity.Transaction)
	if err := u.TransactionRepository.FindByTransactionID(tx, transaction, transactionID); err != nil {
		u.Log.Warnf("Transaction not found: %s, error: %+v", transactionID, err)
		return nil, fiber.NewError(fiber.StatusNotFound, "Transaction not found")
	}

	// Get current balance
	account := new(entity.Account)
	var finalBalance float64
	if err := u.AccountRepository.FindByAccountID(tx, account, transaction.AccountID); err == nil {
		finalBalance = account.Balance
	}

	return &model.TransactionStatusResponse{
		TransactionID: transaction.TransactionID,
		Status:        transaction.Status,
		FinalBalance:  finalBalance,
		Timestamp:     transaction.CreatedAt.Format(time.RFC3339),
	}, nil
}
