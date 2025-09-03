package models

type MessageLogLevel string

const (
	MessageLogLevelInfo    MessageLogLevel = "info"
	MessageLogLevelError   MessageLogLevel = "error"
	MessageLogLevelNotice  MessageLogLevel = "notice"
	MessageLogLevelWarning MessageLogLevel = "warning"
	MessageLogLevelSuccess MessageLogLevel = "success"
)
