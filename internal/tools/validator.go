package tools

import (
	"errors"
	"reflect"
	"strings"

	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	enTranslations "github.com/go-playground/validator/v10/translations/en"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var defaultStructValidator *StructValidator

type StructValidator struct {
	validate *validator.Validate
	trans    ut.Translator
}
type Validatable interface {
	Validate() error
}

func (s *StructValidator) Engine() any {
	return s.validate
}

func (s *StructValidator) Validate(out any) error {
	err := s.validate.Struct(out)
	if err == nil {
		return nil
	}
	var validateErrors validator.ValidationErrors
	if !errors.As(err, &validateErrors) || len(validateErrors) == 0 {
		return createAPIError("Struct parameter error", "", fiber.StatusBadRequest)
	}

	apiErrors := make([]apierror.Error, 0, len(validateErrors))
	for _, validateErr := range validateErrors {
		apiErrors = append(apiErrors, apierror.Error{
			Message: validateErr.Translate(s.trans),
			Field:   validateErr.Field(),
		})
	}

	apiErr := apierror.New(apiErrors...)
	_ = apiErr.SetHttpCode(fiber.StatusUnprocessableEntity)
	res, _ := json.Marshal(apiErr)
	return fiber.NewError(fiber.StatusUnprocessableEntity, string(res))
}

func init() {
	defaultStructValidator = newStruckValidator()
}

func createAPIError(message, field string, code int) error {
	apiErr := apierror.New(apierror.Error{
		Message: message,
		Field:   field,
	})
	_ = apiErr.SetHttpCode(code)
	res, _ := json.Marshal(apiErr)
	return fiber.NewError(code, string(res))
}

func newStruckValidator() *StructValidator {
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	trans, _ := uni.GetTranslator("en")
	validate := validator.New()

	// TODO remove after issue fix https://github.com/go-playground/validator/issues/935
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	validate.RegisterCustomTypeFunc(func(v reflect.Value) any {
		if value, ok := v.Interface().(decimal.Decimal); ok {
			return value.String()
		}

		return nil
	}, decimal.Decimal{})

	if err := validate.RegisterValidation("decimal_gte", func(fl validator.FieldLevel) bool {
		data, ok := fl.Field().Interface().(string)
		if !ok {
			return false
		}

		value, err := decimal.NewFromString(data)
		if err != nil {
			return false
		}

		baseValue, err := decimal.NewFromString(fl.Param())
		if err != nil {
			return false
		}

		return value.GreaterThanOrEqual(baseValue)
	}); err != nil {
		panic(err)
	}

	if err := validate.RegisterTranslation("decimal_gte", trans, func(ut ut.Translator) error {
		return ut.Add("decimal_gte", "{0} must be greater than or equal to {1}", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("decimal_gte", fe.Field(), fe.Param())
		return t
	}); err != nil {
		panic(err)
	}

	_ = enTranslations.RegisterDefaultTranslations(validate, trans)

	return &StructValidator{
		validate: validate,
		trans:    trans,
	}
}

func DefaultStructValidator() *StructValidator {
	return defaultStructValidator
}

func ValidateUUID(id string) (uuid.UUID, error) {
	if len(id) != 36 {
		return uuid.Nil, apierror.New().AddError(errors.New("invalid UUID length")).SetHttpCode(fiber.StatusBadRequest)
	}
	uuidParsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return uuidParsed, nil
}
