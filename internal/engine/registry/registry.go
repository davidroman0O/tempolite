package registry

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sync"

	"github.com/davidroman0O/tempolite/internal/types"
)

/// The only contraint of a Registry is its static nature, it is not meant to be modified at runtime

// RegistryBuildFn is a function that builds a registry
type RegistryBuildFn func() (*Registry, error)

type Registry struct {
	workflows  map[types.HandlerIdentity]types.Workflow
	activities map[types.HandlerIdentity]types.Activity
	mu         sync.Mutex
}

func newRegistry() *Registry {
	return &Registry{
		workflows:  make(map[types.HandlerIdentity]types.Workflow),
		activities: make(map[types.HandlerIdentity]types.Activity),
	}
}

var (
	ErrRegistry         = errors.New("registry error")
	ErrWorkflowNotFound = errors.New("workflow not found")
	ErrActivityNotFound = errors.New("activity not found")
)

func (r *Registry) VerifyParamsMatching(handlerInfo types.HandlerInfo, param ...interface{}) error {

	if len(param) != handlerInfo.NumIn {
		return fmt.Errorf("parameter count mismatch (you probably put the wrong handler): expected %d, got %d", handlerInfo.NumIn, len(param))
	}

	for idx, param := range param {
		if reflect.TypeOf(param) != handlerInfo.ParamTypes[idx] {
			return fmt.Errorf("parameter type mismatch (you probably put the wrong handler) at index %d: expected %s, got %s", idx, handlerInfo.ParamTypes[idx], reflect.TypeOf(param))
		}
	}

	return nil
}

func (r *Registry) GetWorkflow(identity types.HandlerIdentity) (types.Workflow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	w, ok := r.workflows[identity]
	if !ok {
		return types.Workflow{}, errors.Join(ErrRegistry, ErrWorkflowNotFound)
	}
	return w, nil
}

func (r *Registry) GetActivity(identity types.HandlerIdentity) (types.Activity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	a, ok := r.activities[identity]
	if !ok {
		return types.Activity{}, errors.Join(ErrRegistry, ErrActivityNotFound)
	}
	return a, nil
}

func (r *Registry) WorkflowIdentity(w interface{}) (types.HandlerIdentity, error) {
	value := reflect.ValueOf(w)
	ptr := value.Pointer()

	funcName := runtime.FuncForPC(ptr).Name()
	handlerIdentity := types.HandlerIdentity(funcName)

	if _, ok := r.workflows[handlerIdentity]; !ok {
		return "", errors.Join(ErrRegistry, ErrWorkflowNotFound)
	}

	return handlerIdentity, nil
}

func (r *Registry) ActivityIdentity(a interface{}) (types.HandlerIdentity, error) {
	value := reflect.ValueOf(a)
	ptr := value.Pointer()

	funcName := runtime.FuncForPC(ptr).Name()
	handlerIdentity := types.HandlerIdentity(funcName)

	// check exists in registry
	if _, ok := r.activities[handlerIdentity]; !ok {
		return "", errors.Join(ErrRegistry, ErrActivityNotFound)
	}

	return handlerIdentity, nil
}

// RegistryBuilder is used to register workflows, activitiesFunc, and other components
type RegistryBuilder struct {
	workflows  []interface{}
	activities []interface{}
}

// New creates a new RegistryBuilder
func New() *RegistryBuilder {
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
		r := newRegistry()
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
