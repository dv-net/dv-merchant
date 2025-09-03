package processing_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/util"

	"connectrpc.com/connect"
	commonv1 "github.com/dv-net/dv-processing/api/processing/common/v1"
	transferv1 "github.com/dv-net/dv-processing/api/processing/transfer/v1"
	"github.com/stretchr/testify/require"
)

const (
	processingURL = "http://localhost:9000"
)

func newClient() *processing.Processing {
	return processing.NewProcessing(processingURL)
}

func TestCreateTransfer(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip()
	}

	p := newClient()

	resp, err := p.Transfers().Create(context.Background(), connect.NewRequest(&transferv1.CreateRequest{
		OwnerId:         "c012eb2a-54c2-4897-b445-0832b8afbae1",
		RequestId:       "awdawdawdawdawd",
		Blockchain:      commonv1.Blockchain_BLOCKCHAIN_TRON,
		FromAddresses:   []string{"TSRM2HueXrL1ZjqUsRWffareiFbtqPtUUD"},
		ToAddresses:     []string{"TQ6DkBmxz3Zk7neh8mwmmkfJsVjrE9wwjY"},
		AssetIdentifier: "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		WholeAmount:     true,
		Kind:            util.Pointer("resources"),
	}))
	if err != nil {
		if code, ok := processing.ErrorRPCCode(err); ok {
			fmt.Println(code)
		}
	}
	require.NoError(t, err)

	fmt.Printf("response: %+v\n", resp)
}
