package mail_sender

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/service/templater"

	"golang.org/x/net/context"
)

func (svc *Service) handleUserVerificationEmail(_ context.Context, email string, encodedVariables []byte) ([]byte, error) {
	pBody, err := notify.ParseNotificationBody[notify.UserVerificationData](encodedVariables)
	if err != nil {
		return nil, fmt.Errorf("parse user verification email payload: %w", err)
	}

	emailParams := &templater.UserEmailVerification{
		BasePayload: templater.BasePayload{
			UserEmail: email,
			Language:  pBody.Language,
		},
		VerifyEmailCode: strconv.Itoa(pBody.Code),
	}

	body, err := svc.templateSvc.AssembleEmail(emailParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create email template: %w", err)
	}

	bodyBytes := body.Bytes()
	err = svc.mailerClient.Send(svc.mailerSettings.MailerSender, []string{email}, bytes.NewBuffer(bodyBytes))
	return bodyBytes, err
}

func (svc *Service) handleUserForgotPassword(_ context.Context, email string, encodedVariables []byte) ([]byte, error) {
	pBody, err := notify.ParseNotificationBody[notify.UserForgotPassword](encodedVariables)
	if err != nil {
		return nil, fmt.Errorf("parse user forgot password email payload: %w", err)
	}
	emailParams := &templater.UserForgotPassword{
		BasePayload: templater.BasePayload{
			UserEmail: email,
			Language:  pBody.Language,
		},
		UserForgotPasswordCode: pBody.ResetPasswordCode,
	}

	body, err := svc.templateSvc.AssembleEmail(emailParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create email template: %w", err)
	}

	bodyBytes := body.Bytes()
	err = svc.mailerClient.Send(svc.mailerSettings.MailerSender, []string{email}, bytes.NewBuffer(bodyBytes))
	return bodyBytes, err
}

func (svc *Service) handleUserPasswordChanged(_ context.Context, email string, encodedVariables []byte) ([]byte, error) {
	pBody, err := notify.ParseNotificationBody[notify.UserPasswordChanged](encodedVariables)
	if err != nil {
		return nil, fmt.Errorf("parse user password changed email payload: %w", err)
	}
	emailParams := &templater.UserPasswordChanged{
		BasePayload: templater.BasePayload{
			UserEmail: email,
			Language:  pBody.Language,
		},
	}

	body, err := svc.templateSvc.AssembleEmail(emailParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create email template: %w", err)
	}

	bodyBytes := body.Bytes()
	err = svc.mailerClient.Send(svc.mailerSettings.MailerSender, []string{email}, bytes.NewBuffer(bodyBytes))
	return bodyBytes, err
}

func (svc *Service) handleUserChangeEmail(_ context.Context, email string, encodedVariables []byte) ([]byte, error) {
	pBody, err := notify.ParseNotificationBody[notify.UserInitEmailChangeData](encodedVariables)
	if err != nil {
		return nil, fmt.Errorf("parse user change email payload: %w", err)
	}
	emailParams := &templater.UserChangeEmail{
		BasePayload: templater.BasePayload{
			Language:  pBody.Language,
			UserEmail: email,
		},
		ChangeEmailCode: strconv.Itoa(pBody.Code),
	}

	body, err := svc.templateSvc.AssembleEmail(emailParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create email template: %w", err)
	}

	bodyBytes := body.Bytes()
	err = svc.mailerClient.Send(svc.mailerSettings.MailerSender, []string{email}, bytes.NewBuffer(bodyBytes))
	return bodyBytes, err
}

func (svc *Service) handleUserEmailReset(_ context.Context, email string, encodedVariables []byte) ([]byte, error) {
	pBody, err := notify.ParseNotificationBody[notify.EmailChangeConfirmData](encodedVariables)
	if err != nil {
		return nil, fmt.Errorf("parse user email reset payload: %w", err)
	}
	emailParams := &templater.UserEmailReset{
		BasePayload: templater.BasePayload{
			Language:  pBody.Language,
			UserEmail: email,
		},
		NewEmail: pBody.NewEmail,
	}

	body, err := svc.templateSvc.AssembleEmail(emailParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create email template: %w", err)
	}

	bodyBytes := body.Bytes()
	err = svc.mailerClient.Send(svc.mailerSettings.MailerSender, []string{email}, bytes.NewBuffer(bodyBytes))
	return bodyBytes, err
}

