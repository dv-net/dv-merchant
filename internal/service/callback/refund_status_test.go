package callback

import (
	"testing"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/processing_request"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestRefundTransferBroadcast(t *testing.T) {
	tests := []struct {
		name string
		dto  processing_request.TransferStatusWebhook
		want bool
	}{
		{"nothing sent", processing_request.TransferStatusWebhook{}, false},
		{"broadcast flag", processing_request.TransferStatusWebhook{Broadcast: true}, true},
		{
			"transfer system transaction",
			processing_request.TransferStatusWebhook{SystemTransactions: []processing_request.TransferSystemTransactions{
				{TxType: models.TransferTransactionTypeTransfer, TxHash: "0xabc"},
			}},
			true,
		},
		{
			"only an auxiliary transaction",
			processing_request.TransferStatusWebhook{SystemTransactions: []processing_request.TransferSystemTransactions{
				{TxType: models.TransferTransactionTypeSendBurnBaseAsset, TxHash: "0xabc"},
			}},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, refundTransferBroadcast(tt.dto))
		})
	}
}
