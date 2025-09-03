package admin_gateway

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/net/context"

	admin_errors "github.com/dv-net/dv-merchant/pkg/admin_gateway/errors"
	admin_requests "github.com/dv-net/dv-merchant/pkg/admin_gateway/requests"
)

const (
	MethodAccountConfirmation        = "/notify/account-confirmation"
	MethodAccountConfirmed           = "/notify/account-confirmed"
	MethodUserPasswordResetting      = "/notify/password-resetting"
	MethodUserPasswordReset          = "/notify/password-resetted"
	MethodClientTopUpWallets         = "/notify/client-top-up-wallets"
	MethodNotificationEmailResetting = "/notify/email-resetting"
	MethodRemindAccountConfirmation  = "/notify/remind-account-confirmation"
	MethodUpdateSettingsConfirmation = "/notify/update-settings-confirmation"
)

type INotification interface {
	SendUserVerification(ctx context.Context, notification admin_requests.VerifyNotification) error
	SendUserRegistration(ctx context.Context, notification admin_requests.VerifyNotification) error
	SendUserPasswordReset(ctx context.Context, notification admin_requests.VerifyNotification) error
	SendUserForgotPassword(ctx context.Context, notification admin_requests.VerifyNotification) error
	SendTwoFactorAuthentication(ctx context.Context, notification admin_requests.VerifyNotification) error
	SendExternalWalletRequested(ctx context.Context, notification admin_requests.WalletsRequestNotification) error
	SendUserInvite(ctx context.Context, notification admin_requests.VerifyNotification) error
	SendUserEmailReset(ctx context.Context, notification admin_requests.VerifyNotification) error
	SendUserRemindVerification(ctx context.Context, notification admin_requests.VerifyNotification) error
	SendUserUpdateSettingVerification(ctx context.Context, notification admin_requests.VerifyNotification) error
}

var _ INotification = (*Service)(nil)

func (s *Service) SendUserVerification(ctx context.Context, notification admin_requests.VerifyNotification) error {
	encodedReq, err := json.Marshal(notification)
	if err != nil {
		return admin_errors.ErrPrepareRequest
	}

	_, err = s.sendRequest(ctx, MethodAccountConfirmation, http.MethodPost, encodedReq, nil)
	return err
}

func (s *Service) SendUserRegistration(ctx context.Context, notification admin_requests.VerifyNotification) error {
	encodedReq, err := json.Marshal(notification)
	if err != nil {
		return admin_errors.ErrPrepareRequest
	}

	_, err = s.sendRequest(ctx, MethodAccountConfirmed, http.MethodPost, encodedReq, nil)
	return err
}

func (s *Service) SendUserPasswordReset(ctx context.Context, notification admin_requests.VerifyNotification) error {
	encodedReq, err := json.Marshal(notification)
	if err != nil {
		return admin_errors.ErrPrepareRequest
	}

	_, err = s.sendRequest(ctx, MethodUserPasswordReset, http.MethodPost, encodedReq, nil)
	return err
}

func (s *Service) SendUserForgotPassword(ctx context.Context, notification admin_requests.VerifyNotification) error {
	encodedReq, err := json.Marshal(notification)
	if err != nil {
		return admin_errors.ErrPrepareRequest
	}

	_, err = s.sendRequest(ctx, MethodUserPasswordResetting, http.MethodPost, encodedReq, nil)
	return err
}

func (s *Service) SendUserEmailReset(ctx context.Context, notification admin_requests.VerifyNotification) error {
	encodedReq, err := json.Marshal(notification)
	if err != nil {
		return admin_errors.ErrPrepareRequest
	}

	_, err = s.sendRequest(ctx, MethodNotificationEmailResetting, http.MethodPost, encodedReq, nil)
	return err
}

func (s *Service) SendExternalWalletRequested(ctx context.Context, notification admin_requests.WalletsRequestNotification) error {
	encodedReq, err := json.Marshal(notification)
	if err != nil {
		return admin_errors.ErrPrepareRequest
	}

	_, err = s.sendRequest(ctx, MethodClientTopUpWallets, http.MethodPost, encodedReq, nil)
	return err
}

func (s *Service) SendUserUpdateSettingVerification(_ context.Context, _ admin_requests.VerifyNotification) error {
	return errors.New("not implemented")
}

func (s *Service) SendTwoFactorAuthentication(_ context.Context, _ admin_requests.VerifyNotification) error {
	return errors.New("not implemented")
}

func (s *Service) SendUserInvite(_ context.Context, _ admin_requests.VerifyNotification) error {
	return errors.New("not implemented")
}

func (s *Service) SendUserRemindVerification(_ context.Context, _ admin_requests.VerifyNotification) error {
	return errors.New("not implemented")
}
