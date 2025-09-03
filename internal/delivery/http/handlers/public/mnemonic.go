package public

import (
	"errors"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/mnemonic_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/mnemonic_response"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/response"
	"github.com/dv-net/go-bip39"
	"github.com/gofiber/fiber/v3"
)

// acceptInvite
//
//	@Summary		Generate random mnemonic
//	@Description	Generate random mnemonic
//	@Tags			Mnemonic,Public
//	@Accept			json
//	@Produce		json
//	@Param			length	query		string	true	"Length of mnemonic, 12 or 24 words"
//	@Success		200		{object}	response.Result[mnemonic_response.MnemonicResponse]
//	@Router			/v1/public/mnemonic/generate [get]
func (h *Handler) generateMnemonic(c fiber.Ctx) error {
	req := &mnemonic_request.GenerateMnemonicRequest{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}
	bitSize := 256
	switch req.Length {
	case 12:
		bitSize = 128
	case 24:
		bitSize = 256
	}

	entropy, err := bip39.NewEntropy(bitSize) //nolint:mnd
	if err != nil {
		return apierror.New().SetHttpCode(fiber.StatusBadRequest).AddError(errors.New("failed to generate entropy"))
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return apierror.New().SetHttpCode(fiber.StatusBadRequest).AddError(errors.New("failed to generate mnemonic"))
	}

	return c.JSON(response.OkByData(mnemonic_response.MnemonicResponse{
		Mnemonic: mnemonic,
	}))
}

func (h *Handler) initMnemonic(v1 fiber.Router) {
	m := v1.Group("/mnemonic")
	m.Get("/generate", h.generateMnemonic)
}
