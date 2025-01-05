package evbus

import (
	"context"
	"errors"
	"reflect"
	"sync"

	"github.com/kijudev/go-event-bus/typeid"
)

type EventTag reflect.Type
type HandlerTag uint64

type EventBus struct {
	store  *store
	reader *EventReader
	writer *EventWriter
}

type EventReader struct {
	store               *store
	funcTypeidGenerator *typeid.FuncGenerator
}

type EventWriter struct {
	store               *store
	funcTypeidGenerator *typeid.FuncGenerator
}

type store struct {
	guard chan struct{}
	mu    sync.RWMutex
	wg    sync.WaitGroup

	handlers map[HandlerTag]storeHandlerData
	events   map[EventTag]storeEventData

	funcTypeidGenerator *typeid.FuncGenerator
}

type storeHandlerData struct {
	rvalue       reflect.Value
	isSubscribed bool
	evtag        EventTag
}

type storeEventData struct {
	handlers map[HandlerTag]struct{}
}

func NewEventBus(maxGoroutines uint) *EventBus {
	if maxGoroutines == 0 {
		maxGoroutines = 1
	}

	funcTypeidGenerator := typeid.NewFuncGenerator()

	store := &store{
		guard: make(chan struct{}, maxGoroutines),

		handlers: make(map[HandlerTag]storeHandlerData),
		events:   make(map[EventTag]storeEventData),

		funcTypeidGenerator: funcTypeidGenerator,
	}

	evbus := &EventBus{
		store: store,
		reader: &EventReader{
			store:               store,
			funcTypeidGenerator: funcTypeidGenerator,
		},
		writer: &EventWriter{
			store:               store,
			funcTypeidGenerator: funcTypeidGenerator,
		},
	}

	return evbus
}

func (bus *EventBus) Reader() *EventReader {
	return bus.reader
}

func (bus *EventBus) Writer() *EventWriter {
	return bus.writer
}

func (s *store) extractEvent(ev any) (EventTag, reflect.Value, error) {
	evt := elemT(reflect.TypeOf(ev))
	evv := elemV(reflect.ValueOf(ev))

	s.mu.RLock()
	if _, ok := s.events[EventTag(evt)]; !ok {
		s.mu.RUnlock()
		return evt, evv, ErrEventNotRegistered
	}
	s.mu.RUnlock()

	return evt, evv, nil
}

func (s *store) extractHandler(handler any) (HandlerTag, reflect.Value, error) {
	ht := reflect.TypeOf(handler)
	hv := reflect.ValueOf(handler)

	if hv.IsNil() {
		return 0, hv, ErrHandlerInvalid
	}

	if ht.Kind() != reflect.Func {
		return 0, hv, ErrHandlerInvalid
	}

	if ht.NumIn() != 3 {
		return 0, hv, ErrHandlerInvalid
	}

	if ht.In(0) != reflect.TypeFor[context.Context]() {
		return 0, hv, ErrHandlerInvalid
	}

	if ht.In(2) != reflect.TypeFor[EventWriter]() {
		return 0, hv, ErrHandlerInvalid
	}

	evt := ht.In(1)
	if evt.Kind() == reflect.Ptr {
		return 0, hv, ErrHandlerInvalidEventTag
	}

	evtag := EventTag(evt)

	s.mu.RLock()
	if _, ok := s.events[evtag]; !ok {
		s.mu.RUnlock()
		return 0, hv, ErrEventNotRegistered
	}
	s.mu.RUnlock()

	genid, err := s.funcTypeidGenerator.GenID(handler)
	if err != nil {
		return 0, hv, errors.Join(ErrUnknown, err)
	}

	htag := HandlerTag(genid)

	return htag, hv, nil
}

func elemV(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		return v.Elem()
	}

	return v
}

func elemT(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}

	return t
}
