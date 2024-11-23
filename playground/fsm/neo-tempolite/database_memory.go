package tempolite

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/sasha-s/go-deadlock"
)

// DefaultDatabase implementation
// TODO: the default database should constantly clear itself from obsolete data to prevent growing indefinitely
// DefaultDatabase implementation
type DefaultDatabase struct {
	mu deadlock.RWMutex

	// Core maps
	runs        map[int]*Run
	versions    map[int]*Version
	hierarchies map[int]*Hierarchy
	queues      map[int]*Queue

	workflowEntities   map[int]*WorkflowEntity
	activityEntities   map[int]*ActivityEntity
	sagaEntities       map[int]*SagaEntity
	sideEffectEntities map[int]*SideEffectEntity

	workflowExecutions   map[int]*WorkflowExecution
	activityExecutions   map[int]*ActivityExecution
	sagaExecutions       map[int]*SagaExecution
	sideEffectExecutions map[int]*SideEffectExecution

	// New relationship maps
	runToWorkflows map[int][]int // runID -> workflowEntityIDs
	workflowToRun  map[int]int   // workflowEntityID -> runID

	workflowToChildren map[int]map[EntityType][]int // workflowEntityID -> type -> entityIDs
	entityToWorkflow   map[EntityType]map[int]int   // type -> entityID -> workflowEntityID

	queueToWorkflows map[int][]int // queueID -> workflowEntityIDs
	workflowToQueue  map[int]int   // workflowEntityID -> queueID

	workflowToVersions map[int][]int // workflowEntityID -> versionIDs

	// Counters
	runCounter       int64
	versionCounter   int64
	hierarchyCounter int64
	queueCounter     int64

	workflowEntityCounter   int64
	activityEntityCounter   int64
	sagaEntityCounter       int64
	sideEffectEntityCounter int64

	workflowExecutionCounter   int64
	activityExecutionCounter   int64
	sagaExecutionCounter       int64
	sideEffectExecutionCounter int64
}

// NewDefaultDatabase creates a new instance of DefaultDatabase
func NewDefaultDatabase() *DefaultDatabase {
	db := &DefaultDatabase{
		// Core maps
		workflowEntities:   make(map[int]*WorkflowEntity),
		activityEntities:   make(map[int]*ActivityEntity),
		sagaEntities:       make(map[int]*SagaEntity),
		sideEffectEntities: make(map[int]*SideEffectEntity),

		workflowExecutions:   make(map[int]*WorkflowExecution),
		activityExecutions:   make(map[int]*ActivityExecution),
		sagaExecutions:       make(map[int]*SagaExecution),
		sideEffectExecutions: make(map[int]*SideEffectExecution),

		runs:        make(map[int]*Run),
		versions:    make(map[int]*Version),
		hierarchies: make(map[int]*Hierarchy),
		queues:      make(map[int]*Queue),

		// New relationship maps
		runToWorkflows: make(map[int][]int),
		workflowToRun:  make(map[int]int),

		workflowToChildren: make(map[int]map[EntityType][]int),
		entityToWorkflow: map[EntityType]map[int]int{
			EntityTypeActivity:   make(map[int]int),
			EntityTypeSaga:       make(map[int]int),
			EntityTypeSideEffect: make(map[int]int),
		},

		queueToWorkflows: make(map[int][]int),
		workflowToQueue:  make(map[int]int),

		workflowToVersions: make(map[int][]int),

		queueCounter: 1,
	}

	// Initialize default queue
	db.queues[1] = &Queue{
		ID:        1,
		Name:      "default",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Entities:  []*WorkflowEntity{},
	}

	return db
}

// Run methods
func (db *DefaultDatabase) GetRunProperties(id int, getters ...RunPropertyGetter) error {
	db.mu.RLock()
	run, exists := db.runs[id]
	if !exists {
		db.mu.RUnlock()
		return errors.Join(fmt.Errorf("run %d", id), ErrRunNotFound)
	}

	runCopy := copyRun(run)
	db.mu.RUnlock()

	for _, getter := range getters {
		if err := getter(runCopy); err != nil {
			return err
		}
	}
	return nil
}

func (db *DefaultDatabase) SetRunProperties(id int, setters ...RunPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	run, exists := db.runs[id]
	if !exists {
		return errors.Join(fmt.Errorf("run %d", id), ErrRunNotFound)
	}

	runCopy := copyRun(run)
	for _, setter := range setters {
		if err := setter(runCopy); err != nil {
			return err
		}
	}

	runCopy.UpdatedAt = time.Now()
	db.runs[id] = runCopy
	return nil
}

