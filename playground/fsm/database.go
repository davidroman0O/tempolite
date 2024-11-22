package tempolite

import (
	"errors"
	"fmt"
	"sort"
	"sync/atomic"
	"time"

	"github.com/sasha-s/go-deadlock"
)

var (
	ErrQueueExists       = errors.New("queue already exists")
	ErrQueueNotFound     = errors.New("queue not found")
	ErrEntityNotFound    = errors.New("entity not found")
	ErrRunNotFound       = errors.New("run not found")
	ErrVersionNotFound   = errors.New("version not found")
	ErrHierarchyNotFound = errors.New("hierarchy not found")
	ErrExecutionNotFound = errors.New("execution not found")
)

type Database interface {
	// Run methods
	AddRun(run *Run) error
	GetRun(id int) (*Run, error)
	UpdateRun(run *Run) error
	GetRunsCount() (int, error)
	GetRunsByDateRange(start, end time.Time) ([]*Run, error)
	GetActiveRuns() ([]*Run, error)
	UpdateRunStatus(id int, status string) error
	UpdateRunUpdatedAt(id int) error
	BatchUpdateRunStatus(ids []int, status string) error
	DeleteRun(id int) error

	// Version methods
	GetVersion(entityID int, changeID string) (*Version, error)
	SetVersion(version *Version) error
	GetVersionsCount(entityID int) (int, error)
	GetLatestVersion(entityID int, changeID string) (*Version, error)
	GetVersionsByDateRange(entityID int, start, end time.Time) ([]*Version, error)
	BatchDeleteVersions(ids []int) error
	DeleteVersion(id int) error

	// Hierarchy methods
	AddHierarchy(hierarchy *Hierarchy) error
	GetHierarchy(parentID, childID int) (*Hierarchy, error)
	GetHierarchiesByChildEntity(childEntityID int) ([]*Hierarchy, error)
	GetHierarchiesByParentEntity(parentEntityID int) ([]*Hierarchy, error)
	GetHierarchyDepth(entityID int) (int, error)
	GetRootEntity(entityID int) (*Entity, error)
	GetLeafEntities(entityID int) ([]*Entity, error)
	GetSiblingEntities(entityID int) ([]*Entity, error)
	GetDescendants(entityID int) ([]*Entity, error)
	GetAncestors(entityID int) ([]*Entity, error)
	BatchAddHierarchies(hierarchies []*Hierarchy) error
	DeleteHierarchy(id int) error

	// Entity methods
	AddEntity(entity *Entity) error
	HasEntity(id int) (bool, error)
	GetEntity(id int) (*Entity, error)
	GetEntitiesCount(filters ...EntityFilter) (int, error)
	GetEntitiesByStatus(status EntityStatus) ([]*Entity, error)
	GetEntitiesByType(entityType EntityType) ([]*Entity, error)
	GetEntitiesByHandlerName(handlerName string) ([]*Entity, error)
	GetEntitiesByStepID(stepID string) ([]*Entity, error)
	GetEntitiesByRunID(runID int) ([]*Entity, error)
	UpdateEntityStatus(id int, status EntityStatus) error
	BatchUpdateEntityStatus(ids []int, status EntityStatus) error
	UpdateEntityPaused(id int, paused bool) error
	BatchUpdateEntityPaused(ids []int, paused bool) error
	UpdateEntityRetryState(id int, retryState *RetryState) error
	UpdateEntityRetryPolicy(id int, policy *retryPolicyInternal) error
	UpdateEntityResumable(id int, resumable bool) error
	UpdateEntityWorkflowData(id int, workflowData *WorkflowData) error
	UpdateEntityActivityData(id int, activityData *ActivityData) error
	UpdateEntitySideEffectData(id int, sideEffectData *SideEffectData) error
	UpdateEntitySagaData(id int, sagaData *SagaData) error
	UpdateEntityHandlerInfo(id int, handlerInfo *HandlerInfo) error
	GetEntityByWorkflowIDAndStepID(workflowID int, stepID string) (*Entity, error)
	GetChildEntityByParentEntityIDAndStepIDAndType(parentEntityID int, stepID string, entityType EntityType) (*Entity, error)
	FindPendingWorkflowsByQueue(queueID int) ([]*Entity, error)
	GetEntitiesByQueueID(queueID int) ([]*Entity, error)
	GetEntitiesByRetryCount(count int) ([]*Entity, error)
	GetFailedEntities() ([]*Entity, error)
	GetPausedEntities() ([]*Entity, error)
	DeleteEntity(id int) error

	// Execution methods
	AddExecution(execution *Execution) error
	BatchAddExecutions(executions []*Execution) error
	GetExecution(id int) (*Execution, error)
	GetExecutionsCount(filters ...ExecutionFilter) (int, error)
	GetExecutionsByStatus(status ExecutionStatus) ([]*Execution, error)
	GetExecutionsByEntityID(entityID int) ([]*Execution, error)
	GetExecutionsByDateRange(start, end time.Time) ([]*Execution, error)
	UpdateExecutionStatus(id int, status ExecutionStatus) error
	BatchUpdateExecutionStatus(ids []int, status ExecutionStatus) error
	UpdateExecutionError(id int, error string) error
	UpdateExecutionCompletedAt(id int, completedAt time.Time) error
	UpdateExecutionWorkflowExecutionData(id int, data *WorkflowExecutionData) error
	UpdateExecutionActivityExecutionData(id int, data *ActivityExecutionData) error
	UpdateExecutionSideEffectExecutionData(id int, data *SideEffectExecutionData) error
	UpdateExecutionSagaExecutionData(id int, data *SagaExecutionData) error
	GetLatestExecution(entityID int) (*Execution, error)
	GetExecutionHistory(entityID int) ([]*Execution, error)
	GetFailedExecutions(entityID int) ([]*Execution, error)
	GetAverageExecutionTime(entityID int) (time.Duration, error)
	GetExecutionAttempts(entityID int) (int, error)
	DeleteExecution(id int) error

	// Queue methods
	AddQueue(queue *Queue) error
	GetQueue(id int) (*Queue, error)
	GetQueueByName(name string) (*Queue, error)
	UpdateQueue(queue *Queue) error
	UpdateQueueName(id int, name string) error
	BatchUpdateQueueName(ids []int, name string) error
	ListQueues() ([]*Queue, error)
	GetQueueCount() (int, error)
	GetQueuesByEntityCount(minCount, maxCount int) ([]*Queue, error)
	GetEmptyQueues() ([]*Queue, error)
	GetActiveQueues() ([]*Queue, error)
	DeleteQueue(id int) error

	// Maintenance methods
	Clear() error
}

