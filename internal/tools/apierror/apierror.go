package apierror

import "github.com/goccy/go-json"

type Error struct {
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
} // @name APIError

type Errors struct {
	Errors   []Error `json:"errors"`
	HttpCode int     `json:"code"`
} // @name APIErrors

func New(errs ...Error) *Errors {
	if len(errs) == 0 {
		errs = []Error{}
	}

	return &Errors{Errors: errs}
}

func (ae *Errors) AddError(err error, opts ...Option) *Errors {
	e := Error{Message: err.Error()}
	for _, opt := range opts {
		opt(&e)
	}

	ae.Errors = append(ae.Errors, e)
	return ae
}

func (ae Errors) Error() string {
	res, _ := json.Marshal(ae)
	return string(res)
}

func (ae *Errors) SetHttpCode(code int) *Errors {
	ae.HttpCode = code
	return ae
}
