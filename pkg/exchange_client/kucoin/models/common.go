package models

type Permission string

func (o Permission) String() string { return string(o) }

const (
	PermissionGeneral  Permission = "General"
	PermissionSpot     Permission = "Spot"
	PermissionTransfer Permission = "Transfer"
	PermissionMargin   Permission = "Margin"
	PermissionFutures  Permission = "Futures"
)

type AccountType string

const (
	AccountTypeMain  AccountType = "main"
	AccountTypeTrade AccountType = "trade"
)

type TransferAccountType string

const (
	TransferAccountTypeMain  TransferAccountType = "MAIN"
	TransferAccountTypeTrade TransferAccountType = "TRADE"
)

type MarketType string

func (o MarketType) String() string { return string(o) }

const (
	MarketTypeBTC  MarketType = "BTC"
	MarketTypeALTS MarketType = "ALTS"
	MarketTypeUSDS MarketType = "USDS"
	MarketTypeETF  MarketType = "ETF"
)

type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

type SelfTradePreventType string

const (
	SelfTradePreventTypeCN SelfTradePreventType = "CN"
	SelfTradePreventTypeCB SelfTradePreventType = "CB"
	SelfTradePreventTypeCO SelfTradePreventType = "CO"
	SelfTradePreventTypeDC SelfTradePreventType = "DC"
)

type TimeInForce string

const (
	TimeInForceFOK TimeInForce = "FOK"
	TimeInForceGTC TimeInForce = "GTC"
	TimeInForceGTT TimeInForce = "GTT"
	TimeInForceIOC TimeInForce = "IOC"
)

type OrderType string

const (
	OrderTypeLimit  OrderType = "limit"
	OrderTypeMarket OrderType = "market"
)

type WithdrawalStatus string

func (o WithdrawalStatus) String() string { return string(o) }

const (
	WithdrawalStatusFailure          WithdrawalStatus = "FAILURE"
	WithdrawalStatusProcessing       WithdrawalStatus = "PROCESSING"
	WithdrawalStatusReview           WithdrawalStatus = "REVIEW"
	WithdrawalStatusSuccess          WithdrawalStatus = "SUCCESS"
	WithdrawalStatusWalletProcessing WithdrawalStatus = "WALLET_PROCESSING"
)

type WithdrawType string

const (
	WithdrawalTypeAddress WithdrawType = "ADDRESS"
	WithdrawalTypeMail    WithdrawType = "MAIL"
	WithdrawalTypePhone   WithdrawType = "PHONE"
	WithdrawalTypeUID     WithdrawType = "UID"
)

type TransferType string

const (
	TransferTypeInternal TransferType = "INTERNAL"
	TransferTypeP2S      TransferType = "PARENT_TO_SUB"
	TransferTypeS2P      TransferType = "SUB_TO_PARENT"
)

type FeeDeductType string

const (
	FeeDeductTypeInternal FeeDeductType = "INTERNAL"
	FeeDeductTypeExternal FeeDeductType = "EXTERNAL"
)
