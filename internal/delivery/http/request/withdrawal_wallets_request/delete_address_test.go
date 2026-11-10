package withdrawal_wallets_request

import (
	"testing"

	"github.com/dv-net/dv-merchant/internal/tools"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBatchDeleteAddressRequestValidation(t *testing.T) {
	id := uuid.New()
	many := make([]uuid.UUID, 101)
	for i := range many {
		many[i] = uuid.New()
	}

	tests := []struct {
		name    string
		ids     []uuid.UUID
		wantErr bool
	}{
		{name: "one id", ids: []uuid.UUID{id}},
		{name: "nil list", ids: nil, wantErr: true},
		{name: "empty list", ids: []uuid.UUID{}, wantErr: true},
		{name: "duplicates", ids: []uuid.UUID{id, id}, wantErr: true},
		{name: "too many", ids: many, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tools.DefaultStructValidator().Validate(&BatchDeleteAddressRequest{TOTP: "123456", AddressIDs: tt.ids})
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
