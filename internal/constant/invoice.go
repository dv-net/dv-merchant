package constant

type InvoiceStatus string

const (
	InvoiceStatusPending   InvoiceStatus = "pending"
	InvoiceStatusWaiting   InvoiceStatus = "waiting_confirmation"
	InvoiceStausPaid       InvoiceStatus = "paid"
	InvoiceStatusExpired   InvoiceStatus = "expired"
	InvoiceStatusUnderpaid InvoiceStatus = "underpaid"
	InvoiceStatusOverpaid  InvoiceStatus = "overpaid"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
)

func (s InvoiceStatus) CanReleaseWallet() bool {
	switch s {
	case InvoiceStausPaid,
		InvoiceStatusExpired,
		InvoiceStatusCancelled,
		InvoiceStatusOverpaid:
		return true
	case InvoiceStatusPending,
		InvoiceStatusWaiting,
		InvoiceStatusUnderpaid:
		return false
	default:
		return false
	}
}
