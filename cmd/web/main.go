package main

import (
	"fmt"
	"golang-clean-architecture/internal/config"
)

// @title QRIS Payment API
// @version 1.0
// @description QRIS Payment Processing API with HMAC-SHA256 Authentication
// @host localhost:3000
// @BasePath /
func main() {
	viperConfig := config.NewViper()
	log := config.NewLogger(viperConfig)
	db := config.NewDatabase(viperConfig, log)
	validate := config.NewValidator(viperConfig)
	app := config.NewFiber(viperConfig)
	redisClient := config.NewRedis(viperConfig, log)

	config.Bootstrap(&config.BootstrapConfig{
		DB:          db,
		App:         app,
		Log:         log,
		Validate:    validate,
		Config:      viperConfig,
		RedisClient: redisClient,
	})

	webPort := viperConfig.GetInt("web.port")
	err := app.Listen(fmt.Sprintf(":%d", webPort))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
