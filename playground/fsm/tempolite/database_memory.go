package tempolite

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// DefaultDatabase is an in-memory implementation of Database.
type DefaultDatabase struct {
	runs        map[int]*Run
	versions    map[int]*Version
	hierarchies map[int]*Hierarchy
	entities    map[int]*Entity
	executions  map[int]*Execution
	queues      map[int]*Queue
	mu          sync.RWMutex
}

func NewDefaultDatabase() *DefaultDatabase {
	db := &DefaultDatabase{
		runs:        make(map[int]*Run),
		versions:    make(map[int]*Version),
		hierarchies: make(map[int]*Hierarchy),
		entities:    make(map[int]*Entity),
		executions:  make(map[int]*Execution),
		queues:      make(map[int]*Queue),
	}
	db.queues[1] = &Queue{
		ID:        1,
		Name:      "default",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Entities:  []*Entity{},
	}
	return db
}

var (
	runIDCounter       int64
	versionIDCounter   int64
	entityIDCounter    int64
	executionIDCounter int64
	hierarchyIDCounter int64
	queueIDCounter     int64 = 1 // Starting from 1 for the default queue.
)

// Run methods
func (db *DefaultDatabase) AddRun(run *Run) error {
	runID := int(atomic.AddInt64(&runIDCounter, 1))
	run.ID = runID

	db.mu.Lock()
	defer db.mu.Unlock()

	db.runs[run.ID] = run
	return nil
}

func (db *DefaultDatabase) GetRun(id int) (*Run, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	run, exists := db.runs[id]
	if !exists {
		return nil, errors.New("run not found")
	}

	return db.copyRun(run), nil
}

func (db *DefaultDatabase) UpdateRun(run *Run) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.runs[run.ID]; !exists {
		return errors.New("run not found")
	}

	db.runs[run.ID] = run
	return nil
}

func (db *DefaultDatabase) copyRun(run *Run) *Run {
	copyRun := *run
	copyRun.Entities = make([]*Entity, len(run.Entities))
	copy(copyRun.Entities, run.Entities)
	return &copyRun
}

// Version methods
func (db *DefaultDatabase) GetVersion(entityID int, changeID string) (*Version, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, version := range db.versions {
		if version.EntityID == entityID && version.ChangeID == changeID {
			return version, nil
		}
	}
	return nil, errors.New("version not found")
}

func (db *DefaultDatabase) SetVersion(version *Version) error {
	versionID := int(atomic.AddInt64(&versionIDCounter, 1))
	version.ID = versionID

	db.mu.Lock()
	defer db.mu.Unlock()

	db.versions[version.ID] = version
	return nil
}

// Hierarchy methods
func (db *DefaultDatabase) AddHierarchy(hierarchy *Hierarchy) error {
	hierarchyID := int(atomic.AddInt64(&hierarchyIDCounter, 1))
	hierarchy.ID = hierarchyID

	db.mu.Lock()
	defer db.mu.Unlock()

	db.hierarchies[hierarchy.ID] = hierarchy
	return nil
}

func (db *DefaultDatabase) GetHierarchy(parentID, childID int) (*Hierarchy, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, h := range db.hierarchies {
		if h.ParentEntityID == parentID && h.ChildEntityID == childID {
			return h, nil
		}
	}
	return nil, errors.New("hierarchy not found")
}

func (db *DefaultDatabase) GetHierarchiesByChildEntity(childEntityID int) ([]*Hierarchy, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var result []*Hierarchy
	for _, h := range db.hierarchies {
		if h.ChildEntityID == childEntityID {
			result = append(result, h)
		}
	}
	return result, nil
}

// Entity methods
func (db *DefaultDatabase) AddEntity(entity *Entity) error {
	entityID := int(atomic.AddInt64(&entityIDCounter, 1))
	entity.ID = entityID

	db.mu.Lock()
	defer db.mu.Unlock()

	db.entities[entity.ID] = entity

	// Add the entity to its Run
	run, exists := db.runs[entity.RunID]
	if exists {
		run.Entities = append(run.Entities, entity)
		db.runs[entity.RunID] = run
	}

	return nil
}

var ErrEntityNotFound = errors.New("entity not found")

func (db *DefaultDatabase) GetEntity(id int) (*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	entity, exists := db.entities[id]
	if !exists {
		return nil, errors.Join(fmt.Errorf("entity %d", id), ErrEntityNotFound)
	}
	copy := *entity
	return &copy, nil
}

func (db *DefaultDatabase) HasEntity(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	_, exists := db.entities[id]
	return exists, nil
}

func (db *DefaultDatabase) UpdateEntity(entity *Entity) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.entities[entity.ID]; !exists {
		return errors.Join(fmt.Errorf("entity %d", entity.ID), ErrEntityNotFound)
	}

	db.entities[entity.ID] = entity

	// Update the entity in its Run's Entities slice
	run, exists := db.runs[entity.RunID]
	if exists {
		for i, e := range run.Entities {
			if e.ID == entity.ID {
				run.Entities[i] = entity
				db.runs[run.ID] = run
				break
			}
		}
	}

	return nil
}

func (db *DefaultDatabase) GetEntityByWorkflowIDAndStepID(workflowID int, stepID string) (*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, entity := range db.entities {
		if entity.RunID == workflowID && entity.StepID == stepID {
			return entity, nil
		}
	}
	return nil, errors.Join(fmt.Errorf(
		"entity with workflow ID %d and step ID %s", workflowID, stepID,
	), ErrEntityNotFound)
}

