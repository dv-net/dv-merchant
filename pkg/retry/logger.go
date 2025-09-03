package retry

type logger interface {
	Infof(format string, args ...any)
}
