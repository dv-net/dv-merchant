package mail_sender

import (
	"crypto/tls"
	"fmt"
	"io"

	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/service/templater"
	"github.com/dv-net/dv-merchant/internal/settings"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/dv-net/dv-merchant/pkg/mailer"

	"golang.org/x/net/context"
)

type MailerProvider string

func (o MailerProvider) String() string { return string(o) }

func (o MailerProvider) Valid() bool {
	_, ok := validMailingProviders[o]
	return ok
}

var validMailingProviders = map[MailerProvider]bool{
	MailerProviderSMTP:     true,
	MailerProviderSendGrid: true,
}

const (
	MailerProviderSMTP     MailerProvider = "smtp"
	MailerProviderSendGrid MailerProvider = "sendgrid"
)

type IMailSender interface {
	Send(from string, to []string, body io.Reader) error
}

type NotificationHandlerFunc func(context.Context, string, []byte) ([]byte, error)

type Service struct {
	mailerSettings *settings.MailerSettings

	handlers    map[models.NotificationType]NotificationHandlerFunc
	templateSvc templater.ITemplaterService

	mailerClient IMailSender

	log         logger.Logger
	settingsSrv setting.ISettingService
}

func New(log logger.Logger, eventListener event.IListener, templateSvc templater.ITemplaterService, settingSrv setting.ISettingService) (*Service, error) {
	svc := &Service{
		settingsSrv: settingSrv,
		log:         log,
		templateSvc: templateSvc,
	}

	if err := svc.initServices(); err != nil {
		svc.log.Error("failed to initialize mailer", fmt.Errorf("init services: %w", err))
		return nil, err
	}

	svc.handlers = map[models.NotificationType]NotificationHandlerFunc{
		models.NotificationTypeUserVerification:               svc.handleUserVerificationEmail,
		models.NotificationTypeUserRegistration:               svc.handleUserRegistrationEmail,
		models.NotificationTypeUserPasswordChanged:            svc.handleUserPasswordChanged,
		models.NotificationTypeUserForgotPassword:             svc.handleUserForgotPassword,
		models.NotificationTypeExternalWalletRequested:        svc.handleUserExternalWalletRequested,
		models.NotificationTypeUserInvite:                     svc.handleUserInvite,
		models.NotificationTypeUserEmailReset:                 svc.handleUserEmailReset,
		models.NotificationTypeUserEmailChange:                svc.handleUserChangeEmail,
		models.NotificationTypeUserAccessKeyChanged:           svc.handleUserAccessKeyChanged,
		models.NotificationTypeUserAuthorizationFromNewDevice: svc.handleUserAuthorizationFromNewDevice,
		models.NotificationTypeUserRemindVerification:         svc.handleRemindUserVerification,
		models.NotificationTypeUserUpdateSetting:              svc.handleVerifySettingsChange,
		models.NotificationTypeUserTestEmail:                  svc.handleUserTestEmail,
		models.NotificationTypeTwoFactorAuthentication:        svc.handleTwoFactorAuthentication,
		models.NotificationTypeUserCryptoReceipt:              svc.handleUserCryptoReceipt,
	}

	eventListener.Register(setting.MailerSettingsChanged, svc.handleMailerSettingsChanged)

	return svc, nil
}

func (svc *Service) handleMailerSettingsChanged(e event.IEvent) error {
	if _, ok := e.(*setting.MailerSettingChangedEvent); ok {
		return svc.initServices()
	}

	return fmt.Errorf("invalid event-type [%s] fired", e.String())
}

func (svc *Service) Send(
	ctx context.Context,
	notificationType models.NotificationType,
	dest string,
	encodedVars []byte,
) (notify.SendResult, error) {
	if svc == nil || svc.mailerClient == nil {
		return notify.SendResult{}, fmt.Errorf("service is nil")
	}

	sendRes := notify.SendResult{
		Sender: svc.mailerSettings.MailerSender,
	}

	handler, ok := svc.handlers[notificationType]
	if !ok {
		return sendRes, fmt.Errorf("unsupported notification type: %s", notificationType)
	}

	sentBody, err := handler(ctx, dest, encodedVars)
	sendRes.SentBody = sentBody
	if err != nil {
		return sendRes, fmt.Errorf("sending mail: %w", err)
	}

	sendRes.IsSuccess = true
	return sendRes, nil
}

func (svc *Service) initServices() error {
	ctx := context.Background()
	mailSettings, err := svc.settingsSrv.GetMailerSettings(ctx)
	if err != nil || mailSettings == nil {
		return fmt.Errorf("fetch mailer settings")
	}
	svc.mailerSettings = mailSettings

	if svc.mailerSettings.MailerState != setting.FlagValueEnabled {
		svc.log.Info("Mailer service is disabled")
		return nil
	}

	return svc.initMailer(ctx)
}

func (svc *Service) initMailer(ctx context.Context) error {
	svc.log.Info("Initializing SMTP mailer")

	switch svc.mailerSettings.MailerProtocol {
	case MailerProviderSMTP.String():
		opts := &mailer.PoolOptions{
			Address:  svc.mailerSettings.MailerAddress,
			Identity: "github.com/dv-net/dv-merchant",
		}

		if svc.mailerSettings.MailerEncryption == "tls" {
			opts.SSL = true
			opts.TLSConfig = &tls.Config{InsecureSkipVerify: true, ServerName: svc.mailerSettings.MailerAddress} //nolint:gosec
		}

		if svc.mailerSettings.MailerUsername != "" && svc.mailerSettings.MailerPassword != "" {
			opts.Username = &svc.mailerSettings.MailerUsername
			opts.Password = &svc.mailerSettings.MailerPassword
		}

		smtpCl, err := mailer.NewSMTPPool(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to initialize SMTP mailer: %w", err)
		}

		svc.mailerClient = smtpCl
	default:
		err := fmt.Errorf("unsupported mailer provider: %s", svc.mailerSettings.MailerProtocol)
		svc.log.Error("failed to init notification services", err)
		return err
	}
	return nil
}
