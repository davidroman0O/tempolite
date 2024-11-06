package tempolite

import (
	"context"
	"sync"

	"github.com/davidroman0O/tempolite/internal/engine/registry"
)

type tempoliteConfig struct {
	path        *string
	destructive bool
}

type tempoliteOption func(*tempoliteConfig)

type Tempolite struct {
	sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	registry *registry.Registry
}

func New(ctx context.Context, builder registry.RegistryBuildFn, opts ...tempoliteOption) (*Tempolite, error) {
	cfg := tempoliteConfig{}

	for _, opt := range opts {
		opt(&cfg)
	}
	var err error
	var registry *registry.Registry
	if registry, err = builder(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	tp := &Tempolite{
		ctx:      ctx,
		cancel:   cancel,
		registry: registry,
	}

	return tp, nil
}
