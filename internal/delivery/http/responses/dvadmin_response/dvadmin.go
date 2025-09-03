package dvadmin_response

import (
	"github.com/shopspring/decimal"
)

type AuthLinkResponse struct {
	Link string `json:"link"`
} // @name AuthLinkResponse

type OwnerData struct {
	IsAuthorized bool            `json:"is_authorized"`
	Balance      decimal.Decimal `json:"balance"`
	OwnerID      string          `json:"owner_id"`
	Telegram     *string         `json:"telegram"`
} // @name OwnerData
