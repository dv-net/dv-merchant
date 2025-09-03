package apierror

type Option func(*Error)

func WithField(v string) Option {
	return func(e *Error) { e.Field = v }
}