// DefaultDatabase is an in-memory implementation of Database.
type DefaultDatabase struct {
	runs        map[int]*Run
	versions    map[int]*Version
	hierarchies map[int]*Hierarchy
	entities    map[int]*Entity
	executions  map[int]*Execution
	queues      map[int]*Queue
	mu          deadlock.RWMutex

	runIDCounter       int64
	versionIDCounter   int64
	entityIDCounter    int64
	executionIDCounter int64
	hierarchyIDCounter int64
	queueIDCounter     int64
}

func NewDefaultDatabase() *DefaultDatabase {
	db := &DefaultDatabase{
		runs:        make(map[int]*Run),
		versions:    make(map[int]*Version),
		hierarchies: make(map[int]*Hierarchy),
		entities:    make(map[int]*Entity),
		executions:  make(map[int]*Execution),
		queues:      make(map[int]*Queue),

		runIDCounter:       0,
		versionIDCounter:   0,
		entityIDCounter:    0,
		executionIDCounter: 0,
		hierarchyIDCounter: 0,
		queueIDCounter:     1, // Starting from 1 for the default queue
	}

	// Initialize default queue
	db.queues[1] = &Queue{
		ID:        1,
		Name:      "default",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Entities:  []*Entity{},
	}
	return db
}

// Run methods
func (db *DefaultDatabase) AddRun(run *Run) error {
	runID := int(atomic.AddInt64(&db.runIDCounter, 1))
	run.ID = runID
	run.CreatedAt = time.Now()
	run.UpdatedAt = time.Now()

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
		return nil, ErrRunNotFound
	}

	return db.copyRun(run), nil
}

