package setting

import (
	"fmt"

	"github.com/dv-net/dv-merchant/internal/event"
)

const (
	RootSettingsChanged = "SettingChanged"
)

type RootSettingChangedEvent struct {
	SettingName  string
	SettingValue string
}

func (e RootSettingChangedEvent) Type() event.Type {
	return RootSettingsChanged
}

func (e RootSettingChangedEvent) String() string {
	return fmt.Sprintf("SettingChanged: name=%s, value=%s", e.SettingName, e.SettingValue)
}

const (
	MailerSettingsChanged = "MailerSettingChanged"
)

type MailerSettingChangedEvent struct {
	SettingName  string
	SettingValue string
}

func (e MailerSettingChangedEvent) Type() event.Type {
	return MailerSettingsChanged
}

func (e MailerSettingChangedEvent) String() string {
	return fmt.Sprintf("MailerSetting: name=%s, value=%s", e.SettingName, e.SettingValue)
}

const (
	NotificationSenderChanged = "NotificationSenderChanged"
)

type NotificationSenderChangedEvent struct {
	NewValue string
}

func (e NotificationSenderChangedEvent) Type() event.Type {
	return NotificationSenderChanged
}

func (e NotificationSenderChangedEvent) String() string {
	return fmt.Sprintf("Notification sender setting chnged: value=%s", e.NewValue)
}
