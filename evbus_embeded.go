package evbus

import "sync"

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
	lock  sync.Mutex
	wg    sync.WaitGroup
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
