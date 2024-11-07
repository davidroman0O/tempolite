package tempolite

// RegistryBuilder is used to register workflows, activitiesFunc, and other components
type RegistryBuilder struct {
	workflows  []interface{}
	activities []interface{}
}

// NewRegistry creates a new RegistryBuilder
func NewRegistry() *RegistryBuilder {
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
func (b *RegistryBuilder) Build() *Registry {
	return &Registry{
		workflowsFunc: b.workflows,
		activities:    b.activities,
	}
}

// Registry holds the registered components
type Registry struct {
	workflowsFunc []interface{}
	activities    []interface{}
}
