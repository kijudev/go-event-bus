package evbus

import "errors"

var (
	ErrHandlerInvalid             = errors.New("HANDLER_INVALID")
	ErrHandlerInvalidEventTag     = errors.New("HANDLER_INVALID_EVENT_TAG")
	ErrHandlerNotRegistered       = errors.New("HANDLER_NOT_REGISTERED")
	ErrHandlerAlreadyRegistered   = errors.New("HANDLER_ALREADY_REGISTERED")
	ErrHandlerAlreadyUnsubscribed = errors.New("HANDLER_ALREADY_UNSUBSCRIBED")

	ErrEventInvalid           = errors.New("EVENT_INVALID")
	ErrEventNotRegistered     = errors.New("EVENT_NOT_REGISTERED")
	ErrEventAlreadyRegistered = errors.New("EVENT_ALREADY_REGISTERED")

	ErrCmdResursiveDispatchCall = errors.New("CMD_RECURSIVE_DISPATCH_CALL")

	ErrUnknown = errors.New("UNKNOWN")
)
