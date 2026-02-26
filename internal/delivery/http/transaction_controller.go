package http

import (
	"golang-clean-architecture/internal/model"
	"golang-clean-architecture/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type TransactionController struct {
	Log     *logrus.Logger
	UseCase *usecase.TransactionUseCase
}

func NewTransactionController(useCase *usecase.TransactionUseCase, logger *logrus.Logger) *TransactionController {
	return &TransactionController{
		Log:     logger,
		UseCase: useCase,
	}
}

// GetStatus godoc
// @Summary Get Transaction Status
// @Description Retrieve the status of a transaction by its ID
// @Tags Transaction
// @Accept json
// @Produce json
// @Param transaction_id path string true "Transaction ID (UUID)"
// @Param X-Client-Key header string true "Client Key"
// @Param X-Timestamp header string true "Request Timestamp (ISO8601)"
// @Param X-Signature header string true "HMAC-SHA256 Signature"
// @Success 200 {object} model.ApiResponse
// @Failure 401 {object} model.ApiResponse
// @Failure 404 {object} model.ApiResponse
// @Router /api/transaction/status/{transaction_id} [get]
func (c *TransactionController) GetStatus(ctx *fiber.Ctx) error {
	transactionID := ctx.Params("transaction_id")
	if transactionID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(model.ApiResponse{
			Status: "error",
			Errors: "Transaction ID is required",
		})
	}

	response, err := c.UseCase.GetStatus(ctx.UserContext(), transactionID)
	if err != nil {
		c.Log.Warnf("Failed to get transaction status: %+v", err)
		return err
	}

	return ctx.JSON(model.ApiResponse{
		Status: "success",
		Data:   response,
	})
}
