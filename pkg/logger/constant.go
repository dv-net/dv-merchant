package logger

type LogStatus string

const (
	InProgress LogStatus = "in_progress"
	Completed  LogStatus = "completed"
	Failed     LogStatus = "failed"
)

func (s LogStatus) String() string {
	return string(s)
}

const LogBufferSize = 1000
