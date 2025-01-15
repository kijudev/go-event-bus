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

	bus.MustRegister(new(EventA))
	bus.MustRegister(new(EventB))

	_ = bus.MustSubscribe(func(ctx context.Context, event EventA, cmd *evbus.Cmd) {
		fmt.Println("A: ", int(event))
	})

	_ = bus.MustSubscribe(func(ctx context.Context, event EventB, cmd *evbus.Cmd) {
		fmt.Println("A: ", string(event))
	})

	ctx := context.Background()

	// These are squential
	bus.MustDispatchBlocking(ctx, EventA(1))
	bus.MustDispatchBlocking(ctx, EventA(2))
	bus.MustDispatchBlocking(ctx, EventA(3))
	bus.MustDispatchBlocking(ctx, EventA(4))

	// These are not
	bus.MustDispatch(ctx, EventB("1"))
	bus.MustDispatch(ctx, EventB("2"))
	bus.MustDispatch(ctx, EventB("3"))

	bus.Wait()
}
