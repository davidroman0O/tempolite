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
	GetEntityStatus(id int) (EntityStatus, error)
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

// Filter types for entities and executions
type EntityFilter interface {
	Apply(*Entity) bool
}

type ExecutionFilter interface {
	Apply(*Execution) bool
}

// Common entity filters
type EntityStatusFilter struct {
	Status EntityStatus
}

func (f EntityStatusFilter) Apply(e *Entity) bool {
	return e.Status == f.Status
}

type EntityTypeFilter struct {
	Type EntityType
}

func (f EntityTypeFilter) Apply(e *Entity) bool {
	return e.Type == f.Type
}

type EntityHandlerNameFilter struct {
	HandlerName string
}

func (f EntityHandlerNameFilter) Apply(e *Entity) bool {
	return e.HandlerName == f.HandlerName
}

type EntityStepIDFilter struct {
	StepID string
}

func (f EntityStepIDFilter) Apply(e *Entity) bool {
	return e.StepID == f.StepID
}

type EntityRunIDFilter struct {
	RunID int
}

func (f EntityRunIDFilter) Apply(e *Entity) bool {
	return e.RunID == f.RunID
}

type EntityQueueIDFilter struct {
	QueueID int
}

func (f EntityQueueIDFilter) Apply(e *Entity) bool {
	return e.QueueID == f.QueueID
}

type EntityRetryCountFilter struct {
	Count int
}

func (f EntityRetryCountFilter) Apply(e *Entity) bool {
	return e.RetryState != nil && e.RetryState.Attempts == f.Count
}

type EntityPausedFilter struct {
	Paused bool
}

func (f EntityPausedFilter) Apply(e *Entity) bool {
	return e.Paused == f.Paused
}

// Common execution filters
type ExecutionStatusFilter struct {
	Status ExecutionStatus
}

func (f ExecutionStatusFilter) Apply(e *Execution) bool {
	return e.Status == f.Status
}

type ExecutionEntityIDFilter struct {
	EntityID int
}

func (f ExecutionEntityIDFilter) Apply(e *Execution) bool {
	return e.EntityID == f.EntityID
}

type ExecutionDateRangeFilter struct {
	Start time.Time
	End   time.Time
}

func (f ExecutionDateRangeFilter) Apply(e *Execution) bool {
	return (e.StartedAt.Equal(f.Start) || e.StartedAt.After(f.Start)) &&
		(e.StartedAt.Equal(f.End) || e.StartedAt.Before(f.End))
}

type ExecutionAttemptFilter struct {
	Attempt int
}

func (f ExecutionAttemptFilter) Apply(e *Execution) bool {
	return e.Attempt == f.Attempt
}

type ExecutionErrorFilter struct {
	HasError bool
}

func (f ExecutionErrorFilter) Apply(e *Execution) bool {
	return (e.Error != "") == f.HasError
}

type ExecutionCompletedFilter struct {
	IsCompleted bool
}

func (f ExecutionCompletedFilter) Apply(e *Execution) bool {
	return (e.CompletedAt != nil) == f.IsCompleted
}

// Composite filters
type AndEntityFilter struct {
	Filters []EntityFilter
}

func (f AndEntityFilter) Apply(e *Entity) bool {
	for _, filter := range f.Filters {
		if !filter.Apply(e) {
			return false
		}
	}
	return true
}

type OrEntityFilter struct {
	Filters []EntityFilter
}

func (f OrEntityFilter) Apply(e *Entity) bool {
	for _, filter := range f.Filters {
		if filter.Apply(e) {
			return true
		}
	}
	return false
}

type AndExecutionFilter struct {
	Filters []ExecutionFilter
}

func (f AndExecutionFilter) Apply(e *Execution) bool {
	for _, filter := range f.Filters {
		if !filter.Apply(e) {
			return false
		}
	}
	return true
}

type OrExecutionFilter struct {
	Filters []ExecutionFilter
}

