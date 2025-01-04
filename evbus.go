package evbus

import "context"

type HandlerTag string
type EventTag string

type EventReader interface {
	Subscribe(handler any) (HandlerTag, error)
	MustSubscribe(handler any) HandlerTag
	SubscribeTag(tag HandlerTag, rawHandler func(ctx context.Context, event any, eventWriter EventWriter)) (HandlerTag, error)
	MustSubscribeTag(tag HandlerTag, rawHandler func(ctx context.Context, event any, eventWriter EventWriter))

	Unsubscribe(handlerTag string) error
	MustUnsuscribe(HandlerTag string)
}

type EventWriter interface {
	Dispatch(ctx context.Context, event any) error
	MustDipatch(ctx context.Context, event any)
	DispatchTag(ctx context.Context, tag EventTag, rawEvent any) error
	MustDispatchTag(ctx context.Context, tag EventTag, rawEvent any)

	DispatchSync(ctx context.Context, event any) error
	MustDispatchSync(ctx context.Context, event any)
	DispatchTagSync(ctx context.Context, rawEvent any) error
	MustDispatchTagSync(ctx context.Context, rawEvent any)

	Wait(ctx context.Context)
}

type EventBus interface {
	Reader() EventReader
	Writer() EventWriter
}
