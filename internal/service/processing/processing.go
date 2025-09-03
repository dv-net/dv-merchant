package processing

import (
	"net/http"

	"connectrpc.com/connect"
	"github.com/dv-net/dv-processing/api/processing/client/v1/clientv1connect"
	"github.com/dv-net/dv-processing/api/processing/owner/v1/ownerv1connect"
	"github.com/dv-net/dv-processing/api/processing/system/v1/systemv1connect"
	"github.com/dv-net/dv-processing/api/processing/transfer/v1/transferv1connect"
	"github.com/dv-net/dv-processing/api/processing/wallet/v1/walletv1connect"
)

type Processing struct {
	clientService    clientv1connect.ClientServiceClient
	ownerService     ownerv1connect.OwnerServiceClient
	transfersService transferv1connect.TransferServiceClient
	walletService    walletv1connect.WalletServiceClient
	systemService    systemv1connect.SystemServiceClient
}

func NewProcessing(processingURL string, opts ...connect.ClientOption) *Processing {
	return &Processing{
		clientService:    clientv1connect.NewClientServiceClient(http.DefaultClient, processingURL, opts...),
		ownerService:     ownerv1connect.NewOwnerServiceClient(http.DefaultClient, processingURL, opts...),
		transfersService: transferv1connect.NewTransferServiceClient(http.DefaultClient, processingURL, opts...),
		walletService:    walletv1connect.NewWalletServiceClient(http.DefaultClient, processingURL, opts...),
		systemService:    systemv1connect.NewSystemServiceClient(http.DefaultClient, processingURL, opts...),
	}
}

func (p *Processing) Client() clientv1connect.ClientServiceClient { return p.clientService }
func (p *Processing) Owner() ownerv1connect.OwnerServiceClient    { return p.ownerService }
func (p *Processing) Transfers() transferv1connect.TransferServiceClient {
	return p.transfersService
}
func (p *Processing) Wallet() walletv1connect.WalletServiceClient { return p.walletService }
func (p *Processing) System() systemv1connect.SystemServiceClient { return p.systemService }
