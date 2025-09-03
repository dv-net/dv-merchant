package models

type AccountType string

func (o AccountType) String() string { return string(o) }

const (
	AccountTypeUnified AccountType = "UNIFIED"
	AccountTypeFund    AccountType = "FUND"
)

type TransferStatus string

func (o TransferStatus) String() string { return string(o) }

const (
	TransferStatusUnknown TransferStatus = "STATUS_UNKNOWN"
	TransferStatusSuccess TransferStatus = "SUCCESS"
	TransferStatusFailed  TransferStatus = "FAILED"
	TransferStatusPending TransferStatus = "PENDING"
)

type DepositStatus int

const (
	DepositStatusUnknown             DepositStatus = 0
	DepositStatusToBeConfirme        DepositStatus = 1
	DepositStatusProcessing          DepositStatus = 2
	DepositStatusSuccess             DepositStatus = 3
	DepositStatusFailed              DepositStatus = 4
	DepositStatusPendingToBeCredited DepositStatus = 10011
	DepositStatusCredited            DepositStatus = 10012
)

type WithdrawalStatus string

func (o WithdrawalStatus) String() string { return string(o) }

const (
	WithdrawalStatusSecurityCheck           WithdrawalStatus = "SecurityCheck"
	WithdrawalStatusPending                 WithdrawalStatus = "Pending"
	WithdrawalStatusSuccess                 WithdrawalStatus = "success"
	WithdrawalStatusCancelByUser            WithdrawalStatus = "CancelByUser"
	WithdrawalStatusReject                  WithdrawalStatus = "Reject"
	WithdrawalStatusFail                    WithdrawalStatus = "Fail"
	WithdrawalStatusBlockchainConfirmed     WithdrawalStatus = "BlockchainConfirmed"
	WithdrawalStatusMoreInformationRequired WithdrawalStatus = "MoreInformationRequired"
	WithdrawalStatusUnknown                 WithdrawalStatus = "Unknown"
)

type WithdrawType int

const (
	WithdrawTypeOnChain  WithdrawType = 0
	WithdrawTypeOffChain WithdrawType = 1
)

type InstrumentStatus string

func (o InstrumentStatus) String() string { return string(o) }

const (
	InstrumentStatusStatusPreLaunch  InstrumentStatus = "PreLaunch"
	InstrumentStatusStatusTrading    InstrumentStatus = "Trading"
	InstrumentStatusStatusDelivering InstrumentStatus = "Delivering"
	InstrumentStatusStatusClosed     InstrumentStatus = "Closed"
)

type Side string

func (o Side) String() string { return string(o) }

const (
	SideBuy  Side = "Buy"
	SideSell Side = "Sell"
)

type UnifiedUpgradeStatus string

func (o UnifiedUpgradeStatus) String() string { return string(o) }

const (
	UnifiedUpgradeStatusSuccess UnifiedUpgradeStatus = "SUCCESS"
	UnifiedUpgradeStatusProcess UnifiedUpgradeStatus = "PROCESS"
	UnifiedUpgradeStatusFail    UnifiedUpgradeStatus = "FAIL"
)

type UnifiedMarginStatus int

const (
	UnifiedMarginStatusClassicAccount             UnifiedMarginStatus = 1
	UnifiedMarginStatusUnifiedTradingAccount10    UnifiedMarginStatus = 3
	UnifiedMarginStatusUnifiedTradingAccount10Pro UnifiedMarginStatus = 4
	UnifiedMarginStatusUnifiedTradingAccount20    UnifiedMarginStatus = 5
	UnifiedMarginStatusUnifiedTradingAccount20Pro UnifiedMarginStatus = 6
)

type OrderStatus string

func (o OrderStatus) String() string { return string(o) }

const (
	OrderStatusNew                     = "New"
	OrderStatusPartiallyFilled         = "PartiallyFilled"
	OrderStatusFilled                  = "Filled"
	OrderStatusUntriggered             = "Untriggered"
	OrderStatusRejected                = "Rejected"
	OrderStatusPartiallyFilledCanceled = "PartiallyFilledCanceled"
	OrderStatusCancelled               = "Cancelled"
	OrderStatusTriggered               = "Triggered"
	OrderStatusDeactivated             = "Deactivated"
)

type OrderType string

func (o OrderType) String() string { return string(o) }

const (
	OrderTypeMarket OrderType = "Market"
)

// Select the unit for qty when create Spot market orders for UTA account
//
// baseCoin: for example, buy BTCUSDT, then "qty" unit is BTC
//
// quoteCoin: for example, sell BTCUSDT, then "qty" unit is USDT
type MarketUnit string

func (o MarketUnit) String() string { return string(o) }

const (
	MarketUnitBaseCoin  MarketUnit = "baseCoin"
	MarketUnitQuoteCoin MarketUnit = "quoteCoin"
)
