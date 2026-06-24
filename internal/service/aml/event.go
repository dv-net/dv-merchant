package aml

import (
	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/models"
)

const CheckCompletedEventType = "aml_check_completed"

type CheckCompletedEvent struct {
	Check models.AmlCheck
}

func (e CheckCompletedEvent) Type() event.Type {
	return CheckCompletedEventType
}

func (e CheckCompletedEvent) String() string {
	return "aml_check_completed: " + e.Check.ID.String()
}
