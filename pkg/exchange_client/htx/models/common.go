//nolint:tagliatelle
package models

type SymbolStatus string

func (o SymbolStatus) String() string { return string(o) }

const (
	SymbolStatusUnknown       SymbolStatus = "unknown"
	SymbolStatusNotOnline     SymbolStatus = "not-online"
	SymbolStatusPreOnline     SymbolStatus = "pre-online"
	SymbolStatusOnline        SymbolStatus = "online"
	SymbolStatusSuspend       SymbolStatus = "suspend"
	SymbolStatusOffline       SymbolStatus = "offline"
	SymbolStatusTransferBoard SymbolStatus = "transfer-board"
	SymbolStatusFuse          SymbolStatus = "fuse"
)

type Direction int

const (
	DirectionLong  Direction = 1
	DirectionShort Direction = 2
)

type CommonSymbol struct {
	BaseCurrency                    string       `json:"base-currency"`
	QuoteCurrency                   string       `json:"quote-currency"`
	PricePrecision                  int          `json:"price-precision"`
	AmountPrecision                 int          `json:"amount-precision"`
	SymbolPartition                 string       `json:"symbol-partition"`
	Symbol                          string       `json:"symbol"`
	State                           SymbolStatus `json:"state"`
	ValuePrecision                  int          `json:"value-precision"`
	MinOrderAmt                     float64      `json:"min-order-amt"`
	MaxOrderAmt                     float64      `json:"max-order-amt"`
	MinOrderValue                   float64      `json:"min-order-value"`
	LimitOrderMinOrderAmt           float64      `json:"limit-order-min-order-amt"`
	LimitOrderMaxOrderAmt           float64      `json:"limit-order-max-order-amt"`
	LimitOrderMaxBuyAmt             float64      `json:"limit-order-max-buy-amt"`
	LimitOrderMaxSellAmt            float64      `json:"limit-order-max-sell-amt"`
	BuyLimitMustLessThan            float64      `json:"buy-limit-must-less-than"`
	SellLimitMustGreaterThan        float64      `json:"sell-limit-must-greater-than"`
	SellMarketMinOrderAmt           float64      `json:"sell-market-min-order-amt"`
	SellMarketMaxOrderAmt           float64      `json:"sell-market-max-order-amt"`
	BuyMarketMaxOrderValue          float64      `json:"buy-market-max-order-value"`
	MarketSellOrderRateMustLessThan float64      `json:"market-sell-order-rate-must-less-than"`
	MarketBuyOrderRateMustLessThan  float64      `json:"market-buy-order-rate-must-less-than"`
	APITrading                      string       `json:"api-trading"`
	Tags                            string       `json:"tags"`
}

type Symbol struct {
	StateIsolated            string       `json:"si,omitempty"`
	StateCross               string       `json:"scr,omitempty"`
	OutsideSymbol            string       `json:"sc,omitempty"`
	DisplayName              string       `json:"dn,omitempty"`
	BaseCurrency             string       `json:"bc,omitempty"`
	BaseCurrencyDisplayName  string       `json:"bcdn,omitempty"`
	QuoteCurrency            string       `json:"qc,omitempty"`
	QuoteCurrencyDisplayName string       `json:"qcdn,omitempty"`
	SymbolStatus             SymbolStatus `json:"state,omitempty"`
	WhiteEnabled             bool         `json:"whe,omitempty"`
	CountryDisabled          bool         `json:"cd,omitempty"`
	TradeEnabled             bool         `json:"te,omitempty"`
	TradeOpenAt              int64        `json:"toa,omitempty"`
	SymbolPartition          string       `json:"sp,omitempty"`
	WeightSort               int          `json:"w,omitempty"`
	TradeTotalPrecision      float64      `json:"ttp,omitempty"`
	TradeAmountPrecision     float64      `json:"tap,omitempty"`
	TradePricePrecision      float64      `json:"tpp,omitempty"`
	FeePrecision             float64      `json:"fp,omitempty"`
	SuspendDesc              string       `json:"suspend_desc,omitempty"`
	TransferBoardDesc        string       `json:"transfer_board_desc,omitempty"`
	Tags                     string       `json:"tags,omitempty"`
	WithdrawRisk             string       `json:"wr,omitempty"`
}

