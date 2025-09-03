package notification_settings

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_notifications"

	"github.com/jackc/pgx/v5"
)

type INotificationSettings interface {
	AvailableListByUser(ctx context.Context, user *models.User) ([]UserNotification, error)
	UpdateList(ctx context.Context, user *models.User, dto []UpdateDTO) error
}

type Service struct {
	st storage.IStorage
}

func New(st storage.IStorage) *Service {
	return &Service{
		st: st,
	}
}

func (s *Service) AvailableListByUser(ctx context.Context, user *models.User) ([]UserNotification, error) {
	list, err := s.st.UserNotifications().GetUserListWithCategory(ctx, user.ID)
	if err != nil {
		return []UserNotification{}, fmt.Errorf("get notifications list: %w", err)
	}

	res := make([]UserNotification, 0, len(list))
	for _, notification := range list {
		res = append(res, UserNotification{
			ID:           notification.ID,
			Name:         notification.Type.String(),
			Category:     notification.Category,
			EmailEnabled: notification.EmailEnabled,
			TgEnabled:    notification.TgEnabled,
		})
	}

	return res, nil
}

func (s *Service) UpdateList(ctx context.Context, user *models.User, dto []UpdateDTO) error {
	return repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, func(pgx.Tx) error {
		updateSettingsParams := make([]repo_user_notifications.CreateOrUpdateParams, 0, len(dto))
		for _, val := range dto {
			notification, err := s.st.Notifications().GetByID(ctx, val.ID)
			if err != nil {
				return fmt.Errorf("get notification settings: %w", err)
			}

			if val.IsFullDisable() && !notification.Category.IsFullDisableAllowed() {
				return errors.New("full disable system notification is not allowed")
			}

			updateSettingsParams = append(updateSettingsParams, repo_user_notifications.CreateOrUpdateParams{
				UserID:         user.ID,
				NotificationID: val.ID,
				EmailEnabled:   val.EmailEnabled,
				TgEnabled:      val.TgEnabled,
			})
		}

		errChan := make(chan error, len(updateSettingsParams))
		result := s.st.UserNotifications().CreateOrUpdate(ctx, updateSettingsParams)

		wg := sync.WaitGroup{}
		wg.Add(len(updateSettingsParams))

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
				return fmt.Errorf("batch update notifications error: %w", err)
			}
		}

		return nil
	})
}
