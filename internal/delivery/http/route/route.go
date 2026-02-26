package route

import (
	"golang-clean-architecture/internal/delivery/http"

	"github.com/gofiber/fiber/v2"
)

type RouteConfig struct {
	App                   *fiber.App
	QrisController        *http.QrisController
	TransactionController *http.TransactionController
	HMACMiddleware        fiber.Handler
}

func (c *RouteConfig) Setup() {
	// Health check (no auth required)
	c.App.Get("/health", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"status": "ok",
		})
	})

	// API routes with HMAC signature authentication
	api := c.App.Group("/api", c.HMACMiddleware)

	// QRIS endpoints
	api.Get("/qris/inquiry/:qris_payload", c.QrisController.Inquiry)
	api.Post("/qris/payment", c.QrisController.Payment)

	// Transaction endpoints
	api.Get("/transaction/status/:transaction_id", c.TransactionController.GetStatus)
}
