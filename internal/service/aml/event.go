package aml

import (
	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/models"
	amlproviders "github.com/dv-net/dv-merchant/pkg/aml"
)

const CheckCompletedEventType = "aml_check_completed"

type CheckCompletedEvent struct {
	Check   models.AmlCheck
	Signals []amlproviders.SignalContribution
}

func (e CheckCompletedEvent) Type() event.Type {
	return CheckCompletedEventType
}

func (e CheckCompletedEvent) String() string {
	return "aml_check_completed: " + e.Check.ID.String()
}
