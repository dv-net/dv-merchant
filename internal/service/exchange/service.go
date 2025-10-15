package exchange

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/exchange_request"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/exchange_manager"
	"github.com/dv-net/dv-merchant/internal/service/exchange_rules"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_addresses"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_chains"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_orders"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_user_keys"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_withdrawal_settings"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchanges"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_exchange_pairs"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_exchanges"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_users"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/internal/util"
	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/go-mods/excel"
	"github.com/gocarina/gocsv"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

type IExchangeService interface { //nolint:interfacebloat
	Run(ctx context.Context)
	GetAvailableExchangesList(ctx context.Context, userID uuid.UUID) (*ActiveExchangeListDTO, error)
	SetExchangeKeys(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug, keysData map[models.ExchangeKeyName]*string) error
	DeleteExchangeKeys(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug) error
	TestConnection(ctx context.Context, user models.User, slug models.ExchangeSlug) error
	TestConnectionRaw(ctx context.Context, slug models.ExchangeSlug, apiKey, secretKey, passphrase string) error
	GetExchangeBalance(ctx context.Context, slug models.ExchangeSlug, user models.User) ([]*models.AccountBalanceDTO, error)
	GetCurrentExchangeBalance(ctx context.Context, user models.User) ([]*models.AccountBalanceDTO, error)
	SetCurrentExchange(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug) error
	UpdateUserExchangePairs(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug, request *exchange_request.UpdateExchangePairsRequest) error
	GetUserExchangePairs(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug) ([]*models.UserExchangePair, error)
	SubmitExchangeOrder(ctx context.Context, userID uuid.UUID, pair *models.UserExchangePair) error
	GetExchangeWithdrawalRules(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug) ([]*models.WithdrawalRulesDTO, error)
	GetExchangeWithdrawalRule(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug, currencyID string) (*models.WithdrawalRulesDTO, error)
	UpdateDepositAddresses(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug) ([]*models.DepositAddressDTO, error)
	GetDepositExchangeAddresses(ctx context.Context, userID uuid.UUID) ([]models.ExchangeGroup, error)
	GetExchangePairs(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug) ([]*models.ExchangeSymbolDTO, error)
	GetExchangeChains(ctx context.Context, slug models.ExchangeSlug) ([]*models.ExchangeChainShort, error)
	GetExchangeOrdersHistory(ctx context.Context, userID uuid.UUID, req exchange_request.GetExchangeOrdersHistoryRequest) (*storecmn.FindResponseWithFullPagination[*repo_exchange_orders.GetExchangeOrdersByUserAndExchangeIDRow], error)
	DownloadExchangeOrdersHistory(ctx context.Context, userID uuid.UUID, req exchange_request.GetExchangeOrdersHistoryExportedRequest) (*bytes.Buffer, error)
	ToggleExchangeWithdrawals(ctx context.Context, slug models.ExchangeSlug, userID uuid.UUID) (*models.UserExchange, error)
	ChangeExchangeWithdrawalState(ctx context.Context, slug models.ExchangeSlug, userID uuid.UUID, state models.ExchangeWithdrawalState) (*models.UserExchange, error)
	ChangeExchangeSwapsState(ctx context.Context, slug models.ExchangeSlug, userID uuid.UUID, state models.ExchangeSwapState) (*models.UserExchange, error)
}

type TestConnectionResult struct {
	Slug   models.ExchangeSlug
	ErrMsg string
}

type Service struct {
	st         storage.IStorage
	exManager  exchange_manager.IExchangeManager
	log        logger.Logger
	exRulesSvc exchange_rules.IExchangeRules
	settingSvc setting.ISettingService
}

func (s *Service) DeleteExchangeKeys(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug) error {
	err := repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		exID, err := s.st.Exchanges(repos.WithTx(tx)).GetExchangeBySlug(ctx, slug)
		if err != nil {
			return err
		}

		if err = s.st.ExchangeWithdrawalSettings(repos.WithTx(tx)).DeleteByUserAndExchangeID(ctx, repo_exchange_withdrawal_settings.DeleteByUserAndExchangeIDParams{
			UserID:     userID,
			ExchangeID: exID.ID,
		}); err != nil {
			return err
		}

		if err = s.st.UserExchangePairs(repos.WithTx(tx)).DeleteByUserAndExchangeID(ctx, repo_user_exchange_pairs.DeleteByUserAndExchangeIDParams{
			UserID:     userID,
			ExchangeID: exID.ID,
		}); err != nil {
			return err
		}

		if err = s.st.ExchangeAddresses(repos.WithTx(tx)).DeleteByUserAndExchangeID(ctx, repo_exchange_addresses.DeleteByUserAndExchangeIDParams{
			UserID:     userID,
			ExchangeID: exID.ID,
		}); err != nil {
			return err
		}

		err = s.st.UserExchanges(repos.WithTx(tx)).DeleteByUserAndExchangeID(ctx, repo_user_exchanges.DeleteByUserAndExchangeIDParams{
			UserID:     userID,
			ExchangeID: exID.ID,
		})
		if err != nil {
			return err
		}

		keys, err := s.st.ExchangeUserKeys(repos.WithTx(tx)).GetKeysByExchangeSlug(ctx, repo_exchange_user_keys.GetKeysByExchangeSlugParams{
			UserID:       userID,
			ExchangeSlug: slug,
		})
		if err != nil {
			return fmt.Errorf("fetch exchange keys: %w", err)
		}

		batch := s.st.ExchangeUserKeys(repos.WithTx(tx)).BatchDeleteByID(ctx, lo.Map(keys, func(item *repo_exchange_user_keys.GetKeysByExchangeSlugRow, _ int) uuid.UUID {
			return item.ID
		}))
		defer func() {
			if err := batch.Close(); err != nil {
				s.log.Error("batch insert exchange user keys close error", err)
			}
		}()

		batch.Exec(func(_ int, batchErr error) {
			if batchErr != nil {
				err = fmt.Errorf("batch delete exchange keys: %w", err)
				return
			}
		})
		if err != nil {
			return err
		}

		// Clear current_exchange if the deleted exchange was active
		usr, err := s.st.Users(repos.WithTx(tx)).GetByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("fetch user: %w", err)
		}
		if usr.ExchangeSlug != nil && usr.ExchangeSlug.Valid() && *usr.ExchangeSlug == slug {
			_, err = s.st.Users(repos.WithTx(tx)).UpdateExchange(ctx, repo_users.UpdateExchangeParams{
				ID:           userID,
				ExchangeSlug: nil,
			})
			if err != nil {
				return fmt.Errorf("clear current exchange: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("delete exchange keys: %w", err)
	}
	return nil
}

func (s *Service) ToggleExchangeWithdrawals(ctx context.Context, slug models.ExchangeSlug, userID uuid.UUID) (*models.UserExchange, error) {
	var updatedItem *models.UserExchange
	err := repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		ex, err := s.st.Exchanges(repos.WithTx(tx)).GetExchangeBySlug(ctx, slug)
		if err != nil {
			return fmt.Errorf("fetch exchange: %w", err)
		}
		upd, err := s.st.UserExchanges(repos.WithTx(tx)).ToggleWithdrawals(ctx, repo_user_exchanges.ToggleWithdrawalParams{
			UserID:     userID,
			ExchangeID: ex.ID,
		})
		if err != nil {
			return fmt.Errorf("toggle withdrawals: %w", err)
		}
		updatedItem = upd
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("toggle exchange withdrawals: %w", err)
	}
	return updatedItem, nil
}

