//nolint:tagliatelle
package templater

import (
	"bytes"
	"encoding/json"

	"github.com/shopspring/decimal"
)

var validPartials = map[string]struct{}{
	UserVerificationPartialName:               {},
	UserInvitePartialName:                     {},
	UserAccessKeyChangedPartialName:           {},
	UserAuthorizationFromNewDevicePartialName: {},
	UserForgotPasswordPartialName:             {},
	UserRegistrationPartialName:               {},
	UserExternalWalletPartialName:             {},
	UserPasswordChangedPartialName:            {},
	UserRemindVerificationPartialName:         {},
	UserUpdateSettingVerification:             {},
	UserTestEmailPartialName:                  {},
	UserChangeEmailPartialName:                {},
	UserEmailResetPartialName:                 {},
	TwoFactorAuthenticationPartialName:        {},
	UserCryptoReceiptPartialName:              {},
}

const (
	UserVerificationPartialName               = "user_verification"
	UserChangeEmailPartialName                = "user_change_email"
	UserInvitePartialName                     = "user_invite"
	UserAccessKeyChangedPartialName           = "user_access_key_changed"
	UserAuthorizationFromNewDevicePartialName = "user_authorization_from_new_device"
	UserForgotPasswordPartialName             = "user_forgot_password"
	UserRegistrationPartialName               = "user_registration"
	UserExternalWalletPartialName             = "external_wallet_requested"
	UserPasswordChangedPartialName            = "user_password_reset"
	UserEmailResetPartialName                 = "user_email_reset"
	UserRemindVerificationPartialName         = "user_remind_verification"
	UserUpdateSettingVerification             = "user_update_setting_verification"
	UserTestEmailPartialName                  = "user_test_email"
	TwoFactorAuthenticationPartialName        = "two_factor_authentication"
	UserCryptoReceiptPartialName              = "user_crypto_receipt"
)

type IEmailPayload interface {
	GetSubject() string
	GetUserEmail() string
	GetLanguage() string
	GetPayload() []byte
	GetName() string
}

var (
	_ IEmailPayload = (*UserEmailVerification)(nil)
	_ IEmailPayload = (*UserInviteEmail)(nil)
	_ IEmailPayload = (*UserAccessKeyChanged)(nil)
	_ IEmailPayload = (*UserRegistration)(nil)
	_ IEmailPayload = (*UserForgotPassword)(nil)
	_ IEmailPayload = (*UserPasswordChanged)(nil)
	_ IEmailPayload = (*UserExternalWallet)(nil)
	_ IEmailPayload = (*UserRemindVerification)(nil)
	_ IEmailPayload = (*UserChangeEmail)(nil)
	_ IEmailPayload = (*UserCryptoReceiptPayload)(nil)
)

type BasePayload struct {
	Language                  string `json:"language"`
	UserEmail                 string `json:"user_email"`
	DefaultGreeting           string `json:"default_greeting"`
	DefaultRegards            string `json:"default_regards"`
	DefaultSupportInfoText    string `json:"default_support_info_text"`
	DefaultSecurityWarning    string `json:"default_security_warning"`
	DefaultHomeLinkText       string `json:"default_home_link_text"`
	DefaultWarningMessageText string `json:"default_warning_message_text"`
	DefaultSupportActionText  string `json:"default_support_action_text"`
	DefaultFooterHashText     string `json:"default_footer_hash_text"`
	DefaultSupportText        string `json:"default_support_text"`
	DefaultMailText           string `json:"default_mail_text"`
	DefaultTicketText         string `json:"default_ticket_text"`
}

type UserEmailVerification struct {
	BasePayload
	EmailTitle              string `json:"verification_email_title"`
	EmailSubject            string `json:"verification_email_subject"`
	VerificationMessageText string `json:"verification_message_text"`
	VerificationTitle       string `json:"verification_title"`
	VerificationText        string `json:"verification_text"`
	VerifyEmailCode         string `json:"verify_email_code"`
	VerificationMessageTime string `json:"verification_message_time"`
}

