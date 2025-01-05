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
	store      *store
	subscriber *EventSubscriber
	dispatcher *EventDispatcher
}

type EventSubscriber struct {
	store *store
}

type EventDispatcher struct {
	store *store
}

type store struct {
	guard chan struct{}
	mu    sync.RWMutex
	wg    sync.WaitGroup

	handlers map[HandlerTag]storeHandlerData
	events   map[EventTag]storeEventData

	funcTypeidGenerator *typeid.FuncGenerator

	handlerCmd *HandlerCmd
}

type storeHandlerData struct {
	rvalue       reflect.Value
	isSubscribed bool
	evtag        EventTag
}

type storeEventData struct {
	handlers map[HandlerTag]struct{}
}

type HandlerCmd struct {
	store *store
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

	store.handlerCmd = &HandlerCmd{
		store: store,
	}

	evbus := &EventBus{
		store: store,
		subscriber: &EventSubscriber{
			store: store,
		},
		dispatcher: &EventDispatcher{
			store: store,
		},
	}

	return evbus
}

func (bus *EventBus) Subscriber() *EventSubscriber {
	return bus.subscriber
}

func (bus *EventBus) Dispatcher() *EventDispatcher {
	return bus.dispatcher
}

func (bus *EventBus) Wait() {
	bus.store.wait()
}

func (s *store) register(ev any) error {
	evt := elemT(reflect.TypeOf(ev))
	evtag := EventTag(evt)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.events[evtag]; ok {
		return ErrEventAlreadyRegistered
	}

	s.events[evtag] = storeEventData{
		handlers: make(map[HandlerTag]struct{}),
	}

	return nil
}

func (s *store) subscribe(handler any) (HandlerTag, error) {
	htag, hevtag, hv, err := s.extractHandler(handler)
	if err != nil {
		return htag, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.handlers[htag]; ok {
		return htag, ErrHandlerAlreadyRegistered
	}

	s.handlers[htag] = storeHandlerData{
		rvalue:       hv,
		isSubscribed: true,
		evtag:        hevtag,
	}

	s.events[hevtag].handlers[htag] = struct{}{}

	return htag, nil
}

func (s *store) dispatch(ctx context.Context, ev any) error {
	evtag, evv, err := s.extractEvent(ev)
	if err != nil {
		return err
	}

	s.mu.RLock()

	htags := s.events[evtag].handlers
	var hvs []reflect.Value

	for htag := range htags {
		hv := s.handlers[htag].rvalue
		hvs = append(hvs, hv)
	}

	s.mu.RUnlock()

	for _, hv := range hvs {
		err := s.runHandlerAsync(ctx, hv, evv)

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *store) runHandlerAsync(ctx context.Context, hv reflect.Value, evv reflect.Value) error {
	s.wg.Add(1)
	s.guard <- struct{}{}

	go func() {
		hv.Call([]reflect.Value{
			reflect.ValueOf(ctx),
			evv,
			reflect.ValueOf(s.handlerCmd),
		})

		s.wg.Done()
		<-s.guard
	}()

	return nil
}

func (s *store) runHandlerSync(ctx context.Context, hv reflect.Value, evv reflect.Value) error {
	s.wait()

	hv.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		evv,
		reflect.ValueOf(s.handlerCmd),
	})

	return nil
}

func (s *store) wait() {
	s.wg.Wait()
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

func (s *store) extractHandler(handler any) (HandlerTag, EventTag, reflect.Value, error) {
	ht := reflect.TypeOf(handler)
	hv := reflect.ValueOf(handler)
	evtag := EventTag(reflect.TypeOf(nil))

	if hv.IsNil() {
		return 0, evtag, hv, ErrHandlerInvalid
	}

	if ht.Kind() != reflect.Func {
		return 0, evtag, hv, ErrHandlerInvalid
	}

	if ht.NumIn() != 3 {
		return 0, evtag, hv, ErrHandlerInvalid
	}

	if ht.In(0) != reflect.TypeFor[context.Context]() {
		return 0, evtag, hv, ErrHandlerInvalid
	}

	if ht.In(2) != reflect.TypeFor[*HandlerCmd]() {
		return 0, evtag, hv, ErrHandlerInvalid
	}

	evt := ht.In(1)
	if evt.Kind() == reflect.Ptr {
		return 0, evtag, hv, ErrHandlerInvalidEventTag
	}

	evtag = EventTag(evt)

	s.mu.RLock()
	if _, ok := s.events[evtag]; !ok {
		s.mu.RUnlock()
		return 0, evtag, hv, ErrEventNotRegistered
	}
	s.mu.RUnlock()

	genid, err := s.funcTypeidGenerator.GenID(handler)
	if err != nil {
		return 0, evtag, hv, errors.Join(ErrUnknown, err)
	}

	htag := HandlerTag(genid)

	return htag, evtag, hv, nil
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