func (db *DefaultDatabase) UpdateRun(run *Run) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.runs[run.ID]; !exists {
		return ErrRunNotFound
	}

	run.UpdatedAt = time.Now()
	db.runs[run.ID] = run
	return nil
}

func (db *DefaultDatabase) GetRunsCount() (int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return len(db.runs), nil
}

func (db *DefaultDatabase) GetRunsByDateRange(start, end time.Time) ([]*Run, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []*Run
	for _, run := range db.runs {
		if (run.CreatedAt.Equal(start) || run.CreatedAt.After(start)) &&
			(run.CreatedAt.Equal(end) || run.CreatedAt.Before(end)) {
			results = append(results, db.copyRun(run))
		}
	}
	return results, nil
}

func (db *DefaultDatabase) GetActiveRuns() ([]*Run, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []*Run
	for _, run := range db.runs {
		if run.Status != string(StatusCompleted) && run.Status != string(StatusFailed) {
			results = append(results, db.copyRun(run))
		}
	}
	return results, nil
}

func (db *DefaultDatabase) UpdateRunStatus(id int, status string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	run, exists := db.runs[id]
	if !exists {
		return ErrRunNotFound
	}

	run.Status = status
	run.UpdatedAt = time.Now()
	return nil
}

func (db *DefaultDatabase) UpdateRunUpdatedAt(id int) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	run, exists := db.runs[id]
	if !exists {
		return ErrRunNotFound
	}

	run.UpdatedAt = time.Now()
	return nil
}

func (db *DefaultDatabase) BatchUpdateRunStatus(ids []int, status string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, id := range ids {
		run, exists := db.runs[id]
		if !exists {
			continue
		}
		run.Status = status
		run.UpdatedAt = time.Now()
	}
	return nil
}

func (db *DefaultDatabase) DeleteRun(id int) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.runs[id]; !exists {
		return ErrRunNotFound
	}

	delete(db.runs, id)
	return nil
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
	return nil, ErrVersionNotFound
}

func (db *DefaultDatabase) SetVersion(version *Version) error {
	versionID := int(atomic.AddInt64(&db.versionIDCounter, 1))
	version.ID = versionID

	db.mu.Lock()
	defer db.mu.Unlock()

	db.versions[version.ID] = version
	return nil
}

func (db *DefaultDatabase) GetVersionsCount(entityID int) (int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	count := 0
	for _, version := range db.versions {
		if version.EntityID == entityID {
			count++
		}
	}
	return count, nil
}

func (db *DefaultDatabase) GetLatestVersion(entityID int, changeID string) (*Version, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var latest *Version
	for _, version := range db.versions {
		if version.EntityID == entityID && version.ChangeID == changeID {
			if latest == nil || version.ID > latest.ID {
				latest = version
			}
		}
	}

	if latest == nil {
		return nil, ErrVersionNotFound
	}
	return latest, nil
}

func (db *DefaultDatabase) GetVersionsByDateRange(entityID int, start, end time.Time) ([]*Version, error) {
	// Note: Since Version doesn't have timestamps in the original struct,
	// this implementation uses ID as a proxy for time ordering
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []*Version
	for _, version := range db.versions {
		if version.EntityID == entityID {
			results = append(results, version)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ID < results[j].ID
	})

	return results, nil
}

func (db *DefaultDatabase) BatchDeleteVersions(ids []int) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, id := range ids {
		delete(db.versions, id)
	}
	return nil
}

func (db *DefaultDatabase) DeleteVersion(id int) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.versions[id]; !exists {
		return ErrVersionNotFound
	}

	delete(db.versions, id)
	return nil
}

