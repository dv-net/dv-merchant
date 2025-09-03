package settings

type MailerSettings struct {
	MailerState      string `mapstructure:"mailer_state"`
	MailerProtocol   string `mapstructure:"mailer_protocol"`
	MailerAddress    string `mapstructure:"mailer_address"`
	MailerSender     string `mapstructure:"mailer_sender"`
	MailerUsername   string `mapstructure:"mailer_username"`
	MailerPassword   string `mapstructure:"mailer_password"`
	MailerEncryption string `mapstructure:"mailer_encryption"`
}
