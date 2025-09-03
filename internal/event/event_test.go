package event_test

import (
	"fmt"
	"testing"

	"github.com/dv-net/dv-merchant/internal/event"

	"github.com/stretchr/testify/assert"
)

// event1
type event1 struct {
	handler string
}

var _ event.IEvent = (*event1)(nil)

func (e event1) Type() event.Type {
	return "event1"
}

func (e event1) String() string {
	return fmt.Sprintf("%s{%s}", e.Type(), e.handler)
}

// event2
type event2 struct {
	handler string
}

var _ event.IEvent = (*event2)(nil)

func (e event2) Type() event.Type {
	return "event2"
}

func (e event2) String() string {
	return fmt.Sprintf("%s{%s}", e.Type(), e.handler)
}

func TestEventListener(t *testing.T) {
	handledEvents := []event.IEvent{}
	// Listener
	eventListener := event.New()
	// Register
	hidEv1n1 := eventListener.Register("event1", func(ev event.IEvent) error {
		if ev, ok := ev.(event1); ok {
			ev.handler = "ev1:h1"
			handledEvents = append(handledEvents, ev)
			return nil
		}
		return fmt.Errorf("unexpected event object: %s", ev)
	})
	_ = eventListener.Register("event1", func(ev event.IEvent) error {
		if ev, ok := ev.(event1); ok {
			ev.handler = "ev1:h2"
			handledEvents = append(handledEvents, ev)
			return nil
		}
		return fmt.Errorf("unexpected event object: %s", ev)
	})
	hidEv2n1 := eventListener.Register("event2", func(ev event.IEvent) error {
		if ev, ok := ev.(event2); ok {
			ev.handler = "ev2:h1"
			handledEvents = append(handledEvents, ev)
			return nil
		}
		return fmt.Errorf("unexpected event object: %s", ev)
	})
	// Expected
	expectedEvents := []event.IEvent{}
	var err error
	// Fire event1
	expectedEvents = append(expectedEvents, event1{handler: "ev1:h1"})
	expectedEvents = append(expectedEvents, event1{handler: "ev1:h2"})
	err = eventListener.Fire(event1{})
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedEvents, handledEvents)
	// Fire event2
	expectedEvents = append(expectedEvents, event2{handler: "ev2:h1"})
	err = eventListener.Fire(event2{})
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedEvents, handledEvents)
	// Unregister hidEv2n1
	err = eventListener.Unregister(hidEv2n1)
	assert.NoError(t, err)
	// Fire event2
	err = eventListener.Fire(event2{})
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedEvents, handledEvents)
	// Fire event1
	expectedEvents = append(expectedEvents, event1{handler: "ev1:h1"})
	expectedEvents = append(expectedEvents, event1{handler: "ev1:h2"})
	err = eventListener.Fire(event1{})
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedEvents, handledEvents)
	// Unregister hidEv1n1
	err = eventListener.Unregister(hidEv1n1)
	assert.NoError(t, err)
	// Fire event1
	expectedEvents = append(expectedEvents, event1{handler: "ev1:h2"})
	err = eventListener.Fire(event1{})
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedEvents, handledEvents)
}
