package handlers

import (
	"errors"
	"net"
	"regexp"

	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/search_response"
	"github.com/dv-net/dv-merchant/internal/service/transactions"
	"github.com/dv-net/dv-merchant/internal/service/wallet"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

type SearchCriteria string

const (
	SearchTypeTxHash          SearchCriteria = "txHash"
	SearchTypeAddress         SearchCriteria = "address"
	SearchTypeIP              SearchCriteria = "ip"
	SearchTypeEmail           SearchCriteria = "email"
	SearchTypeStoreID         SearchCriteria = "store_id"
	SearchTypeStoreExternalID SearchCriteria = "store_external_id"
)

// searchInfo is a function to search wallet or transaction
//
//	@Summary		Get wallet/tx data
//	@Description	Search tx or wallet information in system
//	@Tags			Search
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[search_response.SearchByCriteriaResponse[any]]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Router			/v1/dv-admin/search/{searchParam} [get]
//	@Security		BearerAuth
func (h *Handler) searchInfo(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	criteria := c.Params("searchParam")
	searchType := determineSearchType(criteria)

	var searchRes any

	switch searchType {
	case SearchTypeTxHash:
		txInfo, err := h.services.TransactionService.GetTransactionInfo(c.Context(), usr.ID, criteria)
		if err != nil {
			if errors.Is(err, transactions.ErrTransactionNotFound) {
				return apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
			}
			return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
		}
		searchRes = converters.NewTransactionInfoResponseFromDto(txInfo)

	default:
		walletsInfo, err := h.services.WalletService.GetWalletsInfo(c.Context(), usr.ID, criteria)
		if err != nil {
			if errors.Is(err, wallet.ErrServiceWalletNotFound) {
				return apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
			}
			return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
		}
		searchRes = walletsInfo
	}

	return c.JSON(response.OkByData(search_response.PrepareByCriteria(string(searchType), searchRes)))
}

func determineSearchType(param string) SearchCriteria {
	if isEmail(param) {
		return SearchTypeEmail
	}
	if isIPAddress(param) {
		return SearchTypeIP
	}
	if isTxHash(param) {
		return SearchTypeTxHash
	}
	if isBlockchainAddress(param) {
		return SearchTypeAddress
	}
	if isUUID(param) {
		return SearchTypeStoreID
	}

	// default search type - contains any string
	return SearchTypeStoreExternalID
}

func isEmail(param string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(param)
}

func isIPAddress(param string) bool {
	return net.ParseIP(param) != nil
}

func isTxHash(param string) bool {
	txHashRegex := regexp.MustCompile(`^(0x)?[0-9a-fA-F]{64}$`)
	return txHashRegex.MatchString(param)
}

func isBlockchainAddress(param string) bool {
	addressRegex := regexp.MustCompile(`^[0-9a-zA-Z]{20,50}$`)
	return addressRegex.MatchString(param)
}

func isUUID(param string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return uuidRegex.MatchString(param)
}

func (h *Handler) initSearchRoutes(v1 fiber.Router) {
	v1.Get("/search/:searchParam", h.searchInfo)
}