func (db *DefaultDatabase) BatchGetRunProperties(ids []int, getters ...RunPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, id := range ids {
		run, exists := db.runs[id]
		if !exists {
			continue
		}

		runCopy := copyRun(run)
		for _, getter := range getters {
			if err := getter(runCopy); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) BatchSetRunProperties(ids []int, setters ...RunPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, id := range ids {
		run, exists := db.runs[id]
		if !exists {
			continue
		}

		runCopy := copyRun(run)
		for _, setter := range setters {
			if err := setter(runCopy); err != nil {
				return err
			}
		}

		runCopy.UpdatedAt = time.Now()
		db.runs[id] = runCopy
	}
	return nil
}

// Version methods
func (db *DefaultDatabase) GetVersionProperties(id int, getters ...VersionPropertyGetter) error {
	db.mu.RLock()
	version, exists := db.versions[id]
	if !exists {
		db.mu.RUnlock()
		return errors.Join(fmt.Errorf("version %d", id), ErrVersionNotFound)
	}

	versionCopy := copyVersion(version)
	db.mu.RUnlock()

	for _, getter := range getters {
		if err := getter(versionCopy); err != nil {
			return err
		}
	}
	return nil
}

func (db *DefaultDatabase) SetVersionProperties(id int, setters ...VersionPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	version, exists := db.versions[id]
	if !exists {
		return errors.Join(fmt.Errorf("version %d", id), ErrVersionNotFound)
	}

	versionCopy := copyVersion(version)
	for _, setter := range setters {
		if err := setter(versionCopy); err != nil {
			return err
		}
	}

	db.versions[id] = versionCopy
	return nil
}

func (db *DefaultDatabase) BatchGetVersionProperties(ids []int, getters ...VersionPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, id := range ids {
		version, exists := db.versions[id]
		if !exists {
			continue
		}

		versionCopy := copyVersion(version)
		for _, getter := range getters {
			if err := getter(versionCopy); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) BatchSetVersionProperties(ids []int, setters ...VersionPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, id := range ids {
		version, exists := db.versions[id]
		if !exists {
			continue
		}

		versionCopy := copyVersion(version)
		for _, setter := range setters {
			if err := setter(versionCopy); err != nil {
				return err
			}
		}

		db.versions[id] = versionCopy
	}
	return nil
}

// Hierarchy methods
func (db *DefaultDatabase) GetHierarchyProperties(id int, getters ...HierarchyPropertyGetter) error {
	db.mu.RLock()
	hierarchy, exists := db.hierarchies[id]
	if !exists {
		db.mu.RUnlock()
		return errors.Join(fmt.Errorf("hierarchy %d", id), ErrHierarchyNotFound)
	}

	hierarchyCopy := copyHierarchy(hierarchy)
	db.mu.RUnlock()

	for _, getter := range getters {
		if err := getter(hierarchyCopy); err != nil {
			return err
		}
	}
	return nil
}

func (db *DefaultDatabase) SetHierarchyProperties(id int, setters ...HierarchyPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	hierarchy, exists := db.hierarchies[id]
	if !exists {
		return errors.Join(fmt.Errorf("hierarchy %d", id), ErrHierarchyNotFound)
	}

	hierarchyCopy := copyHierarchy(hierarchy)
	for _, setter := range setters {
		if err := setter(hierarchyCopy); err != nil {
			return err
		}
	}

	db.hierarchies[id] = hierarchyCopy
	return nil
}

func (db *DefaultDatabase) BatchGetHierarchyProperties(ids []int, getters ...HierarchyPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, id := range ids {
		hierarchy, exists := db.hierarchies[id]
		if !exists {
			continue
		}

		hierarchyCopy := copyHierarchy(hierarchy)
		for _, getter := range getters {
			if err := getter(hierarchyCopy); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) BatchSetHierarchyProperties(ids []int, setters ...HierarchyPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, id := range ids {
		hierarchy, exists := db.hierarchies[id]
		if !exists {
			continue
		}

		hierarchyCopy := copyHierarchy(hierarchy)
		for _, setter := range setters {
			if err := setter(hierarchyCopy); err != nil {
				return err
			}
		}

		db.hierarchies[id] = hierarchyCopy
	}
	return nil
}

// Queue methods
func (db *DefaultDatabase) GetQueueProperties(id int, getters ...QueuePropertyGetter) error {
	db.mu.RLock()
	queue, exists := db.queues[id]
	if !exists {
		db.mu.RUnlock()
		return errors.Join(fmt.Errorf("queue %d", id), ErrQueueNotFound)
	}

	queueCopy := copyQueue(queue)
	db.mu.RUnlock()

	for _, getter := range getters {
		if err := getter(queueCopy); err != nil {
			return err
		}
	}
	return nil
}

func (db *DefaultDatabase) SetQueueProperties(id int, setters ...QueuePropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	queue, exists := db.queues[id]
	if !exists {
		return errors.Join(fmt.Errorf("queue %d", id), ErrQueueNotFound)
	}

	queueCopy := copyQueue(queue)
	for _, setter := range setters {
		if err := setter(queueCopy); err != nil {
			return err
		}
	}

	queueCopy.UpdatedAt = time.Now()
	db.queues[id] = queueCopy
	return nil
}

func (db *DefaultDatabase) BatchGetQueueProperties(ids []int, getters ...QueuePropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, id := range ids {
		queue, exists := db.queues[id]
		if !exists {
			continue
		}

		queueCopy := copyQueue(queue)
		for _, getter := range getters {
			if err := getter(queueCopy); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) BatchSetQueueProperties(ids []int, setters ...QueuePropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, id := range ids {
		queue, exists := db.queues[id]
		if !exists {
			continue
		}

		queueCopy := copyQueue(queue)
		for _, setter := range setters {
			if err := setter(queueCopy); err != nil {
				return err
			}
		}

		queueCopy.UpdatedAt = time.Now()
		db.queues[id] = queueCopy
	}
	return nil
}

// Workflow Entity operations
func (db *DefaultDatabase) AddWorkflowEntity(entity *WorkflowEntity) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	entityCopy := copyWorkflowEntity(entity)
	entityCopy.ID = int(atomic.AddInt64(&db.workflowEntityCounter, 1))
	entityCopy.CreatedAt = time.Now()
	entityCopy.UpdatedAt = time.Now()

	db.workflowEntities[entityCopy.ID] = entityCopy
	return entityCopy.ID, nil
}

func (db *DefaultDatabase) GetWorkflowEntity(id int) (*WorkflowEntity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	entity, exists := db.workflowEntities[id]
	if !exists {
		return nil, ErrWorkflowEntityNotFound
	}

	return copyWorkflowEntity(entity), nil
}

func (db *DefaultDatabase) UpdateWorkflowEntity(entity *WorkflowEntity) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.workflowEntities[entity.ID]; !exists {
		return ErrWorkflowEntityNotFound
	}

	entityCopy := copyWorkflowEntity(entity)
	entityCopy.UpdatedAt = time.Now()
	db.workflowEntities[entity.ID] = entityCopy
	return nil
}

func (db *DefaultDatabase) GetWorkflowEntityProperties(id int, getters ...WorkflowEntityPropertyGetter) error {
	db.mu.RLock()
	entity, exists := db.workflowEntities[id]
	if !exists {
		db.mu.RUnlock()
		return ErrWorkflowEntityNotFound
	}

	entityCopy := copyWorkflowEntity(entity)
	db.mu.RUnlock()

	for _, getter := range getters {
		if err := getter(entityCopy); err != nil {
			return err
		}
	}

	return nil
}

func (db *DefaultDatabase) SetWorkflowEntityProperties(id int, setters ...WorkflowEntityPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.workflowEntities[id]
	if !exists {
		return ErrWorkflowEntityNotFound
	}

	entityCopy := copyWorkflowEntity(entity)
	for _, setter := range setters {
		if err := setter(entityCopy); err != nil {
			return err
		}
	}

	entityCopy.UpdatedAt = time.Now()
	db.workflowEntities[id] = entityCopy
	return nil
}

// Activity Entity operations

func (db *DefaultDatabase) AddActivityEntity(entity *ActivityEntity) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	entityCopy := copyActivityEntity(entity)
	entityCopy.ID = int(atomic.AddInt64(&db.activityEntityCounter, 1))
	entityCopy.CreatedAt = time.Now()
	entityCopy.UpdatedAt = time.Now()

	db.activityEntities[entityCopy.ID] = entityCopy
	return entityCopy.ID, nil
}

func (db *DefaultDatabase) GetActivityEntity(id int) (*ActivityEntity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	entity, exists := db.activityEntities[id]
	if !exists {
		return nil, ErrActivityEntityNotFound
	}

	return copyActivityEntity(entity), nil
}

func (db *DefaultDatabase) UpdateActivityEntity(entity *ActivityEntity) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.activityEntities[entity.ID]; !exists {
		return ErrActivityEntityNotFound
	}

	entityCopy := copyActivityEntity(entity)
	entityCopy.UpdatedAt = time.Now()
	db.activityEntities[entity.ID] = entityCopy
	return nil
}

func (db *DefaultDatabase) GetActivityEntityProperties(id int, getters ...ActivityEntityPropertyGetter) error {
	db.mu.RLock()
	entity, exists := db.activityEntities[id]
	if !exists {
		db.mu.RUnlock()
		return ErrActivityEntityNotFound
	}

	entityCopy := copyActivityEntity(entity)
	db.mu.RUnlock()

	for _, getter := range getters {
		if err := getter(entityCopy); err != nil {
			return err
		}
	}

	return nil
}

func (db *DefaultDatabase) SetActivityEntityProperties(id int, setters ...ActivityEntityPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.activityEntities[id]
	if !exists {
		return ErrActivityEntityNotFound
	}

	entityCopy := copyActivityEntity(entity)
	for _, setter := range setters {
		if err := setter(entityCopy); err != nil {
			return err
		}
	}

	entityCopy.UpdatedAt = time.Now()
	db.activityEntities[id] = entityCopy
	return nil
}

// Saga Entity operations
func (db *DefaultDatabase) AddSagaEntity(entity *SagaEntity) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	entityCopy := copySagaEntity(entity)
	entityCopy.ID = int(atomic.AddInt64(&db.sagaEntityCounter, 1))
	entityCopy.CreatedAt = time.Now()
	entityCopy.UpdatedAt = time.Now()

	db.sagaEntities[entityCopy.ID] = entityCopy
	return entityCopy.ID, nil
}

func (db *DefaultDatabase) GetSagaEntity(id int) (*SagaEntity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	entity, exists := db.sagaEntities[id]
	if !exists {
		return nil, ErrSagaEntityNotFound
	}

	return copySagaEntity(entity), nil
}

func (db *DefaultDatabase) UpdateSagaEntity(entity *SagaEntity) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.sagaEntities[entity.ID]; !exists {
		return ErrSagaEntityNotFound
	}

	entityCopy := copySagaEntity(entity)
	entityCopy.UpdatedAt = time.Now()
	db.sagaEntities[entity.ID] = entityCopy
	return nil
}

func (db *DefaultDatabase) GetSagaEntityProperties(id int, getters ...SagaEntityPropertyGetter) error {
	db.mu.RLock()
	entity, exists := db.sagaEntities[id]
	if !exists {
		db.mu.RUnlock()
		return ErrSagaEntityNotFound
	}

	entityCopy := copySagaEntity(entity)
	db.mu.RUnlock()

	for _, getter := range getters {
		if err := getter(entityCopy); err != nil {
			return err
		}
	}

	return nil
}

func (db *DefaultDatabase) SetSagaEntityProperties(id int, setters ...SagaEntityPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.sagaEntities[id]
	if !exists {
		return ErrSagaEntityNotFound
	}

	entityCopy := copySagaEntity(entity)
	for _, setter := range setters {
		if err := setter(entityCopy); err != nil {
			return err
		}
	}

	entityCopy.UpdatedAt = time.Now()
	db.sagaEntities[id] = entityCopy
	return nil
}

// SideEffect Entity operations
func (db *DefaultDatabase) AddSideEffectEntity(entity *SideEffectEntity) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	entityCopy := copySideEffectEntity(entity)
	entityCopy.ID = int(atomic.AddInt64(&db.sideEffectEntityCounter, 1))
	entityCopy.CreatedAt = time.Now()
	entityCopy.UpdatedAt = time.Now()

	db.sideEffectEntities[entityCopy.ID] = entityCopy
	return entityCopy.ID, nil
}

func (db *DefaultDatabase) GetSideEffectEntity(id int) (*SideEffectEntity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	entity, exists := db.sideEffectEntities[id]
	if !exists {
		return nil, ErrSideEffectEntityNotFound
	}

	return copySideEffectEntity(entity), nil
}

func (db *DefaultDatabase) UpdateSideEffectEntity(entity *SideEffectEntity) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.sideEffectEntities[entity.ID]; !exists {
		return ErrSideEffectEntityNotFound
	}

	entityCopy := copySideEffectEntity(entity)
	entityCopy.UpdatedAt = time.Now()
	db.sideEffectEntities[entity.ID] = entityCopy
	return nil
}