func (s *Service) ChangeExchangeSwapsState(ctx context.Context, slug models.ExchangeSlug, userID uuid.UUID, state models.ExchangeSwapState) (*models.UserExchange, error) { //nolint:dupl
	var updatedItem *models.UserExchange
	err := repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		ex, err := s.st.Exchanges(repos.WithTx(tx)).GetExchangeBySlug(ctx, slug)
		if err != nil {
			return fmt.Errorf("fetch exchange: %w", err)
		}
		upd, err := s.st.UserExchanges(repos.WithTx(tx)).ChangeSwapState(ctx, repo_user_exchanges.ChangeSwapStateParams{
			UserID:     userID,
			ExchangeID: ex.ID,
			State:      state,
		})
		if err != nil {
			return fmt.Errorf("change swaps state: %w", err)
		}
		updatedItem = upd
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("change exchange swaps state: %w", err)
	}
	return updatedItem, nil
}

func (s *Service) ChangeExchangeWithdrawalState(ctx context.Context, slug models.ExchangeSlug, userID uuid.UUID, state models.ExchangeWithdrawalState) (*models.UserExchange, error) { //nolint:dupl
	var updatedItem *models.UserExchange
	err := repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		ex, err := s.st.Exchanges(repos.WithTx(tx)).GetExchangeBySlug(ctx, slug)
		if err != nil {
			return fmt.Errorf("fetch exchange: %w", err)
		}
		upd, err := s.st.UserExchanges(repos.WithTx(tx)).ChangeWithdrawalState(ctx, repo_user_exchanges.ChangeWithdrawalStateParams{
			UserID:     userID,
			ExchangeID: ex.ID,
			State:      state,
		})
		if err != nil {
			return fmt.Errorf("change withdrawal state: %w", err)
		}
		updatedItem = upd
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("change exchange withdrawal state: %w", err)
	}
	return updatedItem, nil
}

func (s *Service) Run(ctx context.Context) {
	spotTicker := time.NewTicker(10 * time.Second)
	ordersTicker := time.NewTicker(20 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-spotTicker.C:
			s.processExchangePairs(ctx)
		case <-ordersTicker.C:
			s.processExchangeOrders(ctx)
		}
	}
}

