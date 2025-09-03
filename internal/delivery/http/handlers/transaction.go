package handlers

import (
	"errors"
	"fmt"

	// Blank import for swag-gen
	_ "github.com/dv-net/dv-merchant/internal/delivery/http/responses/transaction_response"
	_ "github.com/dv-net/dv-merchant/internal/models"
	_ "github.com/dv-net/dv-merchant/internal/storage/repos/repo_transactions"
	_ "github.com/dv-net/dv-merchant/internal/storage/storecmn"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/transactions_request"
	"github.com/dv-net/dv-merchant/internal/service/transactions"
	"github.com/dv-net/dv-merchant/internal/tools"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

// loadUserTransaction is a function to Load user transactions
//
//	@Summary		Load user transactions
//	@Description	Load user transactions
//	@Tags			Transaction
//	@Accept			json
//	@Produce		json
//	@Param			string	query		transactions_request.GetByUser	true	"GetTransactionsByUserParams"
//	@Success		200		{object}	response.Result[storecmn.FindResponseWithFullPagination[repo_transactions.FindRow]]
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Router			/v1/dv-admin/transaction/ [get]
//	@Security		BearerAuth
func (h Handler) loadUserTransaction(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &transactions_request.GetByUser{}
	if err := c.Bind().Query(req); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	dto := transactions.RequestToGetUserTransactionsDTO(req)

	res, err := h.services.TransactionService.GetUserTransactions(c.Context(), user.ID, dto)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(res))
}

// loadUserTransaction is a function to Load user transactions
//
//	@Summary		Load user transactions statistic by date
//	@Description	Load user transactions statistic by date
//	@Tags			Transaction
//	@Accept			json
//	@Produce		json
//	@Param			string	query		transactions_request.GetStatistics	true	"GetStatisticsParams"
//	@Success		200		{object}	response.Result[repo_transactions.StatisticsRow[]]
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Router			/v1/dv-admin/transaction/stats [get]
//	@Security		BearerAuth
func (h Handler) loadUserTransactionStatistics(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	request := &transactions_request.GetStatistics{}
	if err = c.Bind().Query(request); err != nil {
		return err
	}

	res, err := h.services.TransactionService.GetTransactionStats(c.Context(), user, *request)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(res))
}

// downloadUserTranscations is a function to download user transactions
//
//	@Summary		Download user transactions
//	@Description	Download user transactions
//	@Tags			Transaction
//	@Accept			json
//	@Produce		json
//	@Param			string	query		transactions_request.GetByUser	true	"GetTransactionsByUserParams"
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Router			/v1/dv-admin/transaction/export [get]
//	@Security		BearerAuth
func (h Handler) downloadUserTranscations(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	request := &transactions_request.GetByUserExported{}
	if err := c.Bind().Query(request); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	res, err := h.services.TransactionService.DownloadUserTransactions(c.Context(), user.ID, *request)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	c.Response().Header.Set("Content-Type", "application/octet-stream")
	c.Response().Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.%s", "export_user_transactions", request.Format))
	return c.SendStream(res, res.Len())
}

// searchTransaction is a function to Search transaction by hash
//
//	@Summary		Search transactions by hash
//	@Description	Search transactions by hash **Deprecated**: Use `/v1/dv-admin/search/:searchParam instead
//	@Tags			Transaction
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[transaction_response.TransactionInfoResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Router			/v1/dv-admin/transaction/{hash} [get]
//	@Security		BearerAuth
func (h Handler) searchTransaction(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	hash := c.Params("hash")
	txInfo, err := h.services.TransactionService.GetTransactionInfo(c.Context(), user.ID, hash)
	if err != nil {
		if errors.Is(err, transactions.ErrTransactionNotFound) {
			return apierror.New().AddError(errors.New("not found")).SetHttpCode(fiber.StatusNotFound)
		}
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.NewTransactionInfoResponseFromDto(txInfo)))
}

// sendManually is a function to send webhook by transaction
//
//	@Summary		Send webhook
//	@Description	Send webhook manually. Accepts confirmed/unconfirmed tx
//	@Tags			StoreWebhook
//	@Accept			json
//	@Produce		json
//	@Param			ID	path		string	true	"Transaction id"
//	@Success		200	{object}	response.Result[string]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		400	{object}	apierror.Errors
//	@Router			/v1/dv-admin/transaction/{id}/send-webhooks [post]
//	@Security		BearerAuth
func (h *Handler) sendManually(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	txID, err := tools.ValidateUUID(c.Params("ID"))
	if err != nil {
		return err
	}

	if err = h.services.StoreWebhooksService.SendWebhookManual(c.Context(), txID, user.ID); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("ok"))
}

func (h *Handler) initTransactionRoutes(v1 fiber.Router) {
	transaction := v1.Group("/transaction")
	transaction.Get("/", h.loadUserTransaction)
	transaction.Get("/stats", h.loadUserTransactionStatistics)
	transaction.Get("/export", h.downloadUserTranscations)
	transaction.Get("/:hash", h.searchTransaction)
	transaction.Post("/:ID/send-webhooks", h.sendManually)
}
