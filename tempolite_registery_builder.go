package tempolite

// RegistryBuilder is used to register workflows, activitiesFunc, and other components
type RegistryBuilder[T Identifier] struct {
	workflows  []interface{}
	activities []interface{}
}

// NewRegistry creates a new RegistryBuilder
func NewRegistry[T Identifier]() *RegistryBuilder[T] {
	return &RegistryBuilder[T]{
		workflows:  make([]interface{}, 0),
		activities: make([]interface{}, 0),
	}
}

// Workflow adds a workflow to be registered
func (b *RegistryBuilder[T]) Workflow(workflow interface{}) *RegistryBuilder[T] {
	b.workflows = append(b.workflows, workflow)
	return b
}

func (b *RegistryBuilder[T]) Activity(activity interface{}) *RegistryBuilder[T] {
	b.activities = append(b.activities, activity)
	return b
}

// Build finalizes the registry and returns it
func (b *RegistryBuilder[T]) Build() *Registry[T] {
	return &Registry[T]{
		workflowsFunc: b.workflows,
		activities:    b.activities,
	}
}

// Registry holds the registered components
type Registry[T Identifier] struct {
	workflowsFunc []interface{}
	activities    []interface{}
}
