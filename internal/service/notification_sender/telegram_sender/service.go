package telegram_sender

import (
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/notify"

	"golang.org/x/net/context"
)

// Service TODO implement for internal telegram notifications support
type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (svc *Service) Send(_ context.Context, _ models.NotificationType, dest string, _ []byte) (notify.SendResult, error) {
	if svc == nil {
		return notify.SendResult{}, fmt.Errorf("service is nil")
	}
	return notify.SendResult{}, fmt.Errorf("not implemented driver for channel: %s", dest)
}