func (o *UserEmailVerification) GetPayload() []byte {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(o); err != nil {
		return b.Bytes()
	}
	return b.Bytes()
}

func (o *UserEmailVerification) GetSubject() string   { return o.EmailSubject }
func (o *UserEmailVerification) GetUserEmail() string { return o.UserEmail }
func (o *UserEmailVerification) GetLanguage() string  { return o.Language }
func (o *UserEmailVerification) GetName() string      { return UserVerificationPartialName }

type UserChangeEmail struct {
	BasePayload
	ChangeEmailTitle       string `json:"change_email_title"`
	EmailSubject           string `json:"change_email_subject"`
	ChangeEmailText        string `json:"change_email_text"`
	ChangeEmailCode        string `json:"change_email_code"`
	ChangeEmailMessageText string `json:"change_email_message_text"`
	ChangeEmailMessageTime string `json:"change_email_message_time"`
}

func (o *UserChangeEmail) GetPayload() []byte {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(o); err != nil {
		return b.Bytes()
	}
	return b.Bytes()
}

func (o *UserChangeEmail) GetSubject() string   { return o.EmailSubject }
func (o *UserChangeEmail) GetUserEmail() string { return o.UserEmail }
func (o *UserChangeEmail) GetLanguage() string  { return o.Language }
func (o *UserChangeEmail) GetName() string      { return UserChangeEmailPartialName }

type UserAccessKeyChanged struct {
	BasePayload
	EmailTitle                             string `json:"user_access_key_changed_email_title"`
	EmailSubject                           string `json:"user_access_key_changed_email_subject"`
	UserAccessKeyChangedMessageText        string `json:"user_access_key_changed_message_text"`
	UserAccessKeyChangedSuccessMessageText string `json:"user_access_key_changed_success_message_text"`
}

func (o *UserAccessKeyChanged) GetPayload() []byte {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(o); err != nil {
		return b.Bytes()
	}
	return b.Bytes()
}

func (o *UserAccessKeyChanged) GetSubject() string   { return o.EmailSubject }
func (o *UserAccessKeyChanged) GetUserEmail() string { return o.UserEmail }
func (o *UserAccessKeyChanged) GetLanguage() string  { return o.Language }
func (o *UserAccessKeyChanged) GetName() string      { return UserAccessKeyChangedPartialName }

type UserRegistration struct {
	BasePayload
	EmailTitle                  string `json:"user_registration_email_title"`
	EmailSubject                string `json:"user_registration_email_subject"`
	UserRegistrationMessageText string `json:"user_registration_message_text"`
}

func (o *UserRegistration) GetPayload() []byte {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(o); err != nil {
		return b.Bytes()
	}
	return b.Bytes()
}

func (o *UserRegistration) GetSubject() string   { return o.EmailSubject }
func (o *UserRegistration) GetUserEmail() string { return o.UserEmail }
func (o *UserRegistration) GetLanguage() string  { return o.Language }
func (o *UserRegistration) GetName() string      { return UserRegistrationPartialName }

type UserForgotPassword struct {
	BasePayload
	EmailTitle                          string `json:"user_forgot_password_email_title"`
	EmailSubject                        string `json:"user_forgot_password_email_subject"`
	UserForgotPasswordCode              string `json:"user_forgot_password_code"`
	UserForgotPasswordMessageText       string `json:"user_forgot_password_message_text"`
	UserForgotPasswordActionWarningText string `json:"user_forgot_password_action_warning_text"`
}

func (o *UserForgotPassword) GetPayload() []byte {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(o); err != nil {
		return b.Bytes()
	}
	return b.Bytes()
}

func (o *UserForgotPassword) GetSubject() string   { return o.EmailSubject }
func (o *UserForgotPassword) GetUserEmail() string { return o.UserEmail }
func (o *UserForgotPassword) GetLanguage() string  { return o.Language }
func (o *UserForgotPassword) GetName() string      { return UserForgotPasswordPartialName }