func (db *DefaultDatabase) GetSideEffectEntityProperties(id int, getters ...SideEffectEntityPropertyGetter) error {
	db.mu.RLock()
	entity, exists := db.sideEffectEntities[id]
	if !exists {
		db.mu.RUnlock()
		return ErrSideEffectEntityNotFound
	}

	entityCopy := copySideEffectEntity(entity)
	db.mu.RUnlock()

	for _, getter := range getters {
		if err := getter(entityCopy); err != nil {
			return err
		}
	}

	return nil
}

func (db *DefaultDatabase) SetSideEffectEntityProperties(id int, setters ...SideEffectEntityPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.sideEffectEntities[id]
	if !exists {
		return ErrSideEffectEntityNotFound
	}

	entityCopy := copySideEffectEntity(entity)
	for _, setter := range setters {
		if err := setter(entityCopy); err != nil {
			return err
		}
	}

	entityCopy.UpdatedAt = time.Now()
	db.sideEffectEntities[id] = entityCopy
	return nil
}

// Workflow Execution operations
func (db *DefaultDatabase) AddWorkflowExecution(exec *WorkflowExecution) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	execCopy := copyWorkflowExecution(exec)
	execCopy.ID = int(atomic.AddInt64(&db.workflowExecutionCounter, 1))
	execCopy.CreatedAt = time.Now()
	execCopy.UpdatedAt = time.Now()

	db.workflowExecutions[execCopy.ID] = execCopy
	return execCopy.ID, nil
}

