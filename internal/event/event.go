package event

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"sync"

	"github.com/google/uuid"
)

type (
	Type string // Event type

	IEvent interface {
		Type() Type
		String() string
	} // Event interface

	// IListener Event Listener Interface
	//
	// The implementation of asynchronous event processing is assigned
	// to the author of the event handler.
	// For example, a handler can simply send an event to a channel that is being
	// listened to by one or more goroutines performing valid event processing.
	IListener interface {
		// Register registers an event handler
		Register(Type, Handler) HandlerID
		// Unregister deletes the event handler with the specified ID
		Unregister(HandlerID) error
		// Fire starts event processing
		Fire(IEvent) error
	}
)

// New creates new event listener
func New() *Listener {
	return &Listener{
		handlers: make(map[HandlerID]handlerInfo, 32),
		types:    make(map[Type]HandlerIDList, 16),
	}
}

type Listener struct {
	handlers map[HandlerID]handlerInfo
	types    map[Type]HandlerIDList
	mutex    sync.RWMutex
}

var _ IListener = (*Listener)(nil)

type handlerInfo struct {
	eventType Type
	handler   Handler
}

func (l *Listener) Register(t Type, h Handler) HandlerID {
	ID := HandlerID(uuid.New())
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.handlers[ID] = handlerInfo{
		eventType: t,
		handler:   h,
	}
	l.types[t] = append(l.types[t], ID)
	sort.Sort(l.types[t])
	return ID
}

func (l *Listener) Unregister(id HandlerID) error {
	var (
		info handlerInfo
		ok   bool
	)
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if info, ok = l.handlers[id]; !ok {
		return fmt.Errorf("handler[%s] not registered", id)
	}
	delete(l.handlers, id)
	if t, ok := l.types[info.eventType]; ok {
		if index, ok := sort.Find(t.Len(), func(i int) int {
			return -t.Cmp(i, id)
		}); ok {
			t = slices.Delete(t, index, index+1)
			l.types[info.eventType] = t
		}
	}
	return nil
}

func (l *Listener) Fire(ev IEvent) error {
	var errs []error
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	if t, ok := l.types[ev.Type()]; ok {
		for _, ID := range t {
			if h, ok := l.handlers[ID]; ok {
				errs = append(errs, h.handler(ev))
			} else {
				errs = append(errs, fmt.Errorf("internal error: handler[%s] not exists", ID))
			}
		}
	}
	return errors.Join(errs...)
}
