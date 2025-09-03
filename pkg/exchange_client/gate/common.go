package gateio

import "github.com/shopspring/decimal"

type CurrencyDetails struct {
	Currency         string `json:"currency"`
	Name             string `json:"name"`
	Delisted         bool   `json:"delisted"`
	WithdrawDisabled bool   `json:"withdraw_disabled"`
	WithdrawDelayed  bool   `json:"withdraw_delayed"`
	DepositDisabled  bool   `json:"deposit_disabled"`
	TradeDisabled    bool   `json:"trade_disabled"`
	Chain            string `json:"chain"`
	Chains           []struct {
		Name             string `json:"name"`
		Addr             string `json:"addr,omitempty"`
		WithdrawDisabled bool   `json:"withdraw_disabled"`
		WithdrawDelayed  bool   `json:"withdraw_delayed"`
		DepositDisabled  bool   `json:"deposit_disabled"`
	} `json:"chains"`
}

type CurrencyPair struct {
	ID              string `json:"id"`
	Base            string `json:"base"`
	BaseName        string `json:"base_name"`
	Quote           string `json:"quote"`
	QuoteName       string `json:"quote_name"`
	Fee             string `json:"fee"`
	MinBaseAmount   string `json:"min_base_amount"`
	MinQuoteAmount  string `json:"min_quote_amount"`
	MaxBaseAmount   string `json:"max_base_amount"`
	MaxQuoteAmount  string `json:"max_quote_amount"`
	AmountPrecision int    `json:"amount_precision"`
	Precision       int    `json:"precision"`
	TradeStatus     string `json:"trade_status"`
	SellStart       int    `json:"sell_start"`
	BuyStart        int    `json:"buy_start"`
	DelistingTime   int    `json:"delisting_time"`
	TradeURL        string `json:"trade_url"`
}

type TickerInfo struct {
	CurrencyPair string          `json:"currency_pair"`
	Last         decimal.Decimal `json:"last"`
}

type DepositAddress struct {
	Currency            string `json:"currency"`
	Address             string `json:"address"`
	MultichainAddresses []struct {
		Chain        string `json:"chain"`
		Address      string `json:"address"`
		PaymentID    string `json:"payment_id"`
		PaymentName  string `json:"payment_name"`
		ObtainFailed int    `json:"obtain_failed"`
	} `json:"multichain_addresses"`
}

type AccountDetail struct {
	IPWhitelist   []string `json:"ip_whitelist"`
	CurrencyPairs []string `json:"currency_pairs"`
	UserID        int64    `json:"user_id"`
	Tier          int64    `json:"tier"`
	Key           struct {
		Mode            int32 `json:"mode"`
		CopyTradingRole int32 `json:"copy_trading_role"`
	}
}

type CurrencyChain struct {
	Chain              string `json:"chain"`
	NameEn             string `json:"name_en"`
	ContractAddress    string `json:"contract_address,omitempty"`
	IsDisabled         int    `json:"is_disabled"`
	IsDepositDisabled  int    `json:"is_deposit_disabled"`
	IsWithdrawDisabled int    `json:"is_withdraw_disabled"`
	Decimal            int64  `json:"decimal,string"`
}

type SpotAccountBalance struct {
	Currency  string          `json:"currency"`
	Available decimal.Decimal `json:"available"`
	Locked    decimal.Decimal `json:"locked"`
}

type PairTradeStatus string

func (o PairTradeStatus) IsTradable() bool {
	return o == PairTradeStatusTradable
}

func (o PairTradeStatus) String() string { return string(o) }

const (
	PairTradeStatusUntradable PairTradeStatus = "untradable"
	PairTradeStatusTradable   PairTradeStatus = "tradable"
	PairTradeStatusBuyable    PairTradeStatus = "buyable"
	PairTradeStatusSellable   PairTradeStatus = "sellable"
)

type WithdrawalStatus string

func (o WithdrawalStatus) String() string { return string(o) }

const (
	WithdrawalStatusDone    WithdrawalStatus = "DONE"
	WithdrawalStatusCancel  WithdrawalStatus = "CANCEL"
	WithdrawalStatusRequest WithdrawalStatus = "REQUEST"
	WithdrawalStatusManual  WithdrawalStatus = "MANUAL"
	WithdrawalStatusBcode   WithdrawalStatus = "BCODE"
	WithdrawalStatusExtpend WithdrawalStatus = "EXTPEND"
	WithdrawalStatusFail    WithdrawalStatus = "FAIL"
	WithdrawalStatusInvalid WithdrawalStatus = "INVALID"
	WithdrawalStatusVerify  WithdrawalStatus = "VERIFY"
	WithdrawalStatusProces  WithdrawalStatus = "PROCES" //nolint:misspell
	WithdrawalStatusPend    WithdrawalStatus = "PEND"
	WithdrawalStatusDmove   WithdrawalStatus = "DMOVE"
	WithdrawalStatusReview  WithdrawalStatus = "REVIEW"
)

