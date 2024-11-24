package tempolite

import (
	"sync"
	"time"
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

func (db *MemoryDatabase) GetRun(id int) (*Run, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	run, exists := db.runs[id]
	if !exists {
		return nil, ErrRunNotFound
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

func (db *MemoryDatabase) GetQueue(id int) (*Queue, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	queue, exists := db.queues[id]
	if !exists {
		return nil, ErrQueueNotFound
	}

	// Load associated workflows
	if workflowIDs, ok := db.queueToWorkflows[id]; ok {
		workflows := make([]*WorkflowEntity, 0, len(workflowIDs))
		for _, wfID := range workflowIDs {
			if wf, exists := db.workflowEntities[wfID]; exists {
				workflows = append(workflows, copyWorkflowEntity(wf))
			}
		}
		queue.Entities = workflows
	}

	return copyQueue(queue), nil
}

func (db *MemoryDatabase) GetQueueByName(name string) (*Queue, error) {
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

	// Load associated workflows
	if workflowIDs, ok := db.queueToWorkflows[id]; ok {
		workflows := make([]*WorkflowEntity, 0, len(workflowIDs))
		for _, wfID := range workflowIDs {
			if wf, exists := db.workflowEntities[wfID]; exists {
				workflows = append(workflows, copyWorkflowEntity(wf))
			}
		}
		queue.Entities = workflows
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

func (db *MemoryDatabase) GetVersion(id int) (*Version, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	version, exists := db.versions[id]
	if !exists {
		return nil, ErrVersionNotFound
	}
	return copyVersion(version), nil
}

func (db *MemoryDatabase) AddHierarchy(hierarchy *Hierarchy) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.hierarchyCounter++
	hierarchy.ID = db.hierarchyCounter

	db.hierarchies[hierarchy.ID] = copyHierarchy(hierarchy)
	return hierarchy.ID, nil
}

func (db *MemoryDatabase) GetHierarchy(id int) (*Hierarchy, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	hierarchy, exists := db.hierarchies[id]
	if !exists {
		return nil, ErrHierarchyNotFound
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
	}

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
