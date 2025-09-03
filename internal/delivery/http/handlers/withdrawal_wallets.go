package handlers

import (
	"errors"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/dv-net/dv-processing/pkg/avalidator"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/withdrawal_wallets_request"
	"github.com/dv-net/dv-merchant/internal/tools"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

// Actualize wallet addresses list
//
//	@Summary		Update withdrawal wallet addresses
//	@Description	This endpoint for update full addresses list
//	@Tags			WithdrawalWallet
//	@Accept			json
//	@Produce		json
//	@Param			register	body		withdrawal_wallets_request.UpdateAddressesListRequest	true	"Update address"
//	@Param			walletID	path		string													true	"walletID"
//	@Success		200			{object}	response.Result[string]
//	@Failure		400			{object}	apierror.Errors	"Bad request"
//	@Failure		401			{object}	apierror.Errors	"Unauthorized"
//	@Failure		422			{object}	apierror.Errors	"Unprocessable Entity"
//	@Failure		500			{object}	apierror.Errors	"Internal Server Error"
//	@Router			/v1/dv-admin/withdrawal-wallet/{walletID}/addresses [patch]
//	@Security		BearerAuth
func (h *Handler) updateWalletAddresses(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	req := &withdrawal_wallets_request.UpdateAddressesListRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	walletID, err := tools.ValidateUUID(c.Params("walletID"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	wallet, err := h.services.WithdrawalWalletService.GetWalletByID(c.Context(), walletID)
	if err != nil || user.ID != wallet.UserID {
		return apierror.New().
			AddError(errors.New("unauthorized or wallet not found")).
			SetHttpCode(fiber.StatusUnauthorized)
	}

	if err = h.ensureAddrListIsValid(wallet.Blockchain, req.Addresses); err != nil {
		return err
	}

	err = h.services.WithdrawalWalletService.BatchCreateOrUpdateWallet(
		c.Context(),
		user,
		converters.NewWithdrawalWalletAddressesDtoFromRequest(*req, walletID),
	)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("Wallet addresses updated successfully"))
}

func (h Handler) ensureAddrListIsValid(blockchain models.Blockchain, addresses []withdrawal_wallets_request.WalletAddress) error {
	if len(addresses) == 0 {
		return nil
	}

	var apiErr *apierror.Errors
	for _, addr := range addresses {
		if !avalidator.ValidateAddressByBlockchain(addr.Address, blockchain.String()) {
			if apiErr == nil {
				apiErr = apierror.New()
			}
			apiErr = apiErr.AddError(
				fmt.Errorf("invalid address: %s for blockchain: %s", addr.Address, blockchain.String()),
				apierror.WithField(addr.Address),
			)
		}
	}

	if apiErr != nil {
		return apiErr.SetHttpCode(fiber.StatusUnprocessableEntity)
	}

	return nil
}

func (h Handler) initWithdrawalWalletsRoutes(v1 fiber.Router) {
	withdrawal := v1.Group("/withdrawal-wallet")
	withdrawal.Patch("/:walletID/addresses", h.updateWalletAddresses)
}
