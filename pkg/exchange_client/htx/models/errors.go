package models

import (
	"errors"
	"strconv"
)

type HtxErrorCode int

// Wallet errors codes
const (
	HtxInternalError HtxErrorCode = 500
	HtxUnauthorized  HtxErrorCode = 1002
	HtxInvalidSig    HtxErrorCode = 1003
	HtxInvalidField  HtxErrorCode = 2002
	HtxMissingField  HtxErrorCode = 2003
)

func (o HtxErrorCode) String() string {
	return strconv.Itoa(int(o))
}

// type HtxError error

var (
	ErrHtxForbiddenTradeForOpenProtect                          = errors.New("forbidden-trade-for-open-protect")
	ErrHtxBaseArgumentUnsupported                               = errors.New("base-argument-unsupported")
	ErrHtxBaseSystemError                                       = errors.New("base-system-error")
	ErrHtxLoginRequired                                         = errors.New("login-required")
	ErrHtxParameterRequired                                     = errors.New("parameter-required")
	ErrHtxBaseRecordInvalid                                     = errors.New("base-record-invalid")
	ErrHtxOrderAmountOverLimit                                  = errors.New("order-amount-over-limit")
	ErrHtxBaseSymbolTradeDisabled                               = errors.New("base-symbol-trade-disabled")
	ErrHtxBaseOperationForbidden                                = errors.New("base-operation-forbidden")
	ErrHtxAccountGetAccountsInexistent                          = errors.New("account-get-accounts-inexistent-error")
	ErrHtxAccountAccountIDInexistent                            = errors.New("account-account-id-inexistent")
	ErrHtxSubUserAuthRequired                                   = errors.New("sub-user-auth-required")
	ErrHtxOrderDisabled                                         = errors.New("order-disabled")
	ErrHtxCancelDisabled                                        = errors.New("cancel-disabled")
	ErrHtxOrderInvalidPrice                                     = errors.New("order-invalid-price")
	ErrHtxOrderAccountBalanceError                              = errors.New("order-accountbalance-error")
	ErrHtxOrderLimitOrderPriceMinError                          = errors.New("order-limitorder-price-min-error")
	ErrHtxOrderLimitOrderPriceMaxError                          = errors.New("order-limitorder-price-max-error")
	ErrHtxOrderLimitOrderAmountMinError                         = errors.New("order-limitorder-amount-min-error")
	ErrHtxOrderLimitOrderAmountMaxError                         = errors.New("order-limitorder-amount-max-error")
	ErrHtxOrderETPNAVPriceMinError                              = errors.New("order-etp-nav-price-min-error")
	ErrHtxOrderETPNAVPriceMaxError                              = errors.New("order-etp-nav-price-max-error")
	ErrHtxOrderOrderPricePrecisionError                         = errors.New("order-orderprice-precision-error")
	ErrHtxOrderOrderAmountPrecisionError                        = errors.New("order-orderamount-precision-error")
	ErrHtxOrderValueMinError                                    = errors.New("order-value-min-error")
	ErrHtxOrderMarketOrderAmountMinError                        = errors.New("order-marketorder-amount-min-error")
	ErrHtxOrderMarketOrderAmountBuyMaxError                     = errors.New("order-marketorder-amount-buy-max-error")
	ErrHtxOrderMarketOrderAmountSellMaxError                    = errors.New("order-marketorder-amount-sell-max-error")
	ErrHtxOrderHoldingLimitFailed                               = errors.New("order-holding-limit-failed")
	ErrHtxOrderTypeInvalid                                      = errors.New("order-type-invalid")
	ErrHtxOrderOrderStateError                                  = errors.New("order-orderstate-error")
	ErrHtxOrderDateLimitError                                   = errors.New("order-date-limit-error")
	ErrHtxOrderSourceInvalid                                    = errors.New("order-source-invalid")
	ErrHtxOrderUpdateError                                      = errors.New("order-update-error")
	ErrHtxOrderFLCancellationIsDisallowed                       = errors.New("order-fl-cancellation-is-disallowed")
	ErrHtxOperationForbiddenForFLAccountState                   = errors.New("operation-forbidden-for-fl-account-state")
	ErrHtxOperationForbiddenForLockAccountState                 = errors.New("operation-forbidden-for-lock-account-state")
	ErrHtxFLOrderAlreadyExisted                                 = errors.New("fl-order-already-existed")
	ErrHtxOrderUserCancelForbidden                              = errors.New("order-user-cancel-forbidden")
	ErrHtxAccountStateInvalid                                   = errors.New("account-state-invalid")
	ErrHtxOrderPriceGreaterThanLimit                            = errors.New("order-price-greater-than-limit")
	ErrHtxOrderPriceLessThanLimit                               = errors.New("order-price-less-than-limit")
	ErrHtxOrderStopOrderHitTrigger                              = errors.New("order-stop-order-hit-trigger")
	ErrHtxMarketOrdersNotSupportDuringLimitPriceTrading         = errors.New("market-orders-not-support-during-limit-price-trading")
	ErrHtxPriceExceedsTheProtectivePriceDuringLimitPriceTrading = errors.New("price-exceeds-the-protective-price-during-limit-price-trading")
	ErrHtxInvalidClientOrderID                                  = errors.New("invalid-client-order-id")
	ErrHtxInvalidInterval                                       = errors.New("invalid-interval")
	ErrHtxInvalidStartDate                                      = errors.New("invalid-start-date")
	ErrHtxInvalidEndDate                                        = errors.New("invalid-end-date")
	ErrHtxInvalidStartTime                                      = errors.New("invalid-start-time")
	ErrHtxInvalidEndTime                                        = errors.New("invalid-end-time")
	ErrHtxValidationConstraintsRequired                         = errors.New("validation-constraints-required")
	ErrHtxSymbolNotSupport                                      = errors.New("symbol-not-support")
	ErrHtxNotFound                                              = errors.New("not-found")
	ErrHtxBaseNotFound                                          = errors.New("base-not-found")
	ErrHtxAccountGetBalanceAccountInexistentError               = errors.New("account-get-balance-account-inexistent-error")
	ErrHtxRateLimitExceeded                                     = errors.New("rate-too-many-requests")
)
