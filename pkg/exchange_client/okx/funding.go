package okx

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/ulule/limiter/v3"

	okxrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/requests"
	okxresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/responses"
)

const (
	getCurrenciesEndpoint        = "/api/v5/asset/currencies"
	getBalanceEndpoint           = "/api/v5/asset/balances"
	fundsTransferEndpoint        = "/api/v5/asset/transfer"
	assetBillsDetailsEndpoint    = "/api/v5/asset/bills"
	getDepositAddressEndpoint    = "/api/v5/asset/deposit-address"
	getDepositHistoryEndpoint    = "/api/v5/asset/deposit-history"
	withdrawalEndpoint           = "/api/v5/asset/withdrawal"
	getWithdrawalHistoryEndpoint = "/api/v5/asset/withdrawal-history"
)

type IOKXFunding interface {
	GetCurrencies(ctx context.Context, req okxrequests.GetCurrencies) (*okxresponses.GetCurrencies, error)
	GetBalance(ctx context.Context, req okxrequests.GetFundingBalance) (*okxresponses.GetFundingBalance, error)
	FundsTransfer(ctx context.Context, req okxrequests.FundsTransfer) (*okxresponses.FundsTransfer, error)
	AssetBillsDetails(ctx context.Context, req okxrequests.AssetBillsDetails) (*okxresponses.AssetBillsDetails, error)
	GetDepositAddress(ctx context.Context, req okxrequests.GetDepositAddress) (*okxresponses.GetDepositAddress, error)
	GetDepositHistory(ctx context.Context, req okxrequests.GetDepositHistory) (*okxresponses.GetDepositHistory, error)
	Withdrawal(ctx context.Context, req okxrequests.Withdrawal) (*okxresponses.Withdrawal, error)
	GetWithdrawalHistory(ctx context.Context, req *okxrequests.GetWithdrawalHistory) (*okxresponses.GetWithdrawalHistory, error)
}

type Funding struct {
	client *Client
}

func NewFunding(clientOpt *ClientOptions, store limiter.Store, opts ...ClientOption) *Funding {
	funding := &Funding{
		client: NewClient(clientOpt, store, opts...),
	}
	funding.initLimiters()
	return funding
}

func (o *Funding) initLimiters() {
	o.client.limiters = map[string]*limiter.Limiter{
		getCurrenciesEndpoint:        limiter.New(o.client.store, limiter.Rate{Limit: 6, Period: time.Second}),
		getBalanceEndpoint:           limiter.New(o.client.store, limiter.Rate{Limit: 6, Period: time.Second}),
		fundsTransferEndpoint:        limiter.New(o.client.store, limiter.Rate{Limit: 6, Period: 2 * time.Second}),
		assetBillsDetailsEndpoint:    limiter.New(o.client.store, limiter.Rate{Limit: 6, Period: time.Second}),
		getDepositAddressEndpoint:    limiter.New(o.client.store, limiter.Rate{Limit: 6, Period: time.Second}),
		getDepositHistoryEndpoint:    limiter.New(o.client.store, limiter.Rate{Limit: 6, Period: time.Second}),
		withdrawalEndpoint:           limiter.New(o.client.store, limiter.Rate{Limit: 6, Period: time.Second}),
		getWithdrawalHistoryEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 6, Period: time.Second}),
	}
}

func (o *Funding) GetCurrencies(ctx context.Context, req okxrequests.GetCurrencies) (*okxresponses.GetCurrencies, error) {
	response := &okxresponses.GetCurrencies{}
	p := getCurrenciesEndpoint
	m := S2M(req)
	if len(req.Ccy) > 0 {
		m["ccy"] = strings.Join(req.Ccy, ",")
	}

	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Funding) GetBalance(ctx context.Context, req okxrequests.GetFundingBalance) (*okxresponses.GetFundingBalance, error) {
	response := &okxresponses.GetFundingBalance{}
	p := getBalanceEndpoint
	m := S2M(req)
	if len(req.Ccy) > 0 {
		m["ccy"] = strings.Join(req.Ccy, ",")
	}
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Funding) FundsTransfer(ctx context.Context, req okxrequests.FundsTransfer) (*okxresponses.FundsTransfer, error) {
	response := &okxresponses.FundsTransfer{}
	p := fundsTransferEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodPost, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Funding) AssetBillsDetails(ctx context.Context, req okxrequests.AssetBillsDetails) (*okxresponses.AssetBillsDetails, error) {
	response := &okxresponses.AssetBillsDetails{}
	p := assetBillsDetailsEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Funding) GetDepositAddress(ctx context.Context, req okxrequests.GetDepositAddress) (*okxresponses.GetDepositAddress, error) {
	response := &okxresponses.GetDepositAddress{}
	p := getDepositAddressEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Funding) GetDepositHistory(ctx context.Context, req okxrequests.GetDepositHistory) (*okxresponses.GetDepositHistory, error) {
	response := &okxresponses.GetDepositHistory{}
	p := getDepositHistoryEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Funding) Withdrawal(ctx context.Context, req okxrequests.Withdrawal) (*okxresponses.Withdrawal, error) {
	response := &okxresponses.Withdrawal{}
	p := withdrawalEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodPost, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Funding) GetWithdrawalHistory(ctx context.Context, req *okxrequests.GetWithdrawalHistory) (*okxresponses.GetWithdrawalHistory, error) {
	response := &okxresponses.GetWithdrawalHistory{}
	p := getWithdrawalHistoryEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}
