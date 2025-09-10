package notify

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/notification_request"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/permission"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_notification_send_history"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_notification_send_queue"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_notifications"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	DefaultMaxRetries = 2
)

type INotificationSender interface {
	Send(context.Context, models.NotificationSendQueue) SendResult
}

type INotificationService interface {
	Run(ctx context.Context)
	SendUser(ctx context.Context, notificationType models.NotificationType, user *models.User, payload INotificationBody, args *models.NotificationArgs)
	SendSystemEmail(ctx context.Context, notificationType models.NotificationType, email string, payload INotificationBody, args *models.NotificationArgs)
	GetHistoryByParams(ctx context.Context, usr *models.User, params *notification_request.GetNotificationHistoryRequest) (*storecmn.FindResponseWithFullPagination[*models.NotificationSendHistory], error)
	GetTypesList(ctx context.Context) ([]NotificationType, error)
}

type NotificationType struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type SendResult struct {
	IsExternal bool
	IsSuccess  bool
	Sender     string
	SentBody   []byte
}

type Service struct {
	logger            logger.Logger
	storage           storage.IStorage
	settingsSvc       setting.ISettingService
	permissionService permission.IPermission

	sender INotificationSender

	maxRetries int32
	inUse      atomic.Bool
}

func New(
	logger logger.Logger,
	storage storage.IStorage,
	settingsSvc setting.ISettingService,
	sender INotificationSender,
	permSvc permission.IPermission,
) INotificationService {
	svc := &Service{
		logger:            logger,
		storage:           storage,
		settingsSvc:       settingsSvc,
		permissionService: permSvc,
		maxRetries:        DefaultMaxRetries,
		sender:            sender,
	}

	svc.logger.Info("Started notification service")

	return svc
}

func (svc *Service) Run(ctx context.Context) {
	svc.logger.Info("Started to handle notification queue")

	cleanupTicker := time.NewTicker(time.Hour)
	senderTicker := time.NewTicker(time.Second * 10)
	reminderTicker := time.NewTicker(time.Hour * 8)
	defer cleanupTicker.Stop()
	defer senderTicker.Stop()
	defer reminderTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-cleanupTicker.C:
			go func() {
				if err := svc.runCleanHistoryJob(ctx); err != nil {
					svc.logger.Error("failed to run notification cleanup job", err)
				}
			}()
		case <-senderTicker.C:
			go svc.processQueue(ctx)
		case <-reminderTicker.C:
			go svc.processReminds(ctx)
		}
	}
}

// SendUser sends notification by defined user rules
func (svc *Service) SendUser(
	ctx context.Context,
	notificationType models.NotificationType,
	user *models.User,
	payload INotificationBody,
	args *models.NotificationArgs,
) {
	params, err := svc.storage.UserNotifications().GetUserNotificationChannels(ctx,
		repo_user_notifications.GetUserNotificationChannelsParams{
			UserID: user.ID,
			Type:   notificationType,
		},
	)

	if err != nil {
		svc.logger.Error("failed to fetch notification channels", err)
		return
	}

	deliveryChannels := make([]models.DeliveryChannel, 0, 2)
	if params.EmailEnabled {
		deliveryChannels = append(deliveryChannels, models.EmailDeliveryChannel)
	}
	if params.TgEnabled {
		deliveryChannels = append(deliveryChannels, models.TelegramDeliveryChannel)
	}

	preparedPayload, err := payload.Encode()
	if err != nil {
		svc.logger.Error("failed to encode payload", err, "user_id", user.ID, "notification_type", notificationType)
		return
	}

	for _, deliveryChannel := range deliveryChannels {
		dest, err := prepareDestination(deliveryChannel, user)
		if err != nil {
			svc.logger.Error("enqueue notification by channel failed", err)
			continue
		}

		_, err = svc.storage.NotificationSendQueue().Create(ctx, repo_notification_send_queue.CreateParams{
			Destination: dest,
			Type:        notificationType,
			Parameters:  preparedPayload,
			Channel:     deliveryChannel,
			Args:        args,
		})
		if err != nil {
			svc.logger.Error("failed to enqueue notification", err, "channel", deliveryChannel, "destination", dest)
			continue
		}
	}
}

// SendSystemEmail sends plain email to defined destination
func (svc *Service) SendSystemEmail(
	ctx context.Context,
	notificationType models.NotificationType,
	email string,
	payload INotificationBody,
	args *models.NotificationArgs,
) {
	preparedPayload, err := payload.Encode()
	if err != nil {
		svc.logger.Error("failed to encode payload", err, "destination", email, "notification_type", notificationType)
		return
	}

	existHistory, err := svc.storage.NotificationSendHistory().ExistWasSentRecently(ctx, repo_notification_send_history.ExistWasSentRecentlyParams{
		Destination: email,
		Type:        notificationType,
		Channel:     models.EmailDeliveryChannel,
	})
	if err != nil {
		return
	}

	if existHistory && notificationType == models.NotificationTypeExternalWalletRequested {
		svc.logger.Info("notification was sent recently, skipping", "destination", email, "notification_type", notificationType)
		return
	}

	params := repo_notification_send_queue.CreateParams{
		Destination: email,
		Type:        notificationType,
		Parameters:  preparedPayload,
		Channel:     models.EmailDeliveryChannel,
		Args:        args,
	}
	_, err = svc.storage.NotificationSendQueue().Create(ctx, params)

	if err != nil {
		svc.logger.Error("failed to enqueue system email", err, "destination", email, "notification_type", notificationType)
	}
}