func (s *Service) GetExchangeOrdersHistory(ctx context.Context, userID uuid.UUID, req exchange_request.GetExchangeOrdersHistoryRequest) (*storecmn.FindResponseWithFullPagination[*repo_exchange_orders.GetExchangeOrdersByUserAndExchangeIDRow], error) {
	commonParams := storecmn.NewCommonFindParams()

	if req.PageSize != nil {
		commonParams.SetPageSize(req.PageSize)
	}
	if req.Page != nil {
		commonParams.SetPage(req.Page)
	}

	var exchangeID uuid.NullUUID
	if req.Slug != nil {
		exID, err := s.st.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlug(*req.Slug))
		if err != nil {
			return nil, fmt.Errorf("fetch exchange: %w", err)
		}
		exchangeID = uuid.NullUUID{
			UUID:  exID.ID,
			Valid: true,
		}
	}

	params := repo_exchange_orders.GetExchangeOrdersByUserAndExchangeIDParams{
		UserID:           userID,
		ExchangeID:       exchangeID,
		CommonFindParams: *commonParams,
	}

	if req.DateFrom != nil {
		dateFrom, err := util.ParseDate(*req.DateFrom)
		if err != nil {
			return nil, fmt.Errorf("invalid date_from format: %w", err)
		}
		params.DateFrom = dateFrom
	}

	if req.DateTo != nil {
		dateTo, err := util.ParseDate(*req.DateTo)
		if err != nil {
			return nil, fmt.Errorf("invalid date_to format: %w", err)
		}
		params.DateTo = dateTo
	}

	orders, err := s.st.ExchangeOrders().GetByUserAndExchangeID(ctx, params)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *Service) DownloadExchangeOrdersHistory(ctx context.Context, userID uuid.UUID, req exchange_request.GetExchangeOrdersHistoryExportedRequest) (*bytes.Buffer, error) {
	var exchangeID uuid.NullUUID
	if req.Slug != nil {
		exID, err := s.st.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlug(*req.Slug))
		if err != nil {
			return nil, fmt.Errorf("fetch exchange: %w", err)
		}
		exchangeID = uuid.NullUUID{
			UUID:  exID.ID,
			Valid: true,
		}
	}

	params := repo_exchange_orders.GetExportExchangeOrderByUserAndExchangeIDParams{
		UserID:     userID,
		ExchangeID: exchangeID,
	}

	if req.DateFrom != nil {
		dateFrom, err := time.Parse(time.DateTime, *req.DateFrom)
		if err != nil {
			return nil, fmt.Errorf("invalid date_from format: %w", err)
		}
		params.DateFrom = &dateFrom
	}

	if req.DateTo != nil {
		dateTo, err := time.Parse(time.DateTime, *req.DateTo)
		if err != nil {
			return nil, fmt.Errorf("invalid date_to format: %w", err)
		}
		params.DateTo = &dateTo
	}

	orders, err := s.st.ExchangeOrders().GetAllByUserFiltered(ctx, params)
	if err != nil {
		return nil, err
	}

	userOrders, err := s.prepareExchangeOrdersForExport(ctx, orders)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare exchange orders for export: %w", err)
	}
	ordersBuffer := new(bytes.Buffer)
	switch req.Format {
	case "csv":
		if err := gocsv.Marshal(userOrders, ordersBuffer); err != nil {
			return nil, fmt.Errorf("marshal orders: %w", err)
		}
	case "xlsx":
		excelFile := excelize.NewFile()
		if err := excelFile.SetSheetName(excelFile.GetSheetName(excelFile.GetActiveSheetIndex()), "Exchange orders"); err != nil {
			return nil, fmt.Errorf("set active sheet name: %w", err)
		}
		defer func() { _ = excelFile.Close() }()
		excelWriter, err := excel.NewWriter(excelFile)
		if err != nil {
			return nil, fmt.Errorf("create excel writer: %w", err)
		}
		if err := excelWriter.SetActiveSheetName("Exchange orders"); err != nil {
			return nil, fmt.Errorf("set active sheet name: %w", err)
		}
		if err := excelWriter.Marshal(&userOrders); err != nil {
			return nil, fmt.Errorf("marshal orders: %w", err)
		}
		if _, err := excelWriter.File.WriteTo(ordersBuffer); err != nil {
			return nil, fmt.Errorf("write to buffer: %w", err)
		}
	}

	return ordersBuffer, nil
}

func (s *Service) prepareExchangeOrdersForExport(ctx context.Context, orders []*repo_exchange_orders.GetExchangeOrdersByUserAndExchangeIDRow) ([]*UserExchangeOrderModel, error) {
	result := make([]*UserExchangeOrderModel, 0, len(orders))

	for _, order := range orders {
		exchange, err := s.st.Exchanges().GetByID(ctx, order.ExchangeID)
		if err != nil {
			return nil, fmt.Errorf("failed to get exchange: %w", err)
		}
		orderModel := &UserExchangeOrderModel{
			Exchange:        exchange.Name,
			ExchangeID:      order.ExchangeID.String(),
			Symbol:          order.Symbol,
			Amount:          order.Amount.String(),
			ExchangeOrderID: order.ExchangeOrderID.String,
			ClientOrderID:   order.ClientOrderID.String,
			FailReason:      order.FailReason.String,
			Side:            order.Side.String(),
			Status:          order.Status.String(),
			OrderCreatedAt:  order.OrderCreatedAt.Time,
		}
		if order.AmountUsd.Valid {
			orderModel.AmountUsd = order.AmountUsd.Decimal.String()
		}
		result = append(result, orderModel)
	}

	return result, nil
}

func (s *Service) processExchangeOrders(ctx context.Context) {
	orders, err := s.st.ExchangeOrders().GetByStatus(ctx, repo_exchange_orders.GetExchangeOrdersByStatus{
		Statuses: []models.ExchangeOrderStatus{models.ExchangeOrderStatusInProgress},
	})
	if err != nil {
		s.log.Error("failed to fetch orders", err)
	}
	for _, order := range orders {
		go func(order *models.ExchangeOrder) {
			err = repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
				ex, err := s.st.Exchanges(repos.WithTx(tx)).GetByID(ctx, order.ExchangeID)
				if err != nil {
					s.log.Error("failed to get exchange", err)
					return err
				}

				exClient, err := s.exManager.GetDriver(ctx, ex.Slug, order.UserID)
				if err != nil {
					s.log.Error("failed to get exchange client", err)
					return err
				}

				exOrder, err := exClient.GetOrderDetails(ctx, &models.GetOrderByIDParams{
					InstrumentID:    util.Pointer(order.Symbol),
					ExternalOrderID: util.Pointer(order.ExchangeOrderID.String),
					ClientOrderID:   util.Pointer(order.ClientOrderID.String),
					InternalOrder:   order,
				})
				if err != nil {
					s.log.Error("failed to get order details", err)
					return err
				}
				if exOrder == nil {
					err := fmt.Errorf("received nil order details")
					s.log.Error("received nil order details from exchange", err, "order_id", order.ID, "connection_hash", exClient.GetConnectionHash())
					return err
				}
				updateParams := repo_exchange_orders.UpdateParams{
					ID:     order.ID,
					Status: pgtype.Text{Valid: true, String: exOrder.State.String()},
				}

				if exOrder.State != models.ExchangeOrderStatusFailed {
					if exOrder.FailReason != "" {
						updateParams.FailReason = pgtype.Text{Valid: true, String: exOrder.FailReason}
					}
					if !exOrder.Amount.IsZero() {
						updateParams.Amount = decimal.NullDecimal{Valid: true, Decimal: exOrder.Amount}
					}
					if !exOrder.AmountUSD.IsZero() {
						updateParams.AmountUsd = decimal.NullDecimal{Valid: true, Decimal: exOrder.AmountUSD}
					}
				}

				return s.st.ExchangeOrders(repos.WithTx(tx)).Update(ctx, updateParams)
			})

			if err != nil {
				s.log.Error("failed to update exchange order", err)
			}
		}(order)
	}
}

