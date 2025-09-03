package external_sender

import (
	"errors"
	"fmt"

	"github.com/dv-net/dv-merchant/pkg/admin_gateway"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/setting"

	"golang.org/x/net/context"
)

type ExternalNotificationHandlerFunc func(ctx context.Context, params []byte, dest string, channel models.DeliveryChannel) error

type Service struct {
	adminNotifications admin_gateway.INotification
	handlers           map[models.NotificationType]ExternalNotificationHandlerFunc
	settingSvc         setting.ISettingService
}

func New(gateway admin_gateway.INotification, settingSvc setting.ISettingService) *Service {
	svc := &Service{
		adminNotifications: gateway,
		settingSvc:         settingSvc,
	}

	svc.handlers = map[models.NotificationType]ExternalNotificationHandlerFunc{
		models.NotificationTypeUserVerification:        svc.handleUserVerification,
		models.NotificationTypeUserRegistration:        svc.handleUserRegistration,
		models.NotificationTypeUserPasswordChanged:     svc.handleUserPasswordReset,
		models.NotificationTypeUserForgotPassword:      svc.handleUserForgotPassword,
		models.NotificationTypeExternalWalletRequested: svc.handleUserExternalWalletRequest,
		models.NotificationTypeUserEmailReset:          svc.handleUserEmailReset,
		models.NotificationTypeUserRemindVerification:  svc.handleUserRemindVerification,
		models.NotificationTypeUserUpdateSetting:       svc.handleUserUpdateSettingVerification,
	}

	return svc
}

func (svc *Service) Send(ctx context.Context, notification models.NotificationSendQueue) error {
	if svc == nil {
		return errors.New("service is nil")
	}

	handler, ok := svc.handlers[notification.Type]
	if !ok {
		return fmt.Errorf("notification handler not found for type %s", notification.Type)
	}

	clID, err := svc.settingSvc.GetRootSetting(ctx, setting.ProcessingClientID)
	if err != nil {
		return err
	}
	adminSecret, err := svc.settingSvc.GetRootSetting(ctx, setting.DvAdminSecretKey)
	if err != nil {
		return fmt.Errorf("get admin secret key error: %w", err)
	}

	return handler(
		admin_gateway.PrepareServiceContext(ctx, adminSecret.Value, clID.Value),
		notification.Parameters,
		notification.Destination,
		notification.Channel,
	)
}
