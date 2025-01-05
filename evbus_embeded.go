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

	events              map[EventTag]struct{}
	handlers            map[HandlerTag]handlerData
	eventsToHandlerTags map[EventTag](map[HandlerTag]struct{})
}

type handlerData struct {
	fn           reflect.Value
	isSubscribed bool
	evtag        EventTag
	htag         HandlerTag
}

func NewEmbededEventBus(maxGoroutines uint) *EmbededEventBus {
	store := &eventBusStore{
		guard: make(chan struct{}, maxGoroutines),

		events:              make(map[EventTag]struct{}),
		handlers:            make(map[HandlerTag]handlerData),
		eventsToHandlerTags: make(map[EventTag]map[HandlerTag]struct{}),
	}

	evbus := &EmbededEventBus{
		store: store,
		reader: &EmbededEventReader{
			store:         store,
			funcTypeidGen: typeid.NewFuncGenerator(),
		},
		writer: &EmbededEventWriter{
			store: store,
		},
	}

	// Don't forget to register the null event

	return evbus
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

	if _, ok := reader.store.events[evtag]; !ok {
		return 0, ErrEventNotRegisterd
	}

	htaguint, err := reader.funcTypeidGen.GenID(handler)
	if err != nil {
		return 0, errors.Join(ErrUnknown, err)
	}

	htag := HandlerTag(htaguint)
	if _, ok := reader.store.handlers[htag]; ok {
		return 0, ErrHandlerAlreadySubscribed
	}

	reader.store.handlers[htag] = handlerData{
		fn:           hv,
		isSubscribed: true,
		evtag:        evtag,
		htag:         htag,
	}
	reader.store.eventsToHandlerTags[evtag][htag] = struct{}{}

	return htag, nil
}

func (reader *EmbededEventReader) MustSubscribe(handler any) HandlerTag {
	t, err := reader.Subscribe(handler)

	if err != nil {
		panic(err)
	}

	return t
}

func (reader *EmbededEventReader) Unsubscribe(htag HandlerTag) error {
	reader.store.mu.Lock()
	defer reader.store.mu.Unlock()

	if _, ok := reader.store.handlers[htag]; !ok {
		return ErrInvalidHandlerTag
	}

	hdata := reader.store.handlers[htag]
	if !hdata.isSubscribed {
		return ErrHandlerAreadyUnsubscribed
	}

	delete(reader.store.eventsToHandlerTags[hdata.evtag], hdata.htag)
	hdata.isSubscribed = false
	hdata.evtag = NullEvTag

	reader.store.handlers[htag] = hdata

	return nil
}
