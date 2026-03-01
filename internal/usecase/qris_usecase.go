package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"golang-clean-architecture/internal/entity"
	"golang-clean-architecture/internal/model"
	"golang-clean-architecture/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type QrisUseCase struct {
	DB                    *gorm.DB
	Log                   *logrus.Logger
	Validate              *validator.Validate
	RedisClient           *redis.Client
	MerchantRepository    *repository.MerchantRepository
	AccountRepository     *repository.AccountRepository
	TransactionRepository *repository.TransactionRepository
}

func NewQrisUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	redisClient *redis.Client,
	merchantRepo *repository.MerchantRepository,
	accountRepo *repository.AccountRepository,
	transactionRepo *repository.TransactionRepository,
) *QrisUseCase {
	return &QrisUseCase{
		DB:                    db,
		Log:                   log,
		Validate:              validate,
		RedisClient:           redisClient,
		MerchantRepository:    merchantRepo,
		AccountRepository:     accountRepo,
		TransactionRepository: transactionRepo,
	}
}

// Inquiry processes a QRIS payload and returns merchant information
func (u *QrisUseCase) Inquiry(ctx context.Context, qrisPayload string) (*model.InquiryResponse, *model.Metadata, error) {
	start := time.Now()
	source := "database"

	var merchantID, merchantName, city string

	// Try to find merchant data from Redis cache first
	cacheKey := fmt.Sprintf("merchant:%s", qrisPayload)
	cachedData, err := u.RedisClient.Get(ctx, cacheKey).Result()

	if err == nil {
		// Cache hit — only merchant data, NOT inquiry_id
		var merchantCache map[string]string
		if err := json.Unmarshal([]byte(cachedData), &merchantCache); err == nil {
			merchantID = merchantCache["merchant_id"]
			merchantName = merchantCache["merchant_name"]
			city = merchantCache["city"]
			source = "cache"
		}
	}

	// Cache miss — lookup from database
	if merchantID == "" {
		tx := u.DB.WithContext(ctx)
		merchants, err := u.MerchantRepository.FindAll(tx)
		if err != nil || len(merchants) == 0 {
			u.Log.Warnf("No active merchants found: %+v", err)
			return nil, nil, fiber.NewError(fiber.StatusNotFound, "Merchant not found")
		}

		merchant := merchants[0]
		merchantID = merchant.MerchantID
		merchantName = merchant.MerchantName
		city = merchant.City

		// Cache merchant data only (no inquiry_id)
		merchantCache, _ := json.Marshal(map[string]string{
			"merchant_id":   merchantID,
			"merchant_name": merchantName,
			"city":          city,
		})
		u.RedisClient.Set(ctx, cacheKey, merchantCache, 5*time.Minute)
	}

	// Always generate a FRESH inquiry_id (never cached)
	inquiryID := fmt.Sprintf("inq_%s", uuid.New().String()[:6])

	// Store inquiry session in Redis (valid for 5 minutes, one-time use)
	inquiryData, _ := json.Marshal(map[string]interface{}{
		"merchant_id":   merchantID,
		"merchant_name": merchantName,
		"qris_payload":  qrisPayload,
	})
	u.RedisClient.Set(ctx, fmt.Sprintf("inquiry:%s", inquiryID), inquiryData, 5*time.Minute)

	response := &model.InquiryResponse{
		MerchantID:   merchantID,
		MerchantName: merchantName,
		TerminalID:   "T001",
		City:         city,
		FixedAmount:  0,
		InquiryID:    inquiryID,
	}

	latency := time.Since(start).Milliseconds()
	return response, &model.Metadata{
		LatencyMs: latency,
		Source:    source,
	}, nil
}

// Payment processes a QRIS payment
func (u *QrisUseCase) Payment(ctx context.Context, request *model.PaymentRequest) (*model.PaymentResponse, error) {
	// Validate request
	if err := u.Validate.Struct(request); err != nil {
		u.Log.Warnf("Invalid payment request: %+v", err)
		return nil, fiber.ErrBadRequest
	}

	// Validate inquiry_id from Redis
	inquiryKey := fmt.Sprintf("inquiry:%s", request.InquiryID)
	inquiryData, err := u.RedisClient.Get(ctx, inquiryKey).Result()
	if err != nil {
		u.Log.Warnf("Invalid or expired inquiry_id: %s", request.InquiryID)
		return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid or expired inquiry ID")
	}

	// Parse inquiry data to get merchant_id
	var inquiry map[string]interface{}
	if err := json.Unmarshal([]byte(inquiryData), &inquiry); err != nil {
		u.Log.Warnf("Failed to parse inquiry data: %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	merchantID, _ := inquiry["merchant_id"].(string)

	// Start database transaction (explicit, since SkipDefaultTransaction is enabled)
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Find account by user_id
	account := new(entity.Account)
	if err := u.AccountRepository.FindByAccountID(tx, account, request.UserID); err != nil {
		u.Log.Warnf("Account not found for user: %s, error: %+v", request.UserID, err)
		return nil, fiber.NewError(fiber.StatusNotFound, "Account not found")
	}

	// Verify PIN
	if err := bcrypt.CompareHashAndPassword([]byte(account.PinHash), []byte(request.Pincode)); err != nil {
		u.Log.Warnf("Invalid PIN for user: %s", request.UserID)
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid PIN")
	}

	// Check sufficient balance
	if account.Balance < request.Amount {
		u.Log.Warnf("Insufficient balance for user: %s", request.UserID)
		return nil, fiber.NewError(fiber.StatusBadRequest, "Insufficient balance")
	}

	// Create transaction record
	traceID := uuid.New().String()
	transactionID := uuid.New().String()
	transaction := &entity.Transaction{
		TransactionID: transactionID,
		TraceID:       traceID,
		AccountID:     request.UserID,
		MerchantID:    merchantID,
		Amount:        request.Amount,
		Status:        "PENDING",
	}

	if err := u.TransactionRepository.Create(tx, transaction); err != nil {
		u.Log.Warnf("Failed to create transaction: %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	// Deduct balance with optimistic locking
	if err := u.AccountRepository.DeductBalance(tx, request.UserID, request.Amount, account.Version); err != nil {
		u.Log.Warnf("Failed to deduct balance (optimistic lock conflict): %+v", err)
		return nil, fiber.NewError(fiber.StatusConflict, "Transaction conflict, please retry")
	}

	// Update transaction status to SUCCESS
	if err := u.TransactionRepository.UpdateStatus(tx, transactionID, "SUCCESS"); err != nil {
		u.Log.Warnf("Failed to update transaction status: %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Warnf("Failed to commit transaction: %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	// Delete the inquiry from Redis (one-time use)
	u.RedisClient.Del(ctx, inquiryKey)

	// Optimization: Cache transaction status in Redis for fast status lookups
	finalBalance := account.Balance - request.Amount
	statusCache := &model.TransactionStatusResponse{
		TransactionID: transactionID,
		Status:        "SUCCESS",
		FinalBalance:  finalBalance,
		Timestamp:     transaction.CreatedAt.Format(time.RFC3339),
	}
	if cacheJSON, err := json.Marshal(statusCache); err == nil {
		u.RedisClient.Set(ctx, fmt.Sprintf("txstatus:%s", transactionID), cacheJSON, 5*time.Minute)
	}

	return &model.PaymentResponse{
		Status:              "processing",
		TransactionID:       transactionID,
		Message:             "Transaksi sedang diproses",
		EstimatedCompletion: "200ms",
	}, nil
}
