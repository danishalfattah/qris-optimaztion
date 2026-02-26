package config

import (
	"golang-clean-architecture/internal/delivery/http"
	"golang-clean-architecture/internal/delivery/http/middleware"
	"golang-clean-architecture/internal/delivery/http/route"
	"golang-clean-architecture/internal/repository"
	"golang-clean-architecture/internal/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type BootstrapConfig struct {
	DB          *gorm.DB
	App         *fiber.App
	Log         *logrus.Logger
	Validate    *validator.Validate
	Config      *viper.Viper
	RedisClient *redis.Client
}

func Bootstrap(config *BootstrapConfig) {
	// setup repositories
	apiClientRepository := repository.NewApiClientRepository(config.Log)
	merchantRepository := repository.NewMerchantRepository(config.Log)
	accountRepository := repository.NewAccountRepository(config.Log)
	transactionRepository := repository.NewTransactionRepository(config.Log)

	// setup use cases
	qrisUseCase := usecase.NewQrisUseCase(
		config.DB,
		config.Log,
		config.Validate,
		config.RedisClient,
		merchantRepository,
		accountRepository,
		transactionRepository,
	)
	transactionUseCase := usecase.NewTransactionUseCase(
		config.DB,
		config.Log,
		transactionRepository,
		accountRepository,
	)

	// setup controllers
	qrisController := http.NewQrisController(qrisUseCase, config.Log)
	transactionController := http.NewTransactionController(transactionUseCase, config.Log)

	// setup middleware
	hmacMiddleware := middleware.NewHMACAuth(config.DB, apiClientRepository, config.Log)

	routeConfig := route.RouteConfig{
		App:                   config.App,
		QrisController:        qrisController,
		TransactionController: transactionController,
		HMACMiddleware:        hmacMiddleware,
	}
	routeConfig.Setup()
}
