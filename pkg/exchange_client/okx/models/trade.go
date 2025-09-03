//nolint:tagliatelle
package models

type TradingMode string

func (o TradingMode) String() string { return string(o) }

const (
	CashTradingMode TradingMode = "cash"
)

type OrderType string

func (o OrderType) String() string { return string(o) }

const (
	OrderTypeMarket OrderType = "market"
)

type OrderSide string

func (o OrderSide) String() string { return string(o) }

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

type OrderState string

func (o OrderState) String() string { return string(o) }

const (
	OrderStateCanceled        OrderState = "canceled"
	OrderStateLive            OrderState = "live"
	OrderStatePartiallyFilled OrderState = "partially_filled"
	OrderStateFilled          OrderState = "filled"
	OrderStateMmpCanceled     OrderState = "mmp_canceled"
)

type (
	PlaceOrder struct {
		ClientOrderID   string `json:"clOrdId"`
		SystemOrderID   string `json:"ordId"`
		SystemTimestamp string `json:"ts"`
		OrderTag        string `json:"tag,omitempty"`
		SuccessMsg      string `json:"sMsg,omitempty"`
		SuccessCode     string `json:"sCode,omitempty"`
	}
	CancelOrder struct {
		OrdID   string `json:"ordId"`
		ClOrdID string `json:"clOrdId"`
		SMsg    string `json:"sMsg"`
		SCode   string `json:"sCode"`
	}
	AmendOrder struct {
		OrdID   string `json:"ordId"`
		ClOrdID string `json:"clOrdId"`
		ReqID   string `json:"reqId"`
		SMsg    string `json:"sMsg"`
		SCode   string `json:"sCode"`
	}
	ClosePosition struct {
		InstID  string `json:"instId"`
		PosSide string `json:"posSide"`
	}
	Order struct {
		InstID      string     `json:"instId"`
		Ccy         string     `json:"ccy"`
		OrdID       string     `json:"ordId"`
		ClOrdID     string     `json:"clOrdId"`
		TradeID     string     `json:"tradeId"`
		Tag         string     `json:"tag"`
		Category    string     `json:"category"`
		FeeCcy      string     `json:"feeCcy"`
		RebateCcy   string     `json:"rebateCcy"`
		Px          string     `json:"px"`
		Sz          string     `json:"sz"`
		Pnl         string     `json:"pnl"`
		AccFillSz   string     `json:"accFillSz"`
		FillPx      string     `json:"fillPx"`
		FillSz      string     `json:"fillSz"`
		FillTime    string     `json:"fillTime"`
		AvgPx       string     `json:"avgPx"`
		Lever       string     `json:"lever"`
		TpTriggerPx string     `json:"tpTriggerPx"`
		TpOrdPx     string     `json:"tpOrdPx"`
		SlTriggerPx string     `json:"slTriggerPx"`
		SlOrdPx     string     `json:"slOrdPx"`
		Fee         string     `json:"fee"`
		Rebate      string     `json:"rebate"`
		State       OrderState `json:"state"`
		TdMode      string     `json:"tdMode"`
		PosSide     string     `json:"posSide"`
		Side        string     `json:"side"`
		OrdType     string     `json:"ordType"`
		InstType    string     `json:"instType"`
		TgtCcy      string     `json:"tgtCcy"`
		UTime       string     `json:"uTime"`
		CTime       string     `json:"cTime"`
	}
	TransactionDetail struct {
		InstID   string `json:"instId"`
		OrdID    string `json:"ordId"`
		TradeID  string `json:"tradeId"`
		ClOrdID  string `json:"clOrdId"`
		BillID   string `json:"billId"`
		Tag      string `json:"tag"`
		FillPx   string `json:"fillPx"`
		FillSz   string `json:"fillSz"`
		FeeCcy   string `json:"feeCcy"`
		Fee      string `json:"fee"`
		InstType string `json:"instType"`
		Side     string `json:"side"`
		PosSide  string `json:"posSide"`
		ExecType string `json:"execType"`
		TS       int64  `json:"ts"`
	}
	PlaceAlgoOrder struct {
		AlgoID string `json:"algoId"`
		SMsg   string `json:"sMsg"`
		SCode  int64  `json:"sCode"`
	}
	CancelAlgoOrder struct {
		AlgoID string `json:"algoId"`
		SMsg   string `json:"sMsg"`
		SCode  int64  `json:"sCode"`
	}
	AlgoOrder struct {
		InstID       string `json:"instId"`
		Ccy          string `json:"ccy"`
		OrdID        string `json:"ordId"`
		AlgoID       string `json:"algoId"`
		ClOrdID      string `json:"clOrdId"`
		TradeID      string `json:"tradeId"`
		Tag          string `json:"tag"`
		Category     string `json:"category"`
		FeeCcy       string `json:"feeCcy"`
		RebateCcy    string `json:"rebateCcy"`
		TimeInterval string `json:"timeInterval"`
		Px           string `json:"px"`
		PxVar        string `json:"pxVar"`
		PxSpread     string `json:"pxSpread"`
		PxLimit      string `json:"pxLimit"`
		Sz           string `json:"sz"`
		SzLimit      string `json:"szLimit"`
		ActualSz     string `json:"actualSz"`
		ActualPx     string `json:"actualPx"`
		Pnl          string `json:"pnl"`
		AccFillSz    string `json:"accFillSz"`
		FillPx       string `json:"fillPx"`
		FillSz       string `json:"fillSz"`
		FillTime     string `json:"fillTime"`
		AvgPx        string `json:"avgPx"`
		Lever        string `json:"lever"`
		TpTriggerPx  string `json:"tpTriggerPx"`
		TpOrdPx      string `json:"tpOrdPx"`
		SlTriggerPx  string `json:"slTriggerPx"`
		SlOrdPx      string `json:"slOrdPx"`
		OrdPx        string `json:"ordPx"`
		Fee          string `json:"fee"`
		Rebate       string `json:"rebate"`
		State        string `json:"state"`
		TdMode       string `json:"tdMode"`
		ActualSide   string `json:"actualSide"`
		PosSide      string `json:"posSide"`
		Side         string `json:"side"`
		OrdType      string `json:"ordType"`
		InstType     string `json:"instType"`
		TgtCcy       string `json:"tgtCcy"`
		CTime        int64  `json:"cTime"`
		TriggerTime  int64  `json:"triggerTime"`
	}
)
