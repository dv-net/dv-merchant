package models

type NotificationType string // @name NotificationType

func (o NotificationType) String() string { return string(o) }

func (o NotificationType) Valid() bool {
	_, ok := validNotificationTypes[o]
	return ok
}

func (o NotificationType) Label() string {
	switch o {
	case NotificationTypeUserVerification:
		return "User verification"
	case NotificationTypeUserRegistration:
		return "User registration"
	case NotificationTypeUserPasswordChanged:
		return "User password changed"
	case NotificationTypeUserForgotPassword:
		return "User forgot password"
	case NotificationTypeTwoFactorAuthentication:
		return "Two factor authentication"
	case NotificationTypeExternalWalletRequested:
		return "External wallet requested"
	case NotificationTypeUserInvite:
		return "User invite"
	case NotificationTypeUserEmailReset:
		return "User email reset"
	case NotificationTypeUserAccessKeyChanged:
		return "User access key changed"
	case NotificationTypeUserAuthorizationFromNewDevice:
		return "User authorization from new device"
	case NotificationTypeUserRemindVerification:
		return "User remind verification"
	case NotificationTypeUserUpdateSetting:
		return "User update setting verification"
	case NotificationTypeUserTestEmail:
		return "User test email"
	case NotificationTypeUserEmailChange:
		return "User email change"
	case NotificationTypeUserCryptoReceipt:
		return "User crypto receipt"
	default:
		return "Unknown Notification Type"
	}
}

const (
	NotificationTypeUserVerification               NotificationType = "user_verification"
	NotificationTypeUserRegistration               NotificationType = "user_registration"
	NotificationTypeUserPasswordChanged            NotificationType = "user_password_reset"
	NotificationTypeUserForgotPassword             NotificationType = "user_forgot_password"
	NotificationTypeTwoFactorAuthentication        NotificationType = "two_factor_authentication"
	NotificationTypeExternalWalletRequested        NotificationType = "external_wallet_requested"
	NotificationTypeUserInvite                     NotificationType = "user_invite"
	NotificationTypeUserEmailReset                 NotificationType = "user_email_reset"
	NotificationTypeUserEmailChange                NotificationType = "user_change_email"
	NotificationTypeUserAccessKeyChanged           NotificationType = "user_access_key_changed"
	NotificationTypeUserAuthorizationFromNewDevice NotificationType = "user_authorization_from_new_device"
	NotificationTypeUserRemindVerification         NotificationType = "user_remind_verification"
	NotificationTypeUserUpdateSetting              NotificationType = "user_update_setting_verification"
	NotificationTypeUserTestEmail                  NotificationType = "user_test_email"
	NotificationTypeUserCryptoReceipt              NotificationType = "user_crypto_receipt"
)

var validNotificationTypes = map[NotificationType]struct{}{
	NotificationTypeUserVerification:               {},
	NotificationTypeUserRegistration:               {},
	NotificationTypeUserPasswordChanged:            {},
	NotificationTypeUserForgotPassword:             {},
	NotificationTypeTwoFactorAuthentication:        {},
	NotificationTypeExternalWalletRequested:        {},
	NotificationTypeUserInvite:                     {},
	NotificationTypeUserEmailReset:                 {},
	NotificationTypeUserEmailChange:                {},
	NotificationTypeUserAccessKeyChanged:           {},
	NotificationTypeUserAuthorizationFromNewDevice: {},
	NotificationTypeUserRemindVerification:         {},
	NotificationTypeUserUpdateSetting:              {},
	NotificationTypeUserTestEmail:                  {},
	NotificationTypeUserCryptoReceipt:              {},
}
