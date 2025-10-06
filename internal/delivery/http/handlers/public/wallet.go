package public

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	// Blank import for swaggen
	_ "github.com/dv-net/dv-merchant/internal/delivery/http/responses/transaction_response"
	_ "github.com/dv-net/dv-merchant/internal/service/wallet"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/public_request"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// getWalletData is a function to get full wallet data
//
//	@Summary		Get wallet full data
//	@Description	Get wallet full data
//	@Tags			Wallet,Public
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string	true	"Wallet ID"
//	@Param			locale	query		string	false	"Locale"
//	@Success		200		{object}	response.Result[public_request.GetWalletDto]
//	@Failure		401		{object}	apierror.Errors
//	@Failure		410		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/public/wallet/{id} [get]
func (h *Handler) getWalletData(c fiber.Ctx) error {
	walletID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(fmt.Errorf("bad wallet id")).SetHttpCode(fiber.StatusBadRequest)
	}

	// bind query parameters
	req := &public_request.GetWalletRequest{}
	if err := c.Bind().Query(req); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	// get store by wallet ID
	_, err = h.services.StoreService.GetStoreByWalletID(c.Context(), walletID)
	if err != nil {
		if strings.Contains(err.Error(), "store not found") {
			return apierror.New().AddError(errors.New("store not found")).SetHttpCode(fiber.StatusNotFound)
		}
		if strings.Contains(err.Error(), "store is disabled") {
			return apierror.New().AddError(errors.New("store is disabled")).SetHttpCode(fiber.StatusGone)
		}
		return apierror.New().AddError(fmt.Errorf("something went wrong")).SetHttpCode(fiber.StatusBadRequest)
	}

	// update wallet locale if provided
	if req.Locale != nil && *req.Locale != "" {
		err = h.services.WalletService.UpdateLocale(c.Context(), walletID, *req.Locale)
		if err != nil {
			return apierror.New().AddError(fmt.Errorf("something went wrong")).SetHttpCode(fiber.StatusBadRequest)
		}
	}

	// get wallet data
	data, err := h.services.WalletService.GetFullDataByID(c.Context(), walletID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	// convert addresses to dto
	addresses := make([]public_request.WalletAddressDto, len(data.Addresses))
	for idx, address := range data.Addresses {
		wa := public_request.WalletAddressDto{
			Currency: public_request.CurrencyDto{
				ID:         address.CurrencyID,
				Blockchain: address.Blockchain,
			},
			Address: address.Address,
		}

		currencyIdx := slices.IndexFunc(data.AvailableCurrencies, func(item *models.Currency) bool {
			return item.ID == address.CurrencyID
		})

		if currencyIdx >= 0 {
			currency := data.AvailableCurrencies[currencyIdx]
			wa.Currency.Name = currency.Name
			wa.Currency.Code = currency.Code
			wa.Currency.CurrencyLabel = pgtypeutils.DecodeText(currency.CurrencyLabel)
			wa.Currency.TokenLabel = pgtypeutils.DecodeText(currency.TokenLabel)
			wa.Currency.ContractAddress = pgtypeutils.DecodeText(currency.ContractAddress)
			wa.Currency.IsNative = currency.IsNative
		}

		addresses[idx] = wa
	}

	// sort addresses by currency order_idx
	slices.SortFunc(addresses, func(a, b public_request.WalletAddressDto) int {
		aCurrencyIdx := slices.IndexFunc(data.AvailableCurrencies, func(item *models.Currency) bool {
			return item.ID == a.Currency.ID
		})
		bCurrencyIdx := slices.IndexFunc(data.AvailableCurrencies, func(item *models.Currency) bool {
			return item.ID == b.Currency.ID
		})

		if aCurrencyIdx == -1 || bCurrencyIdx == -1 {
			return 0
		}

		aOrderIdx := data.AvailableCurrencies[aCurrencyIdx].SortOrder
		bOrderIdx := data.AvailableCurrencies[bCurrencyIdx].SortOrder

		if aOrderIdx < bOrderIdx {
			return -1
		}
		if aOrderIdx > bOrderIdx {
			return 1
		}
		return 0
	})

	result := public_request.GetWalletDto{
		Store: public_request.StoreDto{
			ID:             data.Store.ID,
			Name:           data.Store.Name,
			CurrencyID:     data.Store.CurrencyID,
			SiteURL:        data.Store.Site,
			ReturnURL:      data.Store.ReturnUrl,
			SuccessURL:     data.Store.SuccessUrl,
			MinimalPayment: data.Store.MinimalPayment.String(),
			Status:         data.Store.Status,
		},
		Addresses: addresses,
		Rates:     data.Rates,
		WalletID:  walletID,
	}

	return c.JSON(response.OkByData(result))
}

// findTransactionsByWallet is a function to fetch recent wallet transactions
//
//	@Summary		Get recent wallet transactions
//	@Description	Get recent wallet transactions
//	@Tags			Wallet,Public
//	@Produce		json
//	@Param			id	path		string	true	"Wallet ID"
//	@Success		200	{object}	response.Result[transaction_response.ShortTransactionInfoListResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/public/wallet/{id}/tx-find [get]
func (h *Handler) findTransactionsByWallet(c fiber.Ctx) error {
	walletID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(fmt.Errorf("bad wallet id")).SetHttpCode(fiber.StatusBadRequest)
	}

	res, err := h.services.WalletTransactionService.GetLastWalletDepositTransactions(c.Context(), walletID)
	if err != nil {
		return apierror.New().AddError(fmt.Errorf("something went wrong")).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.NewShortTransactionsListInfoFromDto(res)))
}

// notifyWalletEmail is a function to notify wallet email
//
//	@Summary		Notify wallet email
//	@Description	Notify wallet email
//	@Tags			Wallet,Public
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string	true	"Wallet ID"
//	@Param			currency_id	query		string	false	"Currency ID"
//	@Success		200			{object}	response.Result[string]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		404			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/public/wallet/{id}/confirm [get]
func (h *Handler) notifyWalletEmail(c fiber.Ctx) error {
	walletID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(fmt.Errorf("bad wallet id")).SetHttpCode(fiber.StatusBadRequest)
	}

	// bind query parameters
	req := &public_request.GetWalletRequest{}
	if err := c.Bind().Query(req); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	_, err = h.services.StoreService.GetStoreByWalletID(c.Context(), walletID)
	if err != nil {
		if strings.Contains(err.Error(), "store not found") {
			return apierror.New().AddError(errors.New("store not found")).SetHttpCode(fiber.StatusNotFound)
		}
		if strings.Contains(err.Error(), "store is disabled") {
			return apierror.New().AddError(errors.New("store is disabled")).SetHttpCode(fiber.StatusGone)
		}
		return apierror.New().AddError(fmt.Errorf("something went wrong")).SetHttpCode(fiber.StatusBadRequest)
	}

	err = h.services.WalletService.SendUserWalletNotification(c.Context(), walletID, req.CurrencyID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByMessage("wallet confirmed successfully"))
}

func (h *Handler) initWalletRoutes(v1 fiber.Router) {
	wallet := v1.Group("/wallet")
	wallet.Get("/:id", h.getWalletData)
	wallet.Get("/:id/tx-find", h.findTransactionsByWallet)
	wallet.Get("/:id/confirm", h.notifyWalletEmail)
}
