package handlers

import (
	"errors"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/withdrawal_requests"
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/withdrawal_wallets_request"
	"github.com/dv-net/dv-merchant/internal/service/withdraw"
	"github.com/dv-net/dv-merchant/internal/service/withdrawal_wallet"
	"github.com/dv-net/dv-merchant/internal/tools"

	// Blank import for swagger
	_ "github.com/dv-net/dv-merchant/internal/delivery/http/responses/withdrawal_response"

	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// getWithdrawalWallets get withdrawal wallets
//
//	@Summary		Get withdrawal wallets
//	@Description	This endpoint returns withdrawal wallet's
//	@Tags			WithdrawalWallet
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[[]withdrawal_response.WithdrawalWithAddressResponse]	"Successful operation"
//	@Failure		401	{object}	apierror.Errors															"Unauthorized"
//	@Failure		500	{object}	apierror.Errors															"Internal Server Error"
//	@Router			/v1/dv-admin/withdrawal/rules [get]
//	@Security		BearerAuth
func (h Handler) getWithdrawalRule(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	wallets, err := h.services.WithdrawalWalletService.GetWithdrawalWallets(c.Context(), user.ID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByData(converters.FromWithdrawalWithAddressToResponses(wallets...)))
}

// getWithdrawalCurrencyRule get withdrawal wallet
//
//	@Summary		Get withdrawal wallet
//	@Description	This endpoint returns withdrawal wallet
//	@Tags			WithdrawalWallet
//	@Accept			json
//	@Produce		json
//	@Param			currencyID	path		string																	true	"Currency ID"
//	@Success		200			{object}	response.Result[withdrawal_response.WithdrawalRulesByCurrencyResponse]	"Successful operation"
//	@Failure		401			{object}	apierror.Errors															"Unauthorized"
//	@Failure		500			{object}	apierror.Errors															"Internal Server Error"
//	@Router			/v1/dv-admin/withdrawal/{currencyID}/rules [get]
//	@Security		BearerAuth
func (h Handler) getWithdrawalCurrencyRule(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	currencyID := c.Params("currencyID")

	wallets, err := h.services.WithdrawalWalletService.GetWithdrawalWalletsByCurrencyID(c.Context(), user.ID, currencyID)
	if err != nil {
		var httpErr *apierror.Errors
		if errors.As(err, &httpErr) {
			return apierror.New().AddError(err).SetHttpCode(httpErr.HttpCode)
		}
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.FromWithdrawalWithAddressToByCurrencyResponse(wallets)))
}

// updateWithdrawalRule update withdrawal wallet
//
//	@Summary		Update withdrawal wallet
//	@Description	This endpoint for update withdrawal wallet
//	@Tags			WithdrawalWallet
//	@Accept			json
//	@Produce		json
//	@Param			walletID	path		string										true	"walletID"
//	@Param			register	body		withdrawal_wallets_request.UpdateRequest	true	"Update withdrawal wallet"
//	@Success		200			{object}	response.Result[string]
//	@Failure		401			{object}	apierror.Errors	"Unauthorized"
//	@Failure		422			{object}	apierror.Errors	"Unauthorized"
//	@Failure		500			{object}	apierror.Errors	"Internal Server Error"
//	@Router			/v1/dv-admin/withdrawal/{currencyID}/rules [patch]
//	@Security		BearerAuth
func (h Handler) updateWithdrawalRule(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	currencyID := c.Params("currencyID")

	req := &withdrawal_wallets_request.UpdateRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	if err = req.Validate(); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	curr, err := h.services.CurrencyService.GetCurrencyByID(c.Context(), currencyID)
	if err != nil {
		return apierror.New().AddError(errors.New("invalid currency")).SetHttpCode(fiber.StatusBadRequest)
	}

	var multiRulesDTO *withdrawal_wallet.MultiWithdrawalRuleDTO
	if req.LowBalanceRules != nil {
		if err = req.LowBalanceRules.ValidateByBlockchain(*curr.Blockchain); err != nil {
			return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
		}

		multiRulesDTO = &withdrawal_wallet.MultiWithdrawalRuleDTO{
			Mode:          req.LowBalanceRules.Mode,
			ManualAddress: req.LowBalanceRules.ManualAddress,
		}
	}

	err = h.services.WithdrawalWalletService.UpdateWithdrawalRules(c.Context(), withdrawal_wallet.UpdateRulesDTO{
		Currency:                curr,
		UserID:                  user.ID,
		WithdrawalMinBalance:    req.MinBalance,
		WithdrawalMinBalanceUsd: req.MinBalanceUSD,
		WithdrawalInterval:      req.Interval.String(),
		WithdrawalEnabled:       req.Status.String(),
		MultiWithdrawal:         multiRulesDTO,
	})
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("Withdrawal wallet rule updated successfully"))
}

