package refund

import (
	"context"
	"errors"
	"testing"

	"github.com/dv-net/dv-merchant/internal/constants"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_blocked_transactions"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_refund_requests"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transactions"
	"github.com/dv-net/dv-merchant/pkg/logger"
	mxlogger "github.com/dv-net/mx/logger"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSenderAddress = "0x742d35cc6634c0532925a3b844bc454e4438f44e"
	testOtherAddress  = "0x8ba1f109551bd432803012645ac136ddd64dba72"
)

// the fakes embed the interfaces: only the methods used by the refund claim are implemented

type fakeStorage struct {
	storage.IStorage
	blocked      *fakeBlockedTransactions
	transactions *fakeTransactions
	refunds      *fakeRefundRequests
}

func (f *fakeStorage) BlockedTransactions(...repos.Option) repo_blocked_transactions.Querier {
	return f.blocked
}

func (f *fakeStorage) Transactions(...repos.Option) repo_transactions.ICustomQuerier {
	return f.transactions
}

func (f *fakeStorage) RefundRequests(...repos.Option) repo_refund_requests.ICustomQuerier {
	return f.refunds
}

type fakeBlockedTransactions struct {
	repo_blocked_transactions.Querier
	btx *models.BlockedTransaction
}

func (f *fakeBlockedTransactions) GetById(_ context.Context, id uuid.UUID) (*models.BlockedTransaction, error) {
	if f.btx == nil || f.btx.ID != id {
		return nil, pgx.ErrNoRows
	}
	return f.btx, nil
}

func (f *fakeBlockedTransactions) GetUnclaimedByWalletID(_ context.Context, walletID uuid.UUID) ([]*repo_blocked_transactions.GetUnclaimedByWalletIDRow, error) {
	if f.btx == nil || f.btx.WalletID != walletID {
		return nil, nil
	}
	// two deposits of the wallet: the owner setting is read once
	rows := make([]*repo_blocked_transactions.GetUnclaimedByWalletIDRow, 0, 2)
	for range 2 {
		rows = append(rows, &repo_blocked_transactions.GetUnclaimedByWalletIDRow{
			ID:       uuid.New(),
			UserID:   f.btx.UserID,
			WalletID: f.btx.WalletID,
		})
	}
	return rows, nil
}

type fakeTransactions struct {
	repo_transactions.ICustomQuerier
	tx *models.Transaction
}

func (f *fakeTransactions) GetById(context.Context, uuid.UUID) (*models.Transaction, error) {
	return f.tx, nil
}

type fakeRefundRequests struct {
	repo_refund_requests.ICustomQuerier
	existing *models.RefundRequest
	created  *repo_refund_requests.CreateParams
}

func (f *fakeRefundRequests) GetByBlockedTransactionID(context.Context, uuid.UUID) (*models.RefundRequest, error) {
	if f.existing == nil {
		return nil, pgx.ErrNoRows
	}
	return f.existing, nil
}

var errCreateStopped = errors.New("create stopped by the test")

// Create records the params and fails, so the test does not reach sending to processing
func (f *fakeRefundRequests) Create(_ context.Context, arg repo_refund_requests.CreateParams) (*models.RefundRequest, error) {
	f.created = &arg
	return nil, errCreateStopped
}

type fakeExecutor struct {
	refundExecutor
	toSender     bool
	checkedDests []string
}

func (f *fakeExecutor) CheckRefund(_ context.Context, ref *models.RefundRequest) (*processing.CheckRefundResult, error) {
	f.checkedDests = append(f.checkedDests, ref.DestinationAddress)
	return &processing.CheckRefundResult{ToSender: f.toSender}, nil
}

type fakeSettings struct {
	value *string // nil - the setting is not set
	users []uuid.UUID
}

func (f *fakeSettings) GetModelSetting(_ context.Context, name string, model setting.IModelSetting) (*models.Setting, error) {
	if name != setting.RefundAuthMode {
		return nil, errors.New("unexpected setting " + name)
	}
	f.users = append(f.users, model.ModelID().UUID)
	if f.value == nil {
		return nil, pgx.ErrNoRows
	}
	return &models.Setting{Name: name, Value: *f.value}, nil
}

type toSenderFixture struct {
	svc      *Service
	btx      *models.BlockedTransaction
	tx       *models.Transaction
	refunds  *fakeRefundRequests
	executor *fakeExecutor
	settings *fakeSettings
}