func (db *DefaultDatabase) GetWorkflowExecution(id int) (*WorkflowExecution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	exec, exists := db.workflowExecutions[id]
	if !exists {
		return nil, ErrWorkflowExecutionNotFound
	}

	return copyWorkflowExecution(exec), nil
}

func (db *DefaultDatabase) GetWorkflowExecutionProperties(id int, getters ...WorkflowExecutionPropertyGetter) error {
	db.mu.RLock()
	exec, exists := db.workflowExecutions[id]
	if !exists {
		db.mu.RUnlock()
		return ErrWorkflowExecutionNotFound
	}

	execCopy := copyWorkflowExecution(exec)
	db.mu.RUnlock()

	for _, getter := range getters {
		if err := getter(execCopy); err != nil {
			return err
		}
	}

	return nil
}

func (db *DefaultDatabase) SetWorkflowExecutionProperties(id int, setters ...WorkflowExecutionPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	exec, exists := db.workflowExecutions[id]
	if !exists {
		return ErrWorkflowExecutionNotFound
	}

	execCopy := copyWorkflowExecution(exec)
	for _, setter := range setters {
		if err := setter(execCopy); err != nil {
			return err
		}
	}

	execCopy.UpdatedAt = time.Now()
	db.workflowExecutions[id] = execCopy
	return nil
}