// Hierarchy methods
func (db *DefaultDatabase) AddHierarchy(hierarchy *Hierarchy) error {
	hierarchyID := int(atomic.AddInt64(&db.hierarchyIDCounter, 1))
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
	return nil, ErrHierarchyNotFound
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

func (db *DefaultDatabase) GetHierarchiesByParentEntity(parentEntityID int) ([]*Hierarchy, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var result []*Hierarchy
	for _, h := range db.hierarchies {
		if h.ParentEntityID == parentEntityID {
			result = append(result, h)
		}
	}
	return result, nil
}

func (db *DefaultDatabase) GetHierarchyDepth(entityID int) (int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	depth := 0
	currentID := entityID

	for {
		var parent *Hierarchy
		for _, h := range db.hierarchies {
			if h.ChildEntityID == currentID {
				parent = h
				break
			}
		}

		if parent == nil {
			break
		}

		depth++
		currentID = parent.ParentEntityID
	}

	return depth, nil
}

func (db *DefaultDatabase) GetRootEntity(entityID int) (*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	currentID := entityID
	for {
		var parent *Hierarchy
		for _, h := range db.hierarchies {
			if h.ChildEntityID == currentID {
				parent = h
				break
			}
		}

		if parent == nil {
			// We've found the root
			if entity, exists := db.entities[currentID]; exists {
				return entity, nil
			}
			return nil, ErrEntityNotFound
		}

		currentID = parent.ParentEntityID
	}
}

func (db *DefaultDatabase) GetLeafEntities(entityID int) ([]*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var leaves []*Entity
	seen := make(map[int]bool)

	var traverse func(int) error
	traverse = func(id int) error {
		if seen[id] {
			return nil
		}
		seen[id] = true

		hasChildren := false
		for _, h := range db.hierarchies {
			if h.ParentEntityID == id {
				hasChildren = true
				if err := traverse(h.ChildEntityID); err != nil {
					return err
				}
			}
		}

		if !hasChildren {
			entity, exists := db.entities[id]
			if !exists {
				return ErrEntityNotFound
			}
			leaves = append(leaves, entity)
		}

		return nil
	}

	if err := traverse(entityID); err != nil {
		return nil, err
	}

	return leaves, nil
}

func (db *DefaultDatabase) GetSiblingEntities(entityID int) ([]*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// First find the parent
	var parent *Hierarchy
	for _, h := range db.hierarchies {
		if h.ChildEntityID == entityID {
			parent = h
			break
		}
	}

	if parent == nil {
		// No parent means no siblings
		return []*Entity{}, nil
	}

	// Now find all children of the same parent
	var siblings []*Entity
	for _, h := range db.hierarchies {
		if h.ParentEntityID == parent.ParentEntityID && h.ChildEntityID != entityID {
			if entity, exists := db.entities[h.ChildEntityID]; exists {
				siblings = append(siblings, entity)
			}
		}
	}

	return siblings, nil
}

func (db *DefaultDatabase) GetDescendants(entityID int) ([]*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var descendants []*Entity
	seen := make(map[int]bool)

	var traverse func(int) error
	traverse = func(id int) error {
		if seen[id] {
			return nil
		}
		seen[id] = true

		for _, h := range db.hierarchies {
			if h.ParentEntityID == id {
				entity, exists := db.entities[h.ChildEntityID]
				if !exists {
					return ErrEntityNotFound
				}
				descendants = append(descendants, entity)
				if err := traverse(h.ChildEntityID); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if err := traverse(entityID); err != nil {
		return nil, err
	}

	return descendants, nil
}

func (db *DefaultDatabase) GetAncestors(entityID int) ([]*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var ancestors []*Entity
	seen := make(map[int]bool)

	currentID := entityID
	for {
		if seen[currentID] {
			break
		}
		seen[currentID] = true

		var parent *Hierarchy
		for _, h := range db.hierarchies {
			if h.ChildEntityID == currentID {
				parent = h
				break
			}
		}

		if parent == nil {
			break
		}

		entity, exists := db.entities[parent.ParentEntityID]
		if !exists {
			return nil, ErrEntityNotFound
		}
		ancestors = append(ancestors, entity)
		currentID = parent.ParentEntityID
	}

	return ancestors, nil
}

func (db *DefaultDatabase) BatchAddHierarchies(hierarchies []*Hierarchy) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, hierarchy := range hierarchies {
		hierarchyID := int(atomic.AddInt64(&db.hierarchyIDCounter, 1))
		hierarchy.ID = hierarchyID
		db.hierarchies[hierarchy.ID] = hierarchy
	}
	return nil
}

func (db *DefaultDatabase) DeleteHierarchy(id int) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.hierarchies[id]; !exists {
		return ErrHierarchyNotFound
	}

	delete(db.hierarchies, id)
	return nil
}

// Entity methods
func (db *DefaultDatabase) AddEntity(entity *Entity) error {
	entityID := int(atomic.AddInt64(&db.entityIDCounter, 1))
	entity.ID = entityID
	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()

	db.mu.Lock()
	defer db.mu.Unlock()

	db.entities[entity.ID] = entity

	// Add the entity to its Run
	if run, exists := db.runs[entity.RunID]; exists {
		run.Entities = append(run.Entities, entity)
	}

	return nil
}

func (db *DefaultDatabase) HasEntity(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	_, exists := db.entities[id]
	return exists, nil
}

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

func (db *DefaultDatabase) GetEntitiesCount(filters ...EntityFilter) (int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	count := 0
	for _, entity := range db.entities {
		matches := true
		for _, filter := range filters {
			if !filter.Apply(entity) {
				matches = false
				break
			}
		}
		if matches {
			count++
		}
	}
	return count, nil
}

func (db *DefaultDatabase) GetEntitiesByStatus(status EntityStatus) ([]*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []*Entity
	for _, entity := range db.entities {
		if entity.Status == status {
			results = append(results, entity)
		}
	}
	return results, nil
}

func (db *DefaultDatabase) GetEntitiesByType(entityType EntityType) ([]*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []*Entity
	for _, entity := range db.entities {
		if entity.Type == entityType {
			results = append(results, entity)
		}
	}
	return results, nil
}

func (db *DefaultDatabase) GetEntitiesByHandlerName(handlerName string) ([]*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []*Entity
	for _, entity := range db.entities {
		if entity.HandlerName == handlerName {
			results = append(results, entity)
		}
	}
	return results, nil
}

func (db *DefaultDatabase) GetEntitiesByStepID(stepID string) ([]*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []*Entity
	for _, entity := range db.entities {
		if entity.StepID == stepID {
			results = append(results, entity)
		}
	}
	return results, nil
}

func (db *DefaultDatabase) GetEntitiesByRunID(runID int) ([]*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []*Entity
	for _, entity := range db.entities {
		if entity.RunID == runID {
			results = append(results, entity)
		}
	}
	return results, nil
}

func (db *DefaultDatabase) UpdateEntityStatus(id int, status EntityStatus) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.entities[id]
	if !exists {
		return errors.Join(fmt.Errorf("entity %d", id), ErrEntityNotFound)
	}
	entity.Status = status
	entity.UpdatedAt = time.Now()
	return nil
}

func (db *DefaultDatabase) BatchUpdateEntityStatus(ids []int, status EntityStatus) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, id := range ids {
		if entity, exists := db.entities[id]; exists {
			entity.Status = status
			entity.UpdatedAt = time.Now()
		}
	}
	return nil
}