func newToSenderFixture(mode *string, processingToSender bool) *toSenderFixture {
	btx := &models.BlockedTransaction{
		ID:            uuid.New(),
		UserID:        uuid.New(),
		StoreID:       uuid.New(),
		WalletID:      uuid.New(),
		TransactionID: uuid.New(),
	}
	tx := &models.Transaction{
		ID:          btx.TransactionID,
		Blockchain:  models.BlockchainEthereum,
		CurrencyID:  "ETH.Ethereum",
		Amount:      decimal.NewFromInt(1),
		FromAddress: testSenderAddress,
	}

	f := &toSenderFixture{
		btx:      btx,
		tx:       tx,
		refunds:  &fakeRefundRequests{},
		executor: &fakeExecutor{toSender: processingToSender},
		settings: &fakeSettings{value: mode},
	}
	st := &fakeStorage{
		blocked:      &fakeBlockedTransactions{btx: btx},
		transactions: &fakeTransactions{tx: tx},
		refunds:      f.refunds,
	}
	f.svc = NewService(logger.New("test", mxlogger.Config{Level: "fatal"}), st, f.executor, nil, f.settings)

	return f
}

func (f *toSenderFixture) dto() ToSenderDTO {
	return ToSenderDTO{WalletID: f.btx.WalletID, BlockedTransactionID: f.btx.ID, DestinationAddress: testSenderAddress}
}

func strPtr(s string) *string { return &s }

func TestRefundToSenderWithoutAuth(t *testing.T) {
	tests := []struct {
		name string
		mode *string
		want bool
	}{
		{"not set", nil, false},
		{"always", strPtr(setting.RefundAuthModeAlways), false},
		{"custom address only", strPtr(setting.RefundAuthModeCustomAddressOnly), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newToSenderFixture(tt.mode, true)
			userID := uuid.New()

			got, err := f.svc.RefundToSenderWithoutAuth(context.Background(), userID)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, []uuid.UUID{userID}, f.settings.users)
		})
	}
}

func TestEstimateRefundToSender(t *testing.T) {
	ctx := context.Background()

	t.Run("auth required when the setting is not set", func(t *testing.T) {
		f := newToSenderFixture(nil, true)

		_, err := f.svc.EstimateRefundToSender(ctx, f.dto())
		require.ErrorIs(t, err, ErrRefundAuthRequired)
		assert.Empty(t, f.executor.checkedDests, "processing must not be called")
	})

	t.Run("auth required when the owner requires it always", func(t *testing.T) {
		f := newToSenderFixture(strPtr(setting.RefundAuthModeAlways), true)

		_, err := f.svc.EstimateRefundToSender(ctx, f.dto())
		require.ErrorIs(t, err, ErrRefundAuthRequired)
		assert.Empty(t, f.executor.checkedDests)
	})

	t.Run("auth required is not disclosed before the wallet check", func(t *testing.T) {
		f := newToSenderFixture(nil, true)
		dto := f.dto()
		dto.WalletID = uuid.New()

		_, err := f.svc.EstimateRefundToSender(ctx, dto)
		require.ErrorIs(t, err, ErrRefundBlockedTxNotFound)
	})

	t.Run("setting of the store owner is checked", func(t *testing.T) {
		f := newToSenderFixture(strPtr(setting.RefundAuthModeCustomAddressOnly), true)

		_, err := f.svc.EstimateRefundToSender(ctx, f.dto())
		require.NoError(t, err)
		assert.Equal(t, []uuid.UUID{f.btx.UserID}, f.settings.users)
	})

	t.Run("destination is the payer address confirmed by processing", func(t *testing.T) {
		f := newToSenderFixture(strPtr(setting.RefundAuthModeCustomAddressOnly), true)

		res, err := f.svc.EstimateRefundToSender(ctx, f.dto())
		require.NoError(t, err)
		assert.Equal(t, testSenderAddress, res.DestinationAddress)
		assert.Equal(t, []string{testSenderAddress}, f.executor.checkedDests)
		assert.False(t, res.RequiresReview)
	})

	t.Run("stored sender address is not used", func(t *testing.T) {
		// btc like deposits have no stored sender, another input of the deposit may be refunded
		f := newToSenderFixture(strPtr(setting.RefundAuthModeCustomAddressOnly), true)
		f.tx.FromAddress = ""
		dto := f.dto()
		dto.DestinationAddress = testOtherAddress

		res, err := f.svc.EstimateRefundToSender(ctx, dto)
		require.NoError(t, err)
		assert.Equal(t, testOtherAddress, res.DestinationAddress)
		assert.Equal(t, []string{testOtherAddress}, f.executor.checkedDests)
	})

	t.Run("address is not a sender of the deposit", func(t *testing.T) {
		f := newToSenderFixture(strPtr(setting.RefundAuthModeCustomAddressOnly), false)

		_, err := f.svc.EstimateRefundToSender(ctx, f.dto())
		require.ErrorIs(t, err, ErrRefundNotSender)
	})

	t.Run("invalid address is rejected before processing", func(t *testing.T) {
		f := newToSenderFixture(strPtr(setting.RefundAuthModeCustomAddressOnly), true)
		dto := f.dto()
		dto.DestinationAddress = "not-an-address"

		_, err := f.svc.EstimateRefundToSender(ctx, dto)
		require.ErrorIs(t, err, ErrRefundInvalidAddress)
		assert.Empty(t, f.executor.checkedDests)
	})

	t.Run("already requested", func(t *testing.T) {
		f := newToSenderFixture(strPtr(setting.RefundAuthModeCustomAddressOnly), true)
		f.refunds.existing = &models.RefundRequest{ID: uuid.New()}

		_, err := f.svc.EstimateRefundToSender(ctx, f.dto())
		require.ErrorIs(t, err, ErrRefundAlreadyRequested)
	})
}

