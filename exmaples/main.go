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

	h1 := bus.Subscriber().MustSubscribe(func(ctx context.Context, event EventA, cmd *evbus.HandlerCmd) {
		fmt.Println("A: ", int(event))
	})

	h2 := bus.Subscriber().MustSubscribe(func(ctx context.Context, event EventB, cmd *evbus.HandlerCmd) {
		fmt.Println("B: ", string(event))
	})

	fmt.Println(h1, h2)

	ctx := context.Background()

	bus.Dispatcher().MustDispatch(ctx, EventC(1.2))

	bus.Wait()
}
