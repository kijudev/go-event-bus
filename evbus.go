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
	store               *store
	funcTypeidGenerator *typeid.FuncGenerator
}

type EventDispatcher struct {
	store               *store
	funcTypeidGenerator *typeid.FuncGenerator
}

type HandlerCmd struct{}

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

		handlerCmd: &HandlerCmd{},
	}

	evbus := &EventBus{
		store: store,
		subscriber: &EventSubscriber{
			store:               store,
			funcTypeidGenerator: funcTypeidGenerator,
		},
		dispatcher: &EventDispatcher{
			store:               store,
			funcTypeidGenerator: funcTypeidGenerator,
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

func (bus *EventBus) Register(ev any) error {
	evt := elemT(reflect.TypeOf(ev))
	evtag := EventTag(evt)

	bus.store.mu.Lock()
	defer bus.store.mu.Unlock()

	if _, ok := bus.store.events[evtag]; ok {
		return ErrEventAlreadyRegistered
	}

	bus.store.events[evtag] = storeEventData{
		handlers: make(map[HandlerTag]struct{}),
	}

	return nil
}

func (bus *EventBus) MustRegister(ev any) {
	if err := bus.Register(ev); err != nil {
		panic(err)
	}
}

func (bus *EventBus) Wait() {
	bus.store.wait()
}

func (sub *EventSubscriber) Subscribe(handler any) (HandlerTag, error) {
	htag, hevtag, hv, err := sub.store.extractHandler(handler)
	if err != nil {
		return htag, err
	}

	sub.store.mu.Lock()
	defer sub.store.mu.Unlock()

	if _, ok := sub.store.handlers[htag]; ok {
		return htag, ErrHandlerAlreadyRegistered
	}

	sub.store.handlers[htag] = storeHandlerData{
		rvalue:       hv,
		isSubscribed: true,
		evtag:        hevtag,
	}

	sub.store.events[hevtag].handlers[htag] = struct{}{}

	return htag, nil
}

func (sub *EventSubscriber) MustSubscribe(handler any) HandlerTag {
	htag, err := sub.Subscribe(handler)

	if err != nil {
		panic(err)
	}

	return htag
}

func (dispatcher *EventDispatcher) Dispatch(ctx context.Context, ev any) error {
	evtag, evv, err := dispatcher.store.extractEvent(ev)
	if err != nil {
		return err
	}

	dispatcher.store.mu.RLock()

	htags := dispatcher.store.events[evtag].handlers
	var hvs []reflect.Value

	for htag := range htags {
		hv := dispatcher.store.handlers[htag].rvalue
		hvs = append(hvs, hv)
	}

	dispatcher.store.mu.RUnlock()

	for _, hv := range hvs {
		err := dispatcher.store.runHandlerAsync(ctx, hv, evv)

		if err != nil {
			return err
		}
	}

	return nil
}

func (dispatcher *EventDispatcher) MustDispatch(ctx context.Context, ev any) {
	if err := dispatcher.Dispatch(ctx, ev); err != nil {
		panic(err)
	}
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