// Activity Execution operations
func (db *DefaultDatabase) AddActivityExecution(exec *ActivityExecution) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	execCopy := copyActivityExecution(exec)
	execCopy.ID = int(atomic.AddInt64(&db.activityExecutionCounter, 1))
	execCopy.CreatedAt = time.Now()
	execCopy.UpdatedAt = time.Now()

	db.activityExecutions[execCopy.ID] = execCopy
	return execCopy.ID, nil
}

func (db *DefaultDatabase) GetActivityExecution(id int) (*ActivityExecution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	exec, exists := db.activityExecutions[id]
	if !exists {
		return nil, ErrActivityExecutionNotFound
	}

	return copyActivityExecution(exec), nil
}

func (db *DefaultDatabase) GetActivityExecutionProperties(id int, getters ...ActivityExecutionPropertyGetter) error {
	db.mu.RLock()
	exec, exists := db.activityExecutions[id]
	if !exists {
		db.mu.RUnlock()
		return ErrActivityExecutionNotFound
	}

	execCopy := copyActivityExecution(exec)
	db.mu.RUnlock()

	for _, getter := range getters {
		if err := getter(execCopy); err != nil {
			return err
		}
	}

	return nil
}

func (db *DefaultDatabase) SetActivityExecutionProperties(id int, setters ...ActivityExecutionPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	exec, exists := db.activityExecutions[id]
	if !exists {
		return ErrActivityExecutionNotFound
	}

	execCopy := copyActivityExecution(exec)
	for _, setter := range setters {
		if err := setter(execCopy); err != nil {
			return err
		}
	}

	execCopy.UpdatedAt = time.Now()
	db.activityExecutions[id] = execCopy
	return nil
}

// Saga Execution operations
func (db *DefaultDatabase) AddSagaExecution(exec *SagaExecution) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	execCopy := copySagaExecution(exec)
	execCopy.ID = int(atomic.AddInt64(&db.sagaExecutionCounter, 1))
	execCopy.CreatedAt = time.Now()
	execCopy.UpdatedAt = time.Now()

	db.sagaExecutions[execCopy.ID] = execCopy
	return execCopy.ID, nil
}

func (db *DefaultDatabase) GetSagaExecution(id int) (*SagaExecution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	exec, exists := db.sagaExecutions[id]
	if !exists {
		return nil, ErrSagaExecutionNotFound
	}

	return copySagaExecution(exec), nil
}

func (db *DefaultDatabase) GetSagaExecutionProperties(id int, getters ...SagaExecutionPropertyGetter) error {
	db.mu.RLock()
	exec, exists := db.sagaExecutions[id]
	if !exists {
		db.mu.RUnlock()
		return ErrSagaExecutionNotFound
	}

	execCopy := copySagaExecution(exec)
	db.mu.RUnlock()

	for _, getter := range getters {
		if err := getter(execCopy); err != nil {
			return err
		}
	}

	return nil
}

func (db *DefaultDatabase) SetSagaExecutionProperties(id int, setters ...SagaExecutionPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	exec, exists := db.sagaExecutions[id]
	if !exists {
		return ErrSagaExecutionNotFound
	}

	execCopy := copySagaExecution(exec)
	for _, setter := range setters {
		if err := setter(execCopy); err != nil {
			return err
		}
	}

	execCopy.UpdatedAt = time.Now()
	db.sagaExecutions[id] = execCopy
	return nil
}

// SideEffect Execution operations
func (db *DefaultDatabase) AddSideEffectExecution(exec *SideEffectExecution) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	execCopy := copySideEffectExecution(exec)
	execCopy.ID = int(atomic.AddInt64(&db.sideEffectExecutionCounter, 1))
	execCopy.CreatedAt = time.Now()
	execCopy.UpdatedAt = time.Now()

	db.sideEffectExecutions[execCopy.ID] = execCopy
	return execCopy.ID, nil
}

func (db *DefaultDatabase) GetSideEffectExecution(id int) (*SideEffectExecution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	exec, exists := db.sideEffectExecutions[id]
	if !exists {
		return nil, ErrSideEffectExecutionNotFound
	}

	return copySideEffectExecution(exec), nil
}

func (db *DefaultDatabase) GetSideEffectExecutionProperties(id int, getters ...SideEffectExecutionPropertyGetter) error {
	db.mu.RLock()
	exec, exists := db.sideEffectExecutions[id]
	if !exists {
		db.mu.RUnlock()
		return ErrSideEffectExecutionNotFound
	}

	execCopy := copySideEffectExecution(exec)
	db.mu.RUnlock()

	for _, getter := range getters {
		if err := getter(execCopy); err != nil {
			return err
		}
	}

	return nil
}

