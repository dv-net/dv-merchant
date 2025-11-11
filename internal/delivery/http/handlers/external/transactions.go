package external

import (
	"errors"
	"net/http"

	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/transaction_response"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/response"
	"github.com/gofiber/fiber/v3"
)

// getUnconfirmedTransfer is a function to get unconfirmed transfer transactions
//
//	@Summary		Get unconfirmed transfer transactions
//	@Description	Get unconfirmed transfer transactions
//	@Tags			Transactions
//	@Accept			json
//	@Produce		json
//	@Param			api_key	query		string	false	"Store API key"
//	@Success		200		{object}	response.Result[transaction_response.UnconfirmedTransactionResponse]
//	@Failure		401		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/external/transactions/unconfirmed/transfer [get]
//	@Security		XApiKey
func (h *Handler) getUnconfirmedTransfer(c fiber.Ctx) error {
	store, err := loadAuthStore(c)
	if err != nil {
		return err
	}
	uTxs, err := h.services.UnconfirmedTransactionService.GetUnconfirmedTransactions(c.Context(), store.UserID, models.TransactionsTypeTransfer)
	if err != nil {
		return apierror.New().AddError(errors.New("error getting unconfirmed transactions")).SetHttpCode(http.StatusInternalServerError)
	}

	return c.JSON(response.OkByData(transaction_response.NewFromUnconfirmedTransactionModels(uTxs)))
}

func (h *Handler) initTransactionsRouter(v1 fiber.Router) {
	w := v1.Group("/transactions")
	w.Get("/unconfirmed/transfer", h.getUnconfirmedTransfer)
}
