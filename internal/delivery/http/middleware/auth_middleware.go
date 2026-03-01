package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

func NewHMACAuth(db *gorm.DB, redisClient *redis.Client, apiClientRepo *repository.ApiClientRepository, log *logrus.Logger) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		clientKey := ctx.Get("X-Client-Key")
		timestamp := ctx.Get("X-Timestamp")
		signature := ctx.Get("X-Signature")

		if clientKey == "" || timestamp == "" || signature == "" {
			log.Warn("Missing required auth headers")
			return ctx.Status(fiber.StatusUnauthorized).JSON(model.ApiResponse{
				Status: "error",
				Errors: "Missing required headers: X-Client-Key, X-Timestamp, X-Signature",
			})
		}

		// Look up client — check Redis cache first
		var clientSecret string
		cacheKey := fmt.Sprintf("apiclient:%s", clientKey)

		cachedSecret, err := redisClient.Get(context.Background(), cacheKey).Result()
		if err == nil {
			// Cache hit
			clientSecret = cachedSecret
		} else {
			// Cache miss — query DB
			client := new(entity.ApiClient)
			if err := apiClientRepo.FindByClientID(db, client, clientKey); err != nil {
				log.Warnf("Invalid client key: %s, error: %+v", clientKey, err)
				return ctx.Status(fiber.StatusUnauthorized).JSON(model.ApiResponse{
					Status: "error",
					Errors: "Invalid client key",
				})
			}
			clientSecret = client.ClientSecret

			// Cache for 10 minutes
			redisClient.Set(context.Background(), cacheKey, clientSecret, 10*time.Minute)
		}

		// Build payload: method + path + timestamp + body
		body := ctx.Body()
		payload := ctx.Method() + ctx.Path() + timestamp + string(body)

		// Compute HMAC-SHA256
		mac := hmac.New(sha256.New, []byte(clientSecret))
		mac.Write([]byte(payload))
		expectedSignature := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			log.Warnf("Invalid HMAC signature for client: %s", clientKey)
			return ctx.Status(fiber.StatusUnauthorized).JSON(model.ApiResponse{
				Status: "error",
				Errors: "Invalid signature",
			})
		}

		log.Debugf("Authenticated client: %s", clientKey)
		ctx.Locals("client_id", clientKey)
		return ctx.Next()
	}
}
