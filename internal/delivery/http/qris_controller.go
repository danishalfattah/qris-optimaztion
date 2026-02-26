package http

import (
	"golang-clean-architecture/internal/model"
	"golang-clean-architecture/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type QrisController struct {
	Log     *logrus.Logger
	UseCase *usecase.QrisUseCase
}

func NewQrisController(useCase *usecase.QrisUseCase, logger *logrus.Logger) *QrisController {
	return &QrisController{
		Log:     logger,
		UseCase: useCase,
	}
}

// Inquiry godoc
// @Summary QRIS Inquiry
// @Description Parse QRIS payload and return merchant information with caching
// @Tags QRIS
// @Accept json
// @Produce json
// @Param qris_payload path string true "QRIS Payload string"
// @Param X-Client-Key header string true "Client Key"
// @Param X-Timestamp header string true "Request Timestamp (ISO8601)"
// @Param X-Signature header string true "HMAC-SHA256 Signature"
// @Success 200 {object} model.ApiResponse
// @Failure 401 {object} model.ApiResponse
// @Failure 404 {object} model.ApiResponse
// @Router /api/qris/inquiry/{qris_payload} [get]
func (c *QrisController) Inquiry(ctx *fiber.Ctx) error {
	qrisPayload := ctx.Params("qris_payload")
	if qrisPayload == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(model.ApiResponse{
			Status: "error",
			Errors: "QRIS payload is required",
		})
	}

	response, metadata, err := c.UseCase.Inquiry(ctx.UserContext(), qrisPayload)
	if err != nil {
		c.Log.Warnf("Failed to process QRIS inquiry: %+v", err)
		return err
	}

	return ctx.JSON(model.ApiResponse{
		Status:   "success",
		Data:     response,
		Metadata: metadata,
	})
}

// Payment godoc
// @Summary QRIS Payment
// @Description Process a QRIS payment with PIN verification and balance deduction
// @Tags QRIS
// @Accept json
// @Produce json
// @Param X-Client-Key header string true "Client Key"
// @Param X-Timestamp header string true "Request Timestamp (ISO8601)"
// @Param X-Signature header string true "HMAC-SHA256 Signature"
// @Param request body model.PaymentRequest true "Payment Request"
// @Success 200 {object} model.ApiResponse
// @Failure 400 {object} model.ApiResponse
// @Failure 401 {object} model.ApiResponse
// @Failure 404 {object} model.ApiResponse
// @Router /api/qris/payment [post]
func (c *QrisController) Payment(ctx *fiber.Ctx) error {
	request := new(model.PaymentRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.Warnf("Failed to parse payment request body: %+v", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(model.ApiResponse{
			Status: "error",
			Errors: "Invalid request body",
		})
	}

	response, err := c.UseCase.Payment(ctx.UserContext(), request)
	if err != nil {
		c.Log.Warnf("Failed to process payment: %+v", err)
		return err
	}

	return ctx.JSON(model.ApiResponse{
		Status: "success",
		Data:   response,
	})
}
