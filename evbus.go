package evbus

import "context"

type HandlerTag string
type EventTag string

type EventReader interface {
	Subscribe(handler any) (HandlerTag, error)
	MustSubscribe(handler any) HandlerTag
	SubscribeTag(tag HandlerTag, rawHandler func(ctx context.Context, event any, eventWriter EventWriter)) (HandlerTag, error)
	MustSubscribeTag(tag HandlerTag, rawHandler func(ctx context.Context, event any, eventWriter EventWriter))
}

type EventWriter interface {
	Dispatch(ctx context.Context, event any) error
	MustDipatch(ctx context.Context, event any)
	DispatchTag(ctx context.Context, tag EventTag, rawEvent any) error
	MustDispatchTag(ctx context.Context, tag EventTag, rawEvent any)

	Wait(ctx context.Context)
}
