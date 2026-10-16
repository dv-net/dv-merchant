package refund_request

import "errors"

type ApproveRefundRequest struct {
	TOTP string `json:"totp" validate:"required,min=6,max=6"`
} //	@name	ApproveRefundRequest

func (req *ApproveRefundRequest) Validate() error {
	if len(req.TOTP) != 6 {
		return errors.New("TOTP must be exactly 6 characters long")
	}

	return nil
}
