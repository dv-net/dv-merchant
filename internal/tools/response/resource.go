package response

import (
	"github.com/gofiber/fiber/v3"
)

type Result[T any] struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data"`
} // @name JSONResponse

func Ok[T any](message string, data T) *Result[T] {
	return &Result[T]{
		Code:    fiber.StatusOK,
		Message: message,
		Data:    data,
	}
}

func OkByMessage(message string) *Result[any] {
	return &Result[any]{
		Code:    fiber.StatusOK,
		Message: message,
	}
}

func OkByData[T any](data T) *Result[T] {
	return &Result[T]{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data:    data,
	}
}

func Fail(code int, message string) *Result[any] {
	return &Result[any]{
		Code:    code,
		Message: message,
	}
}

type ValidationErrorDetail struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	Message string                             `json:"message"`
	Errors  map[string][]ValidationErrorDetail `json:"errors"`
}

func ValidationError(message string, errors map[string][]ValidationErrorDetail) ValidationErrorResponse {
	return ValidationErrorResponse{
		Message: message,
		Errors:  errors,
	}
}

type EmptyBodyResponse struct{} // @name EmptyBodyResponse