type UserPasswordChanged struct {
	BasePayload
	EmailSubject                   string `json:"user_password_changed_email_subject"`
	EmailTitle                     string `json:"user_password_changed_email_title"`
	UserPasswordChangedMessageText string `json:"user_password_changed_message_text"`
}

func (o *UserPasswordChanged) GetPayload() []byte {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(o); err != nil {
		return b.Bytes()
	}
	return b.Bytes()
}

func (o *UserPasswordChanged) GetSubject() string   { return o.EmailSubject }
func (o *UserPasswordChanged) GetUserEmail() string { return o.UserEmail }
func (o *UserPasswordChanged) GetLanguage() string  { return o.Language }
func (o *UserPasswordChanged) GetName() string      { return UserPasswordChangedPartialName }

type UserEmailReset struct {
	BasePayload
	EmailTitle                string `json:"user_email_reset_email_title"`
	EmailSubject              string `json:"user_email_reset_email_subject"`
	UserEmailResetMessageText string `json:"user_email_reset_message_text"`
	NewEmail                  string `json:"new_email"`
}

func (o *UserEmailReset) GetPayload() []byte {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(o); err != nil {
		return b.Bytes()
	}
	return b.Bytes()
}

func (o *UserEmailReset) GetSubject() string   { return o.EmailSubject }
func (o *UserEmailReset) GetUserEmail() string { return o.UserEmail }
func (o *UserEmailReset) GetLanguage() string  { return o.Language }
func (o *UserEmailReset) GetName() string      { return UserEmailResetPartialName }

type ExternalWallet struct {
	WalletCurrencyName   string `json:"wallet_currency_name"`
	WalletCurrency       string `json:"wallet_currency"`
	WalletBlockchain     string `json:"wallet_blockchain"`
	WalletBlockchainName string `json:"wallet_blockchain_name"`
	WalletAddress        string `json:"wallet_address"`
	WalletIsFirst        bool   `json:"wallet_is_first"`
	ShowBlockchain       bool   `json:"show_blockchain"`
}

type UserExternalWallet struct {
	BasePayload
	EmailTitle                         string           `json:"external_wallet_requested_email_title"`
	EmailSubject                       string           `json:"external_wallet_requested_email_subject"`
	ExternalWalletRequestedMessageText string           `json:"external_wallet_requested_message_text"`
	ExternalWalletOtherWallet          string           `json:"external_wallet_other_wallet"`
	Wallets                            []ExternalWallet `json:"wallets"`
	NotificationHash                   string           `json:"notification_hash"`
}

func (o *UserExternalWallet) GetPayload() []byte {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(o); err != nil {
		return b.Bytes()
	}
	return b.Bytes()
}

func (o *UserExternalWallet) GetSubject() string   { return o.EmailSubject }
func (o *UserExternalWallet) GetUserEmail() string { return o.UserEmail }
func (o *UserExternalWallet) GetLanguage() string  { return o.Language }
func (o *UserExternalWallet) GetName() string      { return UserExternalWalletPartialName }

type UserInviteEmail struct {
	BasePayload
	InviteUserRole                  string `json:"invite_user_role"`
	EmailTitle                      string `json:"invite_user_email_title"`
	EmailSubject                    string `json:"invite_user_email_subject"`
	InviteUserLink                  string `json:"invite_user_link"`
	InviteUserMessageText           string `json:"invite_user_message_text"`
	InviteUserActionText            string `json:"invite_user_action_text"`
	InviteUserAlternativeActionText string `json:"invite_user_alternative_action_text"`
}

func (o *UserInviteEmail) GetPayload() []byte {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(o); err != nil {
		return b.Bytes()
	}
	return b.Bytes()
}

func (o *UserInviteEmail) GetSubject() string   { return o.EmailSubject }
func (o *UserInviteEmail) GetUserEmail() string { return o.UserEmail }
func (o *UserInviteEmail) GetLanguage() string  { return o.Language }
func (o *UserInviteEmail) GetName() string      { return UserInvitePartialName }

