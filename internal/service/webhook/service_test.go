package webhook_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/dv-net/dv-merchant/internal/service/webhook"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
)

// rewrite test check sign from header
func TestMessage(t *testing.T) {
	fmt.Printf(`protocol=="socks5" && "Version:5 Method:No Authentication(0x00)" && after="%s" && country="CN"`, time.Now().AddDate(0, -3, 0).Format(time.DateOnly))
	whUUID, _ := uuid.Parse(`3220a124-690a-45a1-8ed4-999258ad9212`)
	testData := []struct {
		name    string
		message webhook.Message
		json    string
	}{
		{
			name: "TEST-1",
			message: webhook.Message{
				WebhookID: whUUID,
				Type:      "msg-type-1",
				TxID:      uuid.MustParse("b98e8188-21b4-11ef-9d04-d74258ea4be5"),
				Data:      []byte(`{"param1": "value1", "param2": "value2"}`),
				Delay:     0,
				Signature: "123456",
			},
			json: `{
  "type": "msg-type-1",
  "tx_id": "b98e8188-21b4-11ef-9d04-d74258ea4be5",
  "data": "eyJwYXJhbTEiOiAidmFsdWUxIiwgInBhcmFtMiI6ICJ2YWx1ZTIifQ==",
  "signature": "123456",
  "delay": 0,
  "webhook_id": "3220a124-690a-45a1-8ed4-999258ad9212"
}`,
		},
	}
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			data, err := json.Marshal(test.message)
			if assert.NoError(t, err) {
				assert.JSONEq(t, test.json, string(data))
			}
			var m webhook.Message
			err = json.Unmarshal(data, &m)
			if assert.NoError(t, err) {
				assert.Equal(t, test.message, m)
			}
		})
	}
}
