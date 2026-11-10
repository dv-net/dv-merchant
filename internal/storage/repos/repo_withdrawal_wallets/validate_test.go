package repo_withdrawal_wallets

import (
	"testing"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestUpdateParamsValidateMinBalance(t *testing.T) {
	dec := func(v string) decimal.NullDecimal {
		return decimal.NullDecimal{Decimal: decimal.RequireFromString(v), Valid: true}
	}

	tests := []struct {
		name       string
		minBalance decimal.NullDecimal
		minUSD     decimal.NullDecimal
		wantErr    bool
	}{
		{name: "positive native", minBalance: dec("0.0001")},
		{name: "positive usd", minUSD: dec("1")},
		{name: "no limits", wantErr: false},
		{name: "zero native", minBalance: dec("0"), wantErr: true},
		{name: "negative native", minBalance: dec("-1"), wantErr: true},
		{name: "zero usd", minUSD: dec("0"), wantErr: true},
		{name: "both set", minBalance: dec("1"), minUSD: dec("1"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UpdateParams{
				WithdrawalMinBalance:    tt.minBalance,
				WithdrawalMinBalanceUsd: tt.minUSD,
				WithdrawalInterval:      models.WithdrawalIntervalNever.String(),
				CurrencyID:              "USDT.Tron",
				UserID:                  uuid.New(),
			}.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
