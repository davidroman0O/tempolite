package context

import "context"

type TempoliteContext struct {
	context.Context
}

func New(
	ctx context.Context,
) TempoliteContext {
	return TempoliteContext{Context: ctx}
}
