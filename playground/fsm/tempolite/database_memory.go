package tempolite

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// DefaultDatabase is an in-memory implementation of Database.
type DefaultDatabase struct {
	runs        map[int]*Run
	versions    map[int]*Version // for in-memory we will basically run the latest all the time, it's just that we don't know if the dev will re-use other workflows. Hopefully we can still force it.
	hierarchies []*Hierarchy
	entities    map[int]*Entity
	executions  map[int]*Execution
	queues      map[int]*Queue
	mu          sync.Mutex
}

func NewDefaultDatabase() *DefaultDatabase {
	return &DefaultDatabase{
		runs:        make(map[int]*Run),
		versions:    make(map[int]*Version),
		entities:    make(map[int]*Entity),
		executions:  make(map[int]*Execution),
		queues:      make(map[int]*Queue),
		hierarchies: []*Hierarchy{},
	}
}

var (
	runIDCounter       int64
	versionIDCounter   int64
	entityIDCounter    int64
	executionIDCounter int64
	hierarchyIDCounter int64
)

func (db *DefaultDatabase) AddRun(run *Run) *Run {
	db.mu.Lock()
	defer db.mu.Unlock()
	runID := int(atomic.AddInt64(&runIDCounter, 1))
	run.ID = runID
	db.runs[run.ID] = run
	return run
}

func (db *DefaultDatabase) GetRun(id int) *Run {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.runs[id]
}

func (db *DefaultDatabase) UpdateRun(run *Run) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.runs[run.ID] = run
}

func (db *DefaultDatabase) GetVersion(entityID int, changeID string) *Version {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, version := range db.versions {
		if version.EntityID == entityID && version.ChangeID == changeID {
			return version
		}
	}
	return nil
}

func (db *DefaultDatabase) SetVersion(version *Version) *Version {
	db.mu.Lock()
	defer db.mu.Unlock()
	versionID := int(atomic.AddInt64(&versionIDCounter, 1))
	version.ID = versionID
	db.versions[version.ID] = version
	return version
}

func (db *DefaultDatabase) AddHierarchy(hierarchy *Hierarchy) {
	db.mu.Lock()
	defer db.mu.Unlock()
	hierarchyID := int(atomic.AddInt64(&hierarchyIDCounter, 1))
	hierarchy.ID = hierarchyID
	db.hierarchies = append(db.hierarchies, hierarchy)
}

func (db *DefaultDatabase) GetHierarchy(parentID, childID int) *Hierarchy {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, h := range db.hierarchies {
		if h.ParentEntityID == parentID && h.ChildEntityID == childID {
			return h
		}
	}
	return nil
}

func (db *DefaultDatabase) GetHierarchiesByChildEntity(childEntityID int) []*Hierarchy {
	db.mu.Lock()
	defer db.mu.Unlock()

	var result []*Hierarchy
	for _, h := range db.hierarchies {
		if h.ChildEntityID == childEntityID {
			result = append(result, h)
		}
	}
	return result
}

// Entity methods
func (db *DefaultDatabase) AddEntity(entity *Entity) *Entity {
	db.mu.Lock()
	defer db.mu.Unlock()
	entityID := int(atomic.AddInt64(&entityIDCounter, 1))
	entity.ID = entityID
	db.entities[entity.ID] = entity

	// Add the entity to its Run
	run := db.runs[entity.RunID]
	if run != nil {
		run.Entities = append(run.Entities, entity)
	}
	return entity
}

func (db *DefaultDatabase) GetEntity(id int) *Entity {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.entities[id]
}

func (db *DefaultDatabase) UpdateEntity(entity *Entity) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.entities[entity.ID] = entity
}

func (db *DefaultDatabase) GetEntityByWorkflowIDAndStepID(workflowID int, stepID string) *Entity {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, entity := range db.entities {
		if entity.RunID == workflowID && entity.StepID == stepID {
			return entity
		}
	}
	return nil
}

func (db *DefaultDatabase) GetChildEntityByParentEntityIDAndStepIDAndType(parentEntityID int, stepID string, entityType EntityType) *Entity {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, hierarchy := range db.hierarchies {
		if hierarchy.ParentEntityID == parentEntityID && hierarchy.ChildStepID == stepID {
			childEntityID := hierarchy.ChildEntityID
			if entity, exists := db.entities[childEntityID]; exists && entity.Type == entityType {
				return entity
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) FindPendingWorkflowsByQueue(queueID int) []*Entity {
	db.mu.Lock()
	defer db.mu.Unlock()

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

	return result
}

// Execution methods
func (db *DefaultDatabase) AddExecution(execution *Execution) *Execution {
	db.mu.Lock()
	defer db.mu.Unlock()
	executionID := int(atomic.AddInt64(&executionIDCounter, 1))
	execution.ID = executionID
	db.executions[execution.ID] = execution
	return execution
}

func (db *DefaultDatabase) GetExecution(id int) *Execution {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.executions[id]
}

func (db *DefaultDatabase) UpdateExecution(execution *Execution) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.executions[execution.ID] = execution
}

func (db *DefaultDatabase) GetLatestExecution(entityID int) *Execution {
	db.mu.Lock()
	defer db.mu.Unlock()
	var latestExecution *Execution
	for _, execution := range db.executions {
		if execution.EntityID == entityID {
			if latestExecution == nil || execution.ID > latestExecution.ID {
				latestExecution = execution
			}
		}
	}
	return latestExecution
}

// Clear removes all Runs that are 'Completed' and their associated data.
func (db *DefaultDatabase) Clear() {
	db.mu.Lock()
	defer db.mu.Unlock()

	runsToDelete := []int{}
	entitiesToDelete := map[int]bool{} // EntityID to delete
	executionsToDelete := map[int]bool{}
	hierarchiesToKeep := []*Hierarchy{}
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
	for _, hierarchy := range db.hierarchies {
		if _, parentExists := entitiesToDelete[hierarchy.ParentEntityID]; parentExists {
			continue
		}
		if _, childExists := entitiesToDelete[hierarchy.ChildEntityID]; childExists {
			continue
		}
		hierarchiesToKeep = append(hierarchiesToKeep, hierarchy)
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
}

func (db *DefaultDatabase) AddQueue(queue *Queue) *Queue {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Check if queue with same name exists
	for _, q := range db.queues {
		if q.Name == queue.Name {
			return q
		}
	}

	queue.ID = len(db.queues) + 1
	queue.CreatedAt = time.Now()
	queue.UpdatedAt = time.Now()

	if db.queues == nil {
		db.queues = make(map[int]*Queue)
	}
	db.queues[queue.ID] = queue
	return queue
}

func (db *DefaultDatabase) GetQueue(id int) *Queue {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.queues[id]
}

func (db *DefaultDatabase) GetQueueByName(name string) *Queue {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, queue := range db.queues {
		if queue.Name == name {
			return queue
		}
	}
	return nil
}

func (db *DefaultDatabase) UpdateQueue(queue *Queue) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if existing, exists := db.queues[queue.ID]; exists {
		existing.UpdatedAt = time.Now()
		existing.Name = queue.Name
		existing.Entities = queue.Entities
	}
}

func (db *DefaultDatabase) ListQueues() []*Queue {
	db.mu.Lock()
	defer db.mu.Unlock()
	queues := make([]*Queue, 0, len(db.queues))
	for _, q := range db.queues {
		queues = append(queues, q)
	}
	return queues
}
