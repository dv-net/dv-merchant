package requests

import "github.com/shopspring/decimal"

const (
	OFlagPostOnly      = "post"
	OFlagFeeInBase     = "fcib"
	OFlagFeeInQuote    = "fciq"
	OFlagNoMarketProtn = "nompp"
)

type AddOrderRequest struct {
	Pair      string
	Type      string          // buy | sell
	OrderType string          // market | limit | ...
	Volume    decimal.Decimal // always in base currency
	Price     decimal.Decimal // limit orders only
	// UUID (with or without dashes) or free text up to 18 ASCII chars.
	ClOrdID     string
	UserRef     *int32
	OFlags      []string
	TimeInForce string // GTC | IOC | GTD
	Validate    bool   // validate only, the order is not placed
}

type QueryOrdersRequest struct {
	TxIDs            []string // up to 50
	Trades           bool
	UserRef          *int32
	ConsolidateTaker *bool
}
