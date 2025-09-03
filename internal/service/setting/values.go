package setting

type TransferTypeValue string

func (o TransferTypeValue) String() string { return string(o) }

const (
	TransferByBurnTRX       TransferTypeValue = "burntrx"
	TransferByResource      TransferTypeValue = "resources"
	TransferByCloudDelegate TransferTypeValue = "cloud_delegate"
)

type FlagValue string

const (
	FlagValueEnabled  = "enabled"
	FlagValueDisabled = "disabled"
)

const (
	MailerEncryptionTypeNone = "none"
	MailerEncryptionTypeTLS  = "tls"
)

const (
	NotificationSenderInternal = "internal"
	NotificationSenderDVNet    = "dv_net"
)

const (
	FlagValueCompleted   = "completed"
	FlagValueIncompleted = "incompleted"
)

const (
	TransferStatusSystemSuspended string = "system_suspended"
)
