package tempolite

// Database interface defines methods for interacting with the data store. Runtime based.
type Database interface {
	// Run methods
	AddRun(run *Run) *Run
	GetRun(id int) *Run
	UpdateRun(run *Run)

	// Version methods
	GetVersion(entityID int, changeID string) *Version
	SetVersion(version *Version) *Version

	// Hierarchy methods
	AddHierarchy(hierarchy *Hierarchy)
	GetHierarchy(parentID, childID int) *Hierarchy
	GetHierarchiesByChildEntity(childEntityID int) []*Hierarchy

	// Entity methods
	AddEntity(entity *Entity) *Entity
	GetEntity(id int) *Entity
	UpdateEntity(entity *Entity)
	GetEntityByWorkflowIDAndStepID(workflowID int, stepID string) *Entity
	GetChildEntityByParentEntityIDAndStepIDAndType(parentEntityID int, stepID string, entityType EntityType) *Entity

	// Execution methods
	AddExecution(execution *Execution) *Execution
	GetExecution(id int) *Execution
	UpdateExecution(execution *Execution)
	GetLatestExecution(entityID int) *Execution

	// Queues
	AddQueue(queue *Queue) *Queue
	GetQueue(id int) *Queue
	GetQueueByName(name string) *Queue
	UpdateQueue(queue *Queue)
	ListQueues() []*Queue
}
