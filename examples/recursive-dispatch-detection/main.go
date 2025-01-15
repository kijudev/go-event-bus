package main

import (
	"context"
	"fmt"

	evbus "github.com/kijudev/go-event-bus"
)

type EventA int

func main() {
	// Creates a new event bus
	// 100 is the maximum number of goroutines that can be utilized processing events
	// If the goroutine number is set to 0 or 1 all events will be proccessed synchronously
	bus := evbus.NewEventBus(100)

	bus.MustRegister(new(EventA))

	_ = bus.MustSubscribe(func(ctx context.Context, event EventA, cmd *evbus.Cmd) {
		fmt.Println("A: ", int(event))

		// Since it's possible to dispatch events from within the subscriber
		// It's possible to create infinite loops like this:
		cmd.MustDispatch(ctx, EventA(2))

		// cmd.MustDistatch will panic if there is a circular dispatch call detected
	})

	ctx := context.Background()

	bus.MustDispatch(ctx, EventA(1))
	bus.Wait()
}
