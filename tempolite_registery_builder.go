package tempolite

// RegistryBuilder is used to register workflows, activitiesFunc, and other components
type RegistryBuilder[T Identifier] struct {
	workflowsFunc  []interface{}
	activitiesFunc []interface{}
	activities     []ActivityRegister
}

// NewRegistry creates a new RegistryBuilder
func NewRegistry[T Identifier]() *RegistryBuilder[T] {
	return &RegistryBuilder[T]{
		workflowsFunc:  make([]interface{}, 0),
		activitiesFunc: make([]interface{}, 0),
		activities:     make([]ActivityRegister, 0),
	}
}

// Workflow adds a workflow to be registered
func (b *RegistryBuilder[T]) Workflow(workflow interface{}) *RegistryBuilder[T] {
	b.workflowsFunc = append(b.workflowsFunc, workflow)
	return b
}

// Activity adds an activity to be registered
func (b *RegistryBuilder[T]) Activity(register ActivityRegister) *RegistryBuilder[T] {
	b.activities = append(b.activities, register)
	return b
}

func (b *RegistryBuilder[T]) ActivityFunc(activity interface{}) *RegistryBuilder[T] {
	b.activitiesFunc = append(b.activitiesFunc, activity)
	return b
}

// Build finalizes the registry and returns it
func (b *RegistryBuilder[T]) Build() *Registry[T] {
	return &Registry[T]{
		workflowsFunc:  b.workflowsFunc,
		activitiesFunc: b.activitiesFunc,
		activities:     b.activities,
	}
}

// Registry holds the registered components
type Registry[T Identifier] struct {
	workflowsFunc  []interface{}
	activitiesFunc []interface{}
	activities     []ActivityRegister
}
