package registry

import (
	"sync"

	"github.com/davidroman0O/tempolite/internal/types"
)

// RegistryBuildFn is a function that builds a registry
type RegistryBuildFn func() (*Registry, error)

type Registry struct {
	workflows  map[types.HandlerIdentity]types.Workflow
	activities map[types.HandlerIdentity]types.Activity
	sync.Mutex
}

func New() *Registry {
	return &Registry{
		workflows:  make(map[types.HandlerIdentity]types.Workflow),
		activities: make(map[types.HandlerIdentity]types.Activity),
	}
}

// RegistryBuilder is used to register workflows, activitiesFunc, and other components
type RegistryBuilder struct {
	workflows  []interface{}
	activities []interface{}
}

// NewBuilder creates a new RegistryBuilder
func NewBuilder() *RegistryBuilder {
	return &RegistryBuilder{
		workflows:  make([]interface{}, 0),
		activities: make([]interface{}, 0),
	}
}

// Workflow adds a workflow to be registered
func (b *RegistryBuilder) Workflow(workflow interface{}) *RegistryBuilder {
	b.workflows = append(b.workflows, workflow)
	return b
}

func (b *RegistryBuilder) Activity(activity interface{}) *RegistryBuilder {
	b.activities = append(b.activities, activity)
	return b
}

// Build finalizes the registry and returns it
func (b *RegistryBuilder) Build() RegistryBuildFn {
	return func() (*Registry, error) {
		r := New()
		for _, w := range b.workflows {
			if err := r.registerWorkflow(w); err != nil {
				return nil, err
			}
		}
		for _, a := range b.activities {
			if err := r.registerActivity(a); err != nil {
				return nil, err
			}
		}
		return r, nil
	}
}
