package migrations

type Logger interface {
	Infof(format string, args ...interface{})
}

type logger struct {
	log Logger
}

func newLogger(log Logger) *logger {
	return &logger{
		log: log,
	}
}

// Printf logs a message at the info level.
func (l *logger) Printf(format string, args ...interface{}) {
	l.log.Infof(format, args...)
}

func (l *logger) Verbose() bool { return false }
