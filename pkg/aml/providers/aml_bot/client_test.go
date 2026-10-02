package aml_bot

import (
	"encoding/json"
	"testing"

	"github.com/dv-net/dv-merchant/pkg/aml"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestPrepareRiskData(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		status    aml.CheckStatus
		wantScore string
		wantLevel aml.CheckRiskLevel
	}{
		{
			name:      "limited flow: level from risk_score_level, score from risky volume share",
			data:      `{"amount": 400000000, "risky_volume": 1600000, "risk_score_level": "low", "hasHighRisk": true, "signals": {"risky_exchange": 0.426, "scam": null}}`,
			status:    aml.CheckStatusSuccess,
			wantScore: "0.4",
			wantLevel: aml.CheckRiskLevelLow,
		},
		{
			name:      "limited flow without risky volume",
			data:      `{"amount": 16175994820000000000, "risky_volume": 0, "risk_score_level": "low"}`,
			status:    aml.CheckStatusSuccess,
			wantScore: "0",
			wantLevel: aml.CheckRiskLevelLow,
		},
		{
			name:      "full flow: riskscore wins over risky volume, level from risk_score_level",
			data:      `{"riskscore": 0.236, "amount": 100, "risky_volume": 50, "risk_score_level": "medium"}`,
			status:    aml.CheckStatusSuccess,
			wantScore: "23.6",
			wantLevel: aml.CheckRiskLevelMedium,
		},
		{
			name:      "full flow without risk_score_level falls back to score thresholds",
			data:      `{"riskscore": 0.85}`,
			status:    aml.CheckStatusSuccess,
			wantScore: "85",
			wantLevel: aml.CheckRiskLevelSevere,
		},
		{
			name:      "blacklist flag forces severe",
			data:      `{"amount": 100, "risky_volume": 0, "risk_score_level": "low", "hasBlackListFlag": true}`,
			status:    aml.CheckStatusSuccess,
			wantScore: "0",
			wantLevel: aml.CheckRiskLevelSevere,
		},
		{
			name:      "pending check without risk data is undefined",
			data:      `{"status": "pending"}`,
			status:    aml.CheckStatusNew,
			wantScore: "0",
			wantLevel: aml.CheckRiskLevelUndefined,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data CheckData
			require.NoError(t, json.Unmarshal([]byte(tt.data), &data))

			score, level := prepareRiskData(&data, tt.status)
			require.True(t, decimal.RequireFromString(tt.wantScore).Equal(score), "score %s", score)
			require.Equal(t, tt.wantLevel, level)
		})
	}
}
