package repo_multi_withdrawal_rules

import (
	"errors"

	"github.com/dv-net/dv-merchant/internal/models"
)

func (p CreateOrUpdateParams) Validate() error {
	if p.Mode == models.MultiWithdrawalModeManual && !p.ManualAddress.Valid {
		return errors.New("manual address is required")
	}

	return nil
}