type UserAuthorizationFromNewDevice struct {
	BasePayload
	EmailTitle                                     string `json:"user_authorization_from_new_device_email_title"`
	EmailSubject                                   string `json:"user_authorization_from_new_device_email_subject"`
	UserAuthorizationFromNewDeviceMessageText      string `json:"user_authorization_from_new_device_message_text"`
	NewAuthorizationTimestamp                      string `json:"new_authorization_timestamp"`
	UserAuthorizationFromNewDeviceAccountText      string `json:"user_authorization_from_new_device_account_text"`
	NewAuthorizationAccountEmail                   string `json:"new_authorization_account_email"`
	UserAuthorizationFromNewDeviceLocationText     string `json:"user_authorization_from_new_device_location_text"`
	NewAuthorizationLocation                       string `json:"new_authorization_location"`
	UserAuthorizationFromNewDeviceIPText           string `json:"user_authorization_from_new_device_ip_text"`
	NewAuthorizationIP                             string `json:"new_authorization_ip"`
	UserAuthorizationFromNewDeviceTimeText         string `json:"user_authorization_from_new_device_time_text"`
	UserAuthorizationFromNewDeviceUnrecognizedText string `json:"user_authorization_from_new_device_unrecognized_text"`
	UserAuthorizationFromNewDeviceSecureText       string `json:"user_authorization_from_new_device_secure_text"`
	UserAuthorizationFromNewDeviceButtonText       string `json:"user_authorization_from_new_device_button_text"`
	SecureAccountURL                               string `json:"secure_account_url"`
}

func (o *UserAuthorizationFromNewDevice) GetPayload() []byte {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(o); err != nil {
		return b.Bytes()
	}
	return b.Bytes()
}

func (o *UserAuthorizationFromNewDevice) GetSubject() string   { return o.EmailSubject }
func (o *UserAuthorizationFromNewDevice) GetUserEmail() string { return o.UserEmail }
func (o *UserAuthorizationFromNewDevice) GetLanguage() string  { return o.Language }
func (o *UserAuthorizationFromNewDevice) GetName() string {
	return UserAuthorizationFromNewDevicePartialName
}

type UserRemindVerification struct {
	BasePayload
	EmailTitle                    string `json:"remind_verification_email_title"`
	EmailSubject                  string `json:"remind_verification_email_subject"`
	RemindVerificationMessageText string `json:"remind_verification_message_text"`
}

func (o *UserRemindVerification) GetPayload() []byte {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(o); err != nil {
		return b.Bytes()
	}
	return b.Bytes()
}

func (o *UserRemindVerification) GetSubject() string {
	return o.EmailSubject
}
func (o *UserRemindVerification) GetUserEmail() string { return o.UserEmail }
func (o *UserRemindVerification) GetLanguage() string  { return o.Language }
func (o *UserRemindVerification) GetName() string {
	return UserRemindVerificationPartialName
}

type UserUpdateSettingsVerification struct {
	BasePayload
	EmailTitle                           string `json:"setting_update_verification_email_title"`
	EmailSubject                         string `json:"setting_update_verification_email_subject"`
	SettingUpdateVerificationMessageText string `json:"setting_update_verification_message_text"`
	Code                                 int    `json:"code"`
}

func (o *UserUpdateSettingsVerification) GetPayload() []byte {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(o); err != nil {
		return b.Bytes()
	}
	return b.Bytes()
}

func (o *UserUpdateSettingsVerification) GetSubject() string {
	return o.EmailSubject
}
func (o *UserUpdateSettingsVerification) GetUserEmail() string { return o.UserEmail }
func (o *UserUpdateSettingsVerification) GetLanguage() string  { return o.Language }
func (o *UserUpdateSettingsVerification) GetName() string {
	return UserUpdateSettingVerification
}