func (db *DefaultDatabase) SetSideEffectExecutionProperties(id int, setters ...SideEffectExecutionPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	exec, exists := db.sideEffectExecutions[id]
	if !exists {
		return ErrSideEffectExecutionNotFound
	}

	execCopy := copySideEffectExecution(exec)
	for _, setter := range setters {
		if err := setter(execCopy); err != nil {
			return err
		}
	}

	execCopy.UpdatedAt = time.Now()
	db.sideEffectExecutions[id] = execCopy
	return nil
}

// Batch Entity operations
func (db *DefaultDatabase) BatchGetWorkflowEntityProperties(ids []int, getters ...WorkflowEntityPropertyGetter) error {
	db.mu.RLock()
	entities := make([]*WorkflowEntity, 0, len(ids))
	for _, id := range ids {
		if entity, exists := db.workflowEntities[id]; exists {
			entities = append(entities, copyWorkflowEntity(entity))
		} else {
			db.mu.RUnlock()
			return fmt.Errorf("workflow entity not found: %d", id)
		}
	}
	db.mu.RUnlock()

	for _, entity := range entities {
		for _, getter := range getters {
			if err := getter(entity); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) BatchSetWorkflowEntityProperties(ids []int, setters ...WorkflowEntityPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entities := make([]*WorkflowEntity, 0, len(ids))
	for _, id := range ids {
		if entity, exists := db.workflowEntities[id]; exists {
			entities = append(entities, copyWorkflowEntity(entity))
		} else {
			return fmt.Errorf("workflow entity not found: %d", id)
		}
	}

	for _, entity := range entities {
		for _, setter := range setters {
			if err := setter(entity); err != nil {
				return err
			}
		}
		entity.UpdatedAt = time.Now()
		db.workflowEntities[entity.ID] = entity
	}
	return nil
}

func (db *DefaultDatabase) BatchGetActivityEntityProperties(ids []int, getters ...ActivityEntityPropertyGetter) error {
	db.mu.RLock()
	entities := make([]*ActivityEntity, 0, len(ids))
	for _, id := range ids {
		if entity, exists := db.activityEntities[id]; exists {
			entities = append(entities, copyActivityEntity(entity))
		} else {
			db.mu.RUnlock()
			return fmt.Errorf("activity entity not found: %d", id)
		}
	}
	db.mu.RUnlock()

	for _, entity := range entities {
		for _, getter := range getters {
			if err := getter(entity); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) BatchSetActivityEntityProperties(ids []int, setters ...ActivityEntityPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entities := make([]*ActivityEntity, 0, len(ids))
	for _, id := range ids {
		if entity, exists := db.activityEntities[id]; exists {
			entities = append(entities, copyActivityEntity(entity))
		} else {
			return fmt.Errorf("activity entity not found: %d", id)
		}
	}

	for _, entity := range entities {
		for _, setter := range setters {
			if err := setter(entity); err != nil {
				return err
			}
		}
		entity.UpdatedAt = time.Now()
		db.activityEntities[entity.ID] = entity
	}
	return nil
}

func (db *DefaultDatabase) BatchGetSagaEntityProperties(ids []int, getters ...SagaEntityPropertyGetter) error {
	db.mu.RLock()
	entities := make([]*SagaEntity, 0, len(ids))
	for _, id := range ids {
		if entity, exists := db.sagaEntities[id]; exists {
			entities = append(entities, copySagaEntity(entity))
		} else {
			db.mu.RUnlock()
			return fmt.Errorf("saga entity not found: %d", id)
		}
	}
	db.mu.RUnlock()

	for _, entity := range entities {
		for _, getter := range getters {
			if err := getter(entity); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) BatchSetSagaEntityProperties(ids []int, setters ...SagaEntityPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entities := make([]*SagaEntity, 0, len(ids))
	for _, id := range ids {
		if entity, exists := db.sagaEntities[id]; exists {
			entities = append(entities, copySagaEntity(entity))
		} else {
			return fmt.Errorf("saga entity not found: %d", id)
		}
	}

	for _, entity := range entities {
		for _, setter := range setters {
			if err := setter(entity); err != nil {
				return err
			}
		}
		entity.UpdatedAt = time.Now()
		db.sagaEntities[entity.ID] = entity
	}
	return nil
}

func (db *DefaultDatabase) BatchGetSideEffectEntityProperties(ids []int, getters ...SideEffectEntityPropertyGetter) error {
	db.mu.RLock()
	entities := make([]*SideEffectEntity, 0, len(ids))
	for _, id := range ids {
		if entity, exists := db.sideEffectEntities[id]; exists {
			entities = append(entities, copySideEffectEntity(entity))
		} else {
			db.mu.RUnlock()
			return fmt.Errorf("side effect entity not found: %d", id)
		}
	}
	db.mu.RUnlock()

	for _, entity := range entities {
		for _, getter := range getters {
			if err := getter(entity); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) BatchSetSideEffectEntityProperties(ids []int, setters ...SideEffectEntityPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entities := make([]*SideEffectEntity, 0, len(ids))
	for _, id := range ids {
		if entity, exists := db.sideEffectEntities[id]; exists {
			entities = append(entities, copySideEffectEntity(entity))
		} else {
			return fmt.Errorf("side effect entity not found: %d", id)
		}
	}

	for _, entity := range entities {
		for _, setter := range setters {
			if err := setter(entity); err != nil {
				return err
			}
		}
		entity.UpdatedAt = time.Now()
		db.sideEffectEntities[entity.ID] = entity
	}
	return nil
}

// Batch Execution operations
func (db *DefaultDatabase) BatchGetWorkflowExecutionProperties(ids []int, getters ...WorkflowExecutionPropertyGetter) error {
	db.mu.RLock()
	executions := make([]*WorkflowExecution, 0, len(ids))
	for _, id := range ids {
		if exec, exists := db.workflowExecutions[id]; exists {
			executions = append(executions, copyWorkflowExecution(exec))
		} else {
			db.mu.RUnlock()
			return fmt.Errorf("workflow execution not found: %d", id)
		}
	}
	db.mu.RUnlock()

	for _, exec := range executions {
		for _, getter := range getters {
			if err := getter(exec); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) BatchSetWorkflowExecutionProperties(ids []int, setters ...WorkflowExecutionPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	executions := make([]*WorkflowExecution, 0, len(ids))
	for _, id := range ids {
		if exec, exists := db.workflowExecutions[id]; exists {
			executions = append(executions, copyWorkflowExecution(exec))
		} else {
			return fmt.Errorf("workflow execution not found: %d", id)
		}
	}

	for _, exec := range executions {
		for _, setter := range setters {
			if err := setter(exec); err != nil {
				return err
			}
		}
		exec.UpdatedAt = time.Now()
		db.workflowExecutions[exec.ID] = exec
	}
	return nil
}

func (db *DefaultDatabase) BatchGetActivityExecutionProperties(ids []int, getters ...ActivityExecutionPropertyGetter) error {
	db.mu.RLock()
	executions := make([]*ActivityExecution, 0, len(ids))
	for _, id := range ids {
		if exec, exists := db.activityExecutions[id]; exists {
			executions = append(executions, copyActivityExecution(exec))
		} else {
			db.mu.RUnlock()
			return fmt.Errorf("activity execution not found: %d", id)
		}
	}
	db.mu.RUnlock()

	for _, exec := range executions {
		for _, getter := range getters {
			if err := getter(exec); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) BatchSetActivityExecutionProperties(ids []int, setters ...ActivityExecutionPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	executions := make([]*ActivityExecution, 0, len(ids))
	for _, id := range ids {
		if exec, exists := db.activityExecutions[id]; exists {
			executions = append(executions, copyActivityExecution(exec))
		} else {
			return fmt.Errorf("activity execution not found: %d", id)
		}
	}

	for _, exec := range executions {
		for _, setter := range setters {
			if err := setter(exec); err != nil {
				return err
			}
		}
		exec.UpdatedAt = time.Now()
		db.activityExecutions[exec.ID] = exec
	}
	return nil
}

func (db *DefaultDatabase) BatchGetSagaExecutionProperties(ids []int, getters ...SagaExecutionPropertyGetter) error {
	db.mu.RLock()
	executions := make([]*SagaExecution, 0, len(ids))
	for _, id := range ids {
		if exec, exists := db.sagaExecutions[id]; exists {
			executions = append(executions, copySagaExecution(exec))
		} else {
			db.mu.RUnlock()
			return fmt.Errorf("saga execution not found: %d", id)
		}
	}
	db.mu.RUnlock()

	for _, exec := range executions {
		for _, getter := range getters {
			if err := getter(exec); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) BatchSetSagaExecutionProperties(ids []int, setters ...SagaExecutionPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	executions := make([]*SagaExecution, 0, len(ids))
	for _, id := range ids {
		if exec, exists := db.sagaExecutions[id]; exists {
			executions = append(executions, copySagaExecution(exec))
		} else {
			return fmt.Errorf("saga execution not found: %d", id)
		}
	}

	for _, exec := range executions {
		for _, setter := range setters {
			if err := setter(exec); err != nil {
				return err
			}
		}
		exec.UpdatedAt = time.Now()
		db.sagaExecutions[exec.ID] = exec
	}
	return nil
}

func (db *DefaultDatabase) BatchGetSideEffectExecutionProperties(ids []int, getters ...SideEffectExecutionPropertyGetter) error {
	db.mu.RLock()
	executions := make([]*SideEffectExecution, 0, len(ids))
	for _, id := range ids {
		if exec, exists := db.sideEffectExecutions[id]; exists {
			executions = append(executions, copySideEffectExecution(exec))
		} else {
			db.mu.RUnlock()
			return fmt.Errorf("side effect execution not found: %d", id)
		}
	}
	db.mu.RUnlock()

	for _, exec := range executions {
		for _, getter := range getters {
			if err := getter(exec); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) BatchSetSideEffectExecutionProperties(ids []int, setters ...SideEffectExecutionPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	executions := make([]*SideEffectExecution, 0, len(ids))
	for _, id := range ids {
		if exec, exists := db.sideEffectExecutions[id]; exists {
			executions = append(executions, copySideEffectExecution(exec))
		} else {
			return fmt.Errorf("side effect execution not found: %d", id)
		}
	}

	for _, exec := range executions {
		for _, setter := range setters {
			if err := setter(exec); err != nil {
				return err
			}
		}
		exec.UpdatedAt = time.Now()
		db.sideEffectExecutions[exec.ID] = exec
	}
	return nil
}

// Data Property operations
func (db *DefaultDatabase) GetWorkflowDataProperties(entityID int, getters ...WorkflowDataPropertyGetter) error {
	db.mu.RLock()
	entity, exists := db.workflowEntities[entityID]
	if !exists {
		db.mu.RUnlock()
		return ErrWorkflowEntityNotFound
	}

	if entity.WorkflowData == nil {
		db.mu.RUnlock()
		return errors.New("workflow data is nil")
	}

	dataCopy := copyWorkflowData(entity.WorkflowData)
	db.mu.RUnlock()

	for _, getter := range getters {
		if err := getter(dataCopy); err != nil {
			return err
		}
	}
	return nil
}

func (db *DefaultDatabase) SetWorkflowDataProperties(entityID int, setters ...WorkflowDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.workflowEntities[entityID]
	if !exists {
		return ErrWorkflowEntityNotFound
	}

	if entity.WorkflowData == nil {
		return errors.New("workflow data is nil")
	}

	dataCopy := copyWorkflowData(entity.WorkflowData)
	for _, setter := range setters {
		if err := setter(dataCopy); err != nil {
			return err
		}
	}

	entity.WorkflowData = dataCopy
	entity.UpdatedAt = time.Now()
	db.workflowEntities[entityID] = entity
	return nil
}

func (db *DefaultDatabase) GetActivityDataProperties(entityID int, getters ...ActivityDataPropertyGetter) error {
	db.mu.RLock()
	entity, exists := db.activityEntities[entityID]
	if !exists {
		db.mu.RUnlock()
		return ErrActivityEntityNotFound
	}

	if entity.ActivityData == nil {
		db.mu.RUnlock()
		return errors.New("activity data is nil")
	}

	dataCopy := copyActivityData(entity.ActivityData)
	db.mu.RUnlock()

	for _, getter := range getters {
		if err := getter(dataCopy); err != nil {
			return err
		}
	}
	return nil
}

func (db *DefaultDatabase) SetActivityDataProperties(entityID int, setters ...ActivityDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.activityEntities[entityID]
	if !exists {
		return ErrActivityEntityNotFound
	}

	if entity.ActivityData == nil {
		return errors.New("activity data is nil")
	}

	dataCopy := copyActivityData(entity.ActivityData)
	for _, setter := range setters {
		if err := setter(dataCopy); err != nil {
			return err
		}
	}

	entity.ActivityData = dataCopy
	entity.UpdatedAt = time.Now()
	db.activityEntities[entityID] = entity
	return nil
}

func (db *DefaultDatabase) GetSagaDataProperties(entityID int, getters ...SagaDataPropertyGetter) error {
	db.mu.RLock()
	entity, exists := db.sagaEntities[entityID]
	if !exists {
		db.mu.RUnlock()
		return ErrSagaEntityNotFound
	}

	if entity.SagaData == nil {
		db.mu.RUnlock()
		return errors.New("saga data is nil")
	}

	dataCopy := copySagaData(entity.SagaData)
	db.mu.RUnlock()

	for _, getter := range getters {
		if err := getter(dataCopy); err != nil {
			return err
		}
	}
	return nil
}

func (db *DefaultDatabase) SetSagaDataProperties(entityID int, setters ...SagaDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.sagaEntities[entityID]
	if !exists {
		return ErrSagaEntityNotFound
	}

	if entity.SagaData == nil {
		return errors.New("saga data is nil")
	}

	dataCopy := copySagaData(entity.SagaData)
	for _, setter := range setters {
		if err := setter(dataCopy); err != nil {
			return err
		}
	}

	entity.SagaData = dataCopy
	entity.UpdatedAt = time.Now()
	db.sagaEntities[entityID] = entity
	return nil
}

func (db *DefaultDatabase) GetSideEffectDataProperties(entityID int, getters ...SideEffectDataPropertyGetter) error {
	db.mu.RLock()
	entity, exists := db.sideEffectEntities[entityID]
	if !exists {
		db.mu.RUnlock()
		return ErrSideEffectEntityNotFound
	}

	if entity.SideEffectData == nil {
		db.mu.RUnlock()
		return errors.New("side effect data is nil")
	}

	dataCopy := copySideEffectData(entity.SideEffectData)
	db.mu.RUnlock()

	for _, getter := range getters {
		if err := getter(dataCopy); err != nil {
			return err
		}
	}
	return nil
}

func (db *DefaultDatabase) SetSideEffectDataProperties(entityID int, setters ...SideEffectDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.sideEffectEntities[entityID]
	if !exists {
		return ErrSideEffectEntityNotFound
	}

	if entity.SideEffectData == nil {
		return errors.New("side effect data is nil")
	}

	dataCopy := copySideEffectData(entity.SideEffectData)
	for _, setter := range setters {
		if err := setter(dataCopy); err != nil {
			return err
		}
	}

	entity.SideEffectData = dataCopy
	entity.UpdatedAt = time.Now()
	db.sideEffectEntities[entityID] = entity
	return nil
}
