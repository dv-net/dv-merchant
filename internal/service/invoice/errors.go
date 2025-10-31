package invoice

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrInvoiceExpired      = errors.New("invoice expired")
	ErrInvoiceAlreadyPaid  = errors.New("invoice already paid")
	ErrInvoiceAmountTooLow = errors.New("invoice amount too low")
)

func NewErrInvoiceNotFound(invoiceID uuid.UUID) error {
	return fmt.Errorf("invoice not found: %s", invoiceID)
}
