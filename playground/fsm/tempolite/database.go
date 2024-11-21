package tempolite

import "errors"

var ErrQueueExists = errors.New("queue already exists")
var ErrQueueNotFound = errors.New("queue not found")

// Database interface defines methods for interacting with the data store.
// All methods return errors to handle future implementations that might have errors.
type Database interface {
	// Run methods
	AddRun(run *Run) error
	GetRun(id int) (*Run, error)
	UpdateRun(run *Run) error

	// Version methods
	GetVersion(entityID int, changeID string) (*Version, error)
	SetVersion(version *Version) error

	// Hierarchy methods
	AddHierarchy(hierarchy *Hierarchy) error
	GetHierarchy(parentID, childID int) (*Hierarchy, error)
	GetHierarchiesByChildEntity(childEntityID int) ([]*Hierarchy, error)

	// Entity methods
	AddEntity(entity *Entity) error
	HasEntity(id int) (bool, error)
	GetEntity(id int) (*Entity, error)
	UpdateEntity(entity *Entity) error
	GetEntityByWorkflowIDAndStepID(workflowID int, stepID string) (*Entity, error)
	GetChildEntityByParentEntityIDAndStepIDAndType(parentEntityID int, stepID string, entityType EntityType) (*Entity, error)
	FindPendingWorkflowsByQueue(queueID int) ([]*Entity, error)

	// Execution methods
	AddExecution(execution *Execution) error
	GetExecution(id int) (*Execution, error)
	UpdateExecution(execution *Execution) error
	GetLatestExecution(entityID int) (*Execution, error)

	// Queue methods
	AddQueue(queue *Queue) error
	GetQueue(id int) (*Queue, error)
	GetQueueByName(name string) (*Queue, error)
	UpdateQueue(queue *Queue) error
	ListQueues() ([]*Queue, error)

	// Clear removes all Runs that are 'Completed' and their associated data.
	Clear() error
}
