package handlers

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/statistics_request"
	"github.com/dv-net/dv-merchant/internal/service/transactions"
	"github.com/dv-net/dv-merchant/internal/service/wallet"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

// depositSummary is a function to summarize deposit statistics in different resolutions
//
//	@Summary		Get deposits summary
//	@Description	Load user deposits summary
//	@Tags			Statistics
//	@Accept			json
//	@Produce		json
//	@Param			string	query		statistics_request.GetDepositsSummaryRequest	true	"GetDepositsSummaryRequest"
//	@Success		200		{object}	response.Result[[]transactions.StatisticsDTO]
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Router			/v1/dv-admin/deposit/summary/ [get]
//	@Security		BearerAuth
func (h *Handler) depositSummary(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	req := &statistics_request.GetDepositsSummaryRequest{}
	if err = c.Bind().Query(req); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	res, err := h.services.TransactionService.DepositStatistics(c.Context(), transactions.StatisticsParams{
		User:        *user,
		Resolution:  req.Resolution,
		DateFrom:    req.DateFrom,
		DateTo:      req.DateTo,
		CurrencyIDs: req.CurrencyIDs,
		StoreUUIDS:  req.StoreUUIDs,
	})
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(res))
}

// tronResources retrieves a summary of Tron resource expenses and deposit counts.
//
//	@Summary		Get Tron resource and deposit statistics
//	@Description	Retrieves aggregated statistics for Tron wallet balances, transfer expenses, and deposit counts for a user over a specified date range.
//	@Tags			Statistics
//	@Accept			json
//	@Produce		json
//	@Param			date_from	query		string												false	"Start date (YYYY-MM-DD)"	example("2025-06-17")
//	@Param			date_to		query		string												false	"End date (YYYY-MM-DD)"		example("2025-06-24")
//	@Param			resolution	query		string												false	"day/hour"					example("hour")
//	@Success		200			{object}	response.Result[map[string]wallet.CombinedStats]	"Successful response with statistics"
//	@Failure		400			{object}	apierror.Errors										"Invalid request parameters or processing error"
//	@Failure		401			{object}	apierror.Errors										"Unauthorized: Invalid or missing Bearer token"
//	@Router			/v1/dv-admin/tron/resource-expenses/ [get]
//	@Security		BearerAuth
func (h *Handler) tronResources(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &statistics_request.FetchTronStatsRequest{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	res, err := h.services.WalletService.FetchTronResourceStatistics(c.Context(), user, wallet.FetchTronStatisticsParams{
		DateFrom:   req.DateFrom,
		DateTo:     req.DateTo,
		Resolution: req.Resolution,
	})
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(res))
}

func (h *Handler) initStatisticsRoutes(v3 fiber.Router) {
	v3.Get("/tron/resource-expenses", h.tronResources)
	v3.Get("/deposit/summary", h.depositSummary)
}
