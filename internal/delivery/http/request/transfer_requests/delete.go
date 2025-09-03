package transfer_requests

import "github.com/google/uuid"

type DeleteTransferRequest struct {
	ID []uuid.UUID `json:"id" format:"uuid" validate:"required,gt=0,dive,required,uuid"`
} // @name DeleteTransferRequest
