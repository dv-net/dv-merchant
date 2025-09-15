package notify

import (
	"bytes"
	"encoding/json"
)

type INotificationBody interface {
	Encode() ([]byte, error)
}

type NotificationBody[T any] struct {
	Body T
}

type UserRegistration struct {
	Language string `json:"language"`
}

func (d *UserRegistration) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(d); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type UserVerificationData struct {
	Language string `json:"language"`
	Code     int    `json:"code"`
}

func (d *UserVerificationData) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(d); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type UserTwoFactorAuth struct {
	Language   string `json:"language"`
	VerifyLink string `json:"verify_link"`
}

func (d *UserTwoFactorAuth) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(d); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type PasswordForgotData struct {
	Language string `json:"language"`
	LinkText string `json:"link_text"`
	Link     string `json:"link"`
	Code     int    `json:"code"`
}

func (d *PasswordForgotData) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(d); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type ExternalWalletRequestedData struct {
	Language         string      `json:"language"`
	Addresses        []WalletDTO `json:"addresses"`
	NotificationHash string      `json:"notification_hash"`
}

type WalletDTO struct {
	CurrencyID     string `json:"currency_id"`
	CurrencyName   string `json:"currency_name"`
	BlockchainID   string `json:"blockchain_id"`
	BlockchainName string `json:"blockchain_name"`
	ShowBlockchain bool   `json:"show_blockchain"`
	Address        string `json:"address"`
	IsFirst        bool   `json:"is_first"`
}

func (d *ExternalWalletRequestedData) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(d); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type UserPasswordChanged struct {
	Language string `json:"language"`
}

func (d *UserPasswordChanged) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(d); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type UserForgotPassword struct {
	Language          string `json:"language"`
	ResetPasswordCode string `json:"reset_password_code"`
}

func (d *UserForgotPassword) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(d); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type UserPasswordReset struct {
	Language          string `json:"language"`
	ResetPasswordCode string `json:"reset_password_code"`
}

func (d *UserPasswordReset) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(d); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type UserInviteData struct {
	Language string `json:"language"`
	Link     string `json:"link"`
	Role     string `json:"role"`
}

func (d *UserInviteData) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(d); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type UserInitEmailChangeData struct {
	Language string `json:"language"`
	Code     int    `json:"code"`
}

func (o *UserInitEmailChangeData) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(o); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type EmailChangeConfirmData struct {
	Language string `json:"language"`
	Email    string `json:"email"`
	NewEmail string `json:"new_email"`
}

func (d *EmailChangeConfirmData) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(d); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type UserAccessKeyChanged struct {
	Language string `json:"language"`
}

func (d *UserAccessKeyChanged) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(d); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type UserAuthorizationFromNewDevice struct {
	Language                     string `json:"language"`
	NewAuthorizationIP           string `json:"new_authorization_ip"`
	NewAuthorizationLocation     string `json:"new_authorization_location"`
	NewAuthorizationAccountEmail string `json:"new_authorization_account_email"`
	NewAuthorizationTimestamp    string `json:"new_authorization_timestamp"`
}

func (d *UserAuthorizationFromNewDevice) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(d); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type UserRemindVerificationData struct {
	Language string `json:"language"`
}

func (d *UserRemindVerificationData) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(d); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type UserUpdateSettingData struct {
	Code       int    `json:"code"`
	Language   string `json:"language"`
	EmailTitle string `json:"email_title"`
}

func (d *UserUpdateSettingData) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(d); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type UserTestEmailData struct {
	Language string `json:"language"`
}

func (d *UserTestEmailData) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(d); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// TwoFactorAuthenticationNotification is sent when a user changes their 2FA status
type TwoFactorAuthenticationNotification struct {
	Email     string `json:"email"`
	Language  string `json:"language"`
	IsEnabled bool   `json:"is_enabled"`
}

func (d *TwoFactorAuthenticationNotification) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(d); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UserAccessKeyChangedNotification is sent when user's API keys are changed
type UserAccessKeyChangedNotification struct {
	Email    string `json:"email"`
	Language string `json:"language"`
}

func (d *UserAccessKeyChangedNotification) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(d); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ParseNotificationBody[T any](data []byte) (T, error) {
	t := NotificationBody[T]{}
	v := &t.Body
	err := json.Unmarshal(data, v)
	if err != nil {
		return *v, err
	}

	return *v, nil
}

type PaymentStatus string

func (o PaymentStatus) String() string { return string(o) }

const (
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusFailed    PaymentStatus = "failed"
)

type PaymentType string

func (o PaymentType) String() string { return string(o) }

const (
	PaymentTypeDeposit  PaymentType = "deposit"
	PaymentTypeTransfer PaymentType = "transfer"
)

// UserCryptoReceiptNotificationData is sent when a user makes a crypto payment
type UserCryptoReceiptNotificationData struct {
	Email                string `json:"email"`
	Language             string `json:"language"`
	PaymentStatus        string `json:"payment_status"` // "completed", "pending", "failed"
	PaymentType          string `json:"payment_type"`   // "deposit", "transfer", etc.
	ReceiptId            string `json:"receipt_id"`     //nolint:revive
	PaymentDate          string `json:"payment_date"`
	UsdAmount            string `json:"usd_amount"`
	TokenAmount          string `json:"token_amount"`
	TokenSymbol          string `json:"token_symbol"`
	TransactionHash      string `json:"transaction_hash"`
	BlockchainCurrencyID string `json:"blockchain_currency_id"`
	BlockchainName       string `json:"blockchain_name"`
	ExchangeRate         string `json:"exchange_rate"`
	NetworkFeeCurrency   string `json:"network_fee_currency"`
	NetworkFeeAmount     string `json:"network_fee_amount"`
	NetworkFeeUSD        string `json:"network_fee_usd"`
	PlatformFeeAmount    string `json:"platform_fee_amount"`
	PlatformFeeUSD       string `json:"platform_fee_usd"`
	PlatformFeeCurrency  string `json:"platform_fee_currency"`
}

func (d *UserCryptoReceiptNotificationData) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(d); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
