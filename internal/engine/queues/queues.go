package queues

import "context"

type Queue struct {
	Name string
}

type QueueRegistry struct {
	ctx    context.Context
	queues map[string]Queue
}

func New(ctx context.Context) (*QueueRegistry, error) {
	q := &QueueRegistry{
		ctx:    ctx,
		queues: make(map[string]Queue),
	}

	return q, nil
}