func (s *Service) processExchangePairs(ctx context.Context) {
	exchangePairs, err := s.st.UserExchangePairs().GetAll(ctx)
	if err != nil {
		s.log.Error("failed to get user exchange pairs", err)
	}
	mPairs := make(map[uuid.UUID]map[uuid.UUID][]*models.UserExchangePair)
	for _, pair := range exchangePairs {
		if mPairs[pair.UserID] == nil {
			mPairs[pair.UserID] = make(map[uuid.UUID][]*models.UserExchangePair)
		}
		mPairs[pair.UserID][pair.ExchangeID] = append(mPairs[pair.UserID][pair.ExchangeID], pair)
	}

	for userID, pairs := range mPairs {
		for exID, p := range pairs {
			swapState, err := s.getSwapState(ctx, userID, exID)
			if err != nil {
				s.log.Errorw("failed to fetch exchange withdrawal state", err, "userID", userID)
				continue
			}
			if *swapState == models.ExchangeSwapStateDisabled {
				s.log.Debugw("skipping exchange swap", "userID", userID, "exchangeID", exID)
				continue
			}
			go func(userID uuid.UUID, userPairs []*models.UserExchangePair) {
				for _, pair := range userPairs {
					err := s.SubmitExchangeOrder(ctx, userID, pair)
					if err != nil { //nolint:nestif
						if errors.Is(err, exchangeclient.ErrInsufficientBalance) {
							continue
						}
						if errors.Is(err, exchangeclient.ErrSymbolTradingHalted) {
							continue
						}
						if errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ETIMEDOUT) {
							continue
						}
						if errors.Is(err, ErrSkipOrder) {
							continue
						}
						s.log.Error("failed to submit exchange order", err, "symbol", pair.Symbol)
						if err := s.suspendTransfers(ctx, userID); err != nil {
							s.log.Error("failed to suspend transfers after unknown state change on exchange", err, "userID", userID)
						}
					}
				}
			}(userID, p)
		}
	}
}

// Suspend transfers on unknown error
func (s *Service) suspendTransfers(ctx context.Context, userID uuid.UUID) error {
	user, err := s.st.Users().GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("fetch user: %w", err)
	}
	transferStatus, err := s.settingSvc.GetModelSetting(ctx, setting.TransfersStatus, setting.IModelSetting(user))
	if err != nil {
		return fmt.Errorf("fetch transfer status: %w", err)
	}
	if transferStatus.Value == setting.FlagValueEnabled {
		if eErr := s.settingSvc.SetModelSetting(ctx, setting.UpdateDTO{
			Name:  setting.TransfersStatus,
			Value: setting.TransferStatusSystemSuspended,
			Model: setting.IModelSetting(user),
		}); eErr != nil {
			return fmt.Errorf("failed to suspend user transfers: %w", eErr)
		}
	}
	return nil
}

func (s *Service) getSwapState(ctx context.Context, userID uuid.UUID, exchangeID uuid.UUID) (*models.ExchangeSwapState, error) {
	ue, err := s.st.UserExchanges().GetByUserAndExchangeID(ctx, repo_user_exchanges.GetByUserAndExchangeIDParams{
		UserID:     userID,
		ExchangeID: exchangeID,
	})
	if err != nil {
		return nil, err
	}
	return &ue.SwapState, nil
}

func (s *Service) GetExchangeChains(ctx context.Context, slug models.ExchangeSlug) ([]*models.ExchangeChainShort, error) {
	chains, err := s.st.ExchangeChains().GetEnabledCurrencies(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("fetch exchange chains: %w", err)
	}
	res := make([]*models.ExchangeChainShort, 0, len(chains))
	for _, chain := range chains {
		shortCurrency := &models.ExchangeChainShort{
			TickerLabel: chain.Name.String,
			Ticker:      chain.Ticker,
			Chain:       chain.Chain,
			CurrencyID:  chain.ID.String,
		}
		if chain.Blockchain != nil {
			shortCurrency.ChainLabel = chain.Blockchain.String()
		}
		res = append(res, shortCurrency)
	}

	return res, nil
}

func (s *Service) GetExchangePairs(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug) ([]*models.ExchangeSymbolDTO, error) {
	exClient, err := s.exManager.GetDriver(ctx, slug, userID)
	if err != nil {
		return nil, err
	}

	symbols, err := exClient.GetExchangeSymbols(ctx)
	if err != nil {
		return nil, err
	}

	return symbols, nil
}

func (s *Service) GetExchangeWithdrawalRules(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug) ([]*models.WithdrawalRulesDTO, error) {
	withdrawalRules, err := s.exRulesSvc.GetWithdrawalRules(ctx, slug, userID.String())
	if err != nil {
		return nil, err
	}
	return withdrawalRules, nil
}