type UserTestEmail struct {
	BasePayload
	EmailTitle                string `json:"user_test_email_title"`
	EmailSubject              string `json:"user_test_email_subject"`
	UserTestEmailMessageText  string `json:"user_test_email_message_text"`
	UserTestEmailNoActionText string `json:"user_test_email_no_action_text"`
}

func (o *UserTestEmail) GetPayload() []byte {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(o); err != nil {
		return b.Bytes()
	}
	return b.Bytes()
}

func (o *UserTestEmail) GetSubject() string   { return o.EmailSubject }
func (o *UserTestEmail) GetUserEmail() string { return o.UserEmail }
func (o *UserTestEmail) GetLanguage() string  { return o.Language }
func (o *UserTestEmail) GetName() string      { return UserTestEmailPartialName }

type TwoFactorAuthentication struct {
	BasePayload
	IsEnabled                              bool   `json:"is_enabled"`
	EmailTitle                             string `json:"two_factor_authentication_email_title"`
	EmailSubject                           string `json:"two_factor_authentication_email_subject"`
	TwoFactorAuthenticationEnabledTitle    string `json:"two_factor_authentication_enabled_title"`
	TwoFactorAuthenticationDisabledTitle   string `json:"two_factor_authentication_disabled_title"`
	TwoFactorAuthenticationEnabledMessage  string `json:"two_factor_authentication_enabled_message"`
	TwoFactorAuthenticationDisabledMessage string `json:"two_factor_authentication_disabled_message"`
}

func (o *TwoFactorAuthentication) GetPayload() []byte {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(o); err != nil {
		return b.Bytes()
	}
	return b.Bytes()
}

func (o *TwoFactorAuthentication) GetSubject() string   { return o.EmailSubject }
func (o *TwoFactorAuthentication) GetUserEmail() string { return o.UserEmail }
func (o *TwoFactorAuthentication) GetLanguage() string  { return o.Language }
func (o *TwoFactorAuthentication) GetName() string      { return TwoFactorAuthenticationPartialName }

// CryptoReceiptI18N represents the nested i18n structure for crypto receipts
type CryptoReceiptI18N struct {
	Email struct {
		Title   string `json:"title"`
		Subject string `json:"subject"`
	} `json:"email"`
	TitleLabel    string `json:"title_label"`
	PaymentStatus struct {
		Pending   string `json:"pending"`
		Completed string `json:"completed"`
		Failed    string `json:"failed"`
	} `json:"payment_status"`
	PaymentType struct {
		Deposit  string `json:"deposit"`
		Transfer string `json:"transfer"`
	} `json:"payment_type"`
	PaymentStatusLabel     string `json:"payment_status_label"`
	PaymentIdLabel         string `json:"payment_id_label"` //nolint:revive
	PaymentDateLabel       string `json:"payment_date_label"`
	PaymentTypeLabel       string `json:"payment_type_label"`
	PaymentBlockchainLabel string `json:"payment_blockchain_label"`
	TransactionHashLabel   string `json:"transaction_hash_label"`
	Payment                struct {
		AmountLabel         string `json:"amount_label"`
		NetworkLabelPrefix  string `json:"network_label_prefix"`
		NetworkLabelPostfix string `json:"network_label_postfix"`
		Exchange            struct {
			ExchangeRateLabel string `json:"exchange_rate_label"`
		} `json:"exchange"`
		Fees struct {
			FreeTierText string `json:"free_tier_text"`
			Platform     struct {
				PlatformFeeLabel string `json:"platform_fee_label"`
			} `json:"platform"`
			Network struct {
				NetworkFeeLabel string `json:"network_fee_label"`
			} `json:"network"`
		} `json:"fees"`
	} `json:"payment"`
}

