package tempolite

import (
	"sync"
	"time"

	"github.com/k0kubun/pp/v3"
)

type MemoryDatabase struct {
	mu sync.RWMutex

	// Core counters
	runCounter       int
	versionCounter   int
	hierarchyCounter int
	queueCounter     int

	// Entity counters
	workflowEntityCounter   int
	activityEntityCounter   int
	sagaEntityCounter       int
	sideEffectEntityCounter int

	// Entity Data counters
	workflowDataCounter   int
	activityDataCounter   int
	sagaDataCounter       int
	sideEffectDataCounter int

	// Execution counters
	workflowExecutionCounter   int
	activityExecutionCounter   int
	sagaExecutionCounter       int
	sideEffectExecutionCounter int

	// Execution Data counters
	workflowExecutionDataCounter   int
	activityExecutionDataCounter   int
	sagaExecutionDataCounter       int
	sideEffectExecutionDataCounter int

	// Core maps
	runs        map[int]*Run
	versions    map[int]*Version
	hierarchies map[int]*Hierarchy
	queues      map[int]*Queue
	queueNames  map[string]int

	// Entity maps
	workflowEntities   map[int]*WorkflowEntity
	activityEntities   map[int]*ActivityEntity
	sagaEntities       map[int]*SagaEntity
	sideEffectEntities map[int]*SideEffectEntity

	// Entity Data maps
	workflowData   map[int]*WorkflowData
	activityData   map[int]*ActivityData
	sagaData       map[int]*SagaData
	sideEffectData map[int]*SideEffectData

	// Execution maps
	workflowExecutions   map[int]*WorkflowExecution
	activityExecutions   map[int]*ActivityExecution
	sagaExecutions       map[int]*SagaExecution
	sideEffectExecutions map[int]*SideEffectExecution

	// Execution Data maps
	workflowExecutionData   map[int]*WorkflowExecutionData
	activityExecutionData   map[int]*ActivityExecutionData
	sagaExecutionData       map[int]*SagaExecutionData
	sideEffectExecutionData map[int]*SideEffectExecutionData

	// Relationship maps
	entityToWorkflow   map[int]int                  // entity ID -> workflow ID
	workflowToChildren map[int]map[EntityType][]int // workflow ID -> type -> child IDs
	workflowToVersion  map[int][]int                // workflow ID -> version IDs
	workflowToQueue    map[int]int                  // workflow ID -> queue ID
	queueToWorkflows   map[int][]int                // queue ID -> workflow IDs
	runToWorkflows     map[int][]int                // run ID -> workflow IDs
}

func NewMemoryDatabase() *MemoryDatabase {
	db := &MemoryDatabase{
		// Core maps
		runs:        make(map[int]*Run),
		versions:    make(map[int]*Version),
		hierarchies: make(map[int]*Hierarchy),
		queues:      make(map[int]*Queue),
		queueNames:  make(map[string]int),

		// Entity maps
		workflowEntities:   make(map[int]*WorkflowEntity),
		activityEntities:   make(map[int]*ActivityEntity),
		sagaEntities:       make(map[int]*SagaEntity),
		sideEffectEntities: make(map[int]*SideEffectEntity),

		// Entity Data maps
		workflowData:   make(map[int]*WorkflowData),
		activityData:   make(map[int]*ActivityData),
		sagaData:       make(map[int]*SagaData),
		sideEffectData: make(map[int]*SideEffectData),

		// Execution maps
		workflowExecutions:   make(map[int]*WorkflowExecution),
		activityExecutions:   make(map[int]*ActivityExecution),
		sagaExecutions:       make(map[int]*SagaExecution),
		sideEffectExecutions: make(map[int]*SideEffectExecution),

		// Execution Data maps
		workflowExecutionData:   make(map[int]*WorkflowExecutionData),
		activityExecutionData:   make(map[int]*ActivityExecutionData),
		sagaExecutionData:       make(map[int]*SagaExecutionData),
		sideEffectExecutionData: make(map[int]*SideEffectExecutionData),

		// Relationship maps
		entityToWorkflow:   make(map[int]int),
		workflowToChildren: make(map[int]map[EntityType][]int),
		workflowToVersion:  make(map[int][]int),
		workflowToQueue:    make(map[int]int),
		queueToWorkflows:   make(map[int][]int),
		runToWorkflows:     make(map[int][]int),

		queueCounter: 1, // Starting from 1 for the default queue
	}

	// Initialize default queue
	now := time.Now()
	db.queues[1] = &Queue{
		ID:        1,
		Name:      "default",
		CreatedAt: now,
		UpdatedAt: now,
		Entities:  make([]*WorkflowEntity, 0),
	}
	db.queueNames["default"] = 1

	return db
}

func (db *MemoryDatabase) AddRun(run *Run) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.runCounter++
	run.ID = db.runCounter
	run.CreatedAt = time.Now()
	run.UpdatedAt = run.CreatedAt

	db.runs[run.ID] = copyRun(run)
	return run.ID, nil
}

func (db *MemoryDatabase) GetRun(id int, opts ...RunGetOption) (*Run, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	run, exists := db.runs[id]
	if !exists {
		return nil, ErrRunNotFound
	}

	cfg := &RunGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeWorkflows {
		if workflowIDs, ok := db.runToWorkflows[id]; ok {
			workflows := make([]*WorkflowEntity, 0, len(workflowIDs))
			for _, wfID := range workflowIDs {
				if wf, exists := db.workflowEntities[wfID]; exists {
					workflows = append(workflows, copyWorkflowEntity(wf))
				}
			}
			run.Entities = workflows
		}
	}

	if cfg.IncludeHierarchies {
		hierarchies := make([]*Hierarchy, 0)
		for _, h := range db.hierarchies {
			if h.RunID == id {
				hierarchies = append(hierarchies, copyHierarchy(h))
			}
		}
		run.Hierarchies = hierarchies
	}

	return copyRun(run), nil
}

func (db *MemoryDatabase) UpdateRun(run *Run) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.runs[run.ID]; !exists {
		return ErrRunNotFound
	}

	run.UpdatedAt = time.Now()
	db.runs[run.ID] = copyRun(run)
	return nil
}