func (s *Service) GetExchangeWithdrawalRule(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug, currencyID string) (*models.WithdrawalRulesDTO, error) {
	withdrawalRule, err := s.exRulesSvc.GetWithdrawalRule(ctx, slug, userID.String(), currencyID)
	if err != nil {
		return nil, err
	}
	return withdrawalRule, nil
}

func (s *Service) UpdateDepositAddresses(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug) ([]*models.DepositAddressDTO, error) {
	var dto []*models.DepositAddressDTO
	if err := repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		enabledCurrencies, err := s.st.ExchangeChains(repos.WithTx(tx)).GetEnabledCurrencies(ctx, slug)
		if err != nil {
			return fmt.Errorf("fetch enabled currencies: %w", err)
		}

		exClient, err := s.exManager.GetDriver(ctx, slug, userID)
		if err != nil {
			return fmt.Errorf("create exchange client: %w", err)
		}

		userExchange, err := s.st.Exchanges(repos.WithTx(tx)).GetExchangeBySlug(ctx, slug)
		if err != nil {
			return fmt.Errorf("fetch user exchange: %w", err)
		}

		err = s.st.ExchangeAddresses(repos.WithTx(tx)).DeleteByUser(ctx, repo_exchange_addresses.DeleteByUserParams{
			UserID:     userID,
			ExchangeID: userExchange.ID,
		})
		if err != nil {
			return fmt.Errorf("delete old addresses: %w", err)
		}

		newAddresses, err := s.fetchNewDepositAddresses(ctx, exClient, enabledCurrencies)
		if err != nil {
			return err
		}

		if err := s.batchInsertAddresses(ctx, tx, userID, userExchange.ID, newAddresses); err != nil {
			return fmt.Errorf("failed to batch insert new addresses")
		}

		dto = newAddresses
		return nil
	}); err != nil {
		return nil, fmt.Errorf("update deposit addresses: %w", err)
	}

	return dto, nil
}

func (s *Service) fetchNewDepositAddresses(ctx context.Context, exClient exchange_manager.IExchangeClient, currencies []*repo_exchange_chains.GetEnabledCurrenciesRow) ([]*models.DepositAddressDTO, error) {
	var addresses []*models.DepositAddressDTO
	for _, curr := range currencies {
		depositAddresses, err := exClient.GetDepositAddresses(ctx, curr.Ticker, curr.Chain)
		if err != nil {
			return nil, fmt.Errorf("fetch deposit addresses: %w", err)
		}
		for _, address := range depositAddresses {
			if address.Chain == curr.Chain && address.InternalCurrency == curr.Ticker && curr.ID.String == address.Currency { // TODO: add converter
				if address.Currency == "BCH.Bitcoincash" {
					if convertedBchAddress, found := strings.CutPrefix(address.Address, "bitcoincash:"); found {
						address.Address = convertedBchAddress
					}
				}
				addresses = append(addresses, address)
			}
		}
	}
	return addresses, nil
}

func (s *Service) batchInsertAddresses(ctx context.Context, tx pgx.Tx, userID uuid.UUID, exchangeID uuid.UUID, addresses []*models.DepositAddressDTO) error {
	params := make([]repo_exchange_addresses.BatchInsertExchangeAddressesParams, 0, len(addresses))
	for _, address := range addresses {
		params = append(params, repo_exchange_addresses.BatchInsertExchangeAddressesParams{
			ExchangeID:  exchangeID,
			Address:     address.Address,
			Chain:       address.Chain,
			Currency:    address.Currency,
			AddressType: models.DepositAddress.String(),
			UserID:      userID,
			CreateType:  "auto",
		})
	}
	batch := s.st.ExchangeAddresses(repos.WithTx(tx)).BatchInsertExchangeAddresses(ctx, params)
	defer func() {
		if err := batch.Close(); err != nil {
			s.log.Error("failed to close batch insert exchange addresses", err)
		}
	}()

	var err error
	batch.Exec(func(_ int, batchErr error) {
		if batchErr != nil {
			err = fmt.Errorf("batch insert exchange addresses: %w", err)
			return
		}
	})
	return err
}

