package models

type ReceiptStatus string // @name ReceiptStatus

const (
	ReceiptStatusPaid     ReceiptStatus = "paid"
	ReceiptStatusCanceled ReceiptStatus = "canceled"
)