func (svc *Service) processQueue(ctx context.Context) {
	if !svc.inUse.CompareAndSwap(false, true) {
		return
	}
	defer svc.inUse.Store(false)

	notifications, err := svc.storage.NotificationSendQueue().GetQueuedNotifications(ctx, svc.maxRetries)
	if err != nil {
		svc.logger.Error("failed to get notification queue", err)
	}

	for _, notification := range notifications {
		res := svc.sender.Send(ctx, *notification)
		if !res.IsExternal {
			svc.createSendHistory(ctx, *notification, res.Sender, string(res.SentBody))
		}

		currentAttempt, err := svc.storage.NotificationSendQueue().IncreaseAttempts(ctx, notification.ID)
		if err != nil {
			svc.logger.Error("failed to increase attempts", err, "id", notification.ID)
		}

		if res.IsExternal || res.IsSuccess || currentAttempt >= svc.maxRetries {
			if deleteErr := svc.storage.NotificationSendQueue().Delete(ctx, notification.ID); deleteErr != nil {
				svc.logger.Error("failed to delete notification queue job", deleteErr)
			}
		}
	}
}

func (svc *Service) createSendHistory(ctx context.Context, notification models.NotificationSendQueue, sender, sentBody string) {
	params := repo_notification_send_history.CreateParams{
		Destination:             notification.Destination,
		MessageText:             pgtype.Text{String: sentBody, Valid: true},
		Sender:                  sender,
		Type:                    notification.Type,
		Channel:                 notification.Channel,
		NotificationSendQueueID: notification.ID,
	}
	if notification.Args != nil {
		if notification.Args.UserID != nil {
			params.UserID = uuid.NullUUID{
				UUID:  *notification.Args.UserID,
				Valid: true,
			}
		}
		if notification.Args.StoreID != nil {
			params.StoreID = uuid.NullUUID{
				UUID:  *notification.Args.StoreID,
				Valid: true,
			}
		}
	}
	err := svc.storage.NotificationSendHistory().Create(ctx, params)
	if err != nil {
		svc.logger.Error("failed to submit notification history", err)
	}
}

func (svc *Service) runCleanHistoryJob(ctx context.Context) error {
	deletedRows, err := svc.storage.NotificationSendHistory().DeleteOldHistory(ctx)
	if err != nil {
		return err
	}
	if deletedRows > 0 {
		svc.logger.Info("Cleaned old notification history", "deletedRows", int(deletedRows))
	}
	return nil
}

func prepareDestination(notificationType models.DeliveryChannel, user *models.User) (string, error) {
	switch notificationType {
	case models.TelegramDeliveryChannel:
		if !user.ProcessingOwnerID.Valid {
			return "", errors.New("owner is unintialized")
		}

		return user.ProcessingOwnerID.UUID.String(), nil
	case models.EmailDeliveryChannel:
		return user.Email, nil
	default:
		return "", fmt.Errorf("unknown notification type:'%s'", notificationType)
	}
}

func (svc *Service) processReminds(ctx context.Context) {
	if !svc.inUse.CompareAndSwap(false, true) {
		return
	}
	defer svc.inUse.Store(false)

	users, err := svc.storage.Users().GetUnverifed(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return
		}
		svc.logger.Error("failed to get unverified users", err)
	}

	for _, user := range users {
		payload := &UserRemindVerificationData{
			Language: user.Language,
		}

		go svc.SendUser(ctx, models.NotificationTypeUserRemindVerification, user, payload, &models.NotificationArgs{
			UserID: &user.ID,
		})
	}
}

func (svc *Service) GetHistoryByParams(ctx context.Context, usr *models.User, params *notification_request.GetNotificationHistoryRequest) (*storecmn.FindResponseWithFullPagination[*models.NotificationSendHistory], error) {
	commonParams := storecmn.NewCommonFindParams()
	if params.PageSize != nil {
		commonParams.PageSize = params.PageSize
	}
	if params.Page != nil {
		commonParams.Page = params.Page
	}

	historyParams := repo_notification_send_history.GetHistoryByUserParams{
		CommonFindParams: commonParams,
		IDs:              params.IDs,
		Destinations:     params.Destinations,
		Types:            params.Types,
		Channels:         params.Channels,
		CreatedFrom:      params.CreatedFrom,
		CreatedTo:        params.CreatedTo,
		SentFrom:         params.SentFrom,
		SentTo:           params.SentTo,
	}

	isRoot, err := svc.permissionService.IsRoot(usr.ID.String())
	if err != nil {
		return nil, fmt.Errorf("check user root permission: %w", err)
	}
	historyParams.IsRoot = isRoot

	history, err := svc.storage.NotificationSendHistory().GetHistoryByUser(ctx, usr.ID, historyParams)
	if err != nil {
		return nil, fmt.Errorf("get notification history: %w", err)
	}

	return history, nil
}

func (svc *Service) GetTypesList(ctx context.Context) ([]NotificationType, error) {
	types, err := svc.storage.Notifications().GetAllTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("get notification types: %w", err)
	}

	result := make([]NotificationType, 0, len(types))
	for _, t := range types {
		result = append(result, NotificationType{
			Label: t.Type.Label(),
			Value: t.Type.String(),
		})
	}

	return result, nil
}