func (s *Service) SubmitExchangeOrder(ctx context.Context, userID uuid.UUID, pair *models.UserExchangePair) error {
	balance := decimal.Zero

	slug, err := s.st.Exchanges().GetByID(ctx, pair.ExchangeID)
	if err != nil {
		return fmt.Errorf("fetch exchange: %w", err)
	}

	exClient, err := s.exManager.GetDriver(ctx, slug.Slug, userID)
	if err != nil {
		return err
	}

	rule, err := exClient.GetOrderRule(ctx, pair.Symbol)
	if err != nil {
		if errors.Is(err, exchangeclient.ErrSymbolTradingHalted) {
			return exchangeclient.ErrSymbolTradingHalted
		}
		return fmt.Errorf("fetch order rules: %w", err)
	}

	switch pair.Type {
	case models.OrderSideSell:
		amt, err := exClient.GetCurrencyBalance(ctx, rule.BaseCurrency)
		if err != nil {
			return fmt.Errorf("fetch currency balance: %w", err)
		}
		balance = balance.Add(*amt)
		minOrderAmount, err := decimal.NewFromString(rule.MinOrderAmount)
		if err != nil {
			return fmt.Errorf("parse min order amount: %w", err)
		}
		if balance.LessThanOrEqual(minOrderAmount) {
			return exchangeclient.ErrInsufficientBalance
		}
	case models.OrderSideBuy:
		amt, err := exClient.GetCurrencyBalance(ctx, rule.QuoteCurrency)
		if err != nil {
			return fmt.Errorf("fetch currency balance: %w", err)
		}
		balance = balance.Add(*amt)
		minOrderValue, err := decimal.NewFromString(rule.MinOrderValue)
		if err != nil {
			return fmt.Errorf("parse min order value: %w", err)
		}
		if balance.LessThanOrEqual(minOrderValue) {
			return exchangeclient.ErrInsufficientBalance
		}
	default:
		return fmt.Errorf("unknown order type: %s", pair.Type)
	}

	err = repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		orderRecord, err := s.createExchangeOrder(ctx, pair.ExchangeID, pair.UserID, pair.Symbol, pair.Type, balance, exClient.GetConnectionHash(), repos.WithTx(tx))
		if err != nil {
			return err
		}

		order, err := exClient.CreateSpotOrder(ctx, "", "", pair.Type.String(), pair.Symbol, nil, rule)
		updateParams := repo_exchange_orders.UpdateParams{
			ID:                     orderRecord.ID,
			ExchangeConnectionHash: pgtype.Text{Valid: true, String: exClient.GetConnectionHash()},
		}

		if err != nil {
			if errors.Is(err, ErrSkipOrder) {
				s.log.Debug("skipping order due to custom error being thrown", "userID", userID, "symbol", pair.Symbol)
				return ErrSkipOrder
			}
			if errors.Is(err, exchangeclient.ErrInsufficientBalance) {
				return exchangeclient.ErrInsufficientBalance
			}
			updateParams.Status = pgtype.Text{Valid: true, String: models.ExchangeOrderStatusFailed.String()}
			updateParams.FailReason = pgtype.Text{Valid: true, String: err.Error()}

			return s.st.ExchangeOrders(repos.WithTx(tx)).Update(ctx, updateParams)
		}

		updateParams.Status = pgtype.Text{Valid: true, String: models.ExchangeOrderStatusInProgress.String()}
		updateParams.Amount = decimal.NullDecimal{Valid: true, Decimal: order.Amount}
		if order.ClientOrderID != "" {
			updateParams.ClientOrderID = pgtype.Text{Valid: true, String: order.ClientOrderID}
		}
		if order.ExchangeOrderID != "" {
			updateParams.ExchangeOrderID = pgtype.Text{Valid: true, String: order.ExchangeOrderID}
		}
		return s.st.ExchangeOrders(repos.WithTx(tx)).Update(ctx, updateParams)
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) createExchangeOrder(ctx context.Context, exchangeID uuid.UUID, userID uuid.UUID, symbol string, side models.OrderSide, amount decimal.Decimal, connHash string, opts ...repos.Option) (*models.ExchangeOrder, error) {
	order, err := s.st.ExchangeOrders(opts...).Create(ctx, repo_exchange_orders.CreateParams{
		ExchangeID: exchangeID,
		UserID:     userID,
		Symbol:     symbol,
		Side:       side,
		Amount:     amount,
		OrderCreatedAt: pgtype.Timestamp{
			Valid: true,
			Time:  time.Now(),
		},
		ExchangeConnectionHash: pgtype.Text{Valid: true, String: connHash},
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *Service) GetUserExchangePairs(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug) ([]*models.UserExchangePair, error) {
	ex, err := s.st.Exchanges().GetExchangeBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("fetch exchange entity: %w", err)
	}

	pairs, err := s.st.UserExchangePairs().Find(ctx, repo_user_exchange_pairs.FindParams{
		UserID:     userID,
		ExchangeID: ex.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("fetch user exchange pairs: %w", err)
	}

	return pairs, nil
}

func (s *Service) GetCurrentExchangeBalance(ctx context.Context, user models.User) ([]*models.AccountBalanceDTO, error) {
	exClient, err := s.exManager.GetDefaultDriver(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("create exchange client: %w", err)
	}

	balances, err := exClient.GetAccountBalance(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch account balance: %w", err)
	}

	return balances, nil
}

func (s *Service) SetCurrentExchange(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug) error {
	err := repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		arg := repo_users.UpdateExchangeParams{
			ID:           userID,
			ExchangeSlug: &slug,
		}
		usr, err := s.st.Users(repos.WithTx(tx)).GetByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("fetch user: %w", err)
		}
		if usr.ExchangeSlug != nil && usr.ExchangeSlug.Valid() && *usr.ExchangeSlug == slug {
			arg.ExchangeSlug = nil
		}
		_, err = s.st.Users(repos.WithTx(tx)).UpdateExchange(ctx, arg)
		if err != nil {
			return fmt.Errorf("update user exchange: %w", err)
		}
		ex, err := s.st.Exchanges(repos.WithTx(tx)).GetExchangeBySlug(ctx, slug)
		if err != nil {
			return fmt.Errorf("fetch exchange: %w", err)
		}
		if arg.ExchangeSlug == nil { //nolint:nestif
			// If the exchange slug is nil, it means we are disabling this exchange
			// Only disable the exchange states - withdrawal settings keep their is_enabled state
			if _, err := s.st.UserExchanges(repos.WithTx(tx)).ChangeSwapState(ctx, repo_user_exchanges.ChangeSwapStateParams{
				UserID:     userID,
				ExchangeID: ex.ID,
				State:      models.ExchangeSwapStateDisabled,
			}); err != nil {
				return fmt.Errorf("disable exchange swap state for user: %w", err)
			}
			if _, err := s.st.UserExchanges(repos.WithTx(tx)).ChangeWithdrawalState(ctx, repo_user_exchanges.ChangeWithdrawalStateParams{
				UserID:     userID,
				ExchangeID: ex.ID,
				State:      models.ExchangeWithdrawalStateDisabled,
			}); err != nil {
				return fmt.Errorf("disable exchange withdrawal state for user: %w", err)
			}
		} else {
			// Enable this exchange - only enable the exchange states
			// Withdrawal settings will be restored automatically based on their saved is_enabled state
			if _, err := s.st.UserExchanges(repos.WithTx(tx)).ChangeSwapState(ctx, repo_user_exchanges.ChangeSwapStateParams{
				UserID:     userID,
				ExchangeID: ex.ID,
				State:      models.ExchangeSwapStateEnabled,
			}); err != nil {
				return fmt.Errorf("enable exchange swap state for user: %w", err)
			}
			if _, err := s.st.UserExchanges(repos.WithTx(tx)).ChangeWithdrawalState(ctx, repo_user_exchanges.ChangeWithdrawalStateParams{
				UserID:     userID,
				ExchangeID: ex.ID,
				State:      models.ExchangeWithdrawalStateEnabled,
			}); err != nil {
				return fmt.Errorf("enable exchange withdrawal state for user: %w", err)
			}
		}
		// Disable other exchanges for the user (maintains single-exchange behavior)
		if err := s.st.UserExchanges(repos.WithTx(tx)).DisableAllPerUserExceptExchange(ctx, userID, ex.ID); err != nil {
			return fmt.Errorf("disable other exchanges: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("set current exchange: %w", err)
	}
	return nil
}

var _ IExchangeService = (*Service)(nil)

func NewService(
	log logger.Logger,
	st storage.IStorage,
	exchangeManager exchange_manager.IExchangeManager,
	exRulesSvc exchange_rules.IExchangeRules,
	settingSvc setting.ISettingService,
) IExchangeService {
	return &Service{
		st:         st,
		exManager:  exchangeManager,
		log:        log,
		exRulesSvc: exRulesSvc,
		settingSvc: settingSvc,
	}
}

func (s *Service) GetExchangeBalance(ctx context.Context, slug models.ExchangeSlug, user models.User) ([]*models.AccountBalanceDTO, error) {
	exClient, err := s.exManager.GetDriver(ctx, slug, user.ID)
	if err != nil {
		return nil, fmt.Errorf("create exchange client: %w", err)
	}

	balances, err := exClient.GetAccountBalance(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch account balance: %w", err)
	}

	return balances, nil
}

func (s *Service) TestConnection(ctx context.Context, user models.User, slug models.ExchangeSlug) error {
	exClient, err := s.exManager.GetDriver(ctx, slug, user.ID)
	if err != nil {
		return err
	}
	if err = exClient.TestConnection(ctx); err != nil {
		if errors.Is(err, exchangeclient.ErrInvalidAPICredentials) {
			return exchangeclient.ErrInvalidAPICredentials
		}
		if errors.Is(err, exchangeclient.ErrInvalidIPAddress) {
			return exchangeclient.ErrInvalidIPAddress
		}
		if errors.Is(err, exchangeclient.ErrIncorrectAPIPermissions) {
			return exchangeclient.ErrIncorrectAPIPermissions
		}
		return err
	}
	return nil
}

func (s *Service) TestConnectionRaw(ctx context.Context, slug models.ExchangeSlug, apiKey, secretKey, passphrase string) error {
	exClient, err := s.exManager.CreateDriver(ctx, slug, apiKey, secretKey, passphrase)
	if err != nil {
		return fmt.Errorf("create exchange client: %w", err)
	}
	if err = exClient.TestConnection(ctx); err != nil {
		if errors.Is(err, exchangeclient.ErrInvalidAPICredentials) {
			return exchangeclient.ErrInvalidAPICredentials
		}
		if errors.Is(err, exchangeclient.ErrInvalidIPAddress) {
			return exchangeclient.ErrInvalidIPAddress
		}
		if errors.Is(err, exchangeclient.ErrIncorrectAPIPermissions) {
			return exchangeclient.ErrIncorrectAPIPermissions
		}
		return err
	}
	return nil
}

func (s *Service) GetAvailableExchangesList(ctx context.Context, userID uuid.UUID) (*ActiveExchangeListDTO, error) {
	r := &ActiveExchangeListDTO{}

	exchangeData, err := s.st.Exchanges().GetAllActiveWithUserKeys(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetch exchanges: %w", err)
	}

	usr, err := s.st.Users().GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetch user: %w", err)
	}

	resMap := make(map[models.ExchangeSlug]*WithExchangeKeysDTO)
	for _, exchange := range exchangeData {
		var exchangeKeysData []KeysExchangeDTO
		existing, ok := resMap[exchange.Slug]
		if !ok {
			exchangeKeysData = make([]KeysExchangeDTO, 0)
			existing = &WithExchangeKeysDTO{
				Exchange: exchange.Name,
				Slug:     exchange.Slug,
				Keys:     exchangeKeysData,
			}
			resMap[exchange.Slug] = existing
		}

		var keyVal *string
		if exchange.Value.Valid {
			keyVal = &exchange.Value.String
		}

		existing.Keys = append(existing.Keys, KeysExchangeDTO{
			Name:  string(exchange.KeyName),
			Title: exchange.KeyTitle.String,
			Value: keyVal,
		})
		if exchange.ExchangeConnectedAt.Valid {
			existing.ExchangeConnectedAt = exchange.ExchangeConnectedAt.Time
		}
	}

	for _, dto := range resMap {
		r.Exchanges = append(r.Exchanges, *dto)
	}

	// Only set current_exchange if the exchange has valid keys
	if usr.ExchangeSlug != nil && usr.ExchangeSlug.Valid() { //nolint:nestif
		// Check if the current exchange has keys
		if exchangeDTO, exists := resMap[*usr.ExchangeSlug]; exists && len(exchangeDTO.Keys) > 0 {
			exchange, err := s.st.Exchanges().GetExchangeBySlug(ctx, *usr.ExchangeSlug)
			if err != nil {
				return nil, fmt.Errorf("fetch exchange by slug: %w", err)
			}
			r.CurrentExchange = util.Pointer(usr.ExchangeSlug.String())
			exInfo, err := s.st.UserExchanges().GetByUserAndExchangeID(ctx, repo_user_exchanges.GetByUserAndExchangeIDParams{
				UserID:     userID,
				ExchangeID: exchange.ID,
			})
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("fetch user exchange info: %w", err)
			}
			if exInfo != nil {
				r.SwapState = util.Pointer(exInfo.SwapState.String())
				r.WithdrawalState = util.Pointer(exInfo.WithdrawalState.String())
			}
		}
	}

	return r, nil
}

