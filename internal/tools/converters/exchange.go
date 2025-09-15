package converters

import (
	"sort"
	"strings"

	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/exchange_response"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/exchange"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_orders"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"

	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

func NewExchangeListResponseFromDto(dto *exchange.ActiveExchangeListDTO) exchange_response.ExchangeListResponse {
	res := exchange_response.ExchangeListResponse{
		CurrentExchange: dto.CurrentExchange,
	}

	if dto.SwapState != nil {
		res.SwapState = dto.SwapState
	}
	if dto.WithdrawalState != nil {
		res.WithdrawalState = dto.WithdrawalState
	}

	// Sort exchanges by name
	sort.Slice(dto.Exchanges, func(i, j int) bool {
		return dto.Exchanges[i].Exchange < dto.Exchanges[j].Exchange
	})

	for _, ex := range dto.Exchanges {
		data := exchange_response.ExchangeData{
			Exchange: ex.Exchange,
			Slug:     ex.Slug.String(),
			Keys:     NewExchangeKeysResponseFromDto(ex.Keys),
		}
		if !ex.ExchangeConnectedAt.IsZero() {
			data.ExchangeConnectedAt = ex.ExchangeConnectedAt
		}
		res.Exchanges = append(res.Exchanges, data)
	}

	return res
}

func NewExchangeKeysResponseFromDto(keys []exchange.KeysExchangeDTO) []exchange_response.ExchangeKey {
	res := make([]exchange_response.ExchangeKey, 0, len(keys))
	for _, key := range keys {
		res = append(res, NewExchangeKeyResponseByDto(key))
	}

	return res
}

func NewExchangeKeyResponseByDto(dto exchange.KeysExchangeDTO) exchange_response.ExchangeKey {
	var maskedValue *string
	if dto.Value != nil && len(*dto.Value) > 4 {
		maskedLength := len(*dto.Value) - 4
		mask := strings.Repeat("*", maskedLength)
		suffix := (*dto.Value)[len(*dto.Value)-4:]
		maskedValue = lo.ToPtr(mask + suffix)
	} else {
		maskedValue = dto.Value
	}
	return exchange_response.ExchangeKey{
		Name:  dto.Name,
		Value: maskedValue,
	}
}

func FromExchangeBalanceModelToResponse(m []*models.AccountBalanceDTO) exchange_response.ExchangeBalanceResponse {
	total := decimal.New(0, 0)
	balances := make([]exchange_response.ExchangeAsset, 0, len(m))
	for _, balance := range m {
		balances = append(balances, exchange_response.ExchangeAsset{
			Currency:  balance.Currency,
			Amount:    balance.Amount,
			AmountUSD: balance.AmountUSD.RoundDown(2),
		})
		total = total.Add(balance.AmountUSD)
	}
	return exchange_response.ExchangeBalanceResponse{
		Balances: balances,
		TotalUSD: total.RoundDown(2),
	}
}

func FromExchangeTestConnectionModelToResponse(m exchange.TestConnectionResult) exchange_response.ExchangeTestConnectionResponse {
	return exchange_response.ExchangeTestConnectionResponse{
		Exchange:     m.Slug.String(),
		ErrorMessage: m.ErrMsg,
	}
}

func FromExchangeTestConnectionModelsToResponses(m []exchange.TestConnectionResult) []exchange_response.ExchangeTestConnectionResponse {
	res := make([]exchange_response.ExchangeTestConnectionResponse, 0, len(m))
	for _, v := range m {
		res = append(res, FromExchangeTestConnectionModelToResponse(v))
	}
	return res
}

func GetUserExchangePairsResponse(m []*models.UserExchangePair) []exchange_response.ExchangeUserPairResponse {
	res := make([]exchange_response.ExchangeUserPairResponse, 0, len(m))
	for _, v := range m {
		r := exchange_response.ExchangeUserPairResponse{}
		switch v.Type {
		case "sell":
			r.DisplayName = v.CurrencyFrom + "/" + v.CurrencyTo
		case "buy":
			r.DisplayName = v.CurrencyTo + "/" + v.CurrencyFrom
		}
		r.DisplayName = strings.ToUpper(r.DisplayName)
		res = append(res, r)
	}
	return res
}

func GetWithdrawalsHistoryResponse(m *storecmn.FindResponseWithFullPagination[*models.ExchangeWithdrawalHistoryDTO]) *storecmn.FindResponseWithFullPagination[*exchange_response.ExchangeWithdrawalHistoryResponse] {
	items := make([]*exchange_response.ExchangeWithdrawalHistoryResponse, 0, len(m.Items))
	for _, v := range m.Items {
		item := &exchange_response.ExchangeWithdrawalHistoryResponse{
			ID:           v.ID.String(),
			CurrencyID:   v.Currency,
			Chain:        v.Chain,
			Status:       v.Status.String(),
			CreatedAt:    v.CreatedAt.Time,
			Address:      v.Address,
			ExchangeID:   v.ExchangeID.String(),
			ExchangeSlug: v.Slug.String(),
		}
		if v.Txid.Valid {
			item.TxID = &v.Txid.String
		}
		if v.NativeAmount.Valid {
			amt := v.NativeAmount.Decimal.String()
			item.AmountNative = &amt
		}
		if v.FiatAmount.Valid {
			amt := v.FiatAmount.Decimal.String()
			item.AmountUSD = &amt
		}
		if v.FailReason.Valid {
			item.FailReason = &v.FailReason.String
		}
		items = append(items, item)
	}

	return &storecmn.FindResponseWithFullPagination[*exchange_response.ExchangeWithdrawalHistoryResponse]{
		Items:      items,
		Pagination: m.Pagination,
	}
}

func GetWithdrawalHistoryResponse(m *models.ExchangeWithdrawalHistoryDTO) *exchange_response.ExchangeWithdrawalHistoryResponse {
	item := &exchange_response.ExchangeWithdrawalHistoryResponse{
		ID:         m.ID.String(),
		CurrencyID: m.Currency,
		Chain:      m.Chain,
		Status:     m.Status.String(),
		CreatedAt:  m.CreatedAt.Time,
		Address:    m.Address,
	}
	if m.Txid.Valid {
		item.TxID = &m.Txid.String
	}
	if m.NativeAmount.Valid {
		amt := m.NativeAmount.Decimal.String()
		item.AmountNative = &amt
	}
	if m.FiatAmount.Valid {
		amt := m.FiatAmount.Decimal.String()
		item.AmountUSD = &amt
	}
	return item
}

func GetOrdersHistoryResponse(m *storecmn.FindResponseWithFullPagination[*repo_exchange_orders.GetExchangeOrdersByUserAndExchangeIDRow]) *storecmn.FindResponseWithFullPagination[*exchange_response.ExchangeOrderHistoryResponse] {
	items := make([]*exchange_response.ExchangeOrderHistoryResponse, 0, len(m.Items))
	for _, v := range m.Items {
		item := &exchange_response.ExchangeOrderHistoryResponse{
			ID:              v.ID.String(),
			UserID:          v.UserID.String(),
			ExchangeOrderID: v.ExchangeOrderID.String,
			ExchangeID:      v.ExchangeID.String(),
			ExchangeSlug:    v.Slug,
			ClientOrderID:   v.ClientOrderID.String,
			CreatedAt:       v.CreatedAt.Time,
			Symbol:          v.Symbol,
			Side:            v.Side.String(),
			Amount:          v.Amount.String(),
			Status:          v.Status.String(),
		}
		if v.AmountUsd.Valid {
			amt := v.AmountUsd.Decimal.String()
			item.AmountUsd = amt
		}
		if v.FailReason.Valid {
			item.FailReason = v.FailReason.String
		}

		items = append(items, item)
	}

	return &storecmn.FindResponseWithFullPagination[*exchange_response.ExchangeOrderHistoryResponse]{
		Items:      items,
		Pagination: m.Pagination,
	}
}

func GetWithdrawalRulesResponse(m []*models.WithdrawalRulesDTO) []*exchange_response.ExchangeWithdrawalRulesResponse {
	res := make([]*exchange_response.ExchangeWithdrawalRulesResponse, 0, len(m))
	for _, v := range m {
		item := &exchange_response.ExchangeWithdrawalRulesResponse{
			Currency:           v.Currency,
			Chain:              v.Chain,
			MinDepositAmount:   v.MinDepositAmount,
			MinWithdrawAmount:  v.MinWithdrawAmount,
			NumOfConfirmations: v.NumOfConfirmations,
			WithdrawPrecision:  v.WithdrawPrecision,
		}
		if v.Fee != "" {
			item.Fee = &v.Fee
		}
		if v.MaxWithdrawAmount != "" {
			item.MaxWithdrawAmount = &v.MaxWithdrawAmount
		}
		if v.WithdrawFeeType != "" {
			fType := v.WithdrawFeeType.String()
			item.WithdrawFeeType = &fType
		}
		if v.WithdrawQuotaPerDay != "" {
			item.WithdrawQuotaPerDay = &v.WithdrawQuotaPerDay
		}
		res = append(res, item)
	}
	return res
}

func GetWithdrawalSettingsResponse(m []*models.ExchangeWithdrawalSetting) []*exchange_response.ExchangeWithdrawalSettingResponse {
	res := make([]*exchange_response.ExchangeWithdrawalSettingResponse, 0, len(m))
	for _, s := range m {
		res = append(res, &exchange_response.ExchangeWithdrawalSettingResponse{
			ID:         s.ID.String(),
			CurrencyID: s.Currency,
			Chain:      s.Chain,
			Address:    s.Address,
			MinAmount:  s.MinAmount,
			Enabled:    s.IsEnabled,
			CreatedAt:  s.CreatedAt.Time,
		})
	}
	return res
}

func GetWithdrawalSettingResponse(m *models.ExchangeWithdrawalSetting) *exchange_response.ExchangeWithdrawalSettingResponse {
	return &exchange_response.ExchangeWithdrawalSettingResponse{
		ID:         m.ID.String(),
		CurrencyID: m.Currency,
		Chain:      m.Chain,
		Address:    m.Address,
		MinAmount:  m.MinAmount,
		Enabled:    m.IsEnabled,
		CreatedAt:  m.CreatedAt.Time,
	}
}

func UpdateDepositAddressesResponse(addresses []*models.DepositAddressDTO, rules []*models.WithdrawalRulesDTO) []*exchange_response.DepositUpdateResponse {
	res := make([]*exchange_response.DepositUpdateResponse, 0, len(addresses))
	for _, a := range addresses {
		dto := &exchange_response.DepositUpdateResponse{
			Address:     a.Address,
			Currency:    a.Currency,
			Chain:       a.Chain,
			AddressType: a.AddressType,
		}
		rule, exists := lo.Find(rules, func(r *models.WithdrawalRulesDTO) bool {
			return r.Currency == a.InternalCurrency && r.Chain == a.Chain
		})
		if exists {
			dto.MinDepositAmount = rule.MinDepositAmount
		}
		res = append(res, dto)
	}
	return res
}

func GetDepositAddressesResponse(addresses []models.ExchangeGroup, rules []*models.WithdrawalRulesDTO) []*exchange_response.GetDepositAddressesResponse {
	res := make([]*exchange_response.GetDepositAddressesResponse, 0, len(addresses))
	for _, group := range addresses {
		dto := &exchange_response.GetDepositAddressesResponse{
			Slug: group.Slug,
			Name: group.Name,
		}
		for _, a := range group.Addresses {
			item := &exchange_response.DepositUpdateResponse{
				Address:     a.Address,
				Currency:    a.Currency,
				Chain:       a.Chain,
				AddressType: a.AddressType,
			}
			rule, exists := lo.Find(rules, func(r *models.WithdrawalRulesDTO) bool {
				return r.Currency == a.InternalCurrency && r.Chain == a.Chain
			})
			if exists {
				item.MinDepositAmount = rule.MinDepositAmount
			}
			dto.Addresses = append(dto.Addresses, *item)
		}
		res = append(res, dto)
	}
	return res
}
