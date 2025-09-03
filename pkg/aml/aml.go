package aml

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/shopspring/decimal"
)

type CheckStatus string

const (
	CheckStatusNew     CheckStatus = "new"
	CheckStatusSuccess CheckStatus = "success"
	CheckStatusFailure CheckStatus = "failure"
)

type CheckRiskLevel string

const (
	CheckRiskLevelLow       CheckRiskLevel = "low"
	CheckRiskLevelMedium    CheckRiskLevel = "medium"
	CheckRiskLevelHigh      CheckRiskLevel = "high"
	CheckRiskLevelSevere    CheckRiskLevel = "severe"
	CheckRiskLevelNone      CheckRiskLevel = "none"
	CheckRiskLevelUndefined CheckRiskLevel = "undefined"
)

type CheckResponse struct {
	ExternalID string          `json:"external_id"`
	Score      decimal.Decimal `json:"score"`
	Status     CheckStatus     `json:"status"`
	RiskLevel  *CheckRiskLevel `json:"risk_level"`
	HTTPStatus int             `json:"http_status"`
	Request    json.RawMessage `json:"request"`
	Response   json.RawMessage `json:"response"`
}

type TokenData struct {
	Blockchain      string `json:"blockchain"`
	ContractAddress string `json:"contract_address"`
}

func (td TokenData) IsNative() bool {
	return td.ContractAddress == ""
}

type Direction string

const (
	DirectionIn  Direction = "in"
	DirectionOut Direction = "ount"
)

type InitCheckDTO struct {
	TokenData           TokenData
	Direction           Direction
	TxID, OutputAddress string
}

type RequestAuthorizer interface {
	Authorize(ctx context.Context, req *http.Request) error
}

type Client interface {
	InitCheckTransaction(ctx context.Context, dto InitCheckDTO, auth RequestAuthorizer) (*CheckResponse, error)
	FetchCheckStatus(ctx context.Context, checkID string, auth RequestAuthorizer) (*CheckResponse, error)
	TestRequestWithAuth(ctx context.Context, auth RequestAuthorizer) error
}
