package tools

import "github.com/gofiber/fiber/v3"

func FiberRequestError(message string) *fiber.Error {
	return fiber.NewError(fiber.StatusBadRequest, message)
}

func FiberAuthError(message string) *fiber.Error {
	return fiber.NewError(fiber.StatusUnauthorized, message)
}

func FiberServerError(message string) *fiber.Error {
	return fiber.NewError(fiber.StatusInternalServerError, message)
}

func FiberUnprocessableEntity(message string) *fiber.Error {
	return fiber.NewError(fiber.StatusUnprocessableEntity, message)
}

func FiberNotFound(message string) *fiber.Error {
	return fiber.NewError(fiber.StatusNotFound, message)
}
