package main

import (
	"context"
	"fmt"

	evbus "github.com/kijudev/go-event-bus"
)

type EventA int
type EventB string
type EventC float32

func main() {
	bus := evbus.NewEventBus(100)

	bus.MustRegister(new(EventA))
	bus.MustRegister(new(EventB))

	_ = bus.Subscriber().MustSubscribe(func(ctx context.Context, event EventA, cmd *evbus.Cmd) {
		fmt.Println("A: ", int(event))

		cmd.MustDispatch(ctx, EventA(2))
	})

	ctx := context.Background()

	bus.Dispatcher().MustDispatch(ctx, EventA(1))

	bus.Wait()
}