// Manual withdrawal wallet address
//
//	@Summary		Create manual withdraw by specific wallet
//	@Description	This endpoint for import manual withdraw
//	@Tags			Withdrawal
//	@Accept			json
//	@Produce		json
//	@Param			register	body		withdrawal_requests.ManualWithdrawRequest	true	"Manual withdraw"
//	@Success		200			{object}	response.Result[[]string]
//	@Failure		400			{object}	apierror.Errors	"Bad request"
//	@Failure		403			{object}	apierror.Errors	"Forbidden"
//	@Failure		422			{object}	apierror.Errors	"Unprocessable Entity"
//	@Failure		500			{object}	apierror.Errors	"Internal Server Error"
//	@Router			/v1/dv-admin/withdrawal/withdraw-manual [Post]
//	@Security		BearerAuth
func (h Handler) manualWithdraw(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	dto := &withdrawal_requests.ManualWithdrawRequest{}
	if err := c.Bind().Body(dto); err != nil {
		return err
	}

	if err = h.services.WithdrawService.WithdrawFromAddress(c.Context(), user, dto.WalletAddressID, dto.CurrencyID); err != nil {
		apiErr := apierror.New().AddError(err)
		if errors.Is(err, withdraw.ErrWalletIsNotOwnedByUser) {
			return apiErr.SetHttpCode(fiber.StatusForbidden)
		}

		return apiErr.SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("ok"))
}

// Manual withdrawal wallet address
//
//	@Summary		Create manual withdraw by specific wallet
//	@Description	This endpoint for import manual withdraw
//	@Tags			Withdrawal
//	@Accept			json
//	@Produce		json
//	@Param			register	body		withdrawal_requests.ManualMultipleWithdrawRequest	true	"Manual withdraw"
//	@Success		200			{object}	response.Result[[]string]
//	@Failure		400			{object}	apierror.Errors	"Bad request"
//	@Failure		403			{object}	apierror.Errors	"Forbidden"
//	@Failure		422			{object}	apierror.Errors	"Unprocessable Entity"
//	@Failure		500			{object}	apierror.Errors	"Internal Server Error"
//	@Router			/v1/dv-admin/withdrawal/withdraw-multiple-manual [Post]
//	@Security		BearerAuth
func (h Handler) manualWithdrawMultiple(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	dto := &withdrawal_requests.ManualMultipleWithdrawRequest{}
	if err := c.Bind().Body(dto); err != nil {
		return err
	}

	if err := h.services.WithdrawService.WithdrawFromAddresses(c.Context(), user, withdraw.MultipleWithdrawalDTO{
		WithdrawalWalletID:         dto.WithdrawalWalletID,
		WalletAddressIDs:           dto.WalletAddressIDs,
		ExcludedWalletAddressesIDs: dto.ExcludedWalletAddressesIDs,
		CurrencyID:                 dto.CurrencyID,
	}); err != nil {
		apiErr := apierror.New().AddError(err)
		if errors.Is(err, withdraw.ErrWalletIsNotOwnedByUser) {
			return apiErr.SetHttpCode(fiber.StatusForbidden)
		}

		return apiErr
	}

	return c.JSON(response.OkByMessage("ok"))
}

