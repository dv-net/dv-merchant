package handlers

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/setting_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/transfer_requests"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/service/user"

	// Blank import for swaggen
	_ "github.com/dv-net/dv-merchant/internal/delivery/http/responses/settings_response"
	_ "github.com/dv-net/dv-merchant/internal/delivery/http/responses/transfer_response"
	_ "github.com/dv-net/dv-merchant/internal/storage/repos/repo_transfers"
	_ "github.com/dv-net/dv-merchant/internal/storage/storecmn"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/withdrawal_wallets_request"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

// Get prefetch Transfers
//
//	@Summary		Get prefetch Transfers
//	@Description	This endpoint for get prefetch transfer
//	@Tags			Transfers
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[[]transfer_response.GetPrefetchDataResponse]
//	@Failure		400	{object}	apierror.Errors	"Bad request"
//	@Failure		401	{object}	apierror.Errors	"Unauthorized"
//	@Failure		422	{object}	apierror.Errors	"Unprocessable Entity"
//	@Failure		500	{object}	apierror.Errors	"Internal Server Error"
//	@Router			/v1/dv-admin/transfer/prefetch [Get]
//	@Security		BearerAuth
func (h Handler) getPrefetchData(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	data, err := h.services.WithdrawService.GetPrefetchWithdrawalAddress(c.Context(), user)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByData(converters.FromTransferPrefetchModelToResponses(data...)))
}

// Get transfer
//
//	@Summary		Get transfer by status
//	@Description	This endpoint for filter transfer by status
//	@Tags			Transfers
//	@Accept			json
//	@Produce		json
//	@Param			register	body		withdrawal_wallets_request.TransferRequest	true	"Transfer Request"
//	@Success		200			{object}	response.Result[storecmn.FindResponseWithFullPagination[transfer_response.GetTransferResponse]]
//	@Failure		400			{object}	apierror.Errors	"Bad request"
//	@Failure		401			{object}	apierror.Errors	"Unauthorized"
//	@Failure		422			{object}	apierror.Errors	"Unprocessable Entity"
//	@Failure		500			{object}	apierror.Errors	"Internal Server Error"
//	@Router			/v1/dv-admin/transfer/ [Get]
//	@Security		BearerAuth
func (h Handler) getTransfer(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	req := &withdrawal_wallets_request.TransferRequest{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	data, err := h.services.WithdrawService.GetTransfers(c.Context(), user.ID, req)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(data))
}

// Get transfer history
//
//	@Summary		Get transfer history
//	@Description	This endpoint for get transfer history
//	@Tags			Transfers
//	@Accept			json
//	@Produce		json
//	@Param			register	body		transfer_requests.TransferHistoryRequest	true	"Transfer History Request"
//	@Success		200			{object}	response.Result[storecmn.FindResponseWithFullPagination[transfer_response.GetTransferResponse]]
//	@Failure		400			{object}	apierror.Errors	"Bad request"
//	@Failure		401			{object}	apierror.Errors	"Unauthorized"
//	@Failure		422			{object}	apierror.Errors	"Unprocessable Entity"
//	@Failure		500			{object}	apierror.Errors	"Internal Server Error"
//	@Router			/v1/dv-admin/transfer/history [Get]
//	@Security		BearerAuth
func (h Handler) getTransferHistory(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	req := &transfer_requests.TransferHistoryRequest{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	data, err := h.services.WithdrawService.GetTransfersHistory(c.Context(), user.ID, req)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(data))
}

// Disable/enable transfers
//
//	@Summary		Disable or enable transfers
//	@Description	This endpoint for on/off transfer
//	@Tags			Transfers
//	@Accept			json
//	@Produce		json
//	@Param			register	body		setting_request.TransfersToggleRequest	true	"Transfer Request"
//	@Success		200			{object}	response.Result[string]
//	@Failure		400			{object}	apierror.Errors	"Bad request"
//	@Failure		401			{object}	apierror.Errors	"Unauthorized"
//	@Failure		422			{object}	apierror.Errors	"Unprocessable Entity"
//	@Failure		500			{object}	apierror.Errors	"Internal Server Error"
//	@Router			/v1/dv-admin/transfer/mode [Post]
//	@Security		BearerAuth
func (h Handler) transferToggle(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &setting_request.TransfersToggleRequest{}
	if err = c.Bind().Body(req); err != nil {
		return err
	}

	if err = h.services.UserService.SettingUpdate(c.Context(), usr, user.SettingUpdateDTO{
		Name:  setting.TransfersStatus,
		Value: &req.Mode,
		Model: setting.IModelSetting(usr),
	}); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("success"))
}

// Update Tron transfer type
//
//	@Summary		Update tron transfer type
//	@Description	This endpoint for update transfer type by resources or burntrx
//	@Tags			Transfers
//	@Accept			json
//	@Produce		json
//	@Param			register	body		setting_request.TronTransfersTypeRequest	true	"Tron Transfers Type Request"
//	@Success		200			{object}	response.Result[settings_response.SettingResponse]
//	@Failure		400			{object}	apierror.Errors	"Bad request"
//	@Failure		401			{object}	apierror.Errors	"Unauthorized"
//	@Failure		422			{object}	apierror.Errors	"Unprocessable Entity"
//	@Failure		500			{object}	apierror.Errors	"Internal Server Error"
//	@Router			/v1/dv-admin/transfer/tron/transfer-type [Post]
//	@Security		BearerAuth
func (h Handler) updateTronTransferType(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &setting_request.TronTransfersTypeRequest{}
	if err = c.Bind().Body(req); err != nil {
		return err
	}

	if err = h.services.UserService.SettingUpdate(c.Context(), usr, user.SettingUpdateDTO{
		Name:  setting.TransferType,
		Value: &req.Type,
		Model: setting.IModelSetting(usr),
	}); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("success"))
}

// Delete transfers by ids
//
//	@Summary		Delete transfers by ids
//	@Description	This endpoint for removal action of transfers by ids
//	@Tags			Transfers
//	@Accept			json
//	@Produce		json
//	@Param			register	body		transfer_requests.DeleteTransferRequest	true	"Delete Transfer Request"
//	@Success		200			{object}	response.Result[string]
//	@Failure		400			{object}	apierror.Errors	"Bad request"
//	@Failure		401			{object}	apierror.Errors	"Unauthorized"
//	@Router			/v1/dv-admin/transfer/ [Delete]
//	@Security		BearerAuth
func (h Handler) deleteTransfer(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &transfer_requests.DeleteTransferRequest{}
	if err = c.Bind().Body(req); err != nil {
		return err
	}

	err = h.services.WithdrawService.DeleteTransfers(c.Context(), usr, req)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("success"))
}

func (h Handler) initTransferRoutes(v1 fiber.Router) {
	transfer := v1.Group("/transfer")
	transfer.Get("/prefetch", h.getPrefetchData)
	transfer.Get("/", h.getTransfer)
	transfer.Get("/history", h.getTransferHistory)
	transfer.Delete("/", h.deleteTransfer)
	transfer.Post("/mode", h.transferToggle)
	transfer.Post("/tron/transfer-type", h.updateTronTransferType)
}
