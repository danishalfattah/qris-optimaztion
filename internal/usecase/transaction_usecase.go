package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"golang-clean-architecture/internal/entity"
	"golang-clean-architecture/internal/model"
	"golang-clean-architecture/internal/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type TransactionUseCase struct {
	DB                    *gorm.DB
	Log                   *logrus.Logger
	RedisClient           *redis.Client
	TransactionRepository *repository.TransactionRepository
	AccountRepository     *repository.AccountRepository
}

func NewTransactionUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	redisClient *redis.Client,
	transactionRepo *repository.TransactionRepository,
	accountRepo *repository.AccountRepository,
) *TransactionUseCase {
	return &TransactionUseCase{
		DB:                    db,
		Log:                   log,
		RedisClient:           redisClient,
		TransactionRepository: transactionRepo,
		AccountRepository:     accountRepo,
	}
}

// GetStatus returns the status of a transaction — checks Redis cache first
func (u *TransactionUseCase) GetStatus(ctx context.Context, transactionID string) (*model.TransactionStatusResponse, error) {
	// Optimization: check Redis cache first
	cacheKey := fmt.Sprintf("txstatus:%s", transactionID)
	cachedData, err := u.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var cached model.TransactionStatusResponse
		if err := json.Unmarshal([]byte(cachedData), &cached); err == nil {
			return &cached, nil
		}
	}

	// Cache miss — query database
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

	response := &model.TransactionStatusResponse{
		TransactionID: transaction.TransactionID,
		Status:        transaction.Status,
		FinalBalance:  finalBalance,
		Timestamp:     transaction.CreatedAt.Format(time.RFC3339),
	}

	// Cache for future lookups
	if cacheJSON, err := json.Marshal(response); err == nil {
		u.RedisClient.Set(ctx, cacheKey, cacheJSON, 5*time.Minute)
	}

	return response, nil
}
