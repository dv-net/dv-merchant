package external

import (
	"errors"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/withdrawal_requests"
	"github.com/dv-net/dv-merchant/internal/service/withdraw"
	"github.com/dv-net/dv-merchant/internal/tools"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	// Blank import for swaggen
	_ "github.com/dv-net/dv-merchant/internal/delivery/http/responses/withdrawal_response"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// createWithdrawalFromProcessingWallet is a function to initialize withdrawal from processing wallet
//
//	@Summary		Initialize withdrawal from processing
//	@Description	Initialize withdrawal from processing
//	@Tags			Withdrawal
//	@Accept			json
//	@Produce		json
//	@Param			api_key		query		string												true	"Store API key"
//	@Param			register	body		withdrawal_requests.CreateProcessingWithdrawRequest	true	"Init withdrawal"
//	@Success		200			{object}	response.Result[withdrawal_response.ProcessingWithdrawalResponse]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		423			{object}	apierror.Errors
//	@Failure		404			{object}	apierror.Errors
//	@Failure		409			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/external/withdrawal-from-processing [post]
//	@Security		XApiKey
func (h Handler) createWithdrawalFromProcessingWallet(c fiber.Ctx) error {
	store, err := loadAuthStore(c)
	if err != nil {
		return err
	}
	req := &withdrawal_requests.CreateProcessingWithdrawRequest{}
	if err = c.Bind().Body(req); err != nil {
		return err
	}
	if err = req.Validate(); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusUnprocessableEntity)
	}
	dto := withdraw.CreateWithdrawalFromProcessingDTO{
		CurrencyID: req.CurrencyID,
		Amount:     req.Amount,
		AddressTo:  req.AddressTo,
		UserID:     store.UserID,
		RequestID:  req.RequestID,
	}
	if store.ID != uuid.Nil {
		dto.StoreID = &store.ID
	}

	res, err := h.services.WithdrawService.CreateWithdrawalFromProcessing(c.Context(), dto)
	if err != nil {
		return prepareWithdrawalHTTPError(err)
	}

	return c.JSON(response.OkByData(converters.FromProcessingWithdrawalToResponse(*res)))
}

// getWithdrawalFromProcessingWallet is a function to get processing withdrawal info
//
//	@Summary		Get withdrawal from processing
//	@Description	Get withdrawal from processing
//	@Tags			Withdrawal
//	@Produce		json
//	@Param			api_key	query		string	true	"Store API key"
//	@Param			id		path		string	true	"Withdrawal ID"
//	@Success		200		{object}	response.Result[withdraw.WithdrawalFromProcessingDto]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		406		{object}	apierror.Errors
//	@Failure		409		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/external/withdrawal-from-processing/{id} [get]
//	@Security		XApiKey
func (h Handler) getWithdrawalFromProcessingWallet(c fiber.Ctx) error {
	store, err := loadAuthStore(c)
	if err != nil {
		return err
	}
	whID, err := tools.ValidateUUID(c.Params("id"))
	if err != nil {
		return err
	}

	res, err := h.services.WithdrawService.GetProcessingWithdrawalWithTransfer(c.Context(), whID, store.ID)
	if err != nil {
		preparedErr := apierror.New().AddError(err)
		if errors.Is(err, withdraw.ErrWithdrawalNotFound) {
			return preparedErr.SetHttpCode(fiber.StatusNotFound)
		}

		return preparedErr.SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(res))
}

// deleteWithdrawalFromProcessingWallet is a function to delete processing withdrawal
//
//	@Summary		Delete withdrawal from processing
//	@Description	Delete withdrawal from processing
//	@Tags			Withdrawal
//	@Produce		json
//	@Param			api_key	query		string	false	"Store API key"
//	@Param			id		path		string	true	"Withdrawal ID"
//	@Success		200		{object}	response.Result[string]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/external/withdrawal-from-processing/{id} [delete]
//	@Security		XApiKey
func (h *Handler) deleteWithdrawalFromProcessingWallet(c fiber.Ctx) error {
	store, err := loadAuthStore(c)
	if err != nil {
		return err
	}
	whID, err := tools.ValidateUUID(c.Params("id"))
	if err != nil {
		return err
	}

	err = h.services.WithdrawService.DeleteWithdrawalFromProcessing(c.Context(), whID, store.ID)
	if err != nil {
		preparedErr := apierror.New().AddError(err)
		if errors.Is(err, withdraw.ErrWithdrawalNotFound) {
			return preparedErr.SetHttpCode(fiber.StatusNotFound)
		}
		return preparedErr.SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByMessage("Withdrawal successfully cancelled"))
}

func prepareWithdrawalHTTPError(err error) error {
	errCode := fiber.StatusBadRequest
	switch {
	case errors.Is(err, withdraw.ErrProcessingUninitialized):
		errCode = fiber.StatusLocked
	case errors.Is(err, withdraw.ErrProcessingWalletNotExists):
		errCode = fiber.StatusNotFound
	case errors.Is(err, withdraw.ErrWithdrawFromProcessingDuplicateRequestID):
		errCode = fiber.StatusUnprocessableEntity
	case errors.Is(err, withdraw.ErrWithdrawFromProcessingToHotNotAllowed):
		errCode = fiber.StatusNotAcceptable
	case errors.Is(err, withdraw.ErrStoreIsNotOwnedByUser):
		errCode = fiber.StatusForbidden
	}

	var targetErr *withdraw.InvalidCurrencyForAddressError
	if errors.As(err, &targetErr) {
		errCode = fiber.StatusConflict
	}

	return apierror.New().AddError(err).SetHttpCode(errCode)
}

func (h Handler) initWithdrawalRoutes(router fiber.Router) {
	router.Post("/withdrawal-from-processing", h.createWithdrawalFromProcessingWallet)
	router.Get("/withdrawal-from-processing/:id", h.getWithdrawalFromProcessingWallet)
	router.Delete("/withdrawal-from-processing/:id", h.deleteWithdrawalFromProcessingWallet)
}