func (db *DefaultDatabase) UpdateEntityPaused(id int, paused bool) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.entities[id]
	if !exists {
		return errors.Join(fmt.Errorf("entity %d", id), ErrEntityNotFound)
	}
	entity.Paused = paused
	entity.UpdatedAt = time.Now()
	return nil
}

func (db *DefaultDatabase) BatchUpdateEntityPaused(ids []int, paused bool) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, id := range ids {
		if entity, exists := db.entities[id]; exists {
			entity.Paused = paused
			entity.UpdatedAt = time.Now()
		}
	}
	return nil
}

func (db *DefaultDatabase) UpdateEntityRetryState(id int, retryState *RetryState) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.entities[id]
	if !exists {
		return errors.Join(fmt.Errorf("entity %d", id), ErrEntityNotFound)
	}
	entity.RetryState = retryState
	entity.UpdatedAt = time.Now()
	return nil
}

func (db *DefaultDatabase) UpdateEntityRetryPolicy(id int, policy *retryPolicyInternal) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.entities[id]
	if !exists {
		return errors.Join(fmt.Errorf("entity %d", id), ErrEntityNotFound)
	}
	entity.RetryPolicy = policy
	entity.UpdatedAt = time.Now()
	return nil
}

func (db *DefaultDatabase) UpdateEntityResumable(id int, resumable bool) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.entities[id]
	if !exists {
		return errors.Join(fmt.Errorf("entity %d", id), ErrEntityNotFound)
	}
	entity.Resumable = resumable
	entity.UpdatedAt = time.Now()
	return nil
}

