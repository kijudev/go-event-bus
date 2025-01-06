package main

import (
	"context"
	"fmt"

	evbus "github.com/kijudev/go-event-bus"
)

type EventA int
type EventB string

func main() {
	// Creates a new event bus
	// 100 is the maximum number of goroutines that can be utilized processing events
	// If the goroutine number is set to 0 or 1 all events will be proccessed synchronously
	bus := evbus.NewEventBus(100)

	// Register events
	bus.MustRegister(new(EventA))
	bus.MustRegister(new(EventB))

	// Creates a subscriber
	// The subscriber will receive all events of type `EventA``
	// If EventA is not registered `MustSubscribe` will panic
	// `MustSubscribe` returns a unique handler tag that can be used to unsubscribe
	// Use `Subscribe`
	// Ise `Dispatch` to obtain an error
	htag1 := bus.MustSubscribe(func(ctx context.Context, event EventA, cmd *evbus.Cmd) {
		fmt.Println("A: ", int(event))
	})

	ctx := context.Background()

	// Dispatches the event to all subscribed handlers, async
	// `MustDispatch` will panic if the `EventA` is not registered
	// Use `Dispatch` to obtain an error
	bus.MustDispatch(ctx, EventA(1))

	// Awaits for all handlers to finish processing events
	bus.Wait()

	// Dispatches the event to all subscribed handlers
	// It will wait until all the events are processed and only then proccess the provided event
	// `MustDispatchBlocking` will panic if the `EventA` is not registered
	// Use `DispatchBlocking` to obtain an error
	bus.MustDispatchBlocking(ctx, EventA(2))

	// U
	bus.MustUnsubscribe(htag1)

}
