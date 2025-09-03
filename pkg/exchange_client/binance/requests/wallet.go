package requests

import (
	binancemodels "github.com/dv-net/dv-merchant/pkg/exchange_client/binance/models"
	"github.com/shopspring/decimal"
)

type GetDefaultDepositAddressRequest struct {
	Coin      string          `json:"coin" url:"coin" validate:"required"`
	Network   string          `json:"network,omitempty" url:"network,omitempty"`
	Amount    decimal.Decimal `json:"amount,omitempty" url:"amount,omitempty"`
	RecvWin   int64           `json:"-" url:"-"`
	Timestamp int64           `json:"-" url:"-"`
}

type WithdrawalRequest struct {
	Coin               string                   `json:"coin" url:"coin" validate:"required"`
	WithdrawOrderId    string                   `json:"withdrawOrderId,omitempty" url:"withdrawOrderId,omitempty"`
	Network            string                   `json:"network,omitempty" url:"network,omitempty"`
	Address            string                   `json:"address" url:"address" validate:"required"`
	AddressTag         string                   `json:"addressTag,omitempty" url:"addressTag,omitempty"`
	Amount             string                   `json:"amount" url:"amount" validate:"required"`
	TransactionFeeFlag bool                     `json:"transactionFeeFlag,omitempty" url:"transactionFeeFlag,omitempty"`
	Name               string                   `json:"name,omitempty" url:"name,omitempty"`
	WalletType         binancemodels.WalletType `json:"walletType,omitempty" url:"walletType,omitempty"`
	RecvWin            int64                    `json:"-" url:"-"`
	Timestamp          int64                    `json:"-" url:"-"`
}

type GetFundingAssetsRequest struct {
	Asset            string `json:"asset,omitempty" url:"asset,omitempty"`
	NeedBTCValuation bool   `json:"needBtcValuation,omitempty" url:"needBtcValuation,omitempty"`
	RecvWin          int64  `json:"-" url:"-"`
	Timestamp        int64  `json:"-" url:"-"`
}

type GetSpotAssetsRequest struct {
	Asset            string `json:"asset,omitempty" url:"asset,omitempty"`
	NeedBTCValuation bool   `json:"needBtcValuation,omitempty" url:"needBtcValuation,omitempty"`
	RecvWin          int64  `json:"-" url:"-"`
	Timestamp        int64  `json:"-" url:"-"`
}

type GetDepositAddressesRequest struct {
	Coin      string `json:"coin" url:"coin" validate:"required"`
	Network   string `json:"network,omitempty" url:"network,omitempty"`
	Timestamp int64  `json:"-" url:"-"`
}

type GetUserBalancesRequest struct {
	RecvWin   int64 `json:"-" url:"-"`
	Timestamp int64 `json:"-" url:"-"`
}

type GetWithdrawalHistoryRequest struct {
	Coin            string                         `json:"coin,omitempty" url:"coin,omitempty"`
	WithdrawOrderId string                         `json:"withdrawOrderId" url:"withdrawOrderId" validate:"required"`
	Status          binancemodels.WithdrawalStatus `json:"status,omitempty" url:"status,omitempty"`
	Offset          int                            `json:"offset,omitempty" url:"offset,omitempty"`
	Limit           int                            `json:"limit,omitempty" url:"limit,omitempty" validate:"max=1000"`
	IDList          string                         `json:"idList,omitempty" url:"idList,omitempty"`
	StartTime       int64                          `json:"startTime,omitempty" url:"startTime,omitempty"`
	EndTime         int64                          `json:"endTime,omitempty" url:"endTime,omitempty"`
	RecvWin         int64                          `json:"-" url:"-"`
	Timestamp       int64                          `json:"-" url:"-"`
}

type UniversalTransferRequest struct {
	Type       binancemodels.TransferType `json:"type" url:"type" validate:"required"`
	Asset      string                     `json:"asset" url:"asset" validate:"required"`
	Amount     string                     `json:"amount" url:"amount" validate:"required"`
	FromSymbol string                     `json:"fromSymbol,omitempty" url:"fromSymbol,omitempty"`
	ToSymbol   string                     `json:"toSymbol,omitempty" url:"toSymbol,omitempty"`
	RecvWin    int64                      `json:"-" url:"-"`
	Timestamp  int64                      `json:"-" url:"-"`
}

// TODO: only required on margin transfers, remove when verified