func (db *DefaultDatabase) UpdateEntityWorkflowData(id int, workflowData *WorkflowData) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.entities[id]
	if !exists {
		return errors.Join(fmt.Errorf("entity %d", id), ErrEntityNotFound)
	}
	entity.WorkflowData = workflowData
	entity.UpdatedAt = time.Now()
	return nil
}

func (db *DefaultDatabase) UpdateEntityActivityData(id int, activityData *ActivityData) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.entities[id]
	if !exists {
		return errors.Join(fmt.Errorf("entity %d", id), ErrEntityNotFound)
	}
	entity.ActivityData = activityData
	entity.UpdatedAt = time.Now()
	return nil
}

func (db *DefaultDatabase) UpdateEntitySideEffectData(id int, sideEffectData *SideEffectData) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.entities[id]
	if !exists {
		return errors.Join(fmt.Errorf("entity %d", id), ErrEntityNotFound)
	}
	entity.SideEffectData = sideEffectData
	entity.UpdatedAt = time.Now()
	return nil
}

func (db *DefaultDatabase) UpdateEntitySagaData(id int, sagaData *SagaData) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.entities[id]
	if !exists {
		return errors.Join(fmt.Errorf("entity %d", id), ErrEntityNotFound)
	}
	entity.SagaData = sagaData
	entity.UpdatedAt = time.Now()
	return nil
}

func (db *DefaultDatabase) UpdateEntityHandlerInfo(id int, handlerInfo *HandlerInfo) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.entities[id]
	if !exists {
		return errors.Join(fmt.Errorf("entity %d", id), ErrEntityNotFound)
	}
	entity.HandlerInfo = handlerInfo
	entity.UpdatedAt = time.Now()
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
			if entity, exists := db.entities[hierarchy.ChildEntityID]; exists && entity.Type == entityType {
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

func (db *DefaultDatabase) GetEntitiesByQueueID(queueID int) ([]*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []*Entity
	for _, entity := range db.entities {
		if entity.QueueID == queueID {
			results = append(results, entity)
		}
	}
	return results, nil
}

func (db *DefaultDatabase) GetEntitiesByRetryCount(count int) ([]*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []*Entity
	for _, entity := range db.entities {
		if entity.RetryState != nil && entity.RetryState.Attempts == count {
			results = append(results, entity)
		}
	}
	return results, nil
}

func (db *DefaultDatabase) GetFailedEntities() ([]*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []*Entity
	for _, entity := range db.entities {
		if entity.Status == StatusFailed {
			results = append(results, entity)
		}
	}
	return results, nil
}

func (db *DefaultDatabase) GetPausedEntities() ([]*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []*Entity
	for _, entity := range db.entities {
		if entity.Status == StatusPaused {
			results = append(results, entity)
		}
	}
	return results, nil
}

func (db *DefaultDatabase) DeleteEntity(id int) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.entities[id]; !exists {
		return errors.Join(fmt.Errorf("entity %d", id), ErrEntityNotFound)
	}

	delete(db.entities, id)
	return nil
}

// Execution methods
func (db *DefaultDatabase) AddExecution(execution *Execution) error {
	executionID := int(atomic.AddInt64(&db.executionIDCounter, 1))
	execution.ID = executionID
	execution.CreatedAt = time.Now()
	execution.UpdatedAt = time.Now()

	db.mu.Lock()
	defer db.mu.Unlock()

	db.executions[execution.ID] = execution
	return nil
}

