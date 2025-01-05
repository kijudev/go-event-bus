package evbus

import "context"

type HandlerTag uint64
type EventTag string

type EventReader interface {
	Subscribe(handler any) (HandlerTag, error)
	MustSubscribe(handler any) HandlerTag

	Unsubscribe(handlerTag string) error
	MustUnsuscribe(HandlerTag string)
}

type EventWriter interface {
	Dispatch(ctx context.Context, event any) error
	MustDipatch(ctx context.Context, event any)

	DispatchSync(ctx context.Context, event any) error
	MustDispatchSync(ctx context.Context, event any)

	Wait(ctx context.Context)
}

type EventBus interface {
	Reader() EventReader
	Writer() EventWriter
}

type Event interface {
	Tag() EventTag
}
