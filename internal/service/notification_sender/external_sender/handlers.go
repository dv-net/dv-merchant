package external_sender

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/tools/url"
	admin_requests "github.com/dv-net/dv-merchant/pkg/admin_gateway/requests"

	"golang.org/x/net/context"
)

func (svc *Service) handleUserVerification(ctx context.Context, encodedVariables []byte, dest string, channel models.DeliveryChannel) error {
	pBody, err := notify.ParseNotificationBody[notify.UserVerificationData](encodedVariables)
	if err != nil {
		return fmt.Errorf("parse user verification body: %w", err)
	}
	clID, domain, err := svc.getBackendSettings(ctx)
	if err != nil {
		return err
	}

	req := admin_requests.VerifyNotification{
		BackendClientID: clID.Value,
		BackendDomain:   domain,
		Locale:          pBody.Language,
		Identity:        svc.prepareIdentityByType(dest, channel),
		Code:            strconv.Itoa(pBody.Code),
	}

	return svc.adminNotifications.SendUserVerification(ctx, req)
}

func (svc *Service) handleUserRegistration(ctx context.Context, encodedVariables []byte, dest string, channel models.DeliveryChannel) error {
	pBody, err := notify.ParseNotificationBody[notify.UserRegistration](encodedVariables)
	if err != nil {
		return fmt.Errorf("parse user verification email payload: %w", err)
	}
	clID, domain, err := svc.getBackendSettings(ctx)
	if err != nil {
		return err
	}

	req := admin_requests.VerifyNotification{
		BackendClientID: clID.Value,
		BackendDomain:   domain,
		Locale:          pBody.Language,
		Identity:        svc.prepareIdentityByType(dest, channel),
	}
	return svc.adminNotifications.SendUserRegistration(ctx, req)
}

func (svc *Service) handleUserForgotPassword(ctx context.Context, encodedVariables []byte, dest string, channel models.DeliveryChannel) error {
	pBody, err := notify.ParseNotificationBody[notify.PasswordForgotData](encodedVariables)
	if err != nil {
		return fmt.Errorf("parse user password forgot email payload: %w", err)
	}
	clID, domain, err := svc.getBackendSettings(ctx)
	if err != nil {
		return err
	}

	req := admin_requests.VerifyNotification{
		BackendClientID: clID.Value,
		BackendDomain:   domain,
		Locale:          pBody.Language,
		Identity:        svc.prepareIdentityByType(dest, channel),
		Code:            pBody.Code,
	}
	return svc.adminNotifications.SendUserForgotPassword(ctx, req)
}

func (svc *Service) handleUserPasswordReset(ctx context.Context, encodedVariables []byte, dest string, channel models.DeliveryChannel) error {
	pBody, err := notify.ParseNotificationBody[notify.UserPasswordChanged](encodedVariables)
	if err != nil {
		return fmt.Errorf("parse user password reset email payload: %w", err)
	}
	clID, domain, err := svc.getBackendSettings(ctx)
	if err != nil {
		return err
	}

	req := admin_requests.VerifyNotification{
		BackendClientID: clID.Value,
		BackendDomain:   domain,
		Locale:          pBody.Language,
		Identity:        svc.prepareIdentityByType(dest, channel),
	}
	return svc.adminNotifications.SendUserPasswordReset(ctx, req)
}

func (svc *Service) handleUserEmailReset(ctx context.Context, encodedVariables []byte, dest string, channel models.DeliveryChannel) error {
	pBody, err := notify.ParseNotificationBody[notify.UserPasswordChanged](encodedVariables)
	if err != nil {
		return fmt.Errorf("parse user password reset email payload: %w", err)
	}
	clID, domain, err := svc.getBackendSettings(ctx)
	if err != nil {
		return err
	}

	req := admin_requests.VerifyNotification{
		BackendClientID: clID.Value,
		BackendDomain:   domain,
		Locale:          pBody.Language,
		Identity:        svc.prepareIdentityByType(dest, channel),
	}
	return svc.adminNotifications.SendUserEmailReset(ctx, req)
}