func (svc *Service) handleUserExternalWalletRequested(_ context.Context, email string, encodedVariables []byte) ([]byte, error) {
	pBody, err := notify.ParseNotificationBody[notify.ExternalWalletRequestedData](encodedVariables)
	if err != nil {
		return nil, fmt.Errorf("parse user verification email payload: %w", err)
	}

	preparedAddresses := make([]templater.ExternalWallet, 0, len(pBody.Addresses))
	for _, v := range pBody.Addresses {
		preparedAddresses = append(preparedAddresses, templater.ExternalWallet{
			WalletCurrency:       v.CurrencyID,
			WalletAddress:        v.Address,
			WalletBlockchain:     v.BlockchainID,
			ShowBlockchain:       v.ShowBlockchain,
			WalletCurrencyName:   v.CurrencyName,
			WalletBlockchainName: v.BlockchainName,
		})
	}
	emailParams := &templater.UserExternalWallet{
		BasePayload: templater.BasePayload{
			UserEmail: email,
			Language:  pBody.Language,
		},
		Wallets:          preparedAddresses,
		NotificationHash: pBody.NotificationHash,
	}

	body, err := svc.templateSvc.AssembleEmail(emailParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create email template: %w", err)
	}

	bodyBytes := body.Bytes()
	err = svc.mailerClient.Send(svc.mailerSettings.MailerSender, []string{email}, bytes.NewBuffer(bodyBytes))
	return bodyBytes, err
}

func (svc *Service) handleUserRegistrationEmail(_ context.Context, email string, encodedVariables []byte) ([]byte, error) {
	pBody, err := notify.ParseNotificationBody[notify.UserRegistration](encodedVariables)
	if err != nil {
		return nil, fmt.Errorf("parse user verification email payload: %w", err)
	}
	emailParams := &templater.UserRegistration{
		BasePayload: templater.BasePayload{
			UserEmail: email,
			Language:  pBody.Language,
		},
	}

	body, err := svc.templateSvc.AssembleEmail(emailParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create email template: %w", err)
	}

	bodyBytes := body.Bytes()
	err = svc.mailerClient.Send(svc.mailerSettings.MailerSender, []string{email}, bytes.NewBuffer(bodyBytes))
	return bodyBytes, err
}

func (svc *Service) handleUserInvite(_ context.Context, email string, encodedVariables []byte) ([]byte, error) {
	pBody, err := notify.ParseNotificationBody[notify.UserInviteData](encodedVariables)
	if err != nil {
		return nil, fmt.Errorf("parse user invite email payload: %w", err)
	}
	emailParams := &templater.UserInviteEmail{
		BasePayload: templater.BasePayload{
			UserEmail: email,
			Language:  pBody.Language,
		},
		InviteUserRole: pBody.Role,
		InviteUserLink: pBody.Link,
	}

	body, err := svc.templateSvc.AssembleEmail(emailParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create email template: %w", err)
	}

	bodyBytes := body.Bytes()
	err = svc.mailerClient.Send(svc.mailerSettings.MailerSender, []string{email}, bytes.NewBuffer(bodyBytes))
	return bodyBytes, err
}

func (svc *Service) handleUserAccessKeyChanged(_ context.Context, email string, encodedVariables []byte) ([]byte, error) {
	pBody, err := notify.ParseNotificationBody[notify.UserAccessKeyChanged](encodedVariables)
	if err != nil {
		return nil, fmt.Errorf("parse user access keys changed payload: %w", err)
	}
	emailParams := &templater.UserAccessKeyChanged{
		BasePayload: templater.BasePayload{
			UserEmail: email,
			Language:  pBody.Language,
		},
	}

	body, err := svc.templateSvc.AssembleEmail(emailParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create email template: %w", err)
	}

	bodyBytes := body.Bytes()
	err = svc.mailerClient.Send(svc.mailerSettings.MailerSender, []string{email}, bytes.NewBuffer(bodyBytes))
	return bodyBytes, err
}

func (svc *Service) handleUserAuthorizationFromNewDevice(_ context.Context, email string, encodedVariables []byte) ([]byte, error) {
	pBody, err := notify.ParseNotificationBody[notify.UserAuthorizationFromNewDevice](encodedVariables)
	if err != nil {
		return nil, fmt.Errorf("parse user authorization from new device payload: %w", err)
	}
	emailParams := &templater.UserAuthorizationFromNewDevice{
		BasePayload: templater.BasePayload{
			UserEmail: email,
			Language:  pBody.Language,
		},
		NewAuthorizationIP:           pBody.NewAuthorizationIP,
		NewAuthorizationTimestamp:    pBody.NewAuthorizationTimestamp,
		NewAuthorizationAccountEmail: pBody.NewAuthorizationAccountEmail,
		NewAuthorizationLocation:     pBody.NewAuthorizationLocation,
	}

	body, err := svc.templateSvc.AssembleEmail(emailParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create email template: %w", err)
	}

	bodyBytes := body.Bytes()
	err = svc.mailerClient.Send(svc.mailerSettings.MailerSender, []string{email}, bytes.NewBuffer(bodyBytes))
	return bodyBytes, err
}

