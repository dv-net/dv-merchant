package exchange_withdrawal

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/dv-net/dv-merchant/internal/models"
	krakensvc "github.com/dv-net/dv-merchant/internal/service/exchange/kraken"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_withdrawal_history"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_users"
)

// Fakes implement only what handleKrakenWithdrawal touches; other calls panic.
type fakeStorage struct {
	storage.IStorage
	users   *fakeUsers
	history *fakeHistory
}

func (f *fakeStorage) Users(...repos.Option) repo_users.ICustomQuerier { return f.users }
func (f *fakeStorage) ExchangeWithdrawalHistory(...repos.Option) repo_exchange_withdrawal_history.ICustomQuerier {
	return f.history
}

type fakeUsers struct {
	repo_users.ICustomQuerier
	user *models.User
}

func (f *fakeUsers) GetByID(context.Context, uuid.UUID) (*models.User, error) { return f.user, nil }

type fakeHistory struct {
	repo_exchange_withdrawal_history.ICustomQuerier
	updates []repo_exchange_withdrawal_history.UpdateParams
}

func (f *fakeHistory) Update(_ context.Context, arg repo_exchange_withdrawal_history.UpdateParams) error {
	f.updates = append(f.updates, arg)
	return nil
}

type fakeSettings struct {
	setting.ISettingService
	transferState string
	set           []setting.UpdateDTO
}

func (f *fakeSettings) GetModelSetting(context.Context, string, setting.IModelSetting) (*models.Setting, error) {
	return &models.Setting{Value: f.transferState}, nil
}

func (f *fakeSettings) SetModelSetting(_ context.Context, dto setting.UpdateDTO, _ ...repos.Option) error {
	f.set = append(f.set, dto)
	return nil
}

func TestHandleKrakenWithdrawal(t *testing.T) {
	cases := []struct {
		name          string
		record        models.WithdrawalStatusDTO
		transferState string
		wantStatus    models.WithdrawalHistoryStatus
		wantTxID      string
		wantReason    string
		wantSetting   string // empty means the transfers setting must not change
	}{
		{
			name:          "in progress",
			record:        models.WithdrawalStatusDTO{Status: models.WithdrawalHistoryStatusInProgress.String()},
			transferState: setting.FlagValueEnabled,
			wantStatus:    models.WithdrawalHistoryStatusInProgress,
		},
		{
			name: "completed",
			record: models.WithdrawalStatusDTO{
				Status: models.WithdrawalHistoryStatusCompleted.String(), TxHash: "0xhash", NativeAmount: decimal.NewFromInt(99),
			},
			transferState: setting.FlagValueEnabled,
			wantStatus:    models.WithdrawalHistoryStatusCompleted,
			wantTxID:      "0xhash",
		},
		{
			name: "completed re-enables transfers suspended by system",
			record: models.WithdrawalStatusDTO{
				Status: models.WithdrawalHistoryStatusCompleted.String(), TxHash: "0xhash", NativeAmount: decimal.NewFromInt(1),
			},
			transferState: setting.TransferStatusSystemSuspended,
			wantStatus:    models.WithdrawalHistoryStatusCompleted,
			wantTxID:      "0xhash",
			wantSetting:   setting.FlagValueEnabled,
		},
		{
			name:          "failed keeps transfers enabled",
			record:        models.WithdrawalStatusDTO{Status: models.WithdrawalHistoryStatusFailed.String(), ErrorMessage: "Failure"},
			transferState: setting.FlagValueEnabled,
			wantStatus:    models.WithdrawalHistoryStatusFailed,
		},
		{
			name: "on hold keeps transfers enabled",
			record: models.WithdrawalStatusDTO{
				Status: models.WithdrawalHistoryStatusRecovery.String(), ErrorMessage: krakensvc.ManualVerificationRequired,
			},
			transferState: setting.FlagValueEnabled,
			wantStatus:    models.WithdrawalHistoryStatusRecovery,
			wantReason:    krakensvc.ManualVerificationRequired,
		},
		{
			name:          "unknown status goes to recovery",
			record:        models.WithdrawalStatusDTO{Status: "something-new"},
			transferState: setting.FlagValueEnabled,
			wantStatus:    models.WithdrawalHistoryStatusRecovery,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			history := &fakeHistory{}
			settings := &fakeSettings{transferState: tc.transferState}
			svc := &Service{
				st: &fakeStorage{
					users:   &fakeUsers{user: &models.User{ID: uuid.New()}},
					history: history,
				},
				settingSvc: settings,
			}
			order := &models.ExchangeWithdrawalHistory{ID: uuid.New(), UserID: uuid.New()}

			require.NoError(t, svc.handleKrakenWithdrawal(t.Context(), &tc.record, order, nil))

			require.Len(t, history.updates, 1)
			upd := history.updates[0]
			require.Equal(t, order.ID, upd.ID)
			require.Equal(t, tc.wantStatus.String(), upd.Status.String)

			require.Equal(t, tc.wantTxID != "", upd.Txid.Valid)
			require.Equal(t, tc.wantTxID, upd.Txid.String)
			if tc.wantStatus == models.WithdrawalHistoryStatusCompleted {
				require.True(t, upd.NativeAmount.Valid)
				require.True(t, tc.record.NativeAmount.Equal(upd.NativeAmount.Decimal))
			}

			require.Equal(t, tc.wantReason != "", upd.FailReason.Valid)
			require.Equal(t, tc.wantReason, upd.FailReason.String)

			if tc.wantSetting == "" {
				require.Empty(t, settings.set)
			} else {
				require.Len(t, settings.set, 1)
				require.Equal(t, tc.wantSetting, settings.set[0].Value)
			}
		})
	}
}