func (svc *Service) handleUserExternalWalletRequest(ctx context.Context, encodedVariables []byte, dest string, channel models.DeliveryChannel) error {
	pBody, err := notify.ParseNotificationBody[notify.ExternalWalletRequestedData](encodedVariables)
	if err != nil {
		return fmt.Errorf("parse user verification email payload: %w", err)
	}
	clID, domain, err := svc.getBackendSettings(ctx)
	if err != nil {
		return err
	}

	req := admin_requests.WalletsRequestNotification{
		BackendClientID: clID.Value,
		BackendDomain:   domain,
		Locale:          pBody.Language,
		Identity:        svc.prepareIdentityByType(dest, channel),
		Addresses:       svc.convertWalletsDto(pBody.Addresses),
	}
	return svc.adminNotifications.SendExternalWalletRequested(ctx, req)
}

func (svc *Service) convertWalletsDto(wallets []notify.WalletDTO) []admin_requests.WalletAddressDTO {
	res := make([]admin_requests.WalletAddressDTO, 0, len(wallets))
	for _, v := range wallets {
		res = append(res, admin_requests.WalletAddressDTO{
			CurrencyID: v.CurrencyID,
			Blockchain: v.BlockchainID,
			Address:    v.Address,
		})
	}

	return res
}

func (svc *Service) getBackendSettings(ctx context.Context) (*models.Setting, string, error) {
	clID, err := svc.settingSvc.GetRootSetting(ctx, setting.ProcessingClientID)
	if err != nil || clID == nil {
		return nil, "", errors.New("invalid client id from settings")
	}
	backendDomain, err := svc.settingSvc.GetRootSetting(ctx, setting.MerchantDomain)
	if err != nil || backendDomain == nil {
		return nil, "", errors.New("invalid domain from settings")
	}
	domain, err := url.GetDomain(backendDomain.Value)
	if err != nil {
		return nil, "", errors.New("error parse domain from shem")
	}
	return clID, domain, nil
}

func (svc *Service) prepareIdentityByType(dest string, channel models.DeliveryChannel) admin_requests.NotificationIdentity {
	return admin_requests.NotificationIdentity{
		Destination: dest,
		Channel:     channel.String(),
	}
}

func (svc *Service) handleUserRemindVerification(ctx context.Context, encodedVariables []byte, dest string, channel models.DeliveryChannel) error {
	pBody, err := notify.ParseNotificationBody[notify.UserRemindVerificationData](encodedVariables)
	if err != nil {
		return fmt.Errorf("parse user remind verification payload: %w", err)
	}

	clID, domain, err := svc.getBackendSettings(ctx)
	if err != nil {
		return err
	}

	req := admin_requests.VerifyNotification{
		BackendClientID: clID.Value,
		BackendDomain:   domain,
		Locale:          pBody.Language,
		Identity:        svc.prepareIdentityByType(dest, channel),
	}

	return svc.adminNotifications.SendUserRemindVerification(ctx, req)
}

func (svc *Service) handleUserUpdateSettingVerification(ctx context.Context, encodedVariables []byte, dest string, channel models.DeliveryChannel) error {
	pBody, err := notify.ParseNotificationBody[notify.UserUpdateSettingData](encodedVariables)
	if err != nil {
		return fmt.Errorf("parse user remind verification payload: %w", err)
	}

	clID, domain, err := svc.getBackendSettings(ctx)
	if err != nil {
		return err
	}

	req := admin_requests.VerifyNotification{
		BackendClientID: clID.Value,
		BackendDomain:   domain,
		Locale:          pBody.Language,
		Identity:        svc.prepareIdentityByType(dest, channel),
		Code:            strconv.Itoa(pBody.Code),
	}

	return svc.adminNotifications.SendUserUpdateSettingVerification(ctx, req)
}