// Manual withdrawal to processing wallet
//
//	@Summary		Create manual withdraw by specific wallet
//	@Description	This endpoint for import manual withdraw
//	@Tags			Withdrawal
//	@Accept			json
//	@Produce		json
//	@Param			register	body		withdrawal_requests.WithdrawalToProcessingRequest	true	"Withdraw to processing"
//	@Success		200			{object}	response.Result[[]string]
//	@Failure		400			{object}	apierror.Errors	"Bad request"
//	@Failure		403			{object}	apierror.Errors	"Forbidden"
//	@Failure		422			{object}	apierror.Errors	"Unprocessable Entity"
//	@Failure		500			{object}	apierror.Errors	"Internal Server Error"
//	@Router			/v1/dv-admin/withdrawal/withdraw-to-processing [Post]
//	@Security		BearerAuth
func (h Handler) withdrawToProcessing(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &withdrawal_requests.WithdrawalToProcessingRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	err = h.services.WithdrawService.WithdrawToProcessingWallet(c.Context(), user, withdraw.WithdrawalToProcessingDTO{
		WalletAddressIDs: []uuid.UUID{req.WalletAddressID},
		CurrencyID:       req.CurrencyID,
	})
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByMessage("ok"))
}

// Manual withdrawal to processing wallet from multiple wallets
//
//	@Summary		Create manual withdraw by specific wallet
//	@Description	This endpoint for import manual withdraw
//	@Tags			Withdrawal
//	@Accept			json
//	@Produce		json
//	@Param			register	body		withdrawal_requests.MultipleWithdrawalToProcessingRequest	true	"Withdraw to processing"
//	@Success		200			{object}	response.Result[[]string]
//	@Failure		400			{object}	apierror.Errors	"Bad request"
//	@Failure		403			{object}	apierror.Errors	"Forbidden"
//	@Failure		422			{object}	apierror.Errors	"Unprocessable Entity"
//	@Failure		500			{object}	apierror.Errors	"Internal Server Error"
//	@Router			/v1/dv-admin/withdraw-multiple-to-processing [Post]
//	@Security		BearerAuth
func (h Handler) withdrawToProcessingMultiple(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &withdrawal_requests.MultipleWithdrawalToProcessingRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	err = h.services.WithdrawService.WithdrawToProcessingWallet(c.Context(), user, withdraw.WithdrawalToProcessingDTO{
		WalletAddressIDs:           req.WalletAddressIDs,
		ExcludedWalletAddressesIDs: req.Exclude,
		CurrencyID:                 req.CurrencyID,
	})
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByMessage("ok"))
}

// createWithdrawalFromProcessingWallet is a function to initialize withdrawal from processing wallet
//
//	@Summary		Initialize withdrawal from processing
//	@Description	Initialize withdrawal from processing
//	@Tags			Withdrawal
//	@Accept			json
//	@Produce		json
//	@Param			register	body		withdrawal_requests.CreateProcessingWithdrawInternalRequest	true	"Init withdrawal"
//	@Success		200			{object}	response.Result[withdrawal_response.ProcessingWithdrawalResponse]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		423			{object}	apierror.Errors
//	@Failure		404			{object}	apierror.Errors
//	@Failure		409			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/dv-admin/withdrawal/withdrawal-from-processing [post]
//	@Security		BearerAuth
func (h Handler) createWithdrawalFromProcessingWallet(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &withdrawal_requests.CreateProcessingWithdrawInternalRequest{}
	if err = c.Bind().Body(req); err != nil {
		return err
	}

	if err = req.Validate(); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	if err := h.services.ProcessingOwnerService.ValidateTwoFactorToken(c.Context(), user.ProcessingOwnerID.UUID, req.TOTP); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	dto := withdraw.CreateWithdrawalFromProcessingDTO{
		CurrencyID: req.CurrencyID,
		Amount:     req.Amount,
		AddressTo:  req.AddressTo,
		UserID:     user.ID,
		RequestID:  req.RequestID,
	}

	res, err := h.services.WithdrawService.CreateWithdrawalFromProcessing(c.Context(), dto)
	if err != nil {
		return prepareWithdrawalHTTPError(err)
	}

	return c.JSON(response.OkByData(converters.FromProcessingWithdrawalToResponse(*res)))
}