func TestCreateRefundToSender(t *testing.T) {
	ctx := context.Background()

	t.Run("created to the sender without email", func(t *testing.T) {
		f := newToSenderFixture(strPtr(setting.RefundAuthModeCustomAddressOnly), true)

		_, err := f.svc.CreateRefundToSender(ctx, f.dto())
		require.ErrorIs(t, err, errCreateStopped)

		created := f.refunds.created
		require.NotNil(t, created)
		assert.Equal(t, testSenderAddress, created.DestinationAddress)
		assert.False(t, created.Email.Valid)
		assert.Equal(t, constants.RefundAuthMethodNone, created.AuthMethod)
		assert.Equal(t, f.btx.ID, created.BlockedTransactionID)
		assert.Equal(t, f.btx.WalletID, created.WalletID)
		assert.Equal(t, f.btx.StoreID, created.StoreID)
	})

	t.Run("not created to an address which is not a sender", func(t *testing.T) {
		f := newToSenderFixture(strPtr(setting.RefundAuthModeCustomAddressOnly), false)
		dto := f.dto()
		dto.DestinationAddress = testOtherAddress

		_, err := f.svc.CreateRefundToSender(ctx, dto)
		require.ErrorIs(t, err, ErrRefundNotSender)
		assert.Nil(t, f.refunds.created)
	})

	t.Run("not created when auth is required", func(t *testing.T) {
		f := newToSenderFixture(strPtr(setting.RefundAuthModeAlways), true)

		_, err := f.svc.CreateRefundToSender(ctx, f.dto())
		require.ErrorIs(t, err, ErrRefundAuthRequired)
		assert.Nil(t, f.refunds.created)
	})
}

func TestCreateRefundWithEmail(t *testing.T) {
	// the verified payer still chooses any address, the setting does not matter
	f := newToSenderFixture(nil, false)

	_, err := f.svc.CreateRefund(context.Background(), CreateRefundDTO{
		WalletID:             f.btx.WalletID,
		BlockedTransactionID: f.btx.ID,
		DestinationAddress:   testOtherAddress,
		Email:                "payer@example.com",
	})
	require.ErrorIs(t, err, errCreateStopped)

	created := f.refunds.created
	require.NotNil(t, created)
	assert.Equal(t, testOtherAddress, created.DestinationAddress)
	assert.Equal(t, "payer@example.com", created.Email.String)
	assert.True(t, created.Email.Valid)
	assert.Equal(t, constants.RefundAuthMethodEmail, created.AuthMethod)
	assert.Empty(t, f.settings.users, "the setting is not read for a verified payer")
}

func TestGetUnclaimedRefundToSenderWithoutAuth(t *testing.T) {
	tests := []struct {
		name string
		mode *string
		want bool
	}{
		{"not set", nil, false},
		{"always", strPtr(setting.RefundAuthModeAlways), false},
		{"custom address only", strPtr(setting.RefundAuthModeCustomAddressOnly), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newToSenderFixture(tt.mode, true)

			items, err := f.svc.GetUnclaimed(context.Background(), f.btx.WalletID)
			require.NoError(t, err)
			require.Len(t, items, 2)
			for _, item := range items {
				assert.Equal(t, tt.want, item.RefundToSenderWithoutAuth)
			}
			assert.Equal(t, []uuid.UUID{f.btx.UserID}, f.settings.users, "the setting is read once per owner")
		})
	}
}
