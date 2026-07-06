package coinkyt

import (
	"time"

	"github.com/dv-net/dv-merchant/pkg/aml"
	"github.com/shopspring/decimal"
)

type Entity struct {
	Name      string `json:"name"`
	TypeLabel string `json:"type_label"`
	Type      string `json:"type"`
}

type Indirect struct {
	TypeLabel      string `json:"type_label"`
	Type           string `json:"type"`
	RiskScoreGrade string `json:"risk_score_grade"`
	TotalCount     string `json:"total_count"`
	TotalCountCoef string `json:"total_count_coef"`
	Total          string `json:"total"`
}

type DiscreteItem struct {
	TypeLabel         string `json:"type_label"`
	Type              string `json:"type"`
	RiskScoreGrade    string `json:"risk_score_grade"`
	ReceivedCountCoef string `json:"received_count_coef"`
	Received          string `json:"received"`
	SentCountCoef     string `json:"sent_count_coef"`
	Sent              string `json:"sent"`
	TotalCountCoef    string `json:"total_count_coef"`
	Total             string `json:"total"`
}

type Discrete struct {
	Received []DiscreteItem `json:"received"`
	Sent     []DiscreteItem `json:"sent"`
}

type AlertTriggeredBy struct {
	TypeLabel      string `json:"type_label"`
	Type           string `json:"type"`
	TotalCount     string `json:"total_count"`
	TotalCountCoef string `json:"total_count_coef"`
	Total          string `json:"total"`
}

type Alert struct {
	ID          string             `json:"id"`
	AlertID     string             `json:"alert_id"`
	RiskProfile string             `json:"risk_profile"`
	RiskGrade   string             `json:"risk_grade"`
	TriggeredBy []AlertTriggeredBy `json:"triggered_by"`
	Date        time.Time          `json:"date"`
	Blockchain  string             `json:"blockchain"`
	Token       string             `json:"token"`
	Address     string             `json:"address"`
	Status      bool               `json:"status"`
}

type TransactionResponse struct {
	ID               string          `json:"id"`
	Blockchain       string          `json:"blockchain"`
	Token            string          `json:"token"`
	Address          string          `json:"address"`
	RiskScore        decimal.Decimal `json:"risk_score"`
	RiskScoreGrade   string          `json:"risk_score_grade"`
	Inputs           int             `json:"inputs"`
	Outputs          int             `json:"outputs"`
	FromEntity       []Entity        `json:"from_entity"`
	ToEntity         []Entity        `json:"to_entity"`
	Indirects        []Indirect      `json:"indirects"`
	BlockNumber      int64           `json:"block_number"`
	Timestamp        time.Time       `json:"timestamp"`
	Link             string          `json:"link"`
	TooManyIndirects bool            `json:"too_many_indirects"`
	Discrete         *Discrete       `json:"discrete"`
	InstantCheck     bool            `json:"instant_check"`
	Alerts           []Alert         `json:"alerts"`
}

func (r *TransactionResponse) ToAMLRiskLevel() aml.CheckRiskLevel {
	switch r.RiskScoreGrade {
	case "low":
		return aml.CheckRiskLevelLow
	case "moderate":
		return aml.CheckRiskLevelMedium
	case "high":
		return aml.CheckRiskLevelHigh
	case "severe":
		return aml.CheckRiskLevelSevere
	default:
		return aml.CheckRiskLevelUndefined
	}
}