func (db *MemoryDatabase) GetRunProperties(id int, getters ...RunPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	run, exists := db.runs[id]
	if !exists {
		return ErrRunNotFound
	}

	opts := &RunGetterOptions{}

	for _, getter := range getters {
		opt, err := getter(run)
		if err != nil {
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	if opts.IncludeWorkflows {
		workflows := make([]*WorkflowEntity, 0)
		if workflowIDs, ok := db.runToWorkflows[id]; ok {
			for _, wfID := range workflowIDs {
				if wf, exists := db.workflowEntities[wfID]; exists {
					workflows = append(workflows, copyWorkflowEntity(wf))
				}
			}
		}
		run.Entities = workflows
	}

	if opts.IncludeHierarchies {
		hierarchies := make([]*Hierarchy, 0)
		for _, h := range db.hierarchies {
			if h.RunID == id {
				hierarchies = append(hierarchies, copyHierarchy(h))
			}
		}
		run.Hierarchies = hierarchies
	}

	return nil
}

func (db *MemoryDatabase) SetRunProperties(id int, setters ...RunPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	run, exists := db.runs[id]
	if !exists {
		return ErrRunNotFound
	}

	opts := &RunSetterOptions{}

	for _, setter := range setters {
		opt, err := setter(run)
		if err != nil {
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	if opts.WorkflowID != nil {
		if _, exists := db.workflowEntities[*opts.WorkflowID]; !exists {
			return ErrWorkflowEntityNotFound
		}
		db.runToWorkflows[id] = append(db.runToWorkflows[id], *opts.WorkflowID)
	}

	run.UpdatedAt = time.Now()
	return nil
}

func (db *MemoryDatabase) AddQueue(queue *Queue) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.queueNames[queue.Name]; exists {
		return 0, ErrQueueExists
	}

	db.queueCounter++
	queue.ID = db.queueCounter
	queue.CreatedAt = time.Now()
	queue.UpdatedAt = queue.CreatedAt
	queue.Entities = make([]*WorkflowEntity, 0)

	db.queues[queue.ID] = copyQueue(queue)
	db.queueNames[queue.Name] = queue.ID
	return queue.ID, nil
}

func (db *MemoryDatabase) GetQueue(id int, opts ...QueueGetOption) (*Queue, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	queue, exists := db.queues[id]
	if !exists {
		return nil, ErrQueueNotFound
	}

	cfg := &QueueGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeWorkflows {
		if workflowIDs, ok := db.queueToWorkflows[id]; ok {
			workflows := make([]*WorkflowEntity, 0, len(workflowIDs))
			for _, wfID := range workflowIDs {
				if wf, exists := db.workflowEntities[wfID]; exists {
					workflows = append(workflows, copyWorkflowEntity(wf))
				}
			}
			queue.Entities = workflows
		}
	}

	return copyQueue(queue), nil
}

func (db *MemoryDatabase) GetQueueByName(name string, opts ...QueueGetOption) (*Queue, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	id, exists := db.queueNames[name]
	if !exists {
		return nil, ErrQueueNotFound
	}

	queue, exists := db.queues[id]
	if !exists {
		return nil, ErrQueueNotFound
	}

	cfg := &QueueGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeWorkflows {
		if workflowIDs, ok := db.queueToWorkflows[id]; ok {
			workflows := make([]*WorkflowEntity, 0, len(workflowIDs))
			for _, wfID := range workflowIDs {
				if wf, exists := db.workflowEntities[wfID]; exists {
					workflows = append(workflows, copyWorkflowEntity(wf))
				}
			}
			queue.Entities = workflows
		}
	}

	return copyQueue(queue), nil
}

func (db *MemoryDatabase) AddVersion(version *Version) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.versionCounter++
	version.ID = db.versionCounter
	version.CreatedAt = time.Now()
	version.UpdatedAt = version.CreatedAt

	db.versions[version.ID] = copyVersion(version)
	if version.EntityID != 0 {
		db.workflowToVersion[version.EntityID] = append(db.workflowToVersion[version.EntityID], version.ID)
	}
	return version.ID, nil
}

func (db *MemoryDatabase) GetVersion(id int, opts ...VersionGetOption) (*Version, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	version, exists := db.versions[id]
	if !exists {
		return nil, ErrVersionNotFound
	}

	cfg := &VersionGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		// Data is already included in the version structure
		// Just ensure we return a copy
		return copyVersion(version), nil
	}

	// If we don't need data, return version without data
	versionCopy := *version
	versionCopy.Data = nil
	return &versionCopy, nil
}

func (db *MemoryDatabase) AddHierarchy(hierarchy *Hierarchy) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.hierarchyCounter++
	hierarchy.ID = db.hierarchyCounter

	db.hierarchies[hierarchy.ID] = copyHierarchy(hierarchy)
	return hierarchy.ID, nil
}

func (db *MemoryDatabase) GetHierarchy(id int, opts ...HierarchyGetOption) (*Hierarchy, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	hierarchy, exists := db.hierarchies[id]
	if !exists {
		return nil, ErrHierarchyNotFound
	}

	cfg := &HierarchyGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	return copyHierarchy(hierarchy), nil
}

func (db *MemoryDatabase) AddWorkflowEntity(entity *WorkflowEntity) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.workflowEntityCounter++
	entity.ID = db.workflowEntityCounter

	if entity.WorkflowData != nil {
		db.workflowDataCounter++
		entity.WorkflowData.ID = db.workflowDataCounter
		entity.WorkflowData.EntityID = entity.ID
		db.workflowData[entity.WorkflowData.ID] = copyWorkflowData(entity.WorkflowData)
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = entity.CreatedAt

	// Add to default queue if no queue specified
	if entity.QueueID == 0 {
		entity.QueueID = 1
		db.queueToWorkflows[1] = append(db.queueToWorkflows[1], entity.ID)
	} else {
		db.workflowToQueue[entity.ID] = entity.QueueID
		db.queueToWorkflows[entity.QueueID] = append(db.queueToWorkflows[entity.QueueID], entity.ID)
	}

	pp.Println("add", entity)

	db.workflowEntities[entity.ID] = copyWorkflowEntity(entity)
	return entity.ID, nil
}

func (db *MemoryDatabase) AddWorkflowExecution(exec *WorkflowExecution) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.workflowExecutionCounter++
	exec.ID = db.workflowExecutionCounter

	if exec.WorkflowExecutionData != nil {
		db.workflowExecutionDataCounter++
		exec.WorkflowExecutionData.ID = db.workflowExecutionDataCounter
		exec.WorkflowExecutionData.ExecutionID = exec.ID
		db.workflowExecutionData[exec.WorkflowExecutionData.ID] = copyWorkflowExecutionData(exec.WorkflowExecutionData)
	}

	exec.CreatedAt = time.Now()
	exec.UpdatedAt = exec.CreatedAt

	db.workflowExecutions[exec.ID] = copyWorkflowExecution(exec)
	return exec.ID, nil
}

func (db *MemoryDatabase) GetWorkflowEntity(id int, opts ...WorkflowEntityGetOption) (*WorkflowEntity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	entity, exists := db.workflowEntities[id]
	if !exists {
		return nil, ErrWorkflowEntityNotFound
	}

	cfg := &WorkflowEntityGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	pp.Println("Get workflow", cfg, entity)

	if cfg.IncludeQueue {
		if queueID, ok := db.workflowToQueue[id]; ok {
			if queue, exists := db.queues[queueID]; exists {
				if entity.Edges == nil {
					entity.Edges = &WorkflowEntityEdges{}
				}
				entity.Edges.Queue = copyQueue(queue)
			}
		}
	}

	if cfg.IncludeData {
		for _, d := range db.workflowData {
			if d.EntityID == id {
				entity.WorkflowData = copyWorkflowData(d)
				break
			}
		}
	}

	return copyWorkflowEntity(entity), nil
}