func (f OrExecutionFilter) Apply(e *Execution) bool {
	for _, filter := range f.Filters {
		if filter.Apply(e) {
			return true
		}
	}
	return false
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

// Copy functions for deep copying

func copyRun(run *Run) *Run {
	if run == nil {
		return nil
	}
	runCopy := *run
	if run.Entities != nil {
		runCopy.Entities = make([]*Entity, len(run.Entities))
		for i, entity := range run.Entities {
			runCopy.Entities[i] = copyEntity(entity)
		}
	}
	if run.Hierarchies != nil {
		runCopy.Hierarchies = make([]*Hierarchy, len(run.Hierarchies))
		for i, hierarchy := range run.Hierarchies {
			runCopy.Hierarchies[i] = copyHierarchy(hierarchy)
		}
	}
	return &runCopy
}

func copyEntity(entity *Entity) *Entity {
	if entity == nil {
		return nil
	}
	entityCopy := *entity

	// Copy slices
	if entity.Executions != nil {
		entityCopy.Executions = make([]*Execution, len(entity.Executions))
		for i, exec := range entity.Executions {
			entityCopy.Executions[i] = copyExecution(exec)
		}
	}
	if entity.Versions != nil {
		entityCopy.Versions = make([]*Version, len(entity.Versions))
		for i, version := range entity.Versions {
			entityCopy.Versions[i] = copyVersion(version)
		}
	}

	// Copy pointers
	entityCopy.Run = nil // Avoid copying the Run to prevent circular references
	entityCopy.Queue = nil
	entityCopy.WorkflowData = copyWorkflowData(entity.WorkflowData)
	entityCopy.ActivityData = copyActivityData(entity.ActivityData)
	entityCopy.SagaData = copySagaData(entity.SagaData)
	entityCopy.SideEffectData = copySideEffectData(entity.SideEffectData)
	entityCopy.HandlerInfo = copyHandlerInfo(entity.HandlerInfo)
	if entity.RetryState != nil {
		retryStateCopy := *entity.RetryState
		entityCopy.RetryState = &retryStateCopy
	}
	if entity.RetryPolicy != nil {
		retryPolicyCopy := *entity.RetryPolicy
		entityCopy.RetryPolicy = &retryPolicyCopy
	}
	return &entityCopy
}

func copyExecutions(executions []*Execution) []*Execution {
	if executions == nil {
		return nil
	}

	// Create a new slice for the copied executions
	copied := make([]*Execution, len(executions))
	for i, exec := range executions {
		copied[i] = copyExecution(exec) // `copyExecution` already exists in your code
	}
	return copied
}

func copyExecution(execution *Execution) *Execution {
	if execution == nil {
		return nil
	}
	execCopy := *execution

	// Copy pointers
	if execution.CompletedAt != nil {
		completedAtCopy := *execution.CompletedAt
		execCopy.CompletedAt = &completedAtCopy
	}

	execCopy.Entity = nil // Avoid copying the Entity to prevent circular references

	// Copy execution data
	execCopy.WorkflowExecutionData = copyWorkflowExecutionData(execution.WorkflowExecutionData)
	execCopy.ActivityExecutionData = copyActivityExecutionData(execution.ActivityExecutionData)
	execCopy.SagaExecutionData = copySagaExecutionData(execution.SagaExecutionData)
	execCopy.SideEffectExecutionData = copySideEffectExecutionData(execution.SideEffectExecutionData)

	return &execCopy
}

func copyVersion(version *Version) *Version {
	if version == nil {
		return nil
	}
	versionCopy := *version
	if version.Data != nil {
		versionCopy.Data = make(map[string]interface{})
		for k, v := range version.Data {
			versionCopy.Data[k] = v
		}
	}
	return &versionCopy
}

func copyHierarchy(h *Hierarchy) *Hierarchy {
	if h == nil {
		return nil
	}
	hCopy := *h
	return &hCopy
}

func copyQueue(queue *Queue) *Queue {
	if queue == nil {
		return nil
	}
	queueCopy := *queue
	if queue.Entities != nil {
		queueCopy.Entities = make([]*Entity, len(queue.Entities))
		for i, entity := range queue.Entities {
			queueCopy.Entities[i] = copyEntity(entity)
		}
	}
	return &queueCopy
}

func copyActivityData(data *ActivityData) *ActivityData {
	if data == nil {
		return nil
	}
	dataCopy := *data

	// Copy Input and Output slices
	if data.Input != nil {
		dataCopy.Input = make([][]byte, len(data.Input))
		for i, input := range data.Input {
			if input != nil {
				inputCopy := make([]byte, len(input))
				copy(inputCopy, input)
				dataCopy.Input[i] = inputCopy
			}
		}
	}
	if data.Output != nil {
		dataCopy.Output = make([][]byte, len(data.Output))
		for i, output := range data.Output {
			if output != nil {
				outputCopy := make([]byte, len(output))
				copy(outputCopy, output)
				dataCopy.Output[i] = outputCopy
			}
		}
	}

	if data.ScheduledFor != nil {
		scheduledForCopy := *data.ScheduledFor
		dataCopy.ScheduledFor = &scheduledForCopy
	}

	return &dataCopy
}

func copySagaData(data *SagaData) *SagaData {
	if data == nil {
		return nil
	}
	dataCopy := *data

	if data.CompensationData != nil {
		dataCopy.CompensationData = make([][]byte, len(data.CompensationData))
		for i, compData := range data.CompensationData {
			if compData != nil {
				compDataCopy := make([]byte, len(compData))
				copy(compDataCopy, compData)
				dataCopy.CompensationData[i] = compDataCopy
			}
		}
	}

	return &dataCopy
}

func copySideEffectData(data *SideEffectData) *SideEffectData {
	if data == nil {
		return nil
	}
	dataCopy := *data
	return &dataCopy
}

func copyWorkflowData(data *WorkflowData) *WorkflowData {
	if data == nil {
		return nil
	}
	dataCopy := *data
	// Copy fields if there are any
	return &dataCopy
}

func copyHandlerInfo(info *HandlerInfo) *HandlerInfo {
	if info == nil {
		return nil
	}
	infoCopy := *info
	// Copy fields if there are any
	return &infoCopy
}

func copyWorkflowExecutionData(data *WorkflowExecutionData) *WorkflowExecutionData {
	if data == nil {
		return nil
	}
	dataCopy := *data
	// Copy fields if there are any
	return &dataCopy
}

func copyActivityExecutionData(data *ActivityExecutionData) *ActivityExecutionData {
	if data == nil {
		return nil
	}
	dataCopy := *data

	if data.LastHeartbeat != nil {
		lastHeartbeatCopy := *data.LastHeartbeat
		dataCopy.LastHeartbeat = &lastHeartbeatCopy
	}

	if data.Outputs != nil {
		dataCopy.Outputs = make([][]byte, len(data.Outputs))
		for i, output := range data.Outputs {
			if output != nil {
				outputCopy := make([]byte, len(output))
				copy(outputCopy, output)
				dataCopy.Outputs[i] = outputCopy
			}
		}
	}
	return &dataCopy
}

func copySagaExecutionData(data *SagaExecutionData) *SagaExecutionData {
	if data == nil {
		return nil
	}
	dataCopy := *data

	if data.LastHeartbeat != nil {
		lastHeartbeatCopy := *data.LastHeartbeat
		dataCopy.LastHeartbeat = &lastHeartbeatCopy
	}

	if data.Output != nil {
		dataCopy.Output = make([][]byte, len(data.Output))
		for i, output := range data.Output {
			if output != nil {
				outputCopy := make([]byte, len(output))
				copy(outputCopy, output)
				dataCopy.Output[i] = outputCopy
			}
		}
	}

	return &dataCopy
}

func copySideEffectExecutionData(data *SideEffectExecutionData) *SideEffectExecutionData {
	if data == nil {
		return nil
	}
	dataCopy := *data

	if data.Outputs != nil {
		dataCopy.Outputs = make([][]byte, len(data.Outputs))
		for i, output := range data.Outputs {
			if output != nil {
				outputCopy := make([]byte, len(output))
				copy(outputCopy, output)
				dataCopy.Outputs[i] = outputCopy
			}
		}
	}

	return &dataCopy
}

// Methods implementation

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

	return copyRun(run), nil
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
			results = append(results, copyRun(run))
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
			results = append(results, copyRun(run))
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
			return copyVersion(version), nil
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
	return copyVersion(latest), nil
}