func (svc *Service) handleRemindUserVerification(_ context.Context, email string, encodedVariables []byte) ([]byte, error) {
	pBody, err := notify.ParseNotificationBody[notify.UserRemindVerificationData](encodedVariables)
	if err != nil {
		return nil, fmt.Errorf("parse user remind verification email payload: %w", err)
	}

	emailParams := &templater.UserRemindVerification{
		BasePayload: templater.BasePayload{
			UserEmail: email,
			Language:  pBody.Language,
		},
	}

	body, err := svc.templateSvc.AssembleEmail(emailParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create email template: %w", err)
	}

	bodyBytes := body.Bytes()
	err = svc.mailerClient.Send(svc.mailerSettings.MailerSender, []string{email}, bytes.NewBuffer(bodyBytes))
	return bodyBytes, err
}

func (svc *Service) handleVerifySettingsChange(_ context.Context, email string, encodedVariables []byte) ([]byte, error) {
	pBody, err := notify.ParseNotificationBody[notify.UserUpdateSettingData](encodedVariables)
	if err != nil {
		return nil, fmt.Errorf("parse user remind verification email payload: %w", err)
	}

	emailParams := &templater.UserUpdateSettingsVerification{
		BasePayload: templater.BasePayload{
			UserEmail: email,
			Language:  pBody.Language,
		},
		Code: pBody.Code,
	}

	body, err := svc.templateSvc.AssembleEmail(emailParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create email template: %w", err)
	}

	bodyBytes := body.Bytes()
	err = svc.mailerClient.Send(svc.mailerSettings.MailerSender, []string{email}, bytes.NewBuffer(bodyBytes))
	return bodyBytes, err
}

func (svc *Service) handleUserTestEmail(_ context.Context, email string, encodedVariables []byte) ([]byte, error) {
	pBody, err := notify.ParseNotificationBody[notify.UserTestEmailData](encodedVariables)
	if err != nil {
		return nil, fmt.Errorf("parse user test email payload: %w", err)
	}

	emailParams := &templater.UserTestEmail{
		BasePayload: templater.BasePayload{
			UserEmail: email,
			Language:  pBody.Language,
		},
	}

	body, err := svc.templateSvc.AssembleEmail(emailParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create email template: %w", err)
	}

	bodyBytes := body.Bytes()
	err = svc.mailerClient.Send(svc.mailerSettings.MailerSender, []string{email}, bytes.NewBuffer(bodyBytes))
	return bodyBytes, err
}

func (svc *Service) handleTwoFactorAuthentication(_ context.Context, email string, encodedVariables []byte) ([]byte, error) {
	pBody, err := notify.ParseNotificationBody[notify.TwoFactorAuthenticationNotification](encodedVariables)
	if err != nil {
		return nil, fmt.Errorf("parse two-factor authentication payload: %w", err)
	}

	emailParams := &templater.TwoFactorAuthentication{
		BasePayload: templater.BasePayload{
			UserEmail: email,
			Language:  pBody.Language,
		},
		IsEnabled: pBody.IsEnabled,
	}

	body, err := svc.templateSvc.AssembleEmail(emailParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create email template: %w", err)
	}

	bodyBytes := body.Bytes()
	err = svc.mailerClient.Send(svc.mailerSettings.MailerSender, []string{email}, bytes.NewBuffer(bodyBytes))
	return bodyBytes, err
}

func (svc *Service) handleUserCryptoReceipt(_ context.Context, email string, encodedVariables []byte) ([]byte, error) {
	pBody, err := notify.ParseNotificationBody[notify.UserCryptoReceiptNotificationData](encodedVariables)
	if err != nil {
		return nil, fmt.Errorf("parse user crypto receipt payload: %w", err)
	}

	emailParams := &templater.UserCryptoReceiptPayload{
		BasePayload: templater.BasePayload{
			UserEmail: email,
			Language:  pBody.Language,
		},
		PaymentStatus:        pBody.PaymentStatus,
		PaymentType:          pBody.PaymentType,
		ReceiptId:            pBody.ReceiptId,
		PaymentDate:          pBody.PaymentDate,
		UsdAmount:            pBody.UsdAmount,
		TokenAmount:          pBody.TokenAmount,
		TokenSymbol:          pBody.TokenSymbol,
		TransactionHash:      pBody.TransactionHash,
		BlockchainName:       pBody.BlockchainName,
		BlockchainCurrencyID: pBody.BlockchainCurrencyID,
		ExchangeRate:         pBody.ExchangeRate,
		NetworkFeeAmount:     pBody.NetworkFeeAmount,
		NetworkFeeCurrency:   pBody.NetworkFeeCurrency,
		NetworkFeeUSD:        pBody.NetworkFeeUSD,
	}

	body, err := svc.templateSvc.AssembleEmail(emailParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create email template: %w", err)
	}

	bodyBytes := body.Bytes()
	err = svc.mailerClient.Send(svc.mailerSettings.MailerSender, []string{email}, bytes.NewBuffer(bodyBytes))
	return bodyBytes, err
}
