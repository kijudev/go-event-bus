package evbus

import (
	"context"
	"errors"
	"reflect"
	"sync"

	"github.com/kijudev/go-event-bus/typeid"
)

type EmbededEventBus struct {
	store  *eventBusStore
	reader *EmbededEventReader
	writer *EmbededEventWriter
}

type EmbededEventReader struct {
	store         *eventBusStore
	funcTypeidGen *typeid.FuncGenerator
}

type EmbededEventWriter struct {
	store *eventBusStore
}

type eventBusStore struct {
	guard chan struct{}
	mu    sync.RWMutex
	wg    sync.WaitGroup

	registry map[EventTag]struct{}
	handlers map[EventTag](map[HandlerTag]reflect.Value)
}

func NewEmbededEventBus(maxGoroutines uint) *EmbededEventBus {
	store := &eventBusStore{
		guard:    make(chan struct{}, maxGoroutines),
		registry: make(map[EventTag]struct{}),
		handlers: make(map[EventTag]map[HandlerTag]reflect.Value),
	}

	return &EmbededEventBus{
		store: store,
		reader: &EmbededEventReader{
			store:         store,
			funcTypeidGen: typeid.NewFuncGenerator(),
		},
		writer: &EmbededEventWriter{
			store: store,
		},
	}
}

func (reader *EmbededEventReader) Subscribe(handler any) (HandlerTag, error) {
	ht := reflect.TypeOf(handler)
	hv := reflect.ValueOf(handler)

	if ht.Kind() != reflect.Func {
		return 0, ErrInvalidEventHandler
	}

	if ht.NumIn() != 3 {
		return 0, ErrInvalidEventHandlerArgs
	}

	if ht.In(0) != reflect.TypeFor[context.Context]() {
		return 0, ErrInvalidEventHandlerArgs
	}

	if ht.In(2) != reflect.TypeFor[*EmbededEventWriter]() {
		return 0, ErrInvalidEventHandlerArgs
	}

	et := ht.In(1)
	if et.Kind() == reflect.Ptr {
		et = et.Elem()
	}

	if !et.Implements(reflect.TypeFor[Event]()) {
		return 0, ErrEventDoesNotImplementTag
	}

	evtagfn := hv.MethodByName("Tag")
	evtagv := evtagfn.Call([]reflect.Value{})[0]
	evtag := EventTag(evtagv.String())

	reader.store.mu.Lock()
	defer reader.store.mu.Unlock()
	if _, ok := reader.store.handlers[evtag]; !ok {
		return 0, ErrEventNotRegisterd
	}

	if reader.store.handlers[evtag] == nil {
		reader.store.handlers[evtag] = make(map[HandlerTag]reflect.Value)
	}

	htaguint, err := reader.funcTypeidGen.GenID(handler)
	if err != nil {
		return 0, errors.Join(ErrUnknown, err)
	}

	htag := HandlerTag(htaguint)
	reader.store.handlers[evtag][htag] = hv

	return htag, nil
}

func (reader *EmbededEventReader) MustSubscribe(handler any) HandlerTag {
	t, err := reader.Subscribe(handler)

	if err != nil {
		panic(err)
	}

	return t
}
