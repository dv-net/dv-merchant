package transactions

import "errors"

var ErrTransactionNotFound = errors.New("transaction not found")
var ErrTransactionAlreadyExists = errors.New("transaction already exists")