type WithdrawalHistory struct {
	ID              string           `json:"id"`
	Currency        string           `json:"currency"`
	Address         string           `json:"address"`
	Amount          decimal.Decimal  `json:"amount"`
	Fee             decimal.Decimal  `json:"fee"`
	Txid            string           `json:"txid"`
	Chain           string           `json:"chain"`
	Status          WithdrawalStatus `json:"status"`
	WithdrawOrderID string           `json:"withdraw_order_id,omitempty"`
	BlockNumber     string           `json:"block_number"`
	FailReason      string           `json:"fail_reason,omitempty"`
}

type WithdrawalRule struct {
	Currency               string                     `json:"currency"`
	Deposit                decimal.Decimal            `json:"deposit"`
	WithdrawPercent        string                     `json:"withdraw_percent"`
	WithdrawFix            decimal.Decimal            `json:"withdraw_fix"`
	WithdrawDayLimit       decimal.Decimal            `json:"withdraw_day_limit"`
	WithdrawDayLimitRemain decimal.Decimal            `json:"withdraw_day_limit_remain"`
	WithdrawAmountMini     decimal.Decimal            `json:"withdraw_amount_mini"`
	WithdrawEachtimeLimit  decimal.Decimal            `json:"withdraw_eachtime_limit"`
	WithdrawFixOnChains    map[string]decimal.Decimal `json:"withdraw_fix_on_chains"`
	// WithdrawPercentOnChains map[string]string          `json:"withdraw_percent_on_chains"`
}

type OrderStatus string

func (o OrderStatus) String() string { return string(o) }

const (
	OrderStatusOpen     OrderStatus = "open"
	OrderStatusClosed   OrderStatus = "closed"
	OrderStatusCanceled OrderStatus = "cancelled"
)

type OrderType string

func (o OrderType) String() string { return string(o) }

const (
	OrderTypeMarket OrderType = "market"
	OrderTypeLimit  OrderType = "limit"
)

type OrderSide string

func (o OrderSide) String() string { return string(o) }

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

type OrderFinishedAs string

func (o OrderFinishedAs) String() string { return string(o) }

// TODO: Does not take into account IOC, POC, FOK, STP
func (o OrderFinishedAs) IsFailed() bool {
	return o != OrderFinishedAsOpen && o != OrderFinishedAsFilled
}

const (
	OrderFinishedAsOpen               OrderFinishedAs = "open"
	OrderFinishedAsFilled             OrderFinishedAs = "filled"
	OrderFinishedAsCancelled          OrderFinishedAs = "cancelled"
	OrderFinishedAsLiquidateCancelled OrderFinishedAs = "liquidate_cancelled"
	OrderFinishedAsSmall              OrderFinishedAs = "small"
	OrderFinishedAsDepthNotEnough     OrderFinishedAs = "depth_not_enough"
	OrderFinishedAsTraderNotEnough    OrderFinishedAs = "trader_not_enough"
	OrderFinishedAsIOC                OrderFinishedAs = "ioc"
	OrderFinishedAsPOC                OrderFinishedAs = "poc"
	OrderFinishedAsFOK                OrderFinishedAs = "fok"
	OrderFinishedAsSTP                OrderFinishedAs = "stp"
	OrderFinishedAsUnknown            OrderFinishedAs = "unknown"
)

type SpotOrder struct {
	ID           string          `json:"id"`
	Text         string          `json:"text,omitempty"`
	AmendText    string          `json:"amend_text,omitempty"`
	CreateTimeMs int64           `json:"create_time_ms"`
	UpdateTimeMs int64           `json:"update_time_ms"`
	Status       OrderStatus     `json:"status"`
	CurrencyPair string          `json:"currency_pair"`
	Type         OrderType       `json:"type"`
	Account      string          `json:"account"`
	Side         OrderSide       `json:"side"`
	Amount       string          `json:"amount"`
	Fee          string          `json:"fee"`
	FeeCurrency  string          `json:"fee_currency"`
	FinishAs     OrderFinishedAs `json:"finish_as"`
}

type Withdrawal struct {
	ID              string           `json:"id"`
	Timestamp       string           `json:"timestamp"`
	WithdrawOrderID string           `json:"withdraw_order_id"`
	Currency        string           `json:"currency"`
	Address         string           `json:"address"`
	Txid            string           `json:"txid"`
	Amount          string           `json:"amount"`
	Memo            string           `json:"memo,omitempty"`
	Status          WithdrawalStatus `json:"status"`
	Chain           string           `json:"chain"`
}

type SavedAddress struct {
	Currency string              `json:"currency"`
	Chain    string              `json:"chain"`
	Address  string              `json:"address"`
	Name     string              `json:"name,omitempty"`
	Tag      string              `json:"tag,omitempty"`
	Verified AddressVerification `json:"verified"`
}

type AddressVerification string

func (o AddressVerification) String() string { return string(o) }

func (o AddressVerification) IsVerified() bool {
	return o == AddressVerificationDone
}

const (
	AddressVerificationNone AddressVerification = "0"
	AddressVerificationDone AddressVerification = "1"
)