func (db *MemoryDatabase) GetWorkflowEntityProperties(id int, getters ...WorkflowEntityPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	entity, exists := db.workflowEntities[id]
	if !exists {
		return ErrWorkflowEntityNotFound
	}

	opts := &WorkflowEntityGetterOptions{}

	for _, getter := range getters {
		opt, err := getter(entity)
		if err != nil {
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	if opts.IncludeVersion {
		if versionIDs, ok := db.workflowToVersion[id]; ok {
			versions := make([]*Version, 0)
			for _, vID := range versionIDs {
				if v, exists := db.versions[vID]; exists {
					versions = append(versions, copyVersion(v))
				}
			}
			if entity.Edges == nil {
				entity.Edges = &WorkflowEntityEdges{}
			}
			entity.Edges.Versions = versions
		}
	}

	if opts.IncludeQueue {
		if queueID, ok := db.workflowToQueue[id]; ok {
			if queue, exists := db.queues[queueID]; exists {
				if entity.Edges == nil {
					entity.Edges = &WorkflowEntityEdges{}
				}
				entity.Edges.Queue = copyQueue(queue)
			}
		}
	}

	if opts.IncludeChildren {
		if childMap, ok := db.workflowToChildren[id]; ok {
			if entity.Edges == nil {
				entity.Edges = &WorkflowEntityEdges{}
			}

			// Load activity children
			if activityIDs, ok := childMap[EntityActivity]; ok {
				activities := make([]*ActivityEntity, 0)
				for _, aID := range activityIDs {
					if a, exists := db.activityEntities[aID]; exists {
						activities = append(activities, copyActivityEntity(a))
					}
				}
				entity.Edges.ActivityChildren = activities
			}

			// Load saga children
			if sagaIDs, ok := childMap[EntitySaga]; ok {
				sagas := make([]*SagaEntity, 0)
				for _, sID := range sagaIDs {
					if s, exists := db.sagaEntities[sID]; exists {
						sagas = append(sagas, copySagaEntity(s))
					}
				}
				entity.Edges.SagaChildren = sagas
			}

			// Load side effect children
			if sideEffectIDs, ok := childMap[EntitySideEffect]; ok {
				sideEffects := make([]*SideEffectEntity, 0)
				for _, seID := range sideEffectIDs {
					if se, exists := db.sideEffectEntities[seID]; exists {
						sideEffects = append(sideEffects, copySideEffectEntity(se))
					}
				}
				entity.Edges.SideEffectChildren = sideEffects
			}
		}
	}

	if opts.IncludeData {
		for _, d := range db.workflowData {
			if d.EntityID == id {
				entity.WorkflowData = copyWorkflowData(d)
				break
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetWorkflowEntityProperties(id int, setters ...WorkflowEntityPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.workflowEntities[id]
	if !exists {
		return ErrWorkflowEntityNotFound
	}

	opts := &WorkflowEntitySetterOptions{}

	for _, setter := range setters {
		opt, err := setter(entity)
		if err != nil {
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	if opts.QueueID != nil {
		if _, exists := db.queues[*opts.QueueID]; !exists {
			return ErrQueueNotFound
		}

		// Remove from old queue if exists
		if oldQueueID, ok := db.workflowToQueue[id]; ok {
			if workflows, exists := db.queueToWorkflows[oldQueueID]; exists {
				for i, wID := range workflows {
					if wID == id {
						db.queueToWorkflows[oldQueueID] = append(workflows[:i], workflows[i+1:]...)
						break
					}
				}
			}
		}

		// Add to new queue
		db.workflowToQueue[id] = *opts.QueueID
		db.queueToWorkflows[*opts.QueueID] = append(db.queueToWorkflows[*opts.QueueID], id)
	}

	if opts.Version != nil {
		db.workflowToVersion[id] = append(db.workflowToVersion[id], opts.Version.ID)
	}

	if opts.ChildID != nil && opts.ChildType != nil {
		if db.workflowToChildren[id] == nil {
			db.workflowToChildren[id] = make(map[EntityType][]int)
		}
		db.workflowToChildren[id][*opts.ChildType] = append(db.workflowToChildren[id][*opts.ChildType], *opts.ChildID)
		db.entityToWorkflow[*opts.ChildID] = id
	}

	entity.UpdatedAt = time.Now()
	return nil
}

func (db *MemoryDatabase) AddActivityEntity(entity *ActivityEntity) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.activityEntityCounter++
	entity.ID = db.activityEntityCounter

	if entity.ActivityData != nil {
		db.activityDataCounter++
		entity.ActivityData.ID = db.activityDataCounter
		entity.ActivityData.EntityID = entity.ID
		db.activityData[entity.ActivityData.ID] = copyActivityData(entity.ActivityData)
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = entity.CreatedAt

	db.activityEntities[entity.ID] = copyActivityEntity(entity)
	return entity.ID, nil
}

func (db *MemoryDatabase) AddActivityExecution(exec *ActivityExecution) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.activityExecutionCounter++
	exec.ID = db.activityExecutionCounter

	if exec.ActivityExecutionData != nil {
		db.activityExecutionDataCounter++
		exec.ActivityExecutionData.ID = db.activityExecutionDataCounter
		exec.ActivityExecutionData.ExecutionID = exec.ID
		db.activityExecutionData[exec.ActivityExecutionData.ID] = copyActivityExecutionData(exec.ActivityExecutionData)
	}

	exec.CreatedAt = time.Now()
	exec.UpdatedAt = exec.CreatedAt

	db.activityExecutions[exec.ID] = copyActivityExecution(exec)
	return exec.ID, nil
}

func (db *MemoryDatabase) GetActivityEntity(id int, opts ...ActivityEntityGetOption) (*ActivityEntity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	entity, exists := db.activityEntities[id]
	if !exists {
		return nil, ErrActivityEntityNotFound
	}

	cfg := &ActivityEntityGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		for _, d := range db.activityData {
			if d.EntityID == id {
				entity.ActivityData = copyActivityData(d)
				break
			}
		}
	}

	return copyActivityEntity(entity), nil
}

func (db *MemoryDatabase) GetActivityEntityProperties(id int, getters ...ActivityEntityPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	entity, exists := db.activityEntities[id]
	if !exists {
		return ErrActivityEntityNotFound
	}

	opts := &ActivityEntityGetterOptions{}

	for _, getter := range getters {
		opt, err := getter(entity)
		if err != nil {
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	if opts.IncludeData {
		for _, d := range db.activityData {
			if d.EntityID == id {
				entity.ActivityData = copyActivityData(d)
				break
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetActivityEntityProperties(id int, setters ...ActivityEntityPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.activityEntities[id]
	if !exists {
		return ErrActivityEntityNotFound
	}

	opts := &ActivityEntitySetterOptions{}

	for _, setter := range setters {
		opt, err := setter(entity)
		if err != nil {
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	if opts.ParentWorkflowID != nil {
		if _, exists := db.workflowEntities[*opts.ParentWorkflowID]; !exists {
			return ErrWorkflowEntityNotFound
		}
		db.entityToWorkflow[id] = *opts.ParentWorkflowID
		if db.workflowToChildren[*opts.ParentWorkflowID] == nil {
			db.workflowToChildren[*opts.ParentWorkflowID] = make(map[EntityType][]int)
		}
		db.workflowToChildren[*opts.ParentWorkflowID][EntityActivity] = append(
			db.workflowToChildren[*opts.ParentWorkflowID][EntityActivity],
			id,
		)
	}

	entity.UpdatedAt = time.Now()
	return nil
}

func (db *MemoryDatabase) AddSagaEntity(entity *SagaEntity) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.sagaEntityCounter++
	entity.ID = db.sagaEntityCounter

	if entity.SagaData != nil {
		db.sagaDataCounter++
		entity.SagaData.ID = db.sagaDataCounter
		entity.SagaData.EntityID = entity.ID
		db.sagaData[entity.SagaData.ID] = copySagaData(entity.SagaData)
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = entity.CreatedAt

	db.sagaEntities[entity.ID] = copySagaEntity(entity)
	return entity.ID, nil
}

func (db *MemoryDatabase) AddSagaExecution(exec *SagaExecution) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.sagaExecutionCounter++
	exec.ID = db.sagaExecutionCounter

	if exec.SagaExecutionData != nil {
		db.sagaExecutionDataCounter++
		exec.SagaExecutionData.ID = db.sagaExecutionDataCounter
		exec.SagaExecutionData.ExecutionID = exec.ID
		db.sagaExecutionData[exec.SagaExecutionData.ID] = copySagaExecutionData(exec.SagaExecutionData)
	}

	exec.CreatedAt = time.Now()
	exec.UpdatedAt = exec.CreatedAt

	db.sagaExecutions[exec.ID] = copySagaExecution(exec)
	return exec.ID, nil
}

func (db *MemoryDatabase) GetSagaEntity(id int, opts ...SagaEntityGetOption) (*SagaEntity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	entity, exists := db.sagaEntities[id]
	if !exists {
		return nil, ErrSagaEntityNotFound
	}

	cfg := &SagaEntityGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		for _, d := range db.sagaData {
			if d.EntityID == id {
				entity.SagaData = copySagaData(d)
				break
			}
		}
	}

	return copySagaEntity(entity), nil
}

func (db *MemoryDatabase) GetSagaEntityProperties(id int, getters ...SagaEntityPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	entity, exists := db.sagaEntities[id]
	if !exists {
		return ErrSagaEntityNotFound
	}

	opts := &SagaEntityGetterOptions{}

	for _, getter := range getters {
		opt, err := getter(entity)
		if err != nil {
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	if opts.IncludeData {
		for _, d := range db.sagaData {
			if d.EntityID == id {
				entity.SagaData = copySagaData(d)
				break
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSagaEntityProperties(id int, setters ...SagaEntityPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.sagaEntities[id]
	if !exists {
		return ErrSagaEntityNotFound
	}

	opts := &SagaEntitySetterOptions{}

	for _, setter := range setters {
		opt, err := setter(entity)
		if err != nil {
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	if opts.ParentWorkflowID != nil {
		if _, exists := db.workflowEntities[*opts.ParentWorkflowID]; !exists {
			return ErrWorkflowEntityNotFound
		}
		db.entityToWorkflow[id] = *opts.ParentWorkflowID
		if db.workflowToChildren[*opts.ParentWorkflowID] == nil {
			db.workflowToChildren[*opts.ParentWorkflowID] = make(map[EntityType][]int)
		}
		db.workflowToChildren[*opts.ParentWorkflowID][EntitySaga] = append(
			db.workflowToChildren[*opts.ParentWorkflowID][EntitySaga],
			id,
		)
	}

	entity.UpdatedAt = time.Now()
	return nil
}

func (db *MemoryDatabase) AddSideEffectEntity(entity *SideEffectEntity) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.sideEffectEntityCounter++
	entity.ID = db.sideEffectEntityCounter

	if entity.SideEffectData != nil {
		db.sideEffectDataCounter++
		entity.SideEffectData.ID = db.sideEffectDataCounter
		entity.SideEffectData.EntityID = entity.ID
		db.sideEffectData[entity.SideEffectData.ID] = copySideEffectData(entity.SideEffectData)
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = entity.CreatedAt

	db.sideEffectEntities[entity.ID] = copySideEffectEntity(entity)
	return entity.ID, nil
}

func (db *MemoryDatabase) AddSideEffectExecution(exec *SideEffectExecution) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.sideEffectExecutionCounter++
	exec.ID = db.sideEffectExecutionCounter

	if exec.SideEffectExecutionData != nil {
		db.sideEffectExecutionDataCounter++
		exec.SideEffectExecutionData.ID = db.sideEffectExecutionDataCounter
		exec.SideEffectExecutionData.ExecutionID = exec.ID
		db.sideEffectExecutionData[exec.SideEffectExecutionData.ID] = copySideEffectExecutionData(exec.SideEffectExecutionData)
	}

	exec.CreatedAt = time.Now()
	exec.UpdatedAt = exec.CreatedAt

	db.sideEffectExecutions[exec.ID] = copySideEffectExecution(exec)
	return exec.ID, nil
}

func (db *MemoryDatabase) GetSideEffectEntity(id int, opts ...SideEffectEntityGetOption) (*SideEffectEntity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	entity, exists := db.sideEffectEntities[id]
	if !exists {
		return nil, ErrSideEffectEntityNotFound
	}

	cfg := &SideEffectEntityGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		for _, d := range db.sideEffectData {
			if d.EntityID == id {
				entity.SideEffectData = copySideEffectData(d)
				break
			}
		}
	}

	return copySideEffectEntity(entity), nil
}

func (db *MemoryDatabase) GetSideEffectEntityProperties(id int, getters ...SideEffectEntityPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	entity, exists := db.sideEffectEntities[id]
	if !exists {
		return ErrSideEffectEntityNotFound
	}

	opts := &SideEffectEntityGetterOptions{}

	for _, getter := range getters {
		opt, err := getter(entity)
		if err != nil {
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	if opts.IncludeData {
		for _, d := range db.sideEffectData {
			if d.EntityID == id {
				entity.SideEffectData = copySideEffectData(d)
				break
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSideEffectEntityProperties(id int, setters ...SideEffectEntityPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.sideEffectEntities[id]
	if !exists {
		return ErrSideEffectEntityNotFound
	}

	opts := &SideEffectEntitySetterOptions{}

	for _, setter := range setters {
		opt, err := setter(entity)
		if err != nil {
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	if opts.ParentWorkflowID != nil {
		if _, exists := db.workflowEntities[*opts.ParentWorkflowID]; !exists {
			return ErrWorkflowEntityNotFound
		}
		db.entityToWorkflow[id] = *opts.ParentWorkflowID
		if db.workflowToChildren[*opts.ParentWorkflowID] == nil {
			db.workflowToChildren[*opts.ParentWorkflowID] = make(map[EntityType][]int)
		}
		db.workflowToChildren[*opts.ParentWorkflowID][EntitySideEffect] = append(
			db.workflowToChildren[*opts.ParentWorkflowID][EntitySideEffect],
			id,
		)
	}

	entity.UpdatedAt = time.Now()
	return nil
}

func (db *MemoryDatabase) GetWorkflowExecution(id int, opts ...WorkflowExecutionGetOption) (*WorkflowExecution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	exec, exists := db.workflowExecutions[id]
	if !exists {
		return nil, ErrWorkflowExecutionNotFound
	}

	cfg := &WorkflowExecutionGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		for _, d := range db.workflowExecutionData {
			if d.ExecutionID == id {
				exec.WorkflowExecutionData = copyWorkflowExecutionData(d)
				break
			}
		}
	}

	return copyWorkflowExecution(exec), nil
}

func (db *MemoryDatabase) GetActivityExecution(id int, opts ...ActivityExecutionGetOption) (*ActivityExecution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	exec, exists := db.activityExecutions[id]
	if !exists {
		return nil, ErrActivityExecutionNotFound
	}

	cfg := &ActivityExecutionGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		for _, d := range db.activityExecutionData {
			if d.ExecutionID == id {
				exec.ActivityExecutionData = copyActivityExecutionData(d)
				break
			}
		}
	}

	return copyActivityExecution(exec), nil
}

func (db *MemoryDatabase) GetSagaExecution(id int, opts ...SagaExecutionGetOption) (*SagaExecution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	exec, exists := db.sagaExecutions[id]
	if !exists {
		return nil, ErrSagaExecutionNotFound
	}

	cfg := &SagaExecutionGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		for _, d := range db.sagaExecutionData {
			if d.ExecutionID == id {
				exec.SagaExecutionData = copySagaExecutionData(d)
				break
			}
		}
	}

	return copySagaExecution(exec), nil
}

func (db *MemoryDatabase) GetSideEffectExecution(id int, opts ...SideEffectExecutionGetOption) (*SideEffectExecution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	exec, exists := db.sideEffectExecutions[id]
	if !exists {
		return nil, ErrSideEffectExecutionNotFound
	}

	cfg := &SideEffectExecutionGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		for _, d := range db.sideEffectExecutionData {
			if d.ExecutionID == id {
				exec.SideEffectExecutionData = copySideEffectExecutionData(d)
				break
			}
		}
	}

	return copySideEffectExecution(exec), nil
}

// Activity Data properties
func (db *MemoryDatabase) GetActivityDataProperties(entityID int, getters ...ActivityDataPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var data *ActivityData
	for _, d := range db.activityData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}

	if data == nil {
		return ErrActivityEntityNotFound
	}

	for _, getter := range getters {
		opt, err := getter(data)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &ActivityDataGetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetActivityDataProperties(entityID int, setters ...ActivityDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	var data *ActivityData
	for _, d := range db.activityData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}

	if data == nil {
		return ErrActivityEntityNotFound
	}

	for _, setter := range setters {
		opt, err := setter(data)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &ActivityDataSetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

// Saga Data properties
func (db *MemoryDatabase) GetSagaDataProperties(entityID int, getters ...SagaDataPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var data *SagaData
	for _, d := range db.sagaData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}

	if data == nil {
		return ErrSagaEntityNotFound
	}

	for _, getter := range getters {
		opt, err := getter(data)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &SagaDataGetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSagaDataProperties(entityID int, setters ...SagaDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	var data *SagaData
	for _, d := range db.sagaData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}

	if data == nil {
		return ErrSagaEntityNotFound
	}

	for _, setter := range setters {
		opt, err := setter(data)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &SagaDataSetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

// SideEffect Data properties
func (db *MemoryDatabase) GetSideEffectDataProperties(entityID int, getters ...SideEffectDataPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var data *SideEffectData
	for _, d := range db.sideEffectData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}

	if data == nil {
		return ErrSideEffectEntityNotFound
	}

	for _, getter := range getters {
		opt, err := getter(data)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &SideEffectDataGetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSideEffectDataProperties(entityID int, setters ...SideEffectDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	var data *SideEffectData
	for _, d := range db.sideEffectData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}

	if data == nil {
		return ErrSideEffectEntityNotFound
	}

	for _, setter := range setters {
		opt, err := setter(data)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &SideEffectDataSetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

// Workflow Data properties
func (db *MemoryDatabase) GetWorkflowDataProperties(entityID int, getters ...WorkflowDataPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var data *WorkflowData
	for _, d := range db.workflowData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}

	if data == nil {
		return ErrWorkflowEntityNotFound
	}

	for _, getter := range getters {
		opt, err := getter(data)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &WorkflowDataGetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetWorkflowDataProperties(entityID int, setters ...WorkflowDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	var data *WorkflowData
	for _, d := range db.workflowData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}

	if data == nil {
		return ErrWorkflowEntityNotFound
	}

	for _, setter := range setters {
		opt, err := setter(data)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &WorkflowDataSetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

// Execution Data properties implementations
func (db *MemoryDatabase) GetWorkflowExecutionDataProperties(entityID int, getters ...WorkflowExecutionDataPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var data *WorkflowExecutionData
	for _, d := range db.workflowExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}

	if data == nil {
		return ErrWorkflowExecutionNotFound
	}

	for _, getter := range getters {
		opt, err := getter(data)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &WorkflowExecutionDataGetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetWorkflowExecutionDataProperties(entityID int, setters ...WorkflowExecutionDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	var data *WorkflowExecutionData
	for _, d := range db.workflowExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}

	if data == nil {
		return ErrWorkflowExecutionNotFound
	}

	for _, setter := range setters {
		opt, err := setter(data)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &WorkflowExecutionDataSetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

// Add similar implementations for Activity, Saga, and SideEffect Execution Data properties
func (db *MemoryDatabase) GetActivityExecutionDataProperties(entityID int, getters ...ActivityExecutionDataPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var data *ActivityExecutionData
	for _, d := range db.activityExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}

	if data == nil {
		return ErrActivityExecutionNotFound
	}

	for _, getter := range getters {
		opt, err := getter(data)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &ActivityExecutionDataGetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetActivityExecutionDataProperties(entityID int, setters ...ActivityExecutionDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	var data *ActivityExecutionData
	for _, d := range db.activityExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}

	if data == nil {
		return ErrActivityExecutionNotFound
	}

	for _, setter := range setters {
		opt, err := setter(data)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &ActivityExecutionDataSetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) GetSagaExecutionDataProperties(entityID int, getters ...SagaExecutionDataPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var data *SagaExecutionData
	for _, d := range db.sagaExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}

	if data == nil {
		return ErrSagaExecutionNotFound
	}

	for _, getter := range getters {
		opt, err := getter(data)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &SagaExecutionDataGetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSagaExecutionDataProperties(entityID int, setters ...SagaExecutionDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	var data *SagaExecutionData
	for _, d := range db.sagaExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}

	if data == nil {
		return ErrSagaExecutionNotFound
	}

	for _, setter := range setters {
		opt, err := setter(data)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &SagaExecutionDataSetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) GetSideEffectExecutionDataProperties(entityID int, getters ...SideEffectExecutionDataPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var data *SideEffectExecutionData
	for _, d := range db.sideEffectExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}

	if data == nil {
		return ErrSideEffectExecutionNotFound
	}

	for _, getter := range getters {
		opt, err := getter(data)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &SideEffectExecutionDataGetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSideEffectExecutionDataProperties(entityID int, setters ...SideEffectExecutionDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	var data *SideEffectExecutionData
	for _, d := range db.sideEffectExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}

	if data == nil {
		return ErrSideEffectExecutionNotFound
	}

	for _, setter := range setters {
		opt, err := setter(data)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &SideEffectExecutionDataSetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

// Workflow Execution properties
func (db *MemoryDatabase) GetWorkflowExecutionProperties(id int, getters ...WorkflowExecutionPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	exec, exists := db.workflowExecutions[id]
	if !exists {
		return ErrWorkflowExecutionNotFound
	}

	for _, getter := range getters {
		opt, err := getter(exec)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &WorkflowExecutionGetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}

			if opts.IncludeData {
				for _, d := range db.workflowExecutionData {
					if d.ExecutionID == id {
						exec.WorkflowExecutionData = copyWorkflowExecutionData(d)
						break
					}
				}
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetWorkflowExecutionProperties(id int, setters ...WorkflowExecutionPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	exec, exists := db.workflowExecutions[id]
	if !exists {
		return ErrWorkflowExecutionNotFound
	}

	for _, setter := range setters {
		opt, err := setter(exec)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &WorkflowExecutionSetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

// Activity Execution properties
func (db *MemoryDatabase) GetActivityExecutionProperties(id int, getters ...ActivityExecutionPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	exec, exists := db.activityExecutions[id]
	if !exists {
		return ErrActivityExecutionNotFound
	}

	for _, getter := range getters {
		opt, err := getter(exec)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &ActivityExecutionGetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}

			if opts.IncludeData {
				for _, d := range db.activityExecutionData {
					if d.ExecutionID == id {
						exec.ActivityExecutionData = copyActivityExecutionData(d)
						break
					}
				}
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetActivityExecutionProperties(id int, setters ...ActivityExecutionPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	exec, exists := db.activityExecutions[id]
	if !exists {
		return ErrActivityExecutionNotFound
	}

	for _, setter := range setters {
		opt, err := setter(exec)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &ActivityExecutionSetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

// Saga Execution properties
func (db *MemoryDatabase) GetSagaExecutionProperties(id int, getters ...SagaExecutionPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	exec, exists := db.sagaExecutions[id]
	if !exists {
		return ErrSagaExecutionNotFound
	}

	for _, getter := range getters {
		opt, err := getter(exec)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &SagaExecutionGetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}

			if opts.IncludeData {
				for _, d := range db.sagaExecutionData {
					if d.ExecutionID == id {
						exec.SagaExecutionData = copySagaExecutionData(d)
						break
					}
				}
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSagaExecutionProperties(id int, setters ...SagaExecutionPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	exec, exists := db.sagaExecutions[id]
	if !exists {
		return ErrSagaExecutionNotFound
	}

	for _, setter := range setters {
		opt, err := setter(exec)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &SagaExecutionSetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

// SideEffect Execution properties
func (db *MemoryDatabase) GetSideEffectExecutionProperties(id int, getters ...SideEffectExecutionPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	exec, exists := db.sideEffectExecutions[id]
	if !exists {
		return ErrSideEffectExecutionNotFound
	}

	for _, getter := range getters {
		opt, err := getter(exec)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &SideEffectExecutionGetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}

			if opts.IncludeData {
				for _, d := range db.sideEffectExecutionData {
					if d.ExecutionID == id {
						exec.SideEffectExecutionData = copySideEffectExecutionData(d)
						break
					}
				}
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSideEffectExecutionProperties(id int, setters ...SideEffectExecutionPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	exec, exists := db.sideEffectExecutions[id]
	if !exists {
		return ErrSideEffectExecutionNotFound
	}

	for _, setter := range setters {
		opt, err := setter(exec)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &SideEffectExecutionSetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

// Hierarchy properties
func (db *MemoryDatabase) GetHierarchyProperties(id int, getters ...HierarchyPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	hierarchy, exists := db.hierarchies[id]
	if !exists {
		return ErrHierarchyNotFound
	}

	for _, getter := range getters {
		opt, err := getter(hierarchy)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &HierarchyGetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) GetHierarchyByParentEntity(parentEntityID int, childStepID string, specificType EntityType) (*Hierarchy, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, hierarchy := range db.hierarchies {
		if hierarchy.ParentEntityID == parentEntityID && hierarchy.ChildStepID == childStepID && hierarchy.ChildType == specificType {
			return copyHierarchy(hierarchy), nil
		}
	}

	return nil, ErrHierarchyNotFound
}

func (db *MemoryDatabase) SetHierarchyProperties(id int, setters ...HierarchyPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	hierarchy, exists := db.hierarchies[id]
	if !exists {
		return ErrHierarchyNotFound
	}

	for _, setter := range setters {
		opt, err := setter(hierarchy)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &HierarchySetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) GetHierarchiesByChildEntity(childEntityID int) ([]*Hierarchy, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	hierarchies := make([]*Hierarchy, 0)
	for _, hierarchy := range db.hierarchies {
		if hierarchy.ChildEntityID == childEntityID {
			hierarchies = append(hierarchies, copyHierarchy(hierarchy))
		}
	}

	return hierarchies, nil
}

func (db *MemoryDatabase) UpdateHierarchy(hierarchy *Hierarchy) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.hierarchies[hierarchy.ID]; !exists {
		return ErrHierarchyNotFound
	}

	db.hierarchies[hierarchy.ID] = copyHierarchy(hierarchy)
	return nil
}

// Queue properties
func (db *MemoryDatabase) GetQueueProperties(id int, getters ...QueuePropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	queue, exists := db.queues[id]
	if !exists {
		return ErrQueueNotFound
	}

	for _, getter := range getters {
		opt, err := getter(queue)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &QueueGetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}

			if opts.IncludeWorkflows {
				if workflowIDs, ok := db.queueToWorkflows[id]; ok {
					workflows := make([]*WorkflowEntity, 0, len(workflowIDs))
					for _, wfID := range workflowIDs {
						if wf, exists := db.workflowEntities[wfID]; exists {
							workflows = append(workflows, copyWorkflowEntity(wf))
						}
					}
					queue.Entities = workflows
				}
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetQueueProperties(id int, setters ...QueuePropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	queue, exists := db.queues[id]
	if !exists {
		return ErrQueueNotFound
	}

	for _, setter := range setters {
		opt, err := setter(queue)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &QueueSetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}

			if opts.WorkflowIDs != nil {
				// Remove existing relationships
				if oldWorkflowIDs, ok := db.queueToWorkflows[id]; ok {
					for _, oldWfID := range oldWorkflowIDs {
						delete(db.workflowToQueue, oldWfID)
					}
				}

				// Add new relationships
				db.queueToWorkflows[id] = opts.WorkflowIDs
				for _, wfID := range opts.WorkflowIDs {
					db.workflowToQueue[wfID] = id
				}
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) UpdateQueue(queue *Queue) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.queues[queue.ID]; !exists {
		return ErrQueueNotFound
	}

	// If name is changing, update the name index
	oldQueue := db.queues[queue.ID]
	if oldQueue.Name != queue.Name {
		delete(db.queueNames, oldQueue.Name)
		db.queueNames[queue.Name] = queue.ID
	}

	queue.UpdatedAt = time.Now()
	db.queues[queue.ID] = copyQueue(queue)
	return nil
}

// Version properties
func (db *MemoryDatabase) GetVersionProperties(id int, getters ...VersionPropertyGetter) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	version, exists := db.versions[id]
	if !exists {
		return ErrVersionNotFound
	}

	for _, getter := range getters {
		opt, err := getter(version)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &VersionGetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}

			// Handle IncludeData option if needed
			if opts.IncludeData {
				// Data is already part of the Version struct
				// No additional loading needed
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetVersionProperties(id int, setters ...VersionPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	version, exists := db.versions[id]
	if !exists {
		return ErrVersionNotFound
	}

	for _, setter := range setters {
		opt, err := setter(version)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &VersionSetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
		}
	}

	version.UpdatedAt = time.Now()
	return nil
}

func (db *MemoryDatabase) UpdateVersion(version *Version) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.versions[version.ID]; !exists {
		return ErrVersionNotFound
	}

	version.UpdatedAt = time.Now()
	db.versions[version.ID] = copyVersion(version)

	return nil
}

// Activity Entity
func (db *MemoryDatabase) UpdateActivityEntity(entity *ActivityEntity) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.activityEntities[entity.ID]; !exists {
		return ErrActivityEntityNotFound
	}

	entity.UpdatedAt = time.Now()
	db.activityEntities[entity.ID] = copyActivityEntity(entity)
	return nil
}

// Saga Entity
func (db *MemoryDatabase) UpdateSagaEntity(entity *SagaEntity) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.sagaEntities[entity.ID]; !exists {
		return ErrSagaEntityNotFound
	}

	entity.UpdatedAt = time.Now()
	db.sagaEntities[entity.ID] = copySagaEntity(entity)
	return nil
}

// SideEffect Entity
func (db *MemoryDatabase) UpdateSideEffectEntity(entity *SideEffectEntity) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.sideEffectEntities[entity.ID]; !exists {
		return ErrSideEffectEntityNotFound
	}

	entity.UpdatedAt = time.Now()
	db.sideEffectEntities[entity.ID] = copySideEffectEntity(entity)
	return nil
}

// Workflow Entity (if not already implemented)
func (db *MemoryDatabase) UpdateWorkflowEntity(entity *WorkflowEntity) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.workflowEntities[entity.ID]; !exists {
		return ErrWorkflowEntityNotFound
	}

	entity.UpdatedAt = time.Now()
	db.workflowEntities[entity.ID] = copyWorkflowEntity(entity)
	return nil
}

// Entity Data operations
func (db *MemoryDatabase) AddWorkflowData(entityID int, data *WorkflowData) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.workflowDataCounter++
	data.ID = db.workflowDataCounter
	data.EntityID = entityID
	db.workflowData[data.ID] = copyWorkflowData(data)
	return data.ID, nil
}

func (db *MemoryDatabase) AddActivityData(entityID int, data *ActivityData) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.activityDataCounter++
	data.ID = db.activityDataCounter
	data.EntityID = entityID
	db.activityData[data.ID] = copyActivityData(data)
	return data.ID, nil
}

func (db *MemoryDatabase) AddSagaData(entityID int, data *SagaData) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.sagaDataCounter++
	data.ID = db.sagaDataCounter
	data.EntityID = entityID
	db.sagaData[data.ID] = copySagaData(data)
	return data.ID, nil
}

func (db *MemoryDatabase) AddSideEffectData(entityID int, data *SideEffectData) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.sideEffectDataCounter++
	data.ID = db.sideEffectDataCounter
	data.EntityID = entityID
	db.sideEffectData[data.ID] = copySideEffectData(data)
	return data.ID, nil
}

func (db *MemoryDatabase) GetWorkflowData(id int) (*WorkflowData, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if data, exists := db.workflowData[id]; exists {
		return copyWorkflowData(data), nil
	}
	return nil, ErrWorkflowEntityNotFound
}

func (db *MemoryDatabase) GetActivityData(id int) (*ActivityData, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if data, exists := db.activityData[id]; exists {
		return copyActivityData(data), nil
	}
	return nil, ErrActivityEntityNotFound
}

func (db *MemoryDatabase) GetSagaData(id int) (*SagaData, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if data, exists := db.sagaData[id]; exists {
		return copySagaData(data), nil
	}
	return nil, ErrSagaEntityNotFound
}

func (db *MemoryDatabase) GetSideEffectData(id int) (*SideEffectData, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if data, exists := db.sideEffectData[id]; exists {
		return copySideEffectData(data), nil
	}
	return nil, ErrSideEffectEntityNotFound
}

func (db *MemoryDatabase) GetWorkflowDataByEntityID(entityID int) (*WorkflowData, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, data := range db.workflowData {
		if data.EntityID == entityID {
			return copyWorkflowData(data), nil
		}
	}
	return nil, ErrWorkflowEntityNotFound
}

func (db *MemoryDatabase) GetActivityDataByEntityID(entityID int) (*ActivityData, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, data := range db.activityData {
		if data.EntityID == entityID {
			return copyActivityData(data), nil
		}
	}
	return nil, ErrActivityEntityNotFound
}

func (db *MemoryDatabase) GetSagaDataByEntityID(entityID int) (*SagaData, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, data := range db.sagaData {
		if data.EntityID == entityID {
			return copySagaData(data), nil
		}
	}
	return nil, ErrSagaEntityNotFound
}

func (db *MemoryDatabase) GetSideEffectDataByEntityID(entityID int) (*SideEffectData, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, data := range db.sideEffectData {
		if data.EntityID == entityID {
			return copySideEffectData(data), nil
		}
	}
	return nil, ErrSideEffectEntityNotFound
}

func (db *MemoryDatabase) AddWorkflowExecutionData(executionID int, data *WorkflowExecutionData) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.workflowExecutionDataCounter++
	data.ID = db.workflowExecutionDataCounter
	data.ExecutionID = executionID
	db.workflowExecutionData[data.ID] = copyWorkflowExecutionData(data)
	return data.ID, nil
}

func (db *MemoryDatabase) AddActivityExecutionData(executionID int, data *ActivityExecutionData) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.activityExecutionDataCounter++
	data.ID = db.activityExecutionDataCounter
	data.ExecutionID = executionID
	db.activityExecutionData[data.ID] = copyActivityExecutionData(data)
	return data.ID, nil
}

func (db *MemoryDatabase) AddSagaExecutionData(executionID int, data *SagaExecutionData) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.sagaExecutionDataCounter++
	data.ID = db.sagaExecutionDataCounter
	data.ExecutionID = executionID
	db.sagaExecutionData[data.ID] = copySagaExecutionData(data)
	return data.ID, nil
}

func (db *MemoryDatabase) AddSideEffectExecutionData(executionID int, data *SideEffectExecutionData) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.sideEffectExecutionDataCounter++
	data.ID = db.sideEffectExecutionDataCounter
	data.ExecutionID = executionID
	db.sideEffectExecutionData[data.ID] = copySideEffectExecutionData(data)
	return data.ID, nil
}

func (db *MemoryDatabase) GetWorkflowExecutionData(id int) (*WorkflowExecutionData, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if data, exists := db.workflowExecutionData[id]; exists {
		return copyWorkflowExecutionData(data), nil
	}
	return nil, ErrWorkflowExecutionNotFound
}

func (db *MemoryDatabase) GetActivityExecutionData(id int) (*ActivityExecutionData, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if data, exists := db.activityExecutionData[id]; exists {
		return copyActivityExecutionData(data), nil
	}
	return nil, ErrActivityExecutionNotFound
}

func (db *MemoryDatabase) GetSagaExecutionData(id int) (*SagaExecutionData, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if data, exists := db.sagaExecutionData[id]; exists {
		return copySagaExecutionData(data), nil
	}
	return nil, ErrSagaExecutionNotFound
}

func (db *MemoryDatabase) GetSideEffectExecutionData(id int) (*SideEffectExecutionData, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if data, exists := db.sideEffectExecutionData[id]; exists {
		return copySideEffectExecutionData(data), nil
	}
	return nil, ErrSideEffectExecutionNotFound
}

func (db *MemoryDatabase) GetWorkflowExecutionDataByExecutionID(executionID int) (*WorkflowExecutionData, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, data := range db.workflowExecutionData {
		if data.ExecutionID == executionID {
			return copyWorkflowExecutionData(data), nil
		}
	}
	return nil, ErrWorkflowExecutionNotFound
}

func (db *MemoryDatabase) GetActivityExecutionDataByExecutionID(executionID int) (*ActivityExecutionData, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, data := range db.activityExecutionData {
		if data.ExecutionID == executionID {
			return copyActivityExecutionData(data), nil
		}
	}
	return nil, ErrActivityExecutionNotFound
}

func (db *MemoryDatabase) GetSagaExecutionDataByExecutionID(executionID int) (*SagaExecutionData, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, data := range db.sagaExecutionData {
		if data.ExecutionID == executionID {
			return copySagaExecutionData(data), nil
		}
	}
	return nil, ErrSagaExecutionNotFound
}

func (db *MemoryDatabase) GetSideEffectExecutionDataByExecutionID(executionID int) (*SideEffectExecutionData, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, data := range db.sideEffectExecutionData {
		if data.ExecutionID == executionID {
			return copySideEffectExecutionData(data), nil
		}
	}
	return nil, ErrSideEffectExecutionNotFound
}

// Has functions for Data operations
func (db *MemoryDatabase) HasRun(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.runs[id]
	return exists, nil
}

func (db *MemoryDatabase) HasVersion(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.versions[id]
	return exists, nil
}

func (db *MemoryDatabase) HasHierarchy(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.hierarchies[id]
	return exists, nil
}

func (db *MemoryDatabase) HasQueue(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.queues[id]
	return exists, nil
}

func (db *MemoryDatabase) HasQueueName(name string) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.queueNames[name]
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowEntity(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.workflowEntities[id]
	return exists, nil
}

func (db *MemoryDatabase) HasActivityEntity(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.activityEntities[id]
	return exists, nil
}

func (db *MemoryDatabase) HasSagaEntity(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.sagaEntities[id]
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectEntity(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.sideEffectEntities[id]
	return exists, nil
}

// Has functions for Execution operations
func (db *MemoryDatabase) HasWorkflowExecution(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.workflowExecutions[id]
	return exists, nil
}

func (db *MemoryDatabase) HasActivityExecution(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.activityExecutions[id]
	return exists, nil
}

func (db *MemoryDatabase) HasSagaExecution(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.sagaExecutions[id]
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectExecution(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.sideEffectExecutions[id]
	return exists, nil
}

// Has functions for Data operations
func (db *MemoryDatabase) HasWorkflowData(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.workflowData[id]
	return exists, nil
}

func (db *MemoryDatabase) HasActivityData(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.activityData[id]
	return exists, nil
}

func (db *MemoryDatabase) HasSagaData(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.sagaData[id]
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectData(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.sideEffectData[id]
	return exists, nil
}

// Has functions for Data by EntityID operations
func (db *MemoryDatabase) HasWorkflowDataByEntityID(entityID int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	for _, data := range db.workflowData {
		if data.EntityID == entityID {
			return true, nil
		}
	}
	return false, nil
}

func (db *MemoryDatabase) HasActivityDataByEntityID(entityID int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	for _, data := range db.activityData {
		if data.EntityID == entityID {
			return true, nil
		}
	}
	return false, nil
}

func (db *MemoryDatabase) HasSagaDataByEntityID(entityID int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	for _, data := range db.sagaData {
		if data.EntityID == entityID {
			return true, nil
		}
	}
	return false, nil
}

func (db *MemoryDatabase) HasSideEffectDataByEntityID(entityID int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	for _, data := range db.sideEffectData {
		if data.EntityID == entityID {
			return true, nil
		}
	}
	return false, nil
}

// Has functions for Execution Data operations
func (db *MemoryDatabase) HasWorkflowExecutionData(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.workflowExecutionData[id]
	return exists, nil
}

func (db *MemoryDatabase) HasActivityExecutionData(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.activityExecutionData[id]
	return exists, nil
}

func (db *MemoryDatabase) HasSagaExecutionData(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.sagaExecutionData[id]
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectExecutionData(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	_, exists := db.sideEffectExecutionData[id]
	return exists, nil
}

// Has functions for Execution Data by ExecutionID operations
func (db *MemoryDatabase) HasWorkflowExecutionDataByExecutionID(executionID int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	for _, data := range db.workflowExecutionData {
		if data.ExecutionID == executionID {
			return true, nil
		}
	}
	return false, nil
}

func (db *MemoryDatabase) HasActivityExecutionDataByExecutionID(executionID int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	for _, data := range db.activityExecutionData {
		if data.ExecutionID == executionID {
			return true, nil
		}
	}
	return false, nil
}

func (db *MemoryDatabase) HasSagaExecutionDataByExecutionID(executionID int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	for _, data := range db.sagaExecutionData {
		if data.ExecutionID == executionID {
			return true, nil
		}
	}
	return false, nil
}

func (db *MemoryDatabase) HasSideEffectExecutionDataByExecutionID(executionID int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	for _, data := range db.sideEffectExecutionData {
		if data.ExecutionID == executionID {
			return true, nil
		}
	}
	return false, nil
}

func (db *MemoryDatabase) GetWorkflowExecutionLatestByEntityID(entityID int) (*WorkflowExecution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var latestExec *WorkflowExecution
	var latestTime time.Time

	for _, exec := range db.workflowExecutions {
		if exec.EntityID == entityID {
			if latestExec == nil || exec.CreatedAt.After(latestTime) {
				latestExec = exec
				latestTime = exec.CreatedAt
			}
		}
	}

	if latestExec == nil {
		return nil, ErrWorkflowExecutionNotFound
	}

	return copyWorkflowExecution(latestExec), nil
}

func (db *MemoryDatabase) GetActivityExecutionLatestByEntityID(entityID int) (*ActivityExecution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var latestExec *ActivityExecution
	var latestTime time.Time

	for _, exec := range db.activityExecutions {
		if exec.EntityID == entityID {
			if latestExec == nil || exec.CreatedAt.After(latestTime) {
				latestExec = exec
				latestTime = exec.CreatedAt
			}
		}
	}

	if latestExec == nil {
		return nil, ErrActivityExecutionNotFound
	}

	return copyActivityExecution(latestExec), nil
}
