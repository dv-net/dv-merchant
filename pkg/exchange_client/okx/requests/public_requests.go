//nolint:tagliatelle
package requests

type (
	GetInstruments struct {
		Uly      string `json:"uly,omitempty"`
		InstID   string `json:"instId,omitempty"`
		InstType string `json:"instType"`
	}
	GetDeliveryExerciseHistory struct {
		Uly      string `json:"uly"`
		After    int64  `json:"after,omitempty,string"`
		Before   int64  `json:"before,omitempty,string"`
		Limit    int64  `json:"limit,omitempty,string"`
		InstType string `json:"instType"`
	}
	GetOpenInterest struct {
		Uly      string `json:"uly,omitempty"`
		InstID   string `json:"instId,omitempty"`
		InstType string `json:"instType"`
	}
	GetFundingRate struct {
		InstID string `json:"instId"`
	}
	GetLimitPrice struct {
		InstID string `json:"instId"`
	}
	GetOptionMarketData struct {
		Uly     string `json:"uly"`
		ExpTime string `json:"expTime,omitempty"`
	}
	GetEstimatedDeliveryExercisePrice struct {
		Uly     string `json:"uly"`
		ExpTime string `json:"expTime,omitempty"`
	}
	GetDiscountRateAndInterestFreeQuota struct {
		Uly        string  `json:"uly"`
		Ccy        string  `json:"ccy,omitempty"`
		DiscountLv float64 `json:"discountLv,string"`
	}
	GetLiquidationOrders struct {
		InstID   string `json:"instId,omitempty"`
		Ccy      string `json:"ccy,omitempty"`
		Uly      string `json:"uly,omitempty"`
		After    int64  `json:"after,omitempty,string"`
		Before   int64  `json:"before,omitempty,string"`
		Limit    int64  `json:"limit,omitempty,string"`
		InstType string `json:"instType"`
		MgnMode  string `json:"mgnMode,omitempty"`
		Alias    string `json:"alias,omitempty"`
		State    string `json:"state,omitempty"`
	}
	GetMarkPrice struct {
		InstID   string `json:"instId,omitempty"`
		Uly      string `json:"uly,omitempty"`
		InstType string `json:"instType"`
	}
	GetPositionTiers struct {
		InstID   string `json:"instId,omitempty"`
		Uly      string `json:"uly,omitempty"`
		InstType string `json:"instType"`
		TdMode   string `json:"tdMode"`
		Tier     int64  `json:"tier,omitempty"`
	}
	GetUnderlying struct {
		InstType string `json:"instType"`
	}
	Status struct {
		State string `json:"state,omitempty"`
	}
)