func (h Handler) initWithdrawalRoutes(v1 fiber.Router) {
	withdrawal := v1.Group("/withdrawal")
	withdrawal.Get("/rules", h.getWithdrawalRule)
	withdrawal.Get("/:currencyID/rules", h.getWithdrawalCurrencyRule)
	withdrawal.Patch("/:currencyID/rules", h.updateWithdrawalRule)
	withdrawal.Post("/withdraw-manual", h.manualWithdraw)
	withdrawal.Post("/withdraw-multiple-manual", h.manualWithdrawMultiple)
	withdrawal.Post("/withdraw-to-processing", h.withdrawToProcessing)
	withdrawal.Post("/withdraw-multiple-to-processing", h.withdrawToProcessingMultiple)
	withdrawal.Post("/withdrawal-from-processing", h.createWithdrawalFromProcessingWallet)

	// Address book routes
	withdrawal.Get("/address-book", h.getUserAddressBook)
	withdrawal.Post("/address-book", h.createAddressBookEntry)
	withdrawal.Put("/address-book/:id", h.updateAddressBookEntry)
	withdrawal.Delete("/address-book", h.deleteAddressBookEntry)

	// Add withdrawal rule routes
	withdrawal.Post("/address-book/withdrawal-rule", h.addWithdrawalRule)
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

// getUserAddressBook gets user's address book entries
//
//	@Summary		Get user address book
//	@Description	Get all address book entries for the current user
//	@Tags			Address Book
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[withdrawal_response.AddressBookListResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		401	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/dv-admin/withdrawal/address-book [get]
//	@Security		BearerAuth
func (h Handler) getUserAddressBook(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	addressListResponse, err := h.services.AddressBookService.GetUserAddresses(c.Context(), user.ID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(addressListResponse))
}

// createAddressBookEntry creates a new address book entry
//
//	@Summary		Create address book entry
//	@Description	Create a new address book entry for the current user. Requires 2FA verification.
//	@Tags			Address Book
//	@Accept			json
//	@Produce		json
//	@Param			request	body		withdrawal_requests.CreateAddressBookRequest	true	"Address book entry data with TOTP"
//	@Success		201		{object}	response.Result[withdrawal_response.AddressBookEntryResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		401		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/dv-admin/withdrawal/address-book [post]
//	@Security		BearerAuth
func (h Handler) createAddressBookEntry(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &withdrawal_requests.CreateAddressBookRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	// Validate request structure and business rules
	if err := req.Validate(); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	// Validate 2FA token
	if err := h.services.ProcessingOwnerService.ValidateTwoFactorToken(c.Context(), user.ProcessingOwnerID.UUID, req.TOTP); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	// Convert request to service dto
	dto := converters.FromCreateAddressBookRequest(*req, user.ID)

	// Create the address entry
	entry, err := h.services.AddressBookService.CreateAddress(c.Context(), dto)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	// Convert to response with withdrawal rule status
	entryResponse := converters.ToAddressBookEntryResponse(entry)
	if withdrawalRuleExists, err := h.services.AddressBookService.CheckWithdrawalRuleExists(c.Context(), entry); err == nil {
		entryResponse.WithdrawalRuleExists = withdrawalRuleExists
	}
	return c.Status(fiber.StatusCreated).JSON(response.OkByData(entryResponse))
}

// updateAddressBookEntry updates an existing address book entry
//
//	@Summary		Update address book entry
//	@Description	Update an existing address book entry. Requires 2FA verification.
//	@Tags			Address Book
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string											true	"Address book entry ID"
//	@Param			request	body		withdrawal_requests.UpdateAddressBookRequest	true	"Update data with TOTP"
//	@Success		200		{object}	response.Result[withdrawal_response.AddressBookEntryResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/dv-admin/withdrawal/address-book/{id} [put]
//	@Security		BearerAuth
func (h Handler) updateAddressBookEntry(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	id, err := tools.ValidateUUID(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	req := &withdrawal_requests.UpdateAddressBookRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	// Validate 2FA token
	if err := h.services.ProcessingOwnerService.ValidateTwoFactorToken(c.Context(), user.ProcessingOwnerID.UUID, req.TOTP); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	// Convert request to service DTO
	dto := converters.FromUpdateAddressBookRequest(*req)

	// Update the address entry with permission check
	entry, err := h.services.AddressBookService.UpdateAddress(c.Context(), user.ID, id, dto)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	// Convert to response with withdrawal rule status
	entryResponse := converters.ToAddressBookEntryResponse(entry)
	if withdrawalRuleExists, err := h.services.AddressBookService.CheckWithdrawalRuleExists(c.Context(), entry); err == nil {
		entryResponse.WithdrawalRuleExists = withdrawalRuleExists
	}
	return c.JSON(response.OkByData(entryResponse))
}

// deleteAddressBookEntry deletes address book entries (individual, universal, or EVM)
//
//	@Summary		Delete address book entries
//	@Description	Delete address book entries based on type (individual, universal, or EVM). Requires 2FA verification.
//	@Tags			Address Book
//	@Accept			json
//	@Produce		json
//	@Param			request	body		withdrawal_requests.DeleteAddressBookRequest	true	"Deletion request with type flags"
//	@Success		200		{object}	response.Result[string]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/dv-admin/withdrawal/address-book [delete]
//	@Security		BearerAuth
func (h Handler) deleteAddressBookEntry(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &withdrawal_requests.DeleteAddressBookRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	// Validate request structure and business rules
	if err := req.Validate(); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	// Validate 2FA token
	if err := h.services.ProcessingOwnerService.ValidateTwoFactorToken(c.Context(), user.ProcessingOwnerID.UUID, req.TOTP); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	// Convert request to service DTO
	dto := converters.FromDeleteAddressBookRequest(*req, user.ID)

	// Delete the address entry
	err = h.services.AddressBookService.DeleteAddress(c.Context(), dto)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("Address book entry deleted successfully"))
}

// addWithdrawalRule adds withdrawal rules (individual, universal, or EVM)
//
//	@Summary		Add withdrawal rules
//	@Description	Add withdrawal rules based on type (individual, universal, or EVM). Requires 2FA verification.
//	@Tags			Address Book
//	@Accept			json
//	@Produce		json
//	@Param			request	body		withdrawal_requests.AddWithdrawalRuleRequest	true	"Withdrawal rule request with type flags"
//	@Success		200		{object}	response.Result[string]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Router			/v1/dv-admin/withdrawal/address-book/withdrawal-rule [post]
//	@Security		BearerAuth
func (h Handler) addWithdrawalRule(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &withdrawal_requests.AddWithdrawalRuleRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	// Validate request structure and business rules
	if err := req.Validate(); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	// Validate 2FA token
	if err := h.services.ProcessingOwnerService.ValidateTwoFactorToken(c.Context(), user.ProcessingOwnerID.UUID, req.TOTP); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	// Convert request to service DTO
	dto := converters.FromAddWithdrawalRuleRequest(*req, user.ID)

	// Add withdrawal rule
	err = h.services.AddressBookService.AddWithdrawalRule(c.Context(), dto)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("Withdrawal rules added successfully"))
}

// deleteWithdrawalRule adds withdrawal rules (individual, universal, or EVM)
//
//	@Summary		Delete withdrawal rules
//	@Description	Delete withdrawal rules based on type (individual, universal, or EVM). Requires 2FA verification.
//	@Tags			Address Book
//	@Accept			json
//	@Produce		json
//	@Param			request	body		withdrawal_requests.DeleteWithdrawalRuleRequest	true	"Withdrawal rule request with type flags"
//	@Success		200		{object}	response.Result[string]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Router			/v1/dv-admin/withdrawal/address-book/withdrawal-rule [delete]
//	@Security		BearerAuth
// func (h Handler) deleteWithdrawalRule(c fiber.Ctx) error {
// 	user, err := loadAuthUser(c)
// 	if err != nil {
// 		return err
// 	}

// 	req := &withdrawal_requests.DeleteWithdrawalRuleRequest{}
// 	if err := c.Bind().Body(req); err != nil {
// 		return err
// 	}

// 	// Validate request structure and business rules
// 	if err := req.Validate(); err != nil {
// 		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
// 	}

// 	// Validate 2FA token
// 	if err := h.services.ProcessingOwnerService.ValidateTwoFactorToken(c.Context(), user.ProcessingOwnerID.UUID, req.TOTP); err != nil {
// 		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
// 	}

// 	// Convert request to service DTO
// 	dto := converters.FromAddWithdrawalRuleRequest(*req, user.ID)

// 	// Add withdrawal rule
// 	err = h.services.AddressBookService.AddWithdrawalRule(c.Context(), dto)
// 	if err != nil {
// 		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
// 	}

// 	return c.JSON(response.OkByMessage("Withdrawal rules added successfully"))
// }
