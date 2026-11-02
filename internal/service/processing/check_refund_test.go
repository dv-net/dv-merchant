package processing

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCheckRefundResultFeeCovered(t *testing.T) {
	tests := []struct {
		name   string
		result CheckRefundResult
		want   bool
	}{
		{name: "native coin, no shortage", result: CheckRefundResult{FeeFromAmount: true}, want: true},
		{name: "token, fee covered", result: CheckRefundResult{FeeShortage: decimal.NewNullDecimal(decimal.Zero)}, want: true},
		{name: "token, fee shortage", result: CheckRefundResult{FeeShortage: decimal.NewNullDecimal(decimal.RequireFromString("0.8"))}, want: false},
		{name: "tron, activation required", result: CheckRefundResult{ActivationRequired: true}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.result.FeeCovered())
		})
	}
}