type MarketSymbol struct {
	Symbol                   string       `json:"symbol,omitempty"`
	State                    SymbolStatus `json:"state,omitempty"`
	BaseCurrency             string       `json:"bc,omitempty"`
	QuoteCurrency            string       `json:"qc,omitempty"`
	PricePrecision           int          `json:"pp,omitempty"`
	AmountPrecision          int          `json:"ap,omitempty"`
	ValuePrecision           int          `json:"vp,omitempty"`
	MinOrderAmount           float64      `json:"minoa,omitempty"`
	MaxOrderAmount           float64      `json:"maxoa,omitempty"`
	MinOrderValue            float64      `json:"minov,omitempty"`
	SellMarketMinOrderAmount float64      `json:"smminoa,omitempty"`
	SellMarketMaxOrderAmount float64      `json:"smmaxoa,omitempty"`
	BuyMarketMaxOrderValue   float64      `json:"bmmaxov,omitempty"`
	APITrading               string       `json:"at,omitempty"`
}

type Currency struct {
	CurrencyCode        string `json:"cc,omitempty"`
	DisplayName         string `json:"dn,omitempty"`
	FullName            string `json:"fn,omitempty"`
	AssetType           int    `json:"at,omitempty"`
	WithdrawPrecision   int    `json:"wp,omitempty"`
	FeeType             string `json:"ft,omitempty"`
	DepositMinAmount    string `json:"dma,omitempty"`
	WithdrawMinAmount   string `json:"wma,omitempty"`
	ShowPrecision       string `json:"sp,omitempty"`
	Weight              int    `json:"w,omitempty"`
	IsQuoteCurrency     bool   `json:"qc,omitempty"`
	State               string `json:"state,omitempty"`
	Visible             bool   `json:"v,omitempty"`
	WhiteEnabled        bool   `json:"whe,omitempty"`
	CountryDisabled     bool   `json:"cd,omitempty"`
	DepositEnabled      bool   `json:"de,omitempty"`
	WithdrawEnabled     bool   `json:"wed,omitempty"`
	CurrencyAddrWithTag bool   `json:"cawt,omitempty"`
	FastConfirms        int    `json:"fc,omitempty"`
	SafeConfirms        int    `json:"sc,omitempty"`
	SuspendWithdrawDesc string `json:"swd,omitempty"`
	WithdrawDesc        string `json:"wd,omitempty"`
	SuspendDepositDesc  string `json:"sdd,omitempty"`
	DepositDesc         string `json:"dd,omitempty"`
	SuspendVisibleDesc  string `json:"svd,omitempty"`
	Tags                string `json:"tags,omitempty"`
}

type CurrencyReference struct {
	Currency string                   `json:"currency"`
	Chains   []CurrencyReferenceChain `json:"chains"`
}

type CurrencyReferenceChain struct {
	Chain                   string `json:"chain"`
	DisplayName             string `json:"displayName"`
	BaseChain               string `json:"baseChain"`
	BaseChainProtocol       string `json:"baseChainProtocol"`
	IsDynamic               bool   `json:"isDynamic"`
	NumOfConfirmations      int    `json:"numOfConfirmations"`
	NumOfFastConfirmations  int    `json:"numOfFastConfirmations"`
	MinDepositAmt           string `json:"minDepositAmt"`
	DepositStatus           string `json:"depositStatus"`
	MinWithdrawAmt          string `json:"minWithdrawAmt"`
	MaxWithdrawAmt          string `json:"maxWithdrawAmt"`
	WithdrawQuotaPerDay     string `json:"withdrawQuotaPerDay"`
	WithdrawQuotaPerYear    string `json:"withdrawQuotaPerYear"`
	WithdrawQuotaTotal      string `json:"withdrawQuotaTotal"`
	WithdrawPrecision       int    `json:"withdrawPrecision"`
	WithdrawFeeType         string `json:"withdrawFeeType"`
	TransactFeeWithdraw     string `json:"transactFeeWithdraw"`
	MinTransactFeeWithdraw  string `json:"minTransactFeeWithdraw"`
	MaxTransactFeeWithdraw  string `json:"maxTransactFeeWithdraw"`
	TransactFeeRateWithdraw string `json:"transactFeeRateWithdraw"`
	WithdrawStatus          string `json:"withdrawStatus"`
}
