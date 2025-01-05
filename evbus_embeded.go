package evbus

import (
	"reflect"
	"sync"
)

type EmbededEventBus struct {
	store  *eventBusStore
	reader *EmbededEventReader
	writer *EmbededEventWriter
}

type EmbededEventReader struct {
	store *eventBusStore
}

type EmbededEventWriter struct {
	store *eventBusStore
}

type eventBusStore struct {
	guard chan struct{}
	lock  sync.RWMutex
	wg    sync.WaitGroup

	registry map[EventTag]struct{}
	handlers map[EventTag](map[HandlerTag]reflect.Value)
}

func NewEmbededEventBus(maxGoroutines uint) *EmbededEventBus {
	store := &eventBusStore{
		guard: make(chan struct{}, maxGoroutines),
	}

	return &EmbededEventBus{
		store: store,
		reader: &EmbededEventReader{
			store: store,
		},
		writer: &EmbededEventWriter{
			store: store,
		},
	}
}

func (reader *EmbededEventReader) Subscribe(handler any) (HandlerTag, error) {
	ht := reflect.TypeOf(handler)
	hv := reflect.ValueOf(handler)

	return "", nil
}

func (reader *EmbededEventReader) MustSubscribe(handler any) HandlerTag {
	t, err := reader.Subscribe(handler)

	if err != nil {
		panic(err)
	}

	return t
}