type UserCryptoReceiptPayload struct {
	BasePayload

	// i18n nested structure (populated from localization files)
	UserCryptoReceipt CryptoReceiptI18N `json:"user_crypto_receipt"`

	// Runtime data fields (generated in code)
	PaymentStatus        string `json:"payment_status"`
	PaymentType          string `json:"payment_type"`
	ReceiptId            string `json:"receipt_id"` //nolint:revive
	PaymentDate          string `json:"payment_date"`
	UsdAmount            string `json:"usd_amount"`
	TokenAmount          string `json:"token_amount"`
	TokenSymbol          string `json:"token_symbol"`
	TransactionHash      string `json:"transaction_hash"`
	BlockchainName       string `json:"blockchain_name"`
	BlockchainCurrencyID string `json:"blockchain_currency_id"`
	ExchangeRate         string `json:"exchange_rate"`
	NetworkFeeAmount     string `json:"network_fee_amount"`
	NetworkFeeCurrency   string `json:"network_fee_currency"`
	NetworkFeeUSD        string `json:"network_fee_usd"`
	PlatformFeeAmount    string `json:"platform_fee_amount"`
	PlatformFeeUSD       string `json:"platform_fee_usd"`
	PlatformFeeCurrency  string `json:"platform_fee_currency"`
}

// Simple methods for dynamic text resolution (called by mustache as simple methods)
func (o UserCryptoReceiptPayload) PaymentStatusText() string {
	switch o.PaymentStatus {
	case "completed":
		return o.UserCryptoReceipt.PaymentStatus.Completed
	case "pending":
		return o.UserCryptoReceipt.PaymentStatus.Pending
	case "failed":
		return o.UserCryptoReceipt.PaymentStatus.Failed
	default:
		return o.UserCryptoReceipt.PaymentStatus.Pending // fallback
	}
}

func (o UserCryptoReceiptPayload) PaymentTypeText() string {
	switch o.PaymentType {
	case "deposit":
		return o.UserCryptoReceipt.PaymentType.Deposit
	case "transfer":
		return o.UserCryptoReceipt.PaymentType.Transfer
	default:
		return o.UserCryptoReceipt.PaymentType.Deposit
	}
}

// Boolean helper methods for mustache conditional sections
func (o UserCryptoReceiptPayload) IsCompleted() bool {
	return o.PaymentStatus == "completed"
}

func (o UserCryptoReceiptPayload) IsPending() bool {
	return o.PaymentStatus == "pending"
}

func (o UserCryptoReceiptPayload) IsFailed() bool {
	return o.PaymentStatus == "failed"
}

// RenderPlatformFreeTierFee returns the platform fee text or free tier text if fee is zero
func (o UserCryptoReceiptPayload) RenderPlatformFreeTierFee() bool {
	platformFee, err := decimal.NewFromString(o.PlatformFeeAmount)
	if err != nil {
		return true
	}
	platformFeeUSD, err := decimal.NewFromString(o.PlatformFeeUSD)
	if err != nil {
		return true
	}
	if platformFee.IsZero() || platformFeeUSD.IsZero() {
		return true
	}
	return false
}

// RenderPlatformFreeTierFee returns the network fee text or free tier text if fee is zero
func (o UserCryptoReceiptPayload) RenderNetworkFreeTierFee() bool {
	networkFee, err := decimal.NewFromString(o.NetworkFeeAmount)
	if err != nil {
		return true
	}
	networkFeeUSD, err := decimal.NewFromString(o.NetworkFeeUSD)
	if err != nil {
		return true
	}
	if networkFee.IsZero() || networkFeeUSD.IsZero() {
		return true
	}
	return false
}

func (o *UserCryptoReceiptPayload) GetPayload() []byte {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(o); err != nil {
		return b.Bytes()
	}
	return b.Bytes()
}

func (o *UserCryptoReceiptPayload) GetSubject() string   { return o.UserCryptoReceipt.Email.Subject }
func (o *UserCryptoReceiptPayload) GetUserEmail() string { return o.UserEmail }
func (o *UserCryptoReceiptPayload) GetLanguage() string  { return o.Language }
func (o *UserCryptoReceiptPayload) GetName() string      { return UserCryptoReceiptPartialName }