func (db *DefaultDatabase) BatchAddExecutions(executions []*Execution) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, execution := range executions {
		executionID := int(atomic.AddInt64(&db.executionIDCounter, 1))
		execution.ID = executionID
		execution.CreatedAt = time.Now()
		execution.UpdatedAt = time.Now()
		db.executions[execution.ID] = execution
	}
	return nil
}

func (db *DefaultDatabase) GetExecution(id int) (*Execution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	execution, exists := db.executions[id]
	if !exists {
		return nil, ErrExecutionNotFound
	}
	return execution, nil
}

func (db *DefaultDatabase) GetExecutionsCount(filters ...ExecutionFilter) (int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	count := 0
	for _, execution := range db.executions {
		matches := true
		for _, filter := range filters {
			if !filter.Apply(execution) {
				matches = false
				break
			}
		}
		if matches {
			count++
		}
	}
	return count, nil
}

func (db *DefaultDatabase) GetExecutionsByStatus(status ExecutionStatus) ([]*Execution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []*Execution
	for _, execution := range db.executions {
		if execution.Status == status {
			results = append(results, execution)
		}
	}
	return results, nil
}

func (db *DefaultDatabase) GetExecutionsByEntityID(entityID int) ([]*Execution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []*Execution
	for _, execution := range db.executions {
		if execution.EntityID == entityID {
			results = append(results, execution)
		}
	}
	return results, nil
}

func (db *DefaultDatabase) GetExecutionsByDateRange(start, end time.Time) ([]*Execution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []*Execution
	for _, execution := range db.executions {
		if (execution.StartedAt.Equal(start) || execution.StartedAt.After(start)) &&
			(execution.StartedAt.Equal(end) || execution.StartedAt.Before(end)) {
			results = append(results, execution)
		}
	}
	return results, nil
}

func (db *DefaultDatabase) UpdateExecutionStatus(id int, status ExecutionStatus) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	execution, exists := db.executions[id]
	if !exists {
		return ErrExecutionNotFound
	}
	execution.Status = status
	execution.UpdatedAt = time.Now()
	return nil
}

func (db *DefaultDatabase) BatchUpdateExecutionStatus(ids []int, status ExecutionStatus) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, id := range ids {
		if execution, exists := db.executions[id]; exists {
			execution.Status = status
			execution.UpdatedAt = time.Now()
		}
	}
	return nil
}

func (db *DefaultDatabase) UpdateExecutionError(id int, err string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	execution, exists := db.executions[id]
	if !exists {
		return ErrExecutionNotFound
	}
	execution.Error = err
	execution.UpdatedAt = time.Now()
	return nil
}

func (db *DefaultDatabase) UpdateExecutionCompletedAt(id int, completedAt time.Time) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	execution, exists := db.executions[id]
	if !exists {
		return ErrExecutionNotFound
	}
	execution.CompletedAt = &completedAt
	execution.UpdatedAt = time.Now()
	return nil
}

func (db *DefaultDatabase) UpdateExecutionWorkflowExecutionData(id int, data *WorkflowExecutionData) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	execution, exists := db.executions[id]
	if !exists {
		return ErrExecutionNotFound
	}
	execution.WorkflowExecutionData = data
	execution.UpdatedAt = time.Now()
	return nil
}

func (db *DefaultDatabase) UpdateExecutionActivityExecutionData(id int, data *ActivityExecutionData) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	execution, exists := db.executions[id]
	if !exists {
		return ErrExecutionNotFound
	}
	execution.ActivityExecutionData = data
	execution.UpdatedAt = time.Now()
	return nil
}

func (db *DefaultDatabase) UpdateExecutionSideEffectExecutionData(id int, data *SideEffectExecutionData) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	execution, exists := db.executions[id]
	if !exists {
		return ErrExecutionNotFound
	}
	execution.SideEffectExecutionData = data
	execution.UpdatedAt = time.Now()
	return nil
}

func (db *DefaultDatabase) UpdateExecutionSagaExecutionData(id int, data *SagaExecutionData) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	execution, exists := db.executions[id]
	if !exists {
		return ErrExecutionNotFound
	}
	execution.SagaExecutionData = data
	execution.UpdatedAt = time.Now()
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
	return latestExecution, nil
}

