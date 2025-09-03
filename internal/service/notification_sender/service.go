package notification_sender

import (
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"golang.org/x/net/context"
)

type Service struct {
	internalDrivers map[models.DeliveryChannel]IInternalSender
	logger          logger.Logger
	settingsSvc     setting.ISettingService
	externalSender  IExternalSender
}

type IInternalSender interface {
	Send(ctx context.Context, notificationType models.NotificationType, dest string, encodedVars []byte) (notify.SendResult, error)
}

type IExternalSender interface {
	Send(ctx context.Context, notification models.NotificationSendQueue) error
}

func New(log logger.Logger, drivers map[models.DeliveryChannel]IInternalSender, settingsSvc setting.ISettingService, externalSender IExternalSender) *Service {
	srv := &Service{
		logger:          log,
		settingsSvc:     settingsSvc,
		internalDrivers: drivers,
		externalSender:  externalSender,
	}

	return srv
}

func (svc *Service) Send(ctx context.Context, queue models.NotificationSendQueue) notify.SendResult {
	sender := setting.NotificationSenderInternal
	senderSetting, err := svc.settingsSvc.GetRootSetting(ctx, setting.NotificationSender)
	if err == nil && senderSetting != nil {
		sender = senderSetting.Value
	}

	switch sender {
	case setting.NotificationSenderInternal:
		return svc.sendInternal(ctx, queue)
	case setting.NotificationSenderDVNet:
		if err = svc.externalSender.Send(ctx, queue); err != nil {
			svc.logger.Error("send external notification failed", err)
		}
		return notify.SendResult{IsExternal: true}
	default:
		svc.logger.Warn("unsupported notification sender setting", "value", sender)
		return notify.SendResult{}
	}
}

func (svc *Service) sendInternal(ctx context.Context, queue models.NotificationSendQueue) notify.SendResult {
	driver, ok := svc.internalDrivers[queue.Channel]
	if !ok {
		svc.logger.Warn("driver not implemented", "channel", queue.Channel)
		return notify.SendResult{}
	}

	sendRes, err := driver.Send(ctx, queue.Type, queue.Destination, queue.Parameters)
	if err != nil {
		svc.logger.Error("send notification error", err, "channel", queue.Channel)
	}

	return sendRes
}