func (db *DefaultDatabase) GetVersionsByDateRange(entityID int, start, end time.Time) ([]*Version, error) {
	// Note: Since Version doesn't have timestamps in the original struct,
	// this implementation uses ID as a proxy for time ordering
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []*Version
	for _, version := range db.versions {
		if version.EntityID == entityID {
			results = append(results, copyVersion(version))
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
			return copyHierarchy(h), nil
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
			result = append(result, copyHierarchy(h))
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
			result = append(result, copyHierarchy(h))
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
				return copyEntity(entity), nil
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
			leaves = append(leaves, copyEntity(entity))
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
				siblings = append(siblings, copyEntity(entity))
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
				descendants = append(descendants, copyEntity(entity))
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
		ancestors = append(ancestors, copyEntity(entity))
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
	return copyEntity(entity), nil
}

func (db *DefaultDatabase) GetEntityStatus(id int) (EntityStatus, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	entity, exists := db.entities[id]
	if !exists {
		return "", errors.Join(fmt.Errorf("entity %d", id), ErrEntityNotFound)
	}
	return entity.Status, nil
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
			results = append(results, copyEntity(entity))
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
			results = append(results, copyEntity(entity))
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
			results = append(results, copyEntity(entity))
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
			results = append(results, copyEntity(entity))
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
			results = append(results, copyEntity(entity))
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
			return copyEntity(entity), nil
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
				entity.mu.RLock()              // Add a read lock specific to the `Entity`
				defer entity.mu.RUnlock()      // Ensure this lock is released after copying
				return copyEntity(entity), nil // Now it's safe to copy the entity
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
			result = append(result, copyEntity(entity))
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
			results = append(results, copyEntity(entity))
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
			results = append(results, copyEntity(entity))
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
			results = append(results, copyEntity(entity))
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
			results = append(results, copyEntity(entity))
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
	return copyExecution(execution), nil
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
			results = append(results, copyExecution(execution))
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
			results = append(results, copyExecution(execution))
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
			results = append(results, copyExecution(execution))
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

	if latestExecution == nil {
		return nil, ErrExecutionNotFound
	}

	return copyExecution(latestExecution), nil
}

func (db *DefaultDatabase) GetExecutionHistory(entityID int) ([]*Execution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var history []*Execution
	for _, execution := range db.executions {
		if execution.EntityID == entityID {
			history = append(history, copyExecution(execution))
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
			failed = append(failed, copyExecution(execution))
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
		return nil, ErrQueueNotFound
	}
	return copyQueue(queue), nil
}

func (db *DefaultDatabase) GetQueueByName(name string) (*Queue, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, queue := range db.queues {
		if queue.Name == name {
			return copyQueue(queue), nil
		}
	}
	return nil, ErrQueueNotFound
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
	return ErrQueueNotFound
}

func (db *DefaultDatabase) UpdateQueueName(id int, name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if queue, exists := db.queues[id]; exists {
		queue.Name = name
		queue.UpdatedAt = time.Now()
		return nil
	}
	return ErrQueueNotFound
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
		queues = append(queues, copyQueue(q))
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
			results = append(results, copyQueue(queue))
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
			results = append(results, copyQueue(queue))
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
			results = append(results, copyQueue(queue))
		}
	}
	return results, nil
}

func (db *DefaultDatabase) DeleteQueue(id int) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.queues[id]; !exists {
		return ErrQueueNotFound
	}

	delete(db.queues, id)
	return nil
}

// Maintenance methods

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
