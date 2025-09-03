package eproxy

import (
	"github.com/dv-net/dv-proto/gen/go/eproxy/addresses/v2/addressesv2connect"
	"github.com/dv-net/dv-proto/gen/go/eproxy/evm/v2/evmv2connect"
	"github.com/dv-net/dv-proto/gen/go/eproxy/transactions/v2/transactionsv2connect"
	"github.com/dv-net/dv-proto/go/eproxy"
)

type IExplorerProxy interface {
	EVM() evmv2connect.EVMServiceClient
	Wallets() addressesv2connect.AddressesServiceClient
	Transactions() transactionsv2connect.TransactionsServiceClient
}

type Service struct {
	client *eproxy.Client
}

func (s *Service) Wallets() addressesv2connect.AddressesServiceClient {
	return s.client.AddressesClient
}

func (s *Service) Transactions() transactionsv2connect.TransactionsServiceClient {
	return s.client.TransactionsClient
}

func (s *Service) EVM() evmv2connect.EVMServiceClient {
	return s.client.EVMClient
}

func New(client *eproxy.Client) *Service {
	return &Service{
		client: client,
	}
}