func (s *Service) SetExchangeKeys(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug, keysData map[models.ExchangeKeyName]*string) error {
	exKeys, err := s.st.Exchanges().GetExchangeKeysBySlug(ctx, repo_exchanges.GetExchangeKeysBySlugParams{
		UserID: userID,
		Slug:   slug,
	})
	if err != nil {
		return fmt.Errorf("fetch exchange keys: %w", err)
	}
	if len(exKeys) == 0 {
		return ErrExchangeNotFound
	}

	for _, exKey := range exKeys {
		val, ok := keysData[exKey.Name]
		if ok {
			updateKeyErr := s.updateKeyByValue(ctx, userID, exKey.ExchangeKeyID, val)
			if updateKeyErr != nil {
				return fmt.Errorf("update exchange key: %w", updateKeyErr)
			}
		}
	}

	return nil
}

func (s *Service) updateKeyByValue(ctx context.Context, userID uuid.UUID, exKeyID uuid.UUID, newValue *string) error {
	err := repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		if newValue != nil {
			_, err := s.st.ExchangeUserKeys(repos.WithTx(tx)).CreateOrUpdateUserKey(ctx, repo_exchange_user_keys.CreateOrUpdateUserKeyParams{
				UserID:        userID,
				ExchangeKeyID: exKeyID,
				Value:         *newValue,
			})
			if err != nil {
				return fmt.Errorf("create or update user key: %w", err)
			}
			if err := s.st.UserExchanges(repos.WithTx(tx)).CreateByUserID(ctx, userID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("update exchange key by value: %w", err)
	}

	return nil
}

func (s *Service) UpdateUserExchangePairs(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug, request *exchange_request.UpdateExchangePairsRequest) error {
	ex, err := s.st.Exchanges().GetExchangeBySlug(ctx, slug)
	if err != nil {
		return err
	}

	err = repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := s.st.UserExchangePairs(repos.WithTx(tx)).Delete(ctx, repo_user_exchange_pairs.DeleteParams{
			UserID:     userID,
			ExchangeID: ex.ID,
		}); err != nil {
			return err
		}
		if err := s.updateUserExchangePairsTx(ctx, userID, ex.ID, request.Pairs, tx); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("update user exchange pairs: %w", err)
	}

	return nil
}

