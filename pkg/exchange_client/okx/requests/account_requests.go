//nolint:tagliatelle
package requests

type (
	GetAccountBalance struct {
		Ccy []string `json:"ccy,omitempty"`
	}
	GetMaxWithdrawal struct {
		Ccy []string `json:"ccy,omitempty"`
	}
	GetPositions struct {
		InstID   []string `json:"instId,omitempty"`
		PosID    []string `json:"posId,omitempty"`
		InstType string   `json:"instType,omitempty"`
	}
	GetAccountAndPositionRisk struct {
		InstType string `json:"instType,omitempty"`
	}
	GetBills struct {
		Ccy      string `json:"ccy,omitempty"`
		After    int64  `json:"after,omitempty,string"`
		Before   int64  `json:"before,omitempty,string"`
		Limit    int64  `json:"limit,omitempty,string"`
		InstType string `json:"instType,omitempty"`
		MgnMode  string `json:"mgnMode,omitempty"`
		CtType   string `json:"ctType,omitempty"`
		Type     string `json:"type,omitempty,string"`
		SubType  string `json:"subType,omitempty,string"`
	}
	SetPositionMode struct {
		PositionMode string `json:"positionMode"`
	}
	SetLeverage struct {
		Lever   int64  `json:"lever,string"`
		InstID  string `json:"instId,omitempty"`
		Ccy     string `json:"ccy,omitempty"`
		MgnMode string `json:"mgnMode"`
		PosSide string `json:"posSide,omitempty"`
	}
	GetMaxBuySellAmount struct {
		Ccy    string   `json:"ccy,omitempty"`
		Px     float64  `json:"px,string,omitempty"`
		InstID []string `json:"instId"`
		TdMode string   `json:"tdMode"`
	}
	GetMaxAvailableTradeAmount struct {
		Ccy        string `json:"ccy,omitempty"`
		InstID     string `json:"instId"`
		ReduceOnly bool   `json:"reduceOnly,omitempty"`
		TdMode     string `json:"tdMode"`
	}
	IncreaseDecreaseMargin struct {
		InstID     string  `json:"instId"`
		Amt        float64 `json:"amt,string"`
		PosSide    string  `json:"posSide"`
		ActionType string  `json:"actionType"`
	}
	GetLeverage struct {
		InstID  []string `json:"instId"`
		MgnMode string   `json:"mgnMode"`
	}
	GetMaxLoan struct {
		InstID  string `json:"instId"`
		MgnCcy  string `json:"mgnCcy,omitempty"`
		MgnMode string `json:"mgnMode"`
	}
	GetFeeRates struct {
		InstID   string `json:"instId,omitempty"`
		Uly      string `json:"uly,omitempty"`
		Category uint8  `json:"category,omitempty,string"`
		InstType string `json:"instType"`
	}
	GetInterestAccrued struct {
		InstID  string `json:"instId,omitempty"`
		Ccy     string `json:"ccy,omitempty"`
		After   int64  `json:"after,omitempty,string"`
		Before  int64  `json:"before,omitempty,string"`
		Limit   int64  `json:"limit,omitempty,string"`
		MgnMode string `json:"mgnMode,omitempty"`
	}
)