func (db *DefaultDatabase) GetChildEntityByParentEntityIDAndStepIDAndType(parentEntityID int, stepID string, entityType EntityType) (*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, hierarchy := range db.hierarchies {
		if hierarchy.ParentEntityID == parentEntityID && hierarchy.ChildStepID == stepID {
			childEntityID := hierarchy.ChildEntityID
			if entity, exists := db.entities[childEntityID]; exists && entity.Type == entityType {
				return entity, nil
			}
		}
	}
	return nil, errors.Join(fmt.Errorf(
		"child entity with parent entity ID %d, step ID %s, and type %s", parentEntityID, stepID, entityType,
	), ErrEntityNotFound)
}

func (db *DefaultDatabase) FindPendingWorkflowsByQueue(queueID int) ([]*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var result []*Entity
	for _, entity := range db.entities {
		if entity.Type == EntityTypeWorkflow &&
			entity.Status == StatusPending &&
			entity.QueueID == queueID {
			result = append(result, entity)
		}
	}

	// Sort by creation time
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	return result, nil
}

// Execution methods
func (db *DefaultDatabase) AddExecution(execution *Execution) error {
	executionID := int(atomic.AddInt64(&executionIDCounter, 1))
	execution.ID = executionID

	db.mu.Lock()
	defer db.mu.Unlock()

	db.executions[execution.ID] = execution
	return nil
}

func (db *DefaultDatabase) GetExecution(id int) (*Execution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	execution, exists := db.executions[id]
	if !exists {
		return nil, errors.New("execution not found")
	}
	return execution, nil
}

func (db *DefaultDatabase) UpdateExecution(execution *Execution) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.executions[execution.ID]; !exists {
		return errors.New("execution not found")
	}

	db.executions[execution.ID] = execution
	return nil
}

func (db *DefaultDatabase) GetLatestExecution(entityID int) (*Execution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var latestExecution *Execution
	for _, execution := range db.executions {
		if execution.EntityID == entityID {
			if latestExecution == nil || execution.ID > latestExecution.ID {
				latestExecution = execution
			}
		}
	}
	if latestExecution != nil {
		return latestExecution, nil
	}
	return nil, errors.New("execution not found")
}

// Queue methods
func (db *DefaultDatabase) AddQueue(queue *Queue) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Check if queue with same name exists
	for _, q := range db.queues {
		if q.Name == queue.Name {
			return errors.New("queue already exists")
		}
	}

	queueID := int(atomic.AddInt64(&queueIDCounter, 1))
	queue.ID = queueID
	queue.CreatedAt = time.Now()
	queue.UpdatedAt = time.Now()

	db.queues[queue.ID] = queue
	return nil
}

func (db *DefaultDatabase) GetQueue(id int) (*Queue, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	queue, exists := db.queues[id]
	if !exists {
		return nil, errors.Join(ErrQueueNotFound)
	}
	copy := *queue
	return &copy, nil
}

func (db *DefaultDatabase) GetQueueByName(name string) (*Queue, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, queue := range db.queues {
		if queue.Name == name {
			return queue, nil
		}
	}
	return nil, errors.Join(ErrQueueNotFound)
}

func (db *DefaultDatabase) UpdateQueue(queue *Queue) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if existing, exists := db.queues[queue.ID]; exists {
		existing.UpdatedAt = time.Now()
		existing.Name = queue.Name
		existing.Entities = queue.Entities
		db.queues[queue.ID] = existing
		return nil
	}
	return errors.Join(ErrQueueNotFound)
}

func (db *DefaultDatabase) ListQueues() ([]*Queue, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	queues := make([]*Queue, 0, len(db.queues))
	for _, q := range db.queues {
		queues = append(queues, q)
	}
	return queues, nil
}

// Clear removes all Runs that are 'Completed' or 'Failed' and their associated data.
func (db *DefaultDatabase) Clear() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	runsToDelete := []int{}
	entitiesToDelete := map[int]bool{}
	executionsToDelete := map[int]bool{}
	hierarchiesToKeep := map[int]*Hierarchy{}
	versionsToDelete := map[int]*Version{}

	// Find Runs to delete
	for runID, run := range db.runs {
		if run.Status == string(StatusCompleted) || run.Status == string(StatusFailed) {
			runsToDelete = append(runsToDelete, runID)
			// Collect Entities associated with the Run
			for _, entity := range run.Entities {
				entitiesToDelete[entity.ID] = true
			}
		}
	}

	// Collect Executions associated with Entities to delete
	for execID, execution := range db.executions {
		if _, exists := entitiesToDelete[execution.EntityID]; exists {
			executionsToDelete[execID] = true
		}
	}

	// Collect Versions associated with Entities to delete
	for versionID, version := range db.versions {
		if _, exists := entitiesToDelete[version.EntityID]; exists {
			versionsToDelete[versionID] = version
		}
	}

	// Filter Hierarchies to keep only those not associated with Entities to delete
	for hid, hierarchy := range db.hierarchies {
		if _, parentExists := entitiesToDelete[hierarchy.ParentEntityID]; parentExists {
			continue
		}
		if _, childExists := entitiesToDelete[hierarchy.ChildEntityID]; childExists {
			continue
		}
		hierarchiesToKeep[hid] = hierarchy
	}

	// Delete Runs
	for _, runID := range runsToDelete {
		delete(db.runs, runID)
	}

	// Delete Entities and Executions
	for entityID := range entitiesToDelete {
		delete(db.entities, entityID)
	}
	for execID := range executionsToDelete {
		delete(db.executions, execID)
	}

	// Delete Versions
	for versionID := range versionsToDelete {
		delete(db.versions, versionID)
	}

	// Replace hierarchies with the filtered ones
	db.hierarchies = hierarchiesToKeep

	return nil
}