func (db *DefaultDatabase) GetExecutionHistory(entityID int) ([]*Execution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var history []*Execution
	for _, execution := range db.executions {
		if execution.EntityID == entityID {
			history = append(history, execution)
		}
	}

	// Sort by ID to ensure chronological order
	sort.Slice(history, func(i, j int) bool {
		return history[i].ID < history[j].ID
	})

	return history, nil
}

func (db *DefaultDatabase) GetFailedExecutions(entityID int) ([]*Execution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var failed []*Execution
	for _, execution := range db.executions {
		if execution.EntityID == entityID && execution.Status == ExecutionStatusFailed {
			failed = append(failed, execution)
		}
	}
	return failed, nil
}

func (db *DefaultDatabase) GetAverageExecutionTime(entityID int) (time.Duration, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var totalDuration time.Duration
	var count int

	for _, execution := range db.executions {
		if execution.EntityID == entityID && execution.CompletedAt != nil {
			duration := execution.CompletedAt.Sub(execution.StartedAt)
			totalDuration += duration
			count++
		}
	}

	if count == 0 {
		return 0, nil
	}

	return totalDuration / time.Duration(count), nil
}

func (db *DefaultDatabase) GetExecutionAttempts(entityID int) (int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	attempts := 0
	for _, execution := range db.executions {
		if execution.EntityID == entityID {
			attempts++
		}
	}
	return attempts, nil
}

func (db *DefaultDatabase) DeleteExecution(id int) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.executions[id]; !exists {
		return ErrExecutionNotFound
	}

	delete(db.executions, id)
	return nil
}

// Queue methods
func (db *DefaultDatabase) AddQueue(queue *Queue) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Check if queue with same name exists
	for _, q := range db.queues {
		if q.Name == queue.Name {
			return ErrQueueExists
		}
	}

	queueID := int(atomic.AddInt64(&db.queueIDCounter, 1))
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

func (db *DefaultDatabase) UpdateQueueName(id int, name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if queue, exists := db.queues[id]; exists {
		queue.Name = name
		queue.UpdatedAt = time.Now()
		return nil
	}
	return errors.Join(ErrQueueNotFound)
}

func (db *DefaultDatabase) BatchUpdateQueueName(ids []int, name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, id := range ids {
		if queue, exists := db.queues[id]; exists {
			queue.Name = name
			queue.UpdatedAt = time.Now()
		}
	}
	return nil
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

func (db *DefaultDatabase) GetQueueCount() (int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return len(db.queues), nil
}

func (db *DefaultDatabase) GetQueuesByEntityCount(minCount, maxCount int) ([]*Queue, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []*Queue
	for _, queue := range db.queues {
		entityCount := len(queue.Entities)
		if entityCount >= minCount && entityCount <= maxCount {
			results = append(results, queue)
		}
	}
	return results, nil
}

func (db *DefaultDatabase) GetEmptyQueues() ([]*Queue, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []*Queue
	for _, queue := range db.queues {
		if len(queue.Entities) == 0 {
			results = append(results, queue)
		}
	}
	return results, nil
}

func (db *DefaultDatabase) GetActiveQueues() ([]*Queue, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []*Queue
	for _, queue := range db.queues {
		hasActiveEntity := false
		for _, entity := range queue.Entities {
			if entity.Status != StatusCompleted && entity.Status != StatusFailed {
				hasActiveEntity = true
				break
			}
		}
		if hasActiveEntity {
			results = append(results, queue)
		}
	}
	return results, nil
}

func (db *DefaultDatabase) DeleteQueue(id int) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.queues[id]; !exists {
		return errors.Join(ErrQueueNotFound)
	}

	delete(db.queues, id)
	return nil
}

func (db *DefaultDatabase) copyRun(run *Run) *Run {
	copyRun := *run
	copyRun.Entities = make([]*Entity, len(run.Entities))
	copy(copyRun.Entities, run.Entities)
	return &copyRun
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