func (s *Service) GetDepositExchangeAddresses(ctx context.Context, userID uuid.UUID) ([]models.ExchangeGroup, error) {
	deposits, err := s.st.ExchangeAddresses().GetAllDepositAddress(ctx, userID)
	if err != nil {
		return nil, err
	}

	groupedData := make(map[string]models.ExchangeGroup)
	for _, deposit := range deposits {
		key := deposit.Slug
		if _, exists := groupedData[key]; !exists {
			groupedData[key] = models.ExchangeGroup{
				Slug:      deposit.Slug,
				Name:      deposit.Name,
				Addresses: []models.DepositAddressDTO{},
			}
		}
		group := groupedData[key]
		group.Addresses = append(group.Addresses, models.DepositAddressDTO{
			Address:          deposit.Address,
			Currency:         deposit.Currency,
			Chain:            deposit.Chain,
			AddressType:      models.AddressType(deposit.AddressType),
			InternalCurrency: deposit.Ticker,
		})

		groupedData[key] = group
	}

	result := make([]models.ExchangeGroup, 0, len(groupedData))
	for _, group := range groupedData {
		result = append(result, group)
	}

	return result, nil
}

func (s *Service) updateUserExchangePairsTx(ctx context.Context, userID uuid.UUID, exchangeID uuid.UUID, pairs []exchange_request.ExchangePair, tx pgx.Tx) error {
	params := make([]repo_user_exchange_pairs.UpdatePairsParams, 0, len(pairs))
	for _, pair := range pairs {
		params = append(params, repo_user_exchange_pairs.UpdatePairsParams{
			UserID:       userID,
			ExchangeID:   exchangeID,
			CurrencyFrom: pair.BaseSymbol,
			CurrencyTo:   pair.QuoteSymbol,
			Symbol:       pair.Symbol,
			Type:         models.OrderSide(pair.Type),
		})
	}

	errChan := make(chan error, len(params))
	result := s.st.UserExchangePairs(repos.WithTx(tx)).UpdatePairs(ctx, params)

	wg := sync.WaitGroup{}
	wg.Add(len(params))

	result.Exec(func(_ int, err error) {
		defer wg.Done()
		errChan <- err
	})

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return fmt.Errorf("batch create user exchange pairs: %w", err)
		}
	}

	return nil
}
