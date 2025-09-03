//nolint:tagliatelle
package responses

import "github.com/dv-net/dv-merchant/pkg/exchange_client/bybit/models"

type GetTransactionLogResponse struct {
	List           []models.TransactionLog `json:"list,omitempty"`
	NextPageCursor string                  `json:"nextPageCursor,omitempty"`
}

type GetTradingBalanceResponse struct {
	List []models.TradingBalance `json:"list,omitempty"`
}

type GetAllCoinsBalanceResponse struct {
	Balance []models.AccountBalance `json:"balance,omitempty"`
}

type GetAPIKeyInfoResponse struct {
	ID          int                 `json:"id,string"`
	Note        string              `json:"note,omitempty"`
	APIKey      string              `json:"apiKey,omitempty"`
	ReadOnly    int                 `json:"readOnly"`
	Permissions map[string][]string `json:"permissions,omitempty"`
}

type GetAccountInfoResponse struct {
	UnifiedMarginStatus models.UnifiedMarginStatus `json:"unifiedMarginStatus"`
}

type UpgradeToUnifiedTradingAccountResponse struct {
	UnifiedUpdateStatus models.UnifiedUpgradeStatus `json:"unifiedUpdateStatus"`
	UnifiedUpdateMsg    []string                    `json:"unifiedUpdateMsg,omitempty"`
}

type GetDepositAddressResponse struct {
	Coin   string                `json:"coin"`
	Chains []models.DepositChain `json:"chains,omitempty"`
}

type CreateInternalTransferResponse struct {
	TransferID string                `json:"transferId"`
	Status     models.TransferStatus `json:"status"`
}

type CreateWithdrawResponse struct {
	ID string `json:"id"`
}

type GetWithdrawResponse struct {
	Rows []models.Withdraw `json:"rows,omitempty"`
}
