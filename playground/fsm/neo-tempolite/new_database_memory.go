package tempolite

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/sasha-s/go-deadlock"
)

type MemoryDatabase struct {
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

	// Signal counters
	signalEntityCounter        int
	signalExecutionCounter     int
	signalDataCounter          int
	signalExecutionDataCounter int

	// Saga specific
	sagaValues map[SagaExecutionID]map[string]*SagaValue

	// Core maps
	runs        map[RunID]*Run
	versions    map[VersionID]*Version
	hierarchies map[HierarchyID]*Hierarchy
	queues      map[QueueID]*Queue
	queueNames  map[string]QueueID

	// Entity maps
	workflowEntities   map[WorkflowEntityID]*WorkflowEntity
	activityEntities   map[ActivityEntityID]*ActivityEntity
	sagaEntities       map[SagaEntityID]*SagaEntity
	sideEffectEntities map[SideEffectEntityID]*SideEffectEntity

	// Entity Data maps
	workflowData   map[WorkflowDataID]*WorkflowData
	activityData   map[ActivityDataID]*ActivityData
	sagaData       map[SagaDataID]*SagaData
	sideEffectData map[SideEffectDataID]*SideEffectData

	// Execution maps
	workflowExecutions   map[WorkflowExecutionID]*WorkflowExecution
	activityExecutions   map[ActivityExecutionID]*ActivityExecution
	sagaExecutions       map[SagaExecutionID]*SagaExecution
	sideEffectExecutions map[SideEffectExecutionID]*SideEffectExecution

	// Execution Data maps
	workflowExecutionData   map[WorkflowExecutionDataID]*WorkflowExecutionData
	activityExecutionData   map[ActivityExecutionDataID]*ActivityExecutionData
	sagaExecutionData       map[SagaExecutionDataID]*SagaExecutionData
	sideEffectExecutionData map[SideEffectExecutionDataID]*SideEffectExecutionData

	// Signal maps
	signalEntities      map[SignalEntityID]*SignalEntity
	signalExecutions    map[SignalExecutionID]*SignalExecution
	signalData          map[SignalDataID]*SignalData
	signalExecutionData map[SignalExecutionDataID]*SignalExecutionData

	// Relationship maps
	entityToWorkflow   map[int]WorkflowEntityID
	workflowToChildren map[WorkflowEntityID]map[EntityType][]int
	workflowToVersion  map[WorkflowEntityID][]VersionID
	workflowToQueue    map[WorkflowEntityID]QueueID
	queueToWorkflows   map[QueueID][]WorkflowEntityID
	runToWorkflows     map[RunID][]WorkflowEntityID
	workflowVersions   map[WorkflowEntityID][]VersionID

	workflowExecToDataMap   map[WorkflowExecutionID]WorkflowExecutionDataID
	activityExecToDataMap   map[ActivityExecutionID]ActivityExecutionDataID
	sagaExecToDataMap       map[SagaExecutionID]SagaExecutionDataID
	sideEffectExecToDataMap map[SideEffectExecutionID]SideEffectExecutionDataID
	signalExecToDataMap     map[SignalExecutionID]SignalExecutionDataID

	// Locks for each category of data
	mu deadlock.RWMutex
}

// NewMemoryDatabase initializes a new memory database with default queue.
func NewMemoryDatabase() *MemoryDatabase {
	db := &MemoryDatabase{
		// Core maps
		runs:        make(map[RunID]*Run),
		versions:    make(map[VersionID]*Version),
		hierarchies: make(map[HierarchyID]*Hierarchy),
		queues:      make(map[QueueID]*Queue),
		queueNames:  make(map[string]QueueID),

		// Entity maps
		workflowEntities:   make(map[WorkflowEntityID]*WorkflowEntity),
		activityEntities:   make(map[ActivityEntityID]*ActivityEntity),
		sagaEntities:       make(map[SagaEntityID]*SagaEntity),
		sideEffectEntities: make(map[SideEffectEntityID]*SideEffectEntity),

		// Entity Data maps
		workflowData:   make(map[WorkflowDataID]*WorkflowData),
		activityData:   make(map[ActivityDataID]*ActivityData),
		sagaData:       make(map[SagaDataID]*SagaData),
		sideEffectData: make(map[SideEffectDataID]*SideEffectData),

		// Execution maps
		workflowExecutions:   make(map[WorkflowExecutionID]*WorkflowExecution),
		activityExecutions:   make(map[ActivityExecutionID]*ActivityExecution),
		sagaExecutions:       make(map[SagaExecutionID]*SagaExecution),
		sideEffectExecutions: make(map[SideEffectExecutionID]*SideEffectExecution),

		// Execution Data maps
		workflowExecutionData:   make(map[WorkflowExecutionDataID]*WorkflowExecutionData),
		activityExecutionData:   make(map[ActivityExecutionDataID]*ActivityExecutionData),
		sagaExecutionData:       make(map[SagaExecutionDataID]*SagaExecutionData),
		sideEffectExecutionData: make(map[SideEffectExecutionDataID]*SideEffectExecutionData),

		// Signal maps
		signalEntities:      make(map[SignalEntityID]*SignalEntity),
		signalExecutions:    make(map[SignalExecutionID]*SignalExecution),
		signalData:          make(map[SignalDataID]*SignalData),
		signalExecutionData: make(map[SignalExecutionDataID]*SignalExecutionData),

		// Saga context
		sagaValues: make(map[SagaExecutionID]map[string]*SagaValue),

		// Relationship maps
		entityToWorkflow:   make(map[int]WorkflowEntityID),
		workflowToChildren: make(map[WorkflowEntityID]map[EntityType][]int),
		workflowToVersion:  make(map[WorkflowEntityID][]VersionID),
		workflowToQueue:    make(map[WorkflowEntityID]QueueID),
		queueToWorkflows:   make(map[QueueID][]WorkflowEntityID),
		runToWorkflows:     make(map[RunID][]WorkflowEntityID),
		workflowVersions:   make(map[WorkflowEntityID][]VersionID),

		workflowExecToDataMap:   make(map[WorkflowExecutionID]WorkflowExecutionDataID),
		activityExecToDataMap:   make(map[ActivityExecutionID]ActivityExecutionDataID),
		sagaExecToDataMap:       make(map[SagaExecutionID]SagaExecutionDataID),
		sideEffectExecToDataMap: make(map[SideEffectExecutionID]SideEffectExecutionDataID),
		signalExecToDataMap:     make(map[SignalExecutionID]SignalExecutionDataID),

		queueCounter: 1,
	}

	// Initialize default queue
	now := time.Now()
	db.mu.Lock()
	db.queues[1] = &Queue{
		ID:        1,
		Name:      DefaultQueue,
		CreatedAt: now,
		UpdatedAt: now,
		Entities:  make([]*WorkflowEntity, 0),
	}
	db.queueNames[DefaultQueue] = 1
	db.mu.Unlock()

	return db
}

func (db *MemoryDatabase) SaveAsJSON(path string) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	// We need to lock all relevant structures for a consistent snapshot.
	// We'll lock everything in alphabetical order to avoid deadlocks.

	data := struct {
		Runs                    map[RunID]*Run
		Versions                map[VersionID]*Version
		Hierarchies             map[HierarchyID]*Hierarchy
		Queues                  map[QueueID]*Queue
		QueueNames              map[string]QueueID
		WorkflowEntities        map[WorkflowEntityID]*WorkflowEntity
		ActivityEntities        map[ActivityEntityID]*ActivityEntity
		SagaEntities            map[SagaEntityID]*SagaEntity
		SideEffectEntities      map[SideEffectEntityID]*SideEffectEntity
		WorkflowData            map[WorkflowDataID]*WorkflowData
		ActivityData            map[ActivityDataID]*ActivityData
		SagaData                map[SagaDataID]*SagaData
		SideEffectData          map[SideEffectDataID]*SideEffectData
		WorkflowExecutions      map[WorkflowExecutionID]*WorkflowExecution
		ActivityExecutions      map[ActivityExecutionID]*ActivityExecution
		SagaExecutions          map[SagaExecutionID]*SagaExecution
		SideEffectExecutions    map[SideEffectExecutionID]*SideEffectExecution
		WorkflowExecutionData   map[WorkflowExecutionDataID]*WorkflowExecutionData
		ActivityExecutionData   map[ActivityExecutionDataID]*ActivityExecutionData
		SagaExecutionData       map[SagaExecutionDataID]*SagaExecutionData
		SideEffectExecutionData map[SideEffectExecutionDataID]*SideEffectExecutionData
		EntityToWorkflow        map[int]WorkflowEntityID
		WorkflowToChildren      map[WorkflowEntityID]map[EntityType][]int
		WorkflowToVersion       map[WorkflowEntityID][]VersionID
		WorkflowToQueue         map[WorkflowEntityID]QueueID
		QueueToWorkflows        map[QueueID][]WorkflowEntityID
		RunToWorkflows          map[RunID][]WorkflowEntityID
		SagaValues              map[SagaExecutionID]map[string]*SagaValue
		SignalEntities          map[SignalEntityID]*SignalEntity
		SignalExecutions        map[SignalExecutionID]*SignalExecution
		SignalData              map[SignalDataID]*SignalData
		SignalExecutionData     map[SignalExecutionDataID]*SignalExecutionData
		// Relationship
		WorkflowVersions        map[WorkflowEntityID][]VersionID
		WorkflowExecToDataMap   map[WorkflowExecutionID]WorkflowExecutionDataID
		ActivityExecToDataMap   map[ActivityExecutionID]ActivityExecutionDataID
		SagaExecToDataMap       map[SagaExecutionID]SagaExecutionDataID
		SideEffectExecToDataMap map[SideEffectExecutionID]SideEffectExecutionDataID
		SignalExecToDataMap     map[SignalExecutionID]SignalExecutionDataID
	}{
		Runs:                    db.runs,
		Versions:                db.versions,
		Hierarchies:             db.hierarchies,
		Queues:                  db.queues,
		QueueNames:              db.queueNames,
		WorkflowEntities:        db.workflowEntities,
		ActivityEntities:        db.activityEntities,
		SagaEntities:            db.sagaEntities,
		SideEffectEntities:      db.sideEffectEntities,
		WorkflowData:            db.workflowData,
		ActivityData:            db.activityData,
		SagaData:                db.sagaData,
		SideEffectData:          db.sideEffectData,
		WorkflowExecutions:      db.workflowExecutions,
		ActivityExecutions:      db.activityExecutions,
		SagaExecutions:          db.sagaExecutions,
		SideEffectExecutions:    db.sideEffectExecutions,
		WorkflowExecutionData:   db.workflowExecutionData,
		ActivityExecutionData:   db.activityExecutionData,
		SagaExecutionData:       db.sagaExecutionData,
		SideEffectExecutionData: db.sideEffectExecutionData,
		EntityToWorkflow:        db.entityToWorkflow,
		WorkflowToChildren:      db.workflowToChildren,
		WorkflowToVersion:       db.workflowToVersion,
		WorkflowToQueue:         db.workflowToQueue,
		QueueToWorkflows:        db.queueToWorkflows,
		RunToWorkflows:          db.runToWorkflows,
		SagaValues:              db.sagaValues,
		SignalEntities:          db.signalEntities,
		SignalExecutions:        db.signalExecutions,
		SignalData:              db.signalData,
		SignalExecutionData:     db.signalExecutionData,
		// Relationship
		WorkflowVersions:        db.workflowVersions,
		WorkflowExecToDataMap:   db.workflowExecToDataMap,
		ActivityExecToDataMap:   db.activityExecToDataMap,
		SagaExecToDataMap:       db.sagaExecToDataMap,
		SideEffectExecToDataMap: db.sideEffectExecToDataMap,
		SignalExecToDataMap:     db.signalExecToDataMap,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Ensure folder exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return os.WriteFile(path, jsonData, 0644)
}

// From here on, we replace all uses of db.mu with fine-grained locks targeting the structures involved.

// AddRun
func (db *MemoryDatabase) AddRun(run *Run) (RunID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()

	db.runCounter++
	run.ID = RunID(db.runCounter)
	run.CreatedAt = time.Now()
	run.UpdatedAt = run.CreatedAt

	db.runs[run.ID] = copyRun(run)
	return run.ID, nil
}

// GetRun
func (db *MemoryDatabase) GetRun(id RunID, opts ...RunGetOption) (*Run, error) {

	db.mu.Lock()
	defer db.mu.Unlock()

	r, exists := db.runs[id]
	if !exists {

		return nil, ErrRunNotFound
	}
	runCopy := copyRun(r) // Copy to avoid race after unlocking

	cfg := &RunGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeWorkflows {
		workflowIDs := db.runToWorkflows[id]

		if len(workflowIDs) > 0 {
			entities := make([]*WorkflowEntity, 0, len(workflowIDs))
			for _, wfID := range workflowIDs {
				if wf, ok := db.workflowEntities[wfID]; ok {
					entities = append(entities, copyWorkflowEntity(wf))
				}
			}
			runCopy.Entities = entities
		}
	}

	if cfg.IncludeHierarchies {
		hierarchies := make([]*Hierarchy, 0)
		for _, h := range db.hierarchies {
			if h.RunID == id {
				hierarchies = append(hierarchies, copyHierarchy(h))
			}
		}
		runCopy.Hierarchies = hierarchies
	}

	return runCopy, nil
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

func (db *MemoryDatabase) GetRunProperties(id RunID, getters ...RunPropertyGetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	run, exists := db.runs[id]
	if !exists {
		return ErrRunNotFound
	}
	runCopy := copyRun(run)

	opts := &RunGetterOptions{}
	for _, getter := range getters {
		opt, err := getter(runCopy)
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
		workflowIDs := db.runToWorkflows[id]

		if len(workflowIDs) > 0 {
			entities := make([]*WorkflowEntity, 0, len(workflowIDs))
			for _, wfID := range workflowIDs {
				if wf, ok := db.workflowEntities[wfID]; ok {
					entities = append(entities, copyWorkflowEntity(wf))
				}
			}
			runCopy.Entities = entities
		}
	}

	if opts.IncludeHierarchies {
		hierarchies := make([]*Hierarchy, 0)
		for _, h := range db.hierarchies {
			if h.RunID == id {
				hierarchies = append(hierarchies, copyHierarchy(h))
			}
		}
		runCopy.Hierarchies = hierarchies
	}

	return nil
}

func (db *MemoryDatabase) SetRunProperties(id RunID, setters ...RunPropertySetter) error {

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
		_, wfExists := db.workflowEntities[*opts.WorkflowID]
		if !wfExists {
			return ErrWorkflowEntityNotFound
		}

		db.runToWorkflows[id] = append(db.runToWorkflows[id], *opts.WorkflowID)
	}

	run.UpdatedAt = time.Now()
	db.runs[id] = copyRun(run)

	return nil
}

func (db *MemoryDatabase) AddQueue(queue *Queue) (QueueID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.queueNames[queue.Name]; exists {
		return QueueID(0), ErrQueueExists
	}

	db.queueCounter++
	queue.ID = QueueID(db.queueCounter)
	queue.CreatedAt = time.Now()
	queue.UpdatedAt = queue.CreatedAt
	queue.Entities = make([]*WorkflowEntity, 0)

	db.queues[queue.ID] = copyQueue(queue)
	db.queueNames[queue.Name] = queue.ID
	return queue.ID, nil
}

func (db *MemoryDatabase) GetQueue(id QueueID, opts ...QueueGetOption) (*Queue, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	q, exists := db.queues[id]
	if !exists {
		return nil, ErrQueueNotFound
	}
	queueCopy := copyQueue(q)

	cfg := &QueueGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeWorkflows {
		workflowIDs := db.queueToWorkflows[id]

		if len(workflowIDs) > 0 {
			workflows := make([]*WorkflowEntity, 0, len(workflowIDs))
			for _, wfID := range workflowIDs {
				if wf, exists := db.workflowEntities[wfID]; exists {
					workflows = append(workflows, copyWorkflowEntity(wf))
				}
			}
			queueCopy.Entities = workflows
		}
	}

	return queueCopy, nil
}

func (db *MemoryDatabase) GetQueueByName(name string, opts ...QueueGetOption) (*Queue, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	id, exists := db.queueNames[name]
	if !exists {
		return nil, ErrQueueNotFound
	}
	q, qexists := db.queues[id]
	if !qexists {
		return nil, ErrQueueNotFound
	}
	queueCopy := copyQueue(q)

	cfg := &QueueGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeWorkflows {
		workflowIDs := db.queueToWorkflows[id]

		if len(workflowIDs) > 0 {
			workflows := make([]*WorkflowEntity, 0, len(workflowIDs))
			for _, wfID := range workflowIDs {
				if wf, exists := db.workflowEntities[wfID]; exists {
					workflows = append(workflows, copyWorkflowEntity(wf))
				}
			}
			queueCopy.Entities = workflows
		}
	}

	return queueCopy, nil
}

func (db *MemoryDatabase) AddVersion(version *Version) (VersionID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()

	db.versionCounter++
	version.ID = VersionID(db.versionCounter)
	version.CreatedAt = time.Now()
	version.UpdatedAt = version.CreatedAt

	db.versions[version.ID] = copyVersion(version)

	if version.EntityID != 0 {
		db.workflowToVersion[version.EntityID] = append(db.workflowToVersion[version.EntityID], version.ID)
	}
	return version.ID, nil
}

func (db *MemoryDatabase) GetVersion(id VersionID, opts ...VersionGetOption) (*Version, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	v, exists := db.versions[id]
	if !exists {
		return nil, ErrVersionNotFound
	}
	versionCopy := copyVersion(v)

	cfg := &VersionGetterOptions{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	return versionCopy, nil
}

func (db *MemoryDatabase) AddHierarchy(hierarchy *Hierarchy) (HierarchyID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()

	db.hierarchyCounter++
	hierarchy.ID = HierarchyID(db.hierarchyCounter)

	db.hierarchies[hierarchy.ID] = copyHierarchy(hierarchy)
	return hierarchy.ID, nil
}

func (db *MemoryDatabase) GetHierarchy(id HierarchyID, opts ...HierarchyGetOption) (*Hierarchy, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	h, exists := db.hierarchies[id]
	if !exists {
		return nil, ErrHierarchyNotFound
	}
	hCopy := copyHierarchy(h)

	cfg := &HierarchyGetterOptions{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	return hCopy, nil
}

func (db *MemoryDatabase) AddWorkflowEntity(entity *WorkflowEntity) (WorkflowEntityID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()

	db.workflowEntityCounter++
	entity.ID = WorkflowEntityID(db.workflowEntityCounter)

	if entity.WorkflowData != nil {
		db.workflowDataCounter++
		entity.WorkflowData.ID = WorkflowDataID(db.workflowDataCounter)
		entity.WorkflowData.EntityID = entity.ID
		db.workflowData[entity.WorkflowData.ID] = copyWorkflowData(entity.WorkflowData)
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = entity.CreatedAt

	// Default queue if none
	defaultQueueID := QueueID(1)
	if entity.QueueID == 0 {
		entity.QueueID = defaultQueueID
		db.queueToWorkflows[defaultQueueID] = append(db.queueToWorkflows[defaultQueueID], entity.ID)
	} else {
		db.workflowToQueue[entity.ID] = entity.QueueID
		db.queueToWorkflows[entity.QueueID] = append(db.queueToWorkflows[entity.QueueID], entity.ID)
	}

	db.workflowEntities[entity.ID] = copyWorkflowEntity(entity)

	return entity.ID, nil
}

func (db *MemoryDatabase) AddWorkflowExecution(entityID WorkflowEntityID, exec *WorkflowExecution) (WorkflowExecutionID, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Verify the entity exists
	if _, exists := db.workflowEntities[entityID]; !exists {
		return WorkflowExecutionID(0), ErrWorkflowEntityNotFound
	}

	db.workflowExecutionCounter++
	exec.ID = WorkflowExecutionID(db.workflowExecutionCounter)
	exec.WorkflowEntityID = entityID // Set the entity ID

	if exec.WorkflowExecutionData != nil {
		db.workflowExecutionDataCounter++
		exec.WorkflowExecutionData.ID = WorkflowExecutionDataID(db.workflowExecutionDataCounter)
		exec.WorkflowExecutionData.ExecutionID = exec.ID
		db.workflowExecutionData[exec.WorkflowExecutionData.ID] = copyWorkflowExecutionData(exec.WorkflowExecutionData)

		db.workflowExecToDataMap[exec.ID] = exec.WorkflowExecutionData.ID
	}

	exec.CreatedAt = time.Now()
	exec.UpdatedAt = exec.CreatedAt

	db.workflowExecutions[exec.ID] = copyWorkflowExecution(exec)
	return exec.ID, nil
}

func (db *MemoryDatabase) GetWorkflowEntity(id WorkflowEntityID, opts ...WorkflowEntityGetOption) (*WorkflowEntity, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	e, exists := db.workflowEntities[id]
	if !exists {
		return nil, ErrWorkflowEntityNotFound
	}
	entityCopy := copyWorkflowEntity(e)

	cfg := &WorkflowEntityGetterOptions{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeQueue {
		queueID, hasQueue := db.workflowToQueue[id]
		if hasQueue {
			if q, qexists := db.queues[queueID]; qexists {
				if entityCopy.Edges == nil {
					entityCopy.Edges = &WorkflowEntityEdges{}
				}
				entityCopy.Edges.Queue = copyQueue(q)
			}
		}
	}

	if cfg.IncludeData {
		for _, d := range db.workflowData {
			if d.EntityID == id {
				entityCopy.WorkflowData = copyWorkflowData(d)
				break
			}
		}
	}

	return entityCopy, nil
}

func (db *MemoryDatabase) GetWorkflowEntityProperties(id WorkflowEntityID, getters ...WorkflowEntityPropertyGetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	entity, exists := db.workflowEntities[id]
	if !exists {
		return ErrWorkflowEntityNotFound
	}
	entityCopy := copyWorkflowEntity(entity)

	opts := &WorkflowEntityGetterOptions{}

	for _, getter := range getters {
		opt, err := getter(entityCopy)
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
		versionIDs := db.workflowToVersion[id]

		if len(versionIDs) > 0 {
			versions := make([]*Version, 0, len(versionIDs))
			for _, vID := range versionIDs {
				if v, vexists := db.versions[vID]; vexists {
					versions = append(versions, copyVersion(v))
				}
			}

			if entityCopy.Edges == nil {
				entityCopy.Edges = &WorkflowEntityEdges{}
			}
			entityCopy.Edges.Versions = versions
		}
	}

	if opts.IncludeQueue {
		queueID, qexists := db.workflowToQueue[id]
		if qexists {
			if q, qfound := db.queues[queueID]; qfound {
				if entityCopy.Edges == nil {
					entityCopy.Edges = &WorkflowEntityEdges{}
				}
				entityCopy.Edges.Queue = copyQueue(q)
			}
		}
	}

	if opts.IncludeChildren {
		childMap, cexists := db.workflowToChildren[id]
		if cexists {
			if entityCopy.Edges == nil {
				entityCopy.Edges = &WorkflowEntityEdges{}
			}

			if activityIDs, ok := childMap[EntityActivity]; ok {
				activities := make([]*ActivityEntity, 0, len(activityIDs))
				for _, aID := range activityIDs {
					if a, aexists := db.activityEntities[ActivityEntityID(aID)]; aexists {
						activities = append(activities, copyActivityEntity(a))
					}
				}
				entityCopy.Edges.ActivityChildren = activities
			}

			if sagaIDs, ok := childMap[EntitySaga]; ok {
				sagas := make([]*SagaEntity, 0, len(sagaIDs))
				for _, sID := range sagaIDs {
					if s, sexists := db.sagaEntities[SagaEntityID(sID)]; sexists {
						sagas = append(sagas, copySagaEntity(s))
					}
				}
				entityCopy.Edges.SagaChildren = sagas
			}

			if sideEffectIDs, ok := childMap[EntitySideEffect]; ok {
				sideEffects := make([]*SideEffectEntity, 0, len(sideEffectIDs))
				for _, seID := range sideEffectIDs {
					if se, seexists := db.sideEffectEntities[SideEffectEntityID(seID)]; seexists {
						sideEffects = append(sideEffects, copySideEffectEntity(se))
					}
				}
				entityCopy.Edges.SideEffectChildren = sideEffects
			}
		}
	}

	if opts.IncludeData {
		for _, d := range db.workflowData {
			if d.EntityID == id {
				entityCopy.WorkflowData = copyWorkflowData(d)
				break
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetWorkflowEntityProperties(id WorkflowEntityID, setters ...WorkflowEntityPropertySetter) error {

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
		_, qexists := db.queues[*opts.QueueID]
		if !qexists {
			return ErrQueueNotFound
		}

		// Remove from old queue
		if oldQueueID, ok := db.workflowToQueue[id]; ok {
			if workflows, exists := db.queueToWorkflows[oldQueueID]; exists {
				newWorkflows := make([]WorkflowEntityID, 0)
				for _, wID := range workflows {
					if wID != id {
						newWorkflows = append(newWorkflows, wID)
					}
				}
				db.queueToWorkflows[oldQueueID] = newWorkflows
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

	if opts.RunID != nil {
		_, exists := db.runs[*opts.RunID]
		if !exists {
			return ErrRunNotFound
		}

		db.runToWorkflows[*opts.RunID] = append(db.runToWorkflows[*opts.RunID], id)
	}

	entity.UpdatedAt = time.Now()
	db.workflowEntities[id] = copyWorkflowEntity(entity)

	return nil
}

func (db *MemoryDatabase) AddActivityEntity(entity *ActivityEntity, parentWorkflowID WorkflowEntityID) (ActivityEntityID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()

	db.activityEntityCounter++
	entity.ID = ActivityEntityID(db.activityEntityCounter)

	if entity.ActivityData != nil {
		db.activityDataCounter++
		entity.ActivityData.ID = ActivityDataID(db.activityDataCounter)
		entity.ActivityData.EntityID = entity.ID
		db.activityData[entity.ActivityData.ID] = copyActivityData(entity.ActivityData)
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = entity.CreatedAt

	db.activityEntities[entity.ID] = copyActivityEntity(entity)

	if _, ok := db.workflowToChildren[parentWorkflowID]; !ok {
		db.workflowToChildren[parentWorkflowID] = make(map[EntityType][]int)
	}
	db.workflowToChildren[parentWorkflowID][EntityActivity] = append(
		db.workflowToChildren[parentWorkflowID][EntityActivity],
		int(entity.ID),
	)

	return entity.ID, nil
}

func (db *MemoryDatabase) AddActivityExecution(exec *ActivityExecution) (ActivityExecutionID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	db.activityExecutionCounter++
	exec.ID = ActivityExecutionID(db.activityExecutionCounter)

	if exec.ActivityExecutionData != nil {
		db.activityExecutionDataCounter++
		exec.ActivityExecutionData.ID = ActivityExecutionDataID(db.activityExecutionDataCounter)
		exec.ActivityExecutionData.ExecutionID = exec.ID

		db.activityExecutionData[exec.ActivityExecutionData.ID] = copyActivityExecutionData(exec.ActivityExecutionData)

		db.activityExecToDataMap[exec.ID] = exec.ActivityExecutionData.ID
	}

	exec.CreatedAt = time.Now()
	exec.UpdatedAt = exec.CreatedAt

	db.activityExecutions[exec.ID] = copyActivityExecution(exec)
	return exec.ID, nil
}

func (db *MemoryDatabase) GetActivityEntity(id ActivityEntityID, opts ...ActivityEntityGetOption) (*ActivityEntity, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	entity, exists := db.activityEntities[id]
	if !exists {
		return nil, ErrActivityEntityNotFound
	}
	eCopy := copyActivityEntity(entity)

	cfg := &ActivityEntityGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		for _, d := range db.activityData {
			if d.EntityID == id {
				eCopy.ActivityData = copyActivityData(d)
				break
			}
		}
	}

	return eCopy, nil
}

func (db *MemoryDatabase) GetActivityEntities(workflowID WorkflowEntityID, opts ...ActivityEntityGetOption) ([]*ActivityEntity, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	activityIDs := db.workflowToChildren[workflowID][EntityActivity]

	entities := make([]*ActivityEntity, 0, len(activityIDs))
	for _, aID := range activityIDs {
		// Instead of calling GetActivityEntity which would try to acquire the lock again,
		// access the data directly since we already have the lock
		entity, exists := db.activityEntities[ActivityEntityID(aID)]
		if !exists {
			continue
		}

		// Create a copy of the entity
		entityCopy := copyActivityEntity(entity)

		// Apply options to the copied entity
		cfg := &ActivityEntityGetterOptions{}
		for _, opt := range opts {
			if err := opt(cfg); err != nil {
				return nil, err
			}
		}

		if cfg.IncludeData {
			for _, d := range db.activityData {
				if d.EntityID == ActivityEntityID(aID) {
					entityCopy.ActivityData = copyActivityData(d)
					break
				}
			}
		}

		entities = append(entities, entityCopy)
	}

	return entities, nil
}

func (db *MemoryDatabase) GetActivityEntityProperties(id ActivityEntityID, getters ...ActivityEntityPropertyGetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	entity, exists := db.activityEntities[id]
	if !exists {
		return ErrActivityEntityNotFound
	}
	eCopy := copyActivityEntity(entity)

	opts := &ActivityEntityGetterOptions{}
	for _, getter := range getters {
		opt, err := getter(eCopy)
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
				eCopy.ActivityData = copyActivityData(d)
				break
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetActivityEntityProperties(id ActivityEntityID, setters ...ActivityEntityPropertySetter) error {

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
		_, wExists := db.workflowEntities[*opts.ParentWorkflowID]
		if !wExists {
			return ErrWorkflowEntityNotFound
		}

		db.entityToWorkflow[int(id)] = *opts.ParentWorkflowID
		if db.workflowToChildren[*opts.ParentWorkflowID] == nil {
			db.workflowToChildren[*opts.ParentWorkflowID] = make(map[EntityType][]int)
		}
		db.workflowToChildren[*opts.ParentWorkflowID][EntityActivity] = append(
			db.workflowToChildren[*opts.ParentWorkflowID][EntityActivity],
			int(id),
		)
	}

	if opts.ParentRunID != nil {
		_, rExists := db.runs[*opts.ParentRunID]
		if !rExists {
			return ErrRunNotFound
		}
	}

	entity.UpdatedAt = time.Now()
	db.activityEntities[id] = copyActivityEntity(entity)

	return nil
}

func (db *MemoryDatabase) AddSagaEntity(entity *SagaEntity, parentWorkflowID WorkflowEntityID) (SagaEntityID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	db.sagaEntityCounter++
	entity.ID = SagaEntityID(db.sagaEntityCounter)

	if entity.SagaData != nil {
		db.sagaDataCounter++
		entity.SagaData.ID = SagaDataID(db.sagaDataCounter)
		entity.SagaData.EntityID = entity.ID
		db.sagaData[entity.SagaData.ID] = copySagaData(entity.SagaData)
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = entity.CreatedAt

	db.sagaEntities[entity.ID] = copySagaEntity(entity)

	if _, ok := db.workflowToChildren[parentWorkflowID]; !ok {
		db.workflowToChildren[parentWorkflowID] = make(map[EntityType][]int)
	}
	db.workflowToChildren[parentWorkflowID][EntitySaga] = append(
		db.workflowToChildren[parentWorkflowID][EntitySaga],
		int(entity.ID),
	)

	return entity.ID, nil
}

func (db *MemoryDatabase) GetSagaEntities(workflowID WorkflowEntityID, opts ...SagaEntityGetOption) ([]*SagaEntity, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	sagaIDs := db.workflowToChildren[workflowID][EntitySaga]
	entities := make([]*SagaEntity, 0, len(sagaIDs))

	for _, sID := range sagaIDs {
		// Access entity directly instead of calling GetSagaEntity
		entity, exists := db.sagaEntities[SagaEntityID(sID)]
		if !exists {
			continue
		}

		// Create a copy of the entity
		entityCopy := copySagaEntity(entity)

		// Apply options to the copied entity
		cfg := &SagaEntityGetterOptions{}
		for _, opt := range opts {
			if err := opt(cfg); err != nil {
				return nil, err
			}
		}

		if cfg.IncludeData {
			for _, d := range db.sagaData {
				if d.EntityID == SagaEntityID(sID) {
					entityCopy.SagaData = copySagaData(d)
					break
				}
			}
		}

		entities = append(entities, entityCopy)
	}

	return entities, nil
}

func (db *MemoryDatabase) AddSagaExecution(exec *SagaExecution) (SagaExecutionID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	db.sagaExecutionCounter++
	exec.ID = SagaExecutionID(db.sagaExecutionCounter)

	if exec.SagaExecutionData != nil {
		db.sagaExecutionDataCounter++
		exec.SagaExecutionData.ID = SagaExecutionDataID(db.sagaExecutionDataCounter)
		exec.SagaExecutionData.ExecutionID = exec.ID
		db.sagaExecutionData[exec.SagaExecutionData.ID] = copySagaExecutionData(exec.SagaExecutionData)
	}

	exec.CreatedAt = time.Now()
	exec.UpdatedAt = exec.CreatedAt

	db.sagaExecutions[exec.ID] = copySagaExecution(exec)
	return exec.ID, nil
}

func (db *MemoryDatabase) GetSagaEntity(id SagaEntityID, opts ...SagaEntityGetOption) (*SagaEntity, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	entity, exists := db.sagaEntities[id]
	if !exists {
		return nil, ErrSagaEntityNotFound
	}
	eCopy := copySagaEntity(entity)

	cfg := &SagaEntityGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		for _, d := range db.sagaData {
			if d.EntityID == id {
				eCopy.SagaData = copySagaData(d)
				break
			}
		}
	}

	return eCopy, nil
}

func (db *MemoryDatabase) GetSagaEntityProperties(id SagaEntityID, getters ...SagaEntityPropertyGetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	entity, exists := db.sagaEntities[id]
	if !exists {
		return ErrSagaEntityNotFound
	}
	eCopy := copySagaEntity(entity)

	opts := &SagaEntityGetterOptions{}
	for _, getter := range getters {
		opt, err := getter(eCopy)
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
				eCopy.SagaData = copySagaData(d)
				break
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSagaEntityProperties(id SagaEntityID, setters ...SagaEntityPropertySetter) error {

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
		_, wExists := db.workflowEntities[*opts.ParentWorkflowID]
		if !wExists {
			return ErrWorkflowEntityNotFound
		}

		db.entityToWorkflow[int(id)] = *opts.ParentWorkflowID
		if db.workflowToChildren[*opts.ParentWorkflowID] == nil {
			db.workflowToChildren[*opts.ParentWorkflowID] = make(map[EntityType][]int)
		}
		db.workflowToChildren[*opts.ParentWorkflowID][EntitySaga] = append(
			db.workflowToChildren[*opts.ParentWorkflowID][EntitySaga],
			int(id),
		)
	}

	entity.UpdatedAt = time.Now()
	db.sagaEntities[id] = copySagaEntity(entity)

	return nil
}

func (db *MemoryDatabase) AddSideEffectEntity(entity *SideEffectEntity, parentWorkflowID WorkflowEntityID) (SideEffectEntityID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	db.sideEffectEntityCounter++
	entity.ID = SideEffectEntityID(db.sideEffectEntityCounter)

	if entity.SideEffectData != nil {
		db.sideEffectDataCounter++
		entity.SideEffectData.ID = SideEffectDataID(db.sideEffectDataCounter)
		entity.SideEffectData.EntityID = entity.ID
		db.sideEffectData[entity.SideEffectData.ID] = copySideEffectData(entity.SideEffectData)
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = entity.CreatedAt

	db.sideEffectEntities[entity.ID] = copySideEffectEntity(entity)

	if _, ok := db.workflowToChildren[parentWorkflowID]; !ok {
		db.workflowToChildren[parentWorkflowID] = make(map[EntityType][]int)
	}
	db.workflowToChildren[parentWorkflowID][EntitySideEffect] = append(
		db.workflowToChildren[parentWorkflowID][EntitySideEffect],
		int(entity.ID),
	)

	return entity.ID, nil
}

func (db *MemoryDatabase) GetSideEffectEntities(workflowID WorkflowEntityID, opts ...SideEffectEntityGetOption) ([]*SideEffectEntity, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	sideEffectIDs := db.workflowToChildren[workflowID][EntitySideEffect]
	entities := make([]*SideEffectEntity, 0, len(sideEffectIDs))

	for _, seID := range sideEffectIDs {
		// Access entity directly instead of calling GetSideEffectEntity
		entity, exists := db.sideEffectEntities[SideEffectEntityID(seID)]
		if !exists {
			continue
		}

		// Create a copy of the entity
		entityCopy := copySideEffectEntity(entity)

		// Apply options to the copied entity
		cfg := &SideEffectEntityGetterOptions{}
		for _, opt := range opts {
			if err := opt(cfg); err != nil {
				return nil, err
			}
		}

		if cfg.IncludeData {
			for _, d := range db.sideEffectData {
				if d.EntityID == SideEffectEntityID(seID) {
					entityCopy.SideEffectData = copySideEffectData(d)
					break
				}
			}
		}

		entities = append(entities, entityCopy)
	}

	return entities, nil
}

func (db *MemoryDatabase) AddSideEffectExecution(exec *SideEffectExecution) (SideEffectExecutionID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	db.sideEffectExecutionCounter++
	exec.ID = SideEffectExecutionID(db.sideEffectExecutionCounter)

	if exec.SideEffectExecutionData != nil {
		db.sideEffectExecutionDataCounter++
		exec.SideEffectExecutionData.ID = SideEffectExecutionDataID(db.sideEffectExecutionDataCounter)
		exec.SideEffectExecutionData.ExecutionID = exec.ID
		db.sideEffectExecutionData[exec.SideEffectExecutionData.ID] = copySideEffectExecutionData(exec.SideEffectExecutionData)
	}

	exec.CreatedAt = time.Now()
	exec.UpdatedAt = exec.CreatedAt

	db.sideEffectExecutions[exec.ID] = copySideEffectExecution(exec)
	return exec.ID, nil
}

func (db *MemoryDatabase) GetSideEffectEntity(id SideEffectEntityID, opts ...SideEffectEntityGetOption) (*SideEffectEntity, error) {

	db.mu.Lock()
	defer db.mu.Unlock()

	entity, exists := db.sideEffectEntities[id]
	if !exists {
		return nil, ErrSideEffectEntityNotFound
	}
	eCopy := copySideEffectEntity(entity)

	cfg := &SideEffectEntityGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		for _, d := range db.sideEffectData {
			if d.EntityID == id {
				eCopy.SideEffectData = copySideEffectData(d)
				break
			}
		}
	}

	return eCopy, nil
}

func (db *MemoryDatabase) GetSideEffectEntityProperties(id SideEffectEntityID, getters ...SideEffectEntityPropertyGetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	entity, exists := db.sideEffectEntities[id]
	if !exists {
		return ErrSideEffectEntityNotFound
	}
	eCopy := copySideEffectEntity(entity)

	opts := &SideEffectEntityGetterOptions{}
	for _, getter := range getters {
		opt, err := getter(eCopy)
		if err != nil {
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				return err
			}
			if opts.IncludeData {
				for _, d := range db.sideEffectData {
					if d.EntityID == id {
						eCopy.SideEffectData = copySideEffectData(d)
						break
					}
				}
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSideEffectEntityProperties(id SideEffectEntityID, setters ...SideEffectEntityPropertySetter) error {

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
		_, wExists := db.workflowEntities[*opts.ParentWorkflowID]
		if !wExists {
			return ErrWorkflowEntityNotFound
		}

		db.entityToWorkflow[int(id)] = *opts.ParentWorkflowID
		if db.workflowToChildren[*opts.ParentWorkflowID] == nil {
			db.workflowToChildren[*opts.ParentWorkflowID] = make(map[EntityType][]int)
		}
		db.workflowToChildren[*opts.ParentWorkflowID][EntitySideEffect] = append(
			db.workflowToChildren[*opts.ParentWorkflowID][EntitySideEffect],
			int(id),
		)
	}

	entity.UpdatedAt = time.Now()
	db.sideEffectEntities[id] = copySideEffectEntity(entity)

	return nil
}

func (db *MemoryDatabase) GetWorkflowExecution(id WorkflowExecutionID, opts ...WorkflowExecutionGetOption) (*WorkflowExecution, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	exec, exists := db.workflowExecutions[id]
	if !exists {
		return nil, ErrWorkflowExecutionNotFound
	}
	execCopy := copyWorkflowExecution(exec)

	cfg := &WorkflowExecutionGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		for _, d := range db.workflowExecutionData {
			if d.ExecutionID == id {
				execCopy.WorkflowExecutionData = copyWorkflowExecutionData(d)
				break
			}
		}
	}

	return execCopy, nil
}

func (db *MemoryDatabase) GetWorkflowExecutions(entityID WorkflowEntityID) ([]*WorkflowExecution, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	results := make([]*WorkflowExecution, 0)
	for _, exec := range db.workflowExecutions {
		if exec.WorkflowEntityID == entityID {
			results = append(results, copyWorkflowExecution(exec))
		}
	}
	return results, nil
}

func (db *MemoryDatabase) GetActivityExecution(id ActivityExecutionID, opts ...ActivityExecutionGetOption) (*ActivityExecution, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	exec, exists := db.activityExecutions[id]
	if !exists {
		return nil, ErrActivityExecutionNotFound
	}
	execCopy := copyActivityExecution(exec)

	cfg := &ActivityExecutionGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		for _, d := range db.activityExecutionData {
			if d.ExecutionID == id {
				execCopy.ActivityExecutionData = copyActivityExecutionData(d)
				break
			}
		}
	}

	return execCopy, nil
}

func (db *MemoryDatabase) GetActivityExecutions(entityID ActivityEntityID) ([]*ActivityExecution, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	results := make([]*ActivityExecution, 0)
	for _, exec := range db.activityExecutions {
		if exec.ActivityEntityID == entityID {
			results = append(results, copyActivityExecution(exec))
		}
	}
	return results, nil
}

func (db *MemoryDatabase) GetSagaExecution(id SagaExecutionID, opts ...SagaExecutionGetOption) (*SagaExecution, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	exec, exists := db.sagaExecutions[id]
	if !exists {
		return nil, ErrSagaExecutionNotFound
	}
	execCopy := copySagaExecution(exec)

	cfg := &SagaExecutionGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		for _, d := range db.sagaExecutionData {
			if d.ExecutionID == id {
				execCopy.SagaExecutionData = copySagaExecutionData(d)
				break
			}
		}
	}

	return execCopy, nil
}

func (db *MemoryDatabase) GetSagaExecutions(entityID SagaEntityID) ([]*SagaExecution, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	results := make([]*SagaExecution, 0)
	for _, exec := range db.sagaExecutions {
		if exec.SagaEntityID == entityID {
			results = append(results, copySagaExecution(exec))
		}
	}

	if len(results) == 0 {
		return nil, ErrSagaExecutionNotFound
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.Before(results[j].CreatedAt)
	})

	return results, nil
}

func (db *MemoryDatabase) GetSideEffectExecution(id SideEffectExecutionID, opts ...SideEffectExecutionGetOption) (*SideEffectExecution, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	exec, exists := db.sideEffectExecutions[id]
	if !exists {
		return nil, ErrSideEffectExecutionNotFound
	}
	execCopy := copySideEffectExecution(exec)

	cfg := &SideEffectExecutionGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		for _, d := range db.sideEffectExecutionData {
			if d.ExecutionID == id {
				execCopy.SideEffectExecutionData = copySideEffectExecutionData(d)
				break
			}
		}
	}

	return execCopy, nil
}

func (db *MemoryDatabase) GetSideEffectExecutions(entityID SideEffectEntityID) ([]*SideEffectExecution, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	results := make([]*SideEffectExecution, 0)
	for _, exec := range db.sideEffectExecutions {
		if exec.SideEffectEntityID == entityID {
			results = append(results, copySideEffectExecution(exec))
		}
	}
	return results, nil
}

func (db *MemoryDatabase) GetActivityDataProperties(id ActivityDataID, getters ...ActivityDataPropertyGetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.activityData[id]
	if !exists {
		return ErrActivityEntityNotFound
	}

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetActivityDataProperties(id ActivityDataID, setters ...ActivityDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.activityData[id]
	if !exists {
		return ErrActivityEntityNotFound
	}

	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) GetActivityDataPropertiesByEntityID(entityID ActivityEntityID, getters ...ActivityDataPropertyGetter) error {
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

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetActivityDataPropertiesByEntityID(entityID ActivityEntityID, setters ...ActivityDataPropertySetter) error {
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
		if _, err := setter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) GetWorkflowDataProperties(id WorkflowDataID, getters ...WorkflowDataPropertyGetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	data, exists := db.workflowData[id]
	if !exists {
		return ErrWorkflowEntityNotFound
	}
	dataCopy := copyWorkflowData(data)

	for _, getter := range getters {
		if _, err := getter(dataCopy); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetWorkflowDataProperties(id WorkflowDataID, setters ...WorkflowDataPropertySetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	data, exists := db.workflowData[id]
	if !exists {
		return ErrWorkflowEntityNotFound
	}

	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			return err
		}
	}

	db.workflowData[id] = copyWorkflowData(data)
	return nil
}

func (db *MemoryDatabase) GetWorkflowDataPropertiesByEntityID(entityID WorkflowEntityID, getters ...WorkflowDataPropertyGetter) error {
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

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetWorkflowDataPropertiesByEntityID(entityID WorkflowEntityID, setters ...WorkflowDataPropertySetter) error {
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
		if _, err := setter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) GetWorkflowExecutionDataIDByEntityID(executionID WorkflowExecutionID) (WorkflowExecutionDataID, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	dataID, exists := db.workflowExecToDataMap[executionID]
	if !exists {
		return WorkflowExecutionDataID(0), ErrWorkflowExecutionNotFound
	}

	return dataID, nil
}

func (db *MemoryDatabase) GetWorkflowExecutionDataProperties(id WorkflowExecutionDataID, getters ...WorkflowExecutionDataPropertyGetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.workflowExecutionData[id]
	if !exists {
		return ErrWorkflowExecutionNotFound
	}

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetWorkflowExecutionDataProperties(id WorkflowExecutionDataID, setters ...WorkflowExecutionDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.workflowExecutionData[id]
	if !exists {
		return ErrWorkflowExecutionNotFound
	}

	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) GetWorkflowExecutionDataPropertiesByExecutionID(executionID WorkflowExecutionID, getters ...WorkflowExecutionDataPropertyGetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	dataID, exists := db.workflowExecToDataMap[executionID]
	if !exists {
		return ErrWorkflowExecutionNotFound
	}

	data, exists := db.workflowExecutionData[dataID]
	if !exists {
		return ErrWorkflowExecutionNotFound
	}

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetWorkflowExecutionDataPropertiesByExecutionID(executionID WorkflowExecutionID, setters ...WorkflowExecutionDataPropertySetter) error {
	db.mu.Lock()

	// First check if data exists
	var data *WorkflowExecutionData
	var dataID WorkflowExecutionDataID
	var exists bool

	dataID, exists = db.workflowExecToDataMap[executionID]
	if exists {
		data = db.workflowExecutionData[dataID]
	}

	// If no data exists, create new data structure directly
	if data == nil {
		db.workflowExecutionDataCounter++
		dataID = WorkflowExecutionDataID(db.workflowExecutionDataCounter)

		data = &WorkflowExecutionData{
			ID:          dataID,
			ExecutionID: executionID,
		}

		db.workflowExecutionData[dataID] = data
		db.workflowExecToDataMap[executionID] = dataID
	}

	// Apply setters
	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			db.mu.Unlock()
			return err
		}
	}

	db.mu.Unlock()
	return nil
}

func (db *MemoryDatabase) GetActivityExecutionDataProperties(id ActivityExecutionDataID, getters ...ActivityExecutionDataPropertyGetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.activityExecutionData[id]
	if !exists {
		return ErrActivityExecutionNotFound
	}

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetActivityExecutionDataProperties(id ActivityExecutionDataID, setters ...ActivityExecutionDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.activityExecutionData[id]
	if !exists {
		return ErrActivityExecutionNotFound
	}

	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) GetActivityExecutionDataPropertiesByExecutionID(executionID ActivityExecutionID, getters ...ActivityExecutionDataPropertyGetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	dataID, exists := db.activityExecToDataMap[executionID]
	if !exists {
		return ErrActivityExecutionNotFound
	}

	data, exists := db.activityExecutionData[dataID]
	if !exists {
		return ErrActivityExecutionNotFound
	}

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetActivityExecutionDataPropertiesByExecutionID(executionID ActivityExecutionID, setters ...ActivityExecutionDataPropertySetter) error {
	db.mu.Lock()

	// First check if data exists
	var data *ActivityExecutionData
	var dataID ActivityExecutionDataID
	var exists bool

	dataID, exists = db.activityExecToDataMap[executionID]
	if exists {
		data = db.activityExecutionData[dataID]
	}

	// If no data exists, create new data structure without calling AddActivityExecutionData
	if data == nil {
		db.activityExecutionDataCounter++
		dataID = ActivityExecutionDataID(db.activityExecutionDataCounter)

		data = &ActivityExecutionData{
			ID:          dataID,
			ExecutionID: executionID,
		}

		db.activityExecutionData[dataID] = data
		db.activityExecToDataMap[executionID] = dataID
	}

	// Apply setters
	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			db.mu.Unlock()
			return err
		}
	}

	db.mu.Unlock()
	return nil
}

func (db *MemoryDatabase) GetSideEffectExecutionDataProperties(id SideEffectExecutionDataID, getters ...SideEffectExecutionDataPropertyGetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	data, exists := db.sideEffectExecutionData[id]
	if !exists {
		return ErrSideEffectExecutionNotFound
	}
	dataCopy := copySideEffectExecutionData(data)

	for _, getter := range getters {
		if _, err := getter(dataCopy); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSideEffectExecutionDataProperties(id SideEffectExecutionDataID, setters ...SideEffectExecutionDataPropertySetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	data, exists := db.sideEffectExecutionData[id]
	if !exists {
		return ErrSideEffectExecutionNotFound
	}

	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			return err
		}
	}

	db.sideEffectExecutionData[id] = copySideEffectExecutionData(data)
	return nil
}

func (db *MemoryDatabase) GetWorkflowExecutionProperties(id WorkflowExecutionID, getters ...WorkflowExecutionPropertyGetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	exec, exists := db.workflowExecutions[id]
	if !exists {
		return ErrWorkflowExecutionNotFound
	}
	execCopy := copyWorkflowExecution(exec)

	for _, getter := range getters {
		opt, err := getter(execCopy)
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
						execCopy.WorkflowExecutionData = copyWorkflowExecutionData(d)
						break
					}
				}
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetWorkflowExecutionProperties(id WorkflowExecutionID, setters ...WorkflowExecutionPropertySetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	exec, exists := db.workflowExecutions[id]
	if !exists {
		return ErrWorkflowExecutionNotFound
	}
	for _, setter := range setters {
		if _, err := setter(exec); err != nil {
			return err
		}
	}
	db.workflowExecutions[id] = copyWorkflowExecution(exec)
	return nil
}

func (db *MemoryDatabase) GetActivityExecutionProperties(id ActivityExecutionID, getters ...ActivityExecutionPropertyGetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	exec, exists := db.activityExecutions[id]
	if !exists {
		return ErrActivityExecutionNotFound
	}
	execCopy := copyActivityExecution(exec)

	for _, getter := range getters {
		opt, err := getter(execCopy)
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
						execCopy.ActivityExecutionData = copyActivityExecutionData(d)
						break
					}
				}
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetActivityExecutionProperties(id ActivityExecutionID, setters ...ActivityExecutionPropertySetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	exec, exists := db.activityExecutions[id]
	if !exists {
		return ErrActivityExecutionNotFound
	}
	for _, setter := range setters {
		if _, err := setter(exec); err != nil {
			return err
		}
	}
	db.activityExecutions[id] = copyActivityExecution(exec)
	return nil
}

func (db *MemoryDatabase) GetSagaExecutionProperties(id SagaExecutionID, getters ...SagaExecutionPropertyGetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	exec, exists := db.sagaExecutions[id]
	if !exists {
		return ErrSagaExecutionNotFound
	}
	execCopy := copySagaExecution(exec)

	for _, getter := range getters {
		opt, err := getter(execCopy)
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
						execCopy.SagaExecutionData = copySagaExecutionData(d)
						break
					}
				}
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSagaExecutionProperties(id SagaExecutionID, setters ...SagaExecutionPropertySetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	exec, exists := db.sagaExecutions[id]
	if !exists {
		return ErrSagaExecutionNotFound
	}
	for _, setter := range setters {
		if _, err := setter(exec); err != nil {
			return err
		}
	}
	db.sagaExecutions[id] = copySagaExecution(exec)
	return nil
}

func (db *MemoryDatabase) GetSideEffectExecutionProperties(id SideEffectExecutionID, getters ...SideEffectExecutionPropertyGetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	exec, exists := db.sideEffectExecutions[id]
	if !exists {
		return ErrSideEffectExecutionNotFound
	}
	execCopy := copySideEffectExecution(exec)

	for _, getter := range getters {
		opt, err := getter(execCopy)
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
						execCopy.SideEffectExecutionData = copySideEffectExecutionData(d)
						break
					}
				}
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSideEffectExecutionProperties(id SideEffectExecutionID, setters ...SideEffectExecutionPropertySetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	exec, exists := db.sideEffectExecutions[id]
	if !exists {
		return ErrSideEffectExecutionNotFound
	}
	for _, setter := range setters {
		if _, err := setter(exec); err != nil {
			return err
		}
	}
	db.sideEffectExecutions[id] = copySideEffectExecution(exec)
	return nil
}

func (db *MemoryDatabase) GetHierarchyProperties(id HierarchyID, getters ...HierarchyPropertyGetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	hierarchy, exists := db.hierarchies[id]
	if !exists {
		return ErrHierarchyNotFound
	}
	hCopy := copyHierarchy(hierarchy)

	for _, getter := range getters {
		if _, err := getter(hCopy); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) GetHierarchyByParentEntity(parentEntityID int, childStepID string, specificType EntityType) (*Hierarchy, error) {

	db.mu.Lock()
	defer db.mu.Unlock()

	for _, h := range db.hierarchies {
		if h.ParentEntityID == parentEntityID && h.ChildStepID == childStepID && h.ChildType == specificType {
			return copyHierarchy(h), nil
		}
	}

	return nil, ErrHierarchyNotFound
}

func (db *MemoryDatabase) GetHierarchiesByParentEntityAndStep(parentEntityID int, childStepID string, specificType EntityType) ([]*Hierarchy, error) {

	db.mu.Lock()
	defer db.mu.Unlock()

	var results []*Hierarchy
	for _, h := range db.hierarchies {
		if h.ParentEntityID == parentEntityID && h.ChildStepID == childStepID && h.ChildType == specificType {
			results = append(results, copyHierarchy(h))
		}
	}

	if len(results) == 0 {
		return nil, ErrHierarchyNotFound
	}

	return results, nil
}

func (db *MemoryDatabase) SetHierarchyProperties(id HierarchyID, setters ...HierarchyPropertySetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	hierarchy, exists := db.hierarchies[id]
	if !exists {
		return ErrHierarchyNotFound
	}
	for _, setter := range setters {
		if _, err := setter(hierarchy); err != nil {
			return err
		}
	}
	db.hierarchies[id] = copyHierarchy(hierarchy)
	return nil
}

func (db *MemoryDatabase) GetHierarchiesByParentEntity(parentEntityID int) ([]*Hierarchy, error) {

	db.mu.Lock()
	defer db.mu.Unlock()

	var results []*Hierarchy
	for _, h := range db.hierarchies {
		if h.ParentEntityID == parentEntityID {
			results = append(results, copyHierarchy(h))
		}
	}
	return results, nil
}

func (db *MemoryDatabase) GetHierarchiesByChildEntity(childEntityID int) ([]*Hierarchy, error) {

	db.mu.Lock()
	defer db.mu.Unlock()

	var results []*Hierarchy
	for _, h := range db.hierarchies {
		if h.ChildEntityID == childEntityID {
			results = append(results, copyHierarchy(h))
		}
	}
	return results, nil
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

func (db *MemoryDatabase) GetQueueProperties(id QueueID, getters ...QueuePropertyGetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	queue, exists := db.queues[id]
	if !exists {
		return ErrQueueNotFound
	}
	qCopy := copyQueue(queue)

	for _, getter := range getters {
		opt, err := getter(qCopy)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &QueueGetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
			if opts.IncludeWorkflows {
				workflowIDs := db.queueToWorkflows[id]

				if len(workflowIDs) > 0 {
					workflows := make([]*WorkflowEntity, 0, len(workflowIDs))
					for _, wfID := range workflowIDs {
						if wf, wexists := db.workflowEntities[wfID]; wexists {
							workflows = append(workflows, copyWorkflowEntity(wf))
						}
					}
					qCopy.Entities = workflows
				}
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetQueueProperties(id QueueID, setters ...QueuePropertySetter) error {

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
				// Remove old workflows
				if oldWorkflowIDs, ok := db.queueToWorkflows[id]; ok {
					for _, oldWfID := range oldWorkflowIDs {
						delete(db.workflowToQueue, oldWfID)
					}
				}
				db.queueToWorkflows[id] = opts.WorkflowIDs
				for _, wfID := range opts.WorkflowIDs {
					db.workflowToQueue[wfID] = id
				}
			}
		}
	}
	db.queues[id] = copyQueue(queue)
	return nil
}

func (db *MemoryDatabase) UpdateQueue(queue *Queue) error {

	db.mu.Lock()
	defer db.mu.Unlock()

	oldQ, exists := db.queues[queue.ID]
	if !exists {
		return ErrQueueNotFound
	}

	if oldQ.Name != queue.Name {
		delete(db.queueNames, oldQ.Name)
		db.queueNames[queue.Name] = queue.ID
	}

	queue.UpdatedAt = time.Now()
	db.queues[queue.ID] = copyQueue(queue)
	return nil
}

// GetVersionByWorkflowAndChangeID
func (db *MemoryDatabase) GetVersionByWorkflowAndChangeID(workflowID WorkflowEntityID, changeID string) (*Version, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	versionIDs := db.workflowVersions[workflowID]

	for _, vID := range versionIDs {
		if v, vexists := db.versions[vID]; vexists && v.ChangeID == changeID {
			return copyVersion(v), nil
		}
	}
	return nil, ErrVersionNotFound
}

func (db *MemoryDatabase) GetVersionsByWorkflowID(workflowID WorkflowEntityID) ([]*Version, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	versionIDs := db.workflowVersions[workflowID]

	versions := make([]*Version, 0, len(versionIDs))
	for _, vID := range versionIDs {
		if version, exists := db.versions[vID]; exists {
			versions = append(versions, copyVersion(version))
		}
	}

	return versions, nil
}

func (db *MemoryDatabase) SetVersion(version *Version) error {

	db.mu.Lock()
	defer db.mu.Unlock()

	if version.ID == 0 {
		db.versionCounter++
		version.ID = VersionID(db.versionCounter)
		version.CreatedAt = time.Now()
	}
	version.UpdatedAt = time.Now()

	db.versions[version.ID] = copyVersion(version)

	found := false
	for _, vID := range db.workflowVersions[version.EntityID] {
		if v, vex := db.versions[vID]; vex && v.ChangeID == version.ChangeID {
			db.versions[vID] = copyVersion(version)
			found = true
			break
		}
	}
	if !found {
		db.workflowVersions[version.EntityID] = append(db.workflowVersions[version.EntityID], version.ID)
	}

	return nil
}

func (db *MemoryDatabase) DeleteVersionsForWorkflow(workflowID WorkflowEntityID) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	versionIDs, ok := db.workflowVersions[workflowID]
	if ok {
		for _, vID := range versionIDs {
			delete(db.versions, vID)
		}
		delete(db.workflowVersions, workflowID)
	}
	return nil
}

func (db *MemoryDatabase) GetVersionProperties(id VersionID, getters ...VersionPropertyGetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	version, exists := db.versions[id]
	if !exists {
		return ErrVersionNotFound
	}
	vCopy := copyVersion(version)

	for _, getter := range getters {
		if _, err := getter(vCopy); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetVersionProperties(id VersionID, setters ...VersionPropertySetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	version, exists := db.versions[id]
	if !exists {
		return ErrVersionNotFound
	}

	for _, setter := range setters {
		if _, err := setter(version); err != nil {
			return err
		}
	}
	version.UpdatedAt = time.Now()
	db.versions[id] = copyVersion(version)
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

func (db *MemoryDatabase) AddWorkflowData(entityID WorkflowEntityID, data *WorkflowData) (WorkflowDataID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	db.workflowDataCounter++
	data.ID = WorkflowDataID(db.workflowDataCounter)
	data.EntityID = entityID
	db.workflowData[data.ID] = copyWorkflowData(data)
	return data.ID, nil
}

func (db *MemoryDatabase) GetWorkflowDataIDByEntityID(entityID WorkflowEntityID) (WorkflowDataID, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, data := range db.workflowData {
		if data.EntityID == entityID {
			return data.ID, nil
		}
	}

	return WorkflowDataID(0), ErrWorkflowEntityNotFound
}

func (db *MemoryDatabase) AddActivityData(entityID ActivityEntityID, data *ActivityData) (ActivityDataID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	db.activityDataCounter++
	data.ID = ActivityDataID(db.activityDataCounter)
	data.EntityID = entityID
	db.activityData[data.ID] = copyActivityData(data)
	return data.ID, nil
}

func (db *MemoryDatabase) AddSagaData(entityID SagaEntityID, data *SagaData) (SagaDataID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	db.sagaDataCounter++
	data.ID = SagaDataID(db.sagaDataCounter)
	data.EntityID = entityID
	db.sagaData[data.ID] = copySagaData(data)
	return data.ID, nil
}

func (db *MemoryDatabase) AddSideEffectData(entityID SideEffectEntityID, data *SideEffectData) (SideEffectDataID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	db.sideEffectDataCounter++
	data.ID = SideEffectDataID(db.sideEffectDataCounter)
	data.EntityID = entityID
	db.sideEffectData[data.ID] = copySideEffectData(data)
	return data.ID, nil
}

func (db *MemoryDatabase) GetWorkflowData(id WorkflowDataID) (*WorkflowData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	d, exists := db.workflowData[id]
	if !exists {
		return nil, ErrWorkflowEntityNotFound
	}
	dCopy := copyWorkflowData(d)
	return dCopy, nil
}

func (db *MemoryDatabase) GetActivityData(id ActivityDataID) (*ActivityData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	d, exists := db.activityData[id]
	if !exists {
		return nil, ErrActivityEntityNotFound
	}
	dCopy := copyActivityData(d)
	return dCopy, nil
}

func (db *MemoryDatabase) GetActivityDataByEntityID(entityID ActivityEntityID) (*ActivityData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, d := range db.activityData {
		if d.EntityID == entityID {
			dCopy := copyActivityData(d)
			return dCopy, nil
		}
	}
	return nil, ErrActivityEntityNotFound
}

func (db *MemoryDatabase) GetSagaData(id SagaDataID) (*SagaData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	d, exists := db.sagaData[id]
	if !exists {
		return nil, ErrSagaEntityNotFound
	}
	dCopy := copySagaData(d)
	return dCopy, nil
}

func (db *MemoryDatabase) GetSideEffectData(id SideEffectDataID) (*SideEffectData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	d, exists := db.sideEffectData[id]
	if !exists {
		return nil, ErrSideEffectEntityNotFound
	}
	dCopy := copySideEffectData(d)
	return dCopy, nil
}

func (db *MemoryDatabase) GetWorkflowDataByEntityID(entityID WorkflowEntityID) (*WorkflowData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, d := range db.workflowData {
		if d.EntityID == entityID {
			dCopy := copyWorkflowData(d)
			return dCopy, nil
		}
	}
	return nil, ErrWorkflowEntityNotFound
}

func (db *MemoryDatabase) GetSagaDataByEntityID(entityID SagaEntityID) (*SagaData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, d := range db.sagaData {
		if d.EntityID == entityID {
			dCopy := copySagaData(d)
			return dCopy, nil
		}
	}
	return nil, ErrSagaEntityNotFound
}

func (db *MemoryDatabase) GetSideEffectDataByEntityID(entityID SideEffectEntityID) (*SideEffectData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, d := range db.sideEffectData {
		if d.EntityID == entityID {
			dCopy := copySideEffectData(d)
			return dCopy, nil
		}
	}
	return nil, ErrSideEffectEntityNotFound
}

func (db *MemoryDatabase) AddWorkflowExecutionData(executionID WorkflowExecutionID, data *WorkflowExecutionData) (WorkflowExecutionDataID, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.workflowExecutionDataCounter++
	data.ID = WorkflowExecutionDataID(db.workflowExecutionDataCounter)
	data.ExecutionID = executionID
	db.workflowExecutionData[data.ID] = copyWorkflowExecutionData(data)

	db.workflowExecToDataMap[executionID] = data.ID

	return data.ID, nil
}

func (db *MemoryDatabase) AddActivityExecutionData(executionID ActivityExecutionID, data *ActivityExecutionData) (ActivityExecutionDataID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	db.activityExecutionDataCounter++
	data.ID = ActivityExecutionDataID(db.activityExecutionDataCounter)
	data.ExecutionID = executionID
	db.activityExecutionData[data.ID] = copyActivityExecutionData(data)

	db.activityExecToDataMap[executionID] = data.ID

	return data.ID, nil
}

func (db *MemoryDatabase) AddSagaExecutionData(executionID SagaExecutionID, data *SagaExecutionData) (SagaExecutionDataID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	db.sagaExecutionDataCounter++
	data.ID = SagaExecutionDataID(db.sagaExecutionDataCounter)
	data.ExecutionID = executionID
	db.sagaExecutionData[data.ID] = copySagaExecutionData(data)

	db.sagaExecToDataMap[executionID] = data.ID

	return data.ID, nil
}

func (db *MemoryDatabase) AddSideEffectExecutionData(executionID SideEffectExecutionID, data *SideEffectExecutionData) (SideEffectExecutionDataID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	db.sideEffectExecutionDataCounter++
	data.ID = SideEffectExecutionDataID(db.sideEffectExecutionDataCounter)
	data.ExecutionID = executionID
	db.sideEffectExecutionData[data.ID] = copySideEffectExecutionData(data)

	db.sideEffectExecToDataMap[executionID] = data.ID

	return data.ID, nil
}

func (db *MemoryDatabase) GetWorkflowExecutionData(id WorkflowExecutionDataID) (*WorkflowExecutionData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	d, exists := db.workflowExecutionData[id]
	if !exists {
		return nil, ErrWorkflowExecutionNotFound
	}
	dCopy := copyWorkflowExecutionData(d)
	return dCopy, nil
}

func (db *MemoryDatabase) GetActivityExecutionData(id ActivityExecutionDataID) (*ActivityExecutionData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	d, exists := db.activityExecutionData[id]
	if !exists {
		return nil, ErrActivityExecutionNotFound
	}
	dCopy := copyActivityExecutionData(d)
	return dCopy, nil
}

func (db *MemoryDatabase) GetSagaExecutionData(id SagaExecutionDataID) (*SagaExecutionData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	d, exists := db.sagaExecutionData[id]
	if !exists {
		return nil, ErrSagaExecutionNotFound
	}
	dCopy := copySagaExecutionData(d)
	return dCopy, nil
}

func (db *MemoryDatabase) GetSideEffectExecutionData(id SideEffectExecutionDataID) (*SideEffectExecutionData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	d, exists := db.sideEffectExecutionData[id]
	if !exists {
		return nil, ErrSideEffectExecutionNotFound
	}
	dCopy := copySideEffectExecutionData(d)
	return dCopy, nil
}

func (db *MemoryDatabase) GetWorkflowExecutionDataByExecutionID(executionID WorkflowExecutionID) (*WorkflowExecutionData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	dataID, exists := db.workflowExecToDataMap[executionID]

	if !exists {
		return nil, ErrWorkflowExecutionNotFound
	}

	if data, ok := db.workflowExecutionData[dataID]; ok {
		return copyWorkflowExecutionData(data), nil
	}

	return nil, ErrWorkflowExecutionNotFound
}

func (db *MemoryDatabase) GetActivityExecutionDataByExecutionID(executionID ActivityExecutionID) (*ActivityExecutionData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, d := range db.activityExecutionData {
		if d.ExecutionID == executionID {
			dCopy := copyActivityExecutionData(d)
			return dCopy, nil
		}
	}
	return nil, ErrActivityExecutionNotFound
}

func (db *MemoryDatabase) GetSagaExecutionDataByExecutionID(executionID SagaExecutionID) (*SagaExecutionData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, d := range db.sagaExecutionData {
		if d.ExecutionID == executionID {
			dCopy := copySagaExecutionData(d)
			return dCopy, nil
		}
	}
	return nil, ErrSagaExecutionNotFound
}

func (db *MemoryDatabase) GetSideEffectExecutionDataByExecutionID(executionID SideEffectExecutionID) (*SideEffectExecutionData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, d := range db.sideEffectExecutionData {
		if d.ExecutionID == executionID {
			dCopy := copySideEffectExecutionData(d)
			return dCopy, nil
		}
	}
	return nil, ErrSideEffectExecutionNotFound
}

func (db *MemoryDatabase) HasRun(id RunID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.runs[id]
	return exists, nil
}

func (db *MemoryDatabase) HasVersion(id VersionID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.versions[id]
	return exists, nil
}

func (db *MemoryDatabase) HasHierarchy(id HierarchyID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.hierarchies[id]
	return exists, nil
}

func (db *MemoryDatabase) HasQueue(id QueueID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.queues[id]
	return exists, nil
}

func (db *MemoryDatabase) HasQueueName(name string) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.queueNames[name]
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowEntity(id WorkflowEntityID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.workflowEntities[id]
	return exists, nil
}

func (db *MemoryDatabase) HasActivityEntity(id ActivityEntityID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.activityEntities[id]
	return exists, nil
}

func (db *MemoryDatabase) HasSagaEntity(id SagaEntityID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.sagaEntities[id]
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectEntity(id SideEffectEntityID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.sideEffectEntities[id]
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowExecution(id WorkflowExecutionID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.workflowExecutions[id]
	return exists, nil
}

func (db *MemoryDatabase) HasActivityExecution(id ActivityExecutionID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.activityExecutions[id]
	return exists, nil
}

func (db *MemoryDatabase) HasSagaExecution(id SagaExecutionID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.sagaExecutions[id]
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectExecution(id SideEffectExecutionID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.sideEffectExecutions[id]
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowData(id WorkflowDataID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.workflowData[id]
	return exists, nil
}

func (db *MemoryDatabase) HasActivityData(id ActivityDataID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.activityData[id]
	return exists, nil
}

func (db *MemoryDatabase) GetActivityDataIDByEntityID(entityID ActivityEntityID) (ActivityDataID, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, data := range db.activityData {
		if data.EntityID == entityID {
			return data.ID, nil
		}
	}

	return ActivityDataID(0), ErrActivityEntityNotFound
}

func (db *MemoryDatabase) HasSagaData(id SagaDataID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.sagaData[id]
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectData(id SideEffectDataID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.sideEffectData[id]
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowDataByEntityID(entityID WorkflowEntityID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, d := range db.workflowData {
		if d.EntityID == entityID {
			return true, nil
		}
	}
	return false, nil
}

func (db *MemoryDatabase) HasActivityDataByEntityID(entityID ActivityEntityID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, d := range db.activityData {
		if d.EntityID == entityID {
			return true, nil
		}
	}
	return false, nil
}

func (db *MemoryDatabase) HasSagaDataByEntityID(entityID SagaEntityID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, d := range db.sagaData {
		if d.EntityID == entityID {
			return true, nil
		}
	}
	return false, nil
}

func (db *MemoryDatabase) HasSideEffectDataByEntityID(entityID SideEffectEntityID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, d := range db.sideEffectData {
		if d.EntityID == entityID {
			return true, nil
		}
	}
	return false, nil
}

func (db *MemoryDatabase) HasWorkflowExecutionData(id WorkflowExecutionDataID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.workflowExecutionData[id]
	return exists, nil
}

func (db *MemoryDatabase) HasActivityExecutionData(id ActivityExecutionDataID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.activityExecutionData[id]
	return exists, nil
}

func (db *MemoryDatabase) GetActivityExecutionDataIDByEntityID(executionID ActivityExecutionID) (ActivityExecutionDataID, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	dataID, exists := db.activityExecToDataMap[executionID]
	if !exists {
		return ActivityExecutionDataID(0), ErrActivityExecutionNotFound
	}

	return dataID, nil
}

func (db *MemoryDatabase) HasSagaExecutionData(id SagaExecutionDataID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.sagaExecutionData[id]
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectExecutionData(id SideEffectExecutionDataID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.sideEffectExecutionData[id]
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowExecutionDataByExecutionID(executionID WorkflowExecutionID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, d := range db.workflowExecutionData {
		if d.ExecutionID == executionID {
			return true, nil
		}
	}
	return false, nil
}

func (db *MemoryDatabase) HasActivityExecutionDataByExecutionID(executionID ActivityExecutionID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, d := range db.activityExecutionData {
		if d.ExecutionID == executionID {
			return true, nil
		}
	}
	return false, nil
}

func (db *MemoryDatabase) HasSagaExecutionDataByExecutionID(executionID SagaExecutionID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, d := range db.sagaExecutionData {
		if d.ExecutionID == executionID {
			return true, nil
		}
	}
	return false, nil
}

func (db *MemoryDatabase) HasSideEffectExecutionDataByExecutionID(executionID SideEffectExecutionID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, d := range db.sideEffectExecutionData {
		if d.ExecutionID == executionID {
			return true, nil
		}
	}
	return false, nil
}

func (db *MemoryDatabase) GetWorkflowExecutionLatestByEntityID(entityID WorkflowEntityID) (*WorkflowExecution, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	var latestExec *WorkflowExecution
	var latestTime time.Time
	for _, exec := range db.workflowExecutions {
		if exec.WorkflowEntityID == entityID {
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

func (db *MemoryDatabase) GetActivityExecutionLatestByEntityID(entityID ActivityEntityID) (*ActivityExecution, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	var latestExec *ActivityExecution
	var latestTime time.Time
	for _, exec := range db.activityExecutions {
		if exec.ActivityEntityID == entityID {
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

func (db *MemoryDatabase) SetSagaValue(executionID SagaExecutionID, key string, value []byte) (SagaValueID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.sagaValues[executionID]; !exists {
		db.sagaValues[executionID] = make(map[string]*SagaValue)
	}

	sv := &SagaValue{
		ID:          SagaValueID(len(db.sagaValues[executionID]) + 1),
		ExecutionID: executionID,
		Key:         key,
		Value:       value,
	}
	db.sagaValues[executionID][key] = sv
	return sv.ID, nil
}

func (db *MemoryDatabase) GetSagaValue(id SagaValueID, key string) ([]byte, error) {

	db.mu.Lock()
	defer db.mu.Unlock()

	for _, keyMap := range db.sagaValues {
		if val, exists := keyMap[key]; exists && val.ID == id {
			return val.Value, nil
		}
	}
	return nil, ErrSagaValueNotFound
}

func (db *MemoryDatabase) GetSagaValueByExecutionID(executionID SagaExecutionID, key string) ([]byte, error) {

	db.mu.Lock()
	defer db.mu.Unlock()

	if keyMap, exists := db.sagaValues[executionID]; exists {
		if val, vexists := keyMap[key]; vexists {
			return val.Value, nil
		}
	}
	return nil, ErrSagaValueNotFound
}

func (db *MemoryDatabase) AddSignalEntity(entity *SignalEntity, parentWorkflowID WorkflowEntityID) (SignalEntityID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	db.signalEntityCounter++
	entity.ID = SignalEntityID(db.signalEntityCounter)

	if entity.SignalData != nil {
		db.signalDataCounter++
		entity.SignalData.ID = SignalDataID(db.signalDataCounter)
		entity.SignalData.EntityID = entity.ID
		db.signalData[entity.SignalData.ID] = copySignalData(entity.SignalData)
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = entity.CreatedAt

	db.signalEntities[entity.ID] = copySignalEntity(entity)

	// Signal relationships to workflow are through hierarchies,
	// but the code provided doesn't explicitly add them here.
	return entity.ID, nil
}

func (db *MemoryDatabase) GetSignalEntity(id SignalEntityID, opts ...SignalEntityGetOption) (*SignalEntity, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	entity, exists := db.signalEntities[id]
	if !exists {
		return nil, fmt.Errorf("signal entity not found")
	}
	eCopy := copySignalEntity(entity)

	cfg := &SignalEntityGetterOptions{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		for _, d := range db.signalData {
			if d.EntityID == id {
				eCopy.SignalData = copySignalData(d)
				break
			}
		}
	}

	return eCopy, nil
}

func (db *MemoryDatabase) GetSignalEntities(workflowID WorkflowEntityID, opts ...SignalEntityGetOption) ([]*SignalEntity, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	var entities []*SignalEntity
	for _, h := range db.hierarchies {
		if h.ParentEntityID == int(workflowID) && h.ChildType == EntitySignal {
			if e, ex := db.signalEntities[SignalEntityID(h.ChildEntityID)]; ex {
				entities = append(entities, copySignalEntity(e))
			}
		}
	}

	return entities, nil
}

func (db *MemoryDatabase) AddSignalExecution(exec *SignalExecution) (SignalExecutionID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	db.signalExecutionCounter++
	exec.ID = SignalExecutionID(db.signalExecutionCounter)

	if exec.SignalExecutionData != nil {
		db.signalExecutionDataCounter++
		exec.SignalExecutionData.ID = SignalExecutionDataID(db.signalExecutionDataCounter)
		exec.SignalExecutionData.ExecutionID = exec.ID
		db.signalExecutionData[exec.SignalExecutionData.ID] = copySignalExecutionData(exec.SignalExecutionData)
	}

	exec.CreatedAt = time.Now()
	exec.UpdatedAt = exec.CreatedAt

	db.signalExecutions[exec.ID] = copySignalExecution(exec)
	return exec.ID, nil
}

func (db *MemoryDatabase) GetSignalExecution(id SignalExecutionID, opts ...SignalExecutionGetOption) (*SignalExecution, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	exec, exists := db.signalExecutions[id]
	if !exists {
		return nil, fmt.Errorf("signal execution not found")
	}
	execCopy := copySignalExecution(exec)

	cfg := &SignalExecutionGetterOptions{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		for _, d := range db.signalExecutionData {
			if d.ExecutionID == id {
				execCopy.SignalExecutionData = copySignalExecutionData(d)
				break
			}
		}
	}

	return execCopy, nil
}

func (db *MemoryDatabase) GetSignalExecutions(entityID SignalEntityID) ([]*SignalExecution, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	var executions []*SignalExecution
	for _, exec := range db.signalExecutions {
		if exec.EntityID == entityID {
			executions = append(executions, copySignalExecution(exec))
		}
	}
	return executions, nil
}

func (db *MemoryDatabase) GetSignalExecutionLatestByEntityID(entityID SignalEntityID) (*SignalExecution, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	var latest *SignalExecution
	var latestTime time.Time
	for _, exec := range db.signalExecutions {
		if exec.EntityID == entityID {
			if latest == nil || exec.CreatedAt.After(latestTime) {
				latest = exec
				latestTime = exec.CreatedAt
			}
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("no signal execution found")
	}

	return copySignalExecution(latest), nil
}

func (db *MemoryDatabase) HasSignalEntity(id SignalEntityID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.signalEntities[id]
	return exists, nil
}

func (db *MemoryDatabase) HasSignalExecution(id SignalExecutionID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.signalExecutions[id]
	return exists, nil
}

func (db *MemoryDatabase) HasSignalData(id SignalDataID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.signalData[id]
	return exists, nil
}

func (db *MemoryDatabase) HasSignalExecutionData(id SignalExecutionDataID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	_, exists := db.signalExecutionData[id]
	return exists, nil
}

func (db *MemoryDatabase) GetSignalEntityProperties(id SignalEntityID, getters ...SignalEntityPropertyGetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	entity, exists := db.signalEntities[id]
	if !exists {
		return fmt.Errorf("signal entity not found")
	}
	eCopy := copySignalEntity(entity)

	for _, getter := range getters {
		opt, err := getter(eCopy)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &SignalEntityGetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
			if opts.IncludeData {
				for _, d := range db.signalData {
					if d.EntityID == id {
						eCopy.SignalData = copySignalData(d)
						break
					}
				}
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSignalEntityProperties(id SignalEntityID, setters ...SignalEntityPropertySetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	entity, exists := db.signalEntities[id]
	if !exists {
		return fmt.Errorf("signal entity not found")
	}

	for _, setter := range setters {
		if _, err := setter(entity); err != nil {
			return err
		}
	}
	entity.UpdatedAt = time.Now()
	db.signalEntities[id] = copySignalEntity(entity)

	return nil
}

func (db *MemoryDatabase) GetSignalExecutionProperties(id SignalExecutionID, getters ...SignalExecutionPropertyGetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	exec, exists := db.signalExecutions[id]
	if !exists {
		return fmt.Errorf("signal execution not found")
	}
	execCopy := copySignalExecution(exec)

	for _, getter := range getters {
		opt, err := getter(execCopy)
		if err != nil {
			return err
		}
		if opt != nil {
			opts := &SignalExecutionGetterOptions{}
			if err := opt(opts); err != nil {
				return err
			}
			if opts.IncludeData {
				for _, d := range db.signalExecutionData {
					if d.ExecutionID == id {
						execCopy.SignalExecutionData = copySignalExecutionData(d)
						break
					}
				}
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSignalExecutionProperties(id SignalExecutionID, setters ...SignalExecutionPropertySetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	exec, exists := db.signalExecutions[id]
	if !exists {
		return fmt.Errorf("signal execution not found")
	}

	for _, setter := range setters {
		if _, err := setter(exec); err != nil {
			return err
		}
	}
	exec.UpdatedAt = time.Now()
	db.signalExecutions[id] = copySignalExecution(exec)
	return nil
}

func (db *MemoryDatabase) AddSignalData(entityID SignalEntityID, data *SignalData) (SignalDataID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	db.signalDataCounter++
	data.ID = SignalDataID(db.signalDataCounter)
	data.EntityID = entityID
	db.signalData[data.ID] = copySignalData(data)
	return data.ID, nil
}

func (db *MemoryDatabase) GetSignalData(id SignalDataID) (*SignalData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	d, exists := db.signalData[id]
	if !exists {
		return nil, fmt.Errorf("signal data not found")
	}
	dCopy := copySignalData(d)
	return dCopy, nil
}

func (db *MemoryDatabase) GetSignalDataByEntityID(entityID SignalEntityID) (*SignalData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, d := range db.signalData {
		if d.EntityID == entityID {
			dCopy := copySignalData(d)
			return dCopy, nil
		}
	}
	return nil, fmt.Errorf("signal data not found for entity")
}

func (db *MemoryDatabase) HasSignalDataByEntityID(entityID SignalEntityID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, d := range db.signalData {
		if d.EntityID == entityID {
			return true, nil
		}
	}
	return false, nil
}

func (db *MemoryDatabase) GetSignalDataProperties(entityID SignalEntityID, getters ...SignalDataPropertyGetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	var data *SignalData
	for _, d := range db.signalData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}

	if data == nil {
		return fmt.Errorf("signal data not found")
	}

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSignalDataProperties(entityID SignalEntityID, setters ...SignalDataPropertySetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	var data *SignalData
	for _, d := range db.signalData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}
	if data == nil {
		return fmt.Errorf("signal data not found")
	}

	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			return err
		}
	}
	return nil
}

func (db *MemoryDatabase) AddSignalExecutionData(executionID SignalExecutionID, data *SignalExecutionData) (SignalExecutionDataID, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	db.signalExecutionDataCounter++
	data.ID = SignalExecutionDataID(db.signalExecutionDataCounter)
	data.ExecutionID = executionID
	db.signalExecutionData[data.ID] = copySignalExecutionData(data)

	db.signalExecToDataMap[executionID] = data.ID

	return data.ID, nil
}

func (db *MemoryDatabase) GetSignalExecutionData(id SignalExecutionDataID) (*SignalExecutionData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	d, exists := db.signalExecutionData[id]
	if !exists {
		return nil, fmt.Errorf("signal execution data not found")
	}
	dCopy := copySignalExecutionData(d)
	return dCopy, nil
}

func (db *MemoryDatabase) GetSignalExecutionDataByExecutionID(executionID SignalExecutionID) (*SignalExecutionData, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, d := range db.signalExecutionData {
		if d.ExecutionID == executionID {
			dCopy := copySignalExecutionData(d)
			return dCopy, nil
		}
	}
	return nil, fmt.Errorf("signal execution data not found for execution")
}

func (db *MemoryDatabase) HasSignalExecutionDataByExecutionID(executionID SignalExecutionID) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, d := range db.signalExecutionData {
		if d.ExecutionID == executionID {
			return true, nil
		}
	}
	return false, nil
}

func (db *MemoryDatabase) GetSignalExecutionDataProperties(entityID SignalExecutionID, getters ...SignalExecutionDataPropertyGetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	var data *SignalExecutionData
	for _, d := range db.signalExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}

	if data == nil {
		return fmt.Errorf("signal execution data not found")
	}

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSignalExecutionDataProperties(entityID SignalExecutionID, setters ...SignalExecutionDataPropertySetter) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	var data *SignalExecutionData
	for _, d := range db.signalExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}
	if data == nil {
		return fmt.Errorf("signal execution data not found")
	}

	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			return err
		}
	}
	return nil
}

// //////////////
// Saga Data related functions
func (db *MemoryDatabase) GetSagaDataIDByEntityID(entityID SagaEntityID) (SagaDataID, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, data := range db.sagaData {
		if data.EntityID == entityID {
			return data.ID, nil
		}
	}

	return SagaDataID(0), ErrSagaEntityNotFound
}

func (db *MemoryDatabase) GetSagaDataProperties(id SagaDataID, getters ...SagaDataPropertyGetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.sagaData[id]
	if !exists {
		return ErrSagaEntityNotFound
	}

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSagaDataProperties(id SagaDataID, setters ...SagaDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.sagaData[id]
	if !exists {
		return ErrSagaEntityNotFound
	}

	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) GetSagaDataPropertiesByEntityID(entityID SagaEntityID, getters ...SagaDataPropertyGetter) error {
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

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSagaDataPropertiesByEntityID(entityID SagaEntityID, setters ...SagaDataPropertySetter) error {
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
		if _, err := setter(data); err != nil {
			return err
		}
	}

	return nil
}

// Saga Execution Data related functions
func (db *MemoryDatabase) GetSagaExecutionDataIDByEntityID(executionID SagaExecutionID) (SagaExecutionDataID, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	dataID, exists := db.sagaExecToDataMap[executionID]
	if !exists {
		return SagaExecutionDataID(0), ErrSagaExecutionNotFound
	}

	return dataID, nil
}

func (db *MemoryDatabase) GetSagaExecutionDataProperties(id SagaExecutionDataID, getters ...SagaExecutionDataPropertyGetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.sagaExecutionData[id]
	if !exists {
		return ErrSagaExecutionNotFound
	}

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSagaExecutionDataProperties(id SagaExecutionDataID, setters ...SagaExecutionDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.sagaExecutionData[id]
	if !exists {
		return ErrSagaExecutionNotFound
	}

	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) GetSagaExecutionDataPropertiesByEntities(executionID SagaExecutionID, getters ...SagaExecutionDataPropertyGetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	dataID, exists := db.sagaExecToDataMap[executionID]
	if !exists {
		return ErrSagaExecutionNotFound
	}

	data, exists := db.sagaExecutionData[dataID]
	if !exists {
		return ErrSagaExecutionNotFound
	}

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSagaExecutionDataPropertiesByEntities(executionID SagaExecutionID, setters ...SagaExecutionDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	dataID, exists := db.sagaExecToDataMap[executionID]
	if !exists {
		return ErrSagaExecutionNotFound
	}

	data, exists := db.sagaExecutionData[dataID]
	if !exists {
		return ErrSagaExecutionNotFound
	}

	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			return err
		}
	}

	return nil
}

// Side Effect Data related functions
func (db *MemoryDatabase) GetSideEffectDataIDByEntityID(entityID SideEffectEntityID) (SideEffectDataID, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, data := range db.sideEffectData {
		if data.EntityID == entityID {
			return data.ID, nil
		}
	}

	return SideEffectDataID(0), ErrSideEffectEntityNotFound
}

func (db *MemoryDatabase) GetSideEffectDataProperties(id SideEffectDataID, getters ...SideEffectDataPropertyGetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.sideEffectData[id]
	if !exists {
		return ErrSideEffectEntityNotFound
	}

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSideEffectDataProperties(id SideEffectDataID, setters ...SideEffectDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.sideEffectData[id]
	if !exists {
		return ErrSideEffectEntityNotFound
	}

	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) GetSideEffectDataPropertiesByEntityID(entityID SideEffectEntityID, getters ...SideEffectDataPropertyGetter) error {
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

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSideEffectDataPropertiesByEntityID(entityID SideEffectEntityID, setters ...SideEffectDataPropertySetter) error {
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
		if _, err := setter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) GetSideEffectExecutionDataIDByExecutionID(executionID SideEffectEntityID) (SideEffectDataID, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	dataID, exists := db.sideEffectExecToDataMap[SideEffectExecutionID(executionID)]
	if !exists {
		return SideEffectDataID(0), ErrSideEffectExecutionNotFound
	}

	return SideEffectDataID(dataID), nil
}

func (db *MemoryDatabase) GetSideEffectExecutionDataPropertiesByExecutionID(executionID SideEffectExecutionID, getters ...SideEffectExecutionDataPropertyGetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	dataID, exists := db.sideEffectExecToDataMap[executionID]
	if !exists {
		return ErrSideEffectExecutionNotFound
	}

	data, exists := db.sideEffectExecutionData[dataID]
	if !exists {
		return ErrSideEffectExecutionNotFound
	}

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSideEffectExecutionDataPropertiesByExecutionID(executionID SideEffectExecutionID, setters ...SideEffectExecutionDataPropertySetter) error {
	db.mu.Lock()

	// First check if data exists
	var data *SideEffectExecutionData
	var dataID SideEffectExecutionDataID
	var exists bool

	dataID, exists = db.sideEffectExecToDataMap[executionID]
	if exists {
		data = db.sideEffectExecutionData[dataID]
	}

	// If no data exists, create new data structure directly
	if data == nil {
		db.sideEffectExecutionDataCounter++
		dataID = SideEffectExecutionDataID(db.sideEffectExecutionDataCounter)

		data = &SideEffectExecutionData{
			ID:          dataID,
			ExecutionID: executionID,
		}

		db.sideEffectExecutionData[dataID] = data
		db.sideEffectExecToDataMap[executionID] = dataID
	}

	// Apply setters
	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			db.mu.Unlock()
			return err
		}
	}

	db.mu.Unlock()
	return nil
}

////////////////

func (db *MemoryDatabase) UpdateSignalEntity(entity *SignalEntity) error {

	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.signalEntities[entity.ID]; !exists {
		return fmt.Errorf("signal entity not found")
	}

	entity.UpdatedAt = time.Now()
	db.signalEntities[entity.ID] = copySignalEntity(entity)
	return nil
}

func (db *MemoryDatabase) DeleteRuns(ids ...RunID) error {
	if len(ids) == 0 {
		return nil
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	// For each run to delete
	for _, runID := range ids {
		// Skip if run doesn't exist
		if _, exists := db.runs[runID]; !exists {
			continue
		}

		// Clean up Activities by RunID
		for activityID, activity := range db.activityEntities {
			if activity.RunID == runID {
				// Clean up all activity executions
				for execID, exec := range db.activityExecutions {
					if exec.ActivityEntityID == activityID {
						// Delete activity execution data
						if dataID, exists := db.activityExecToDataMap[execID]; exists {
							delete(db.activityExecutionData, dataID)
							delete(db.activityExecToDataMap, execID)
						}
						delete(db.activityExecutions, execID)
					}
				}

				// Clean up activity data
				for dataID, data := range db.activityData {
					if data.EntityID == activityID {
						delete(db.activityData, dataID)
					}
				}

				delete(db.activityEntities, activityID)
			}
		}

		// Clean up Sagas by RunID
		for sagaID, saga := range db.sagaEntities {
			if saga.RunID == runID {
				// Clean up all saga executions
				for execID, exec := range db.sagaExecutions {
					if exec.SagaEntityID == sagaID {
						// Delete saga execution data
						if dataID, exists := db.sagaExecToDataMap[execID]; exists {
							delete(db.sagaExecutionData, dataID)
							delete(db.sagaExecToDataMap, execID)
						}
						// Delete saga values
						delete(db.sagaValues, execID)
						delete(db.sagaExecutions, execID)
					}
				}

				// Clean up all saga data
				for dataID, data := range db.sagaData {
					if data.EntityID == sagaID {
						delete(db.sagaData, dataID)
					}
				}

				// Clean up saga execution data
				for dataID := range db.sagaExecutionData {
					delete(db.sagaExecutionData, dataID)
				}

				delete(db.sagaEntities, sagaID)
			}
		}

		// Clean up SideEffects by RunID
		for sideEffectID, sideEffect := range db.sideEffectEntities {
			if sideEffect.RunID == runID {
				// Clean up all side effect executions
				for execID, exec := range db.sideEffectExecutions {
					if exec.SideEffectEntityID == sideEffectID {
						// Delete side effect execution data
						if dataID, exists := db.sideEffectExecToDataMap[execID]; exists {
							delete(db.sideEffectExecutionData, dataID)
							delete(db.sideEffectExecToDataMap, execID)
						}
						delete(db.sideEffectExecutions, execID)
					}
				}

				// Clean up side effect data
				for dataID, data := range db.sideEffectData {
					if data.EntityID == sideEffectID {
						delete(db.sideEffectData, dataID)
					}
				}

				// Clean up side effect execution data
				for dataID := range db.sideEffectExecutionData {
					delete(db.sideEffectExecutionData, dataID)
				}

				delete(db.sideEffectEntities, sideEffectID)
			}
		}

		// Clean up Signals by RunID
		for signalID, signal := range db.signalEntities {
			if signal.RunID == runID {
				// Clean up all signal executions
				for execID, exec := range db.signalExecutions {
					if exec.EntityID == signalID {
						// Delete signal execution data
						if dataID, exists := db.signalExecToDataMap[execID]; exists {
							delete(db.signalExecutionData, dataID)
							delete(db.signalExecToDataMap, execID)
						}
						delete(db.signalExecutions, execID)
					}
				}

				// Clean up signal data
				for dataID, data := range db.signalData {
					if data.EntityID == signalID {
						delete(db.signalData, dataID)
					}
				}

				delete(db.signalEntities, signalID)
			}
		}

		// Clean up Workflows by RunID
		for wfID, wf := range db.workflowEntities {
			if wf.RunID == runID {
				// Clean up all workflow executions
				for execID, exec := range db.workflowExecutions {
					if exec.WorkflowEntityID == wfID {
						// Delete workflow execution data
						if dataID, exists := db.workflowExecToDataMap[execID]; exists {
							delete(db.workflowExecutionData, dataID)
							delete(db.workflowExecToDataMap, execID)
						}
						delete(db.workflowExecutions, execID)
					}
				}

				// Clean up workflow data
				for dataID, data := range db.workflowData {
					if data.EntityID == wfID {
						delete(db.workflowData, dataID)
					}
				}

				// Clean up workflow versions
				if versions, ok := db.workflowToVersion[wfID]; ok {
					for _, versionID := range versions {
						delete(db.versions, versionID)
					}
				}
				delete(db.workflowToVersion, wfID)
				delete(db.workflowVersions, wfID)

				// Clean up queue associations
				if queueID, ok := db.workflowToQueue[wfID]; ok {
					if workflows, exists := db.queueToWorkflows[queueID]; exists {
						newWorkflows := make([]WorkflowEntityID, 0)
						for _, id := range workflows {
							if id != wfID {
								newWorkflows = append(newWorkflows, id)
							}
						}
						if len(newWorkflows) > 0 {
							db.queueToWorkflows[queueID] = newWorkflows
						} else {
							delete(db.queueToWorkflows, queueID)
						}
					}
					delete(db.workflowToQueue, wfID)
				}

				// Clean up workflow children mappings
				delete(db.workflowToChildren, wfID)

				delete(db.workflowEntities, wfID)
			}
		}

		// Clean up entity to workflow mappings
		for entityID, workflowID := range db.entityToWorkflow {
			if _, exists := db.workflowEntities[workflowID]; !exists {
				delete(db.entityToWorkflow, entityID)
			}
		}

		// Clean up hierarchies
		for id, h := range db.hierarchies {
			if h.RunID == runID {
				delete(db.hierarchies, id)
			}
		}

		// Clean up run mappings
		delete(db.runToWorkflows, runID)
		delete(db.runs, runID)
	}

	return nil
}

func (db *MemoryDatabase) DeleteRunsByStatus(status RunStatus) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Collect all runs with matching status
	runsToDelete := make([]RunID, 0)
	for _, run := range db.runs {
		if run.Status == status {
			runsToDelete = append(runsToDelete, run.ID)
		}
	}

	if len(runsToDelete) == 0 {
		return nil
	}

	// For each run to delete
	for _, runID := range runsToDelete {
		// Clean up Activities by RunID
		for activityID, activity := range db.activityEntities {
			if activity.RunID == runID {
				// Clean up all activity executions
				for execID, exec := range db.activityExecutions {
					if exec.ActivityEntityID == activityID {
						// Delete activity execution data
						if dataID, exists := db.activityExecToDataMap[execID]; exists {
							delete(db.activityExecutionData, dataID)
							delete(db.activityExecToDataMap, execID)
						}
						delete(db.activityExecutions, execID)
					}
				}

				// Clean up activity data
				for dataID, data := range db.activityData {
					if data.EntityID == activityID {
						delete(db.activityData, dataID)
					}
				}

				delete(db.activityEntities, activityID)
			}
		}

		// Clean up Sagas by RunID
		for sagaID, saga := range db.sagaEntities {
			if saga.RunID == runID {
				// Clean up all saga executions
				for execID, exec := range db.sagaExecutions {
					if exec.SagaEntityID == sagaID {
						// Delete saga execution data
						if dataID, exists := db.sagaExecToDataMap[execID]; exists {
							delete(db.sagaExecutionData, dataID)
							delete(db.sagaExecToDataMap, execID)
						}
						// Delete saga values
						delete(db.sagaValues, execID)
						delete(db.sagaExecutions, execID)
					}
				}

				// Clean up all saga data
				for dataID, data := range db.sagaData {
					if data.EntityID == sagaID {
						delete(db.sagaData, dataID)
					}
				}

				// Clean up saga execution data
				for dataID := range db.sagaExecutionData {
					delete(db.sagaExecutionData, dataID)
				}

				delete(db.sagaEntities, sagaID)
			}
		}

		// Clean up SideEffects by RunID
		for sideEffectID, sideEffect := range db.sideEffectEntities {
			if sideEffect.RunID == runID {
				// Clean up all side effect executions
				for execID, exec := range db.sideEffectExecutions {
					if exec.SideEffectEntityID == sideEffectID {
						// Delete side effect execution data
						if dataID, exists := db.sideEffectExecToDataMap[execID]; exists {
							delete(db.sideEffectExecutionData, dataID)
							delete(db.sideEffectExecToDataMap, execID)
						}
						delete(db.sideEffectExecutions, execID)
					}
				}

				// Clean up side effect data
				for dataID, data := range db.sideEffectData {
					if data.EntityID == sideEffectID {
						delete(db.sideEffectData, dataID)
					}
				}

				// Clean up side effect execution data
				for dataID := range db.sideEffectExecutionData {
					delete(db.sideEffectExecutionData, dataID)
				}

				delete(db.sideEffectEntities, sideEffectID)
			}
		}

		// Clean up Signals by RunID
		for signalID, signal := range db.signalEntities {
			if signal.RunID == runID {
				// Clean up all signal executions
				for execID, exec := range db.signalExecutions {
					if exec.EntityID == signalID {
						// Delete signal execution data
						if dataID, exists := db.signalExecToDataMap[execID]; exists {
							delete(db.signalExecutionData, dataID)
							delete(db.signalExecToDataMap, execID)
						}
						delete(db.signalExecutions, execID)
					}
				}

				// Clean up signal data
				for dataID, data := range db.signalData {
					if data.EntityID == signalID {
						delete(db.signalData, dataID)
					}
				}

				delete(db.signalEntities, signalID)
			}
		}

		// Clean up Workflows by RunID
		for wfID, wf := range db.workflowEntities {
			if wf.RunID == runID {
				// Clean up all workflow executions
				for execID, exec := range db.workflowExecutions {
					if exec.WorkflowEntityID == wfID {
						// Delete workflow execution data
						if dataID, exists := db.workflowExecToDataMap[execID]; exists {
							delete(db.workflowExecutionData, dataID)
							delete(db.workflowExecToDataMap, execID)
						}
						delete(db.workflowExecutions, execID)
					}
				}

				// Clean up workflow data
				for dataID, data := range db.workflowData {
					if data.EntityID == wfID {
						delete(db.workflowData, dataID)
					}
				}

				// Clean up workflow versions
				if versions, ok := db.workflowToVersion[wfID]; ok {
					for _, versionID := range versions {
						delete(db.versions, versionID)
					}
				}
				delete(db.workflowToVersion, wfID)
				delete(db.workflowVersions, wfID)

				// Clean up queue associations
				if queueID, ok := db.workflowToQueue[wfID]; ok {
					if workflows, exists := db.queueToWorkflows[queueID]; exists {
						newWorkflows := make([]WorkflowEntityID, 0)
						for _, id := range workflows {
							if id != wfID {
								newWorkflows = append(newWorkflows, id)
							}
						}
						if len(newWorkflows) > 0 {
							db.queueToWorkflows[queueID] = newWorkflows
						} else {
							delete(db.queueToWorkflows, queueID)
						}
					}
					delete(db.workflowToQueue, wfID)
				}

				// Clean up workflow children mappings
				delete(db.workflowToChildren, wfID)

				delete(db.workflowEntities, wfID)
			}
		}

		// Clean up entity to workflow mappings
		for entityID, workflowID := range db.entityToWorkflow {
			if _, exists := db.workflowEntities[workflowID]; !exists {
				delete(db.entityToWorkflow, entityID)
			}
		}

		// Clean up hierarchies
		for id, h := range db.hierarchies {
			if h.RunID == runID {
				delete(db.hierarchies, id)
			}
		}

		// Clean up run mappings
		delete(db.runToWorkflows, runID)
		delete(db.runs, runID)
	}

	return nil
}

func (db *MemoryDatabase) deleteWorkflowAndChildren(workflowID WorkflowEntityID) error {

	var hierarchies []*Hierarchy
	for _, h := range db.hierarchies {
		if h.ParentEntityID == int(workflowID) {
			hierarchies = append(hierarchies, copyHierarchy(h))
		}
	}

	for _, h := range hierarchies {
		switch h.ChildType {
		case EntityActivity:
			if err := db.deleteActivityEntity(ActivityEntityID(h.ChildEntityID)); err != nil {
				return err
			}
		case EntitySaga:
			if err := db.deleteSagaEntity(SagaEntityID(h.ChildEntityID)); err != nil {
				return err
			}
		case EntitySideEffect:
			if err := db.deleteSideEffectEntity(SideEffectEntityID(h.ChildEntityID)); err != nil {
				return err
			}
		case EntitySignal:
			if err := db.deleteSignalEntity(SignalEntityID(h.ChildEntityID)); err != nil {
				return err
			}
		}
	}

	if err := db.deleteWorkflowEntity(workflowID); err != nil {
		return err
	}

	for id, hh := range db.hierarchies {
		if hh.ParentEntityID == int(workflowID) || hh.ChildEntityID == int(workflowID) {
			delete(db.hierarchies, id)
		}
	}

	return nil
}

func (db *MemoryDatabase) deleteWorkflowEntity(workflowID WorkflowEntityID) error {

	for id, exec := range db.workflowExecutions {
		if exec.WorkflowEntityID == workflowID {
			if dataID, exists := db.workflowExecToDataMap[id]; exists {
				delete(db.workflowExecutionData, dataID)
				delete(db.workflowExecToDataMap, id)
			}
		}
	}

	for id, data := range db.workflowData {
		if data.EntityID == workflowID {
			delete(db.workflowData, id)
		}
	}

	versionIDs := db.workflowToVersion[workflowID]
	delete(db.workflowToVersion, workflowID)
	delete(db.workflowVersions, workflowID)

	for _, vID := range versionIDs {
		delete(db.versions, vID)
	}

	if queueID, ok := db.workflowToQueue[workflowID]; ok {
		if workflows, exists := db.queueToWorkflows[queueID]; exists {
			newWorkflows := make([]WorkflowEntityID, 0, len(workflows)-1)
			for _, wID := range workflows {
				if wID != workflowID {
					newWorkflows = append(newWorkflows, wID)
				}
			}
			db.queueToWorkflows[queueID] = newWorkflows
		}
		delete(db.workflowToQueue, workflowID)
	}

	delete(db.workflowEntities, workflowID)

	return nil
}

func (db *MemoryDatabase) deleteActivityEntity(entityID ActivityEntityID) error {

	for id, exec := range db.activityExecutions {
		if exec.ActivityEntityID == entityID {
			if dataID, exists := db.activityExecToDataMap[id]; exists {
				delete(db.activityExecutionData, dataID)
				delete(db.activityExecToDataMap, id)
			}
			delete(db.activityExecutions, id)
		}
	}

	for id, data := range db.activityData {
		if data.EntityID == entityID {
			delete(db.activityData, id)
		}
	}

	delete(db.activityEntities, entityID)

	return nil
}

func (db *MemoryDatabase) deleteSagaEntity(entityID SagaEntityID) error {

	for id, exec := range db.sagaExecutions {
		if exec.SagaEntityID == entityID {
			if dataID, exists := db.sagaExecToDataMap[id]; exists {
				delete(db.sagaExecutionData, dataID)
				delete(db.sagaExecToDataMap, id)
			}

			delete(db.sagaValues, exec.ID)

			delete(db.sagaExecutions, id)
		}
	}

	for id, data := range db.sagaData {
		if data.EntityID == entityID {
			delete(db.sagaData, id)
		}
	}

	delete(db.sagaEntities, entityID)

	return nil
}

func (db *MemoryDatabase) deleteSideEffectEntity(entityID SideEffectEntityID) error {

	for id, exec := range db.sideEffectExecutions {
		if exec.SideEffectEntityID == entityID {
			if dataID, exists := db.sideEffectExecToDataMap[id]; exists {
				delete(db.sideEffectExecutionData, dataID)
				delete(db.sideEffectExecToDataMap, id)
			}
			delete(db.sideEffectExecutions, id)
		}
	}

	for id, data := range db.sideEffectData {
		if data.EntityID == entityID {
			delete(db.sideEffectData, id)
		}
	}

	delete(db.sideEffectEntities, entityID)

	return nil
}

func (db *MemoryDatabase) deleteSignalEntity(entityID SignalEntityID) error {

	db.mu.Lock()
	defer db.mu.Unlock()
	for id, exec := range db.signalExecutions {
		if exec.EntityID == entityID {
			if dataID, exists := db.signalExecToDataMap[id]; exists {
				delete(db.signalExecutionData, dataID)
				delete(db.signalExecToDataMap, id)
			}
			delete(db.signalExecutions, id)
		}
	}

	for id, data := range db.signalData {
		if data.EntityID == entityID {
			delete(db.signalData, id)
		}
	}

	delete(db.signalEntities, entityID)

	return nil
}

func (db *MemoryDatabase) GetRunsPaginated(page, pageSize int, filter *RunFilter, sortCriteria *RunSort) (*PaginatedRuns, error) {

	db.mu.Lock()
	defer db.mu.Unlock()
	matchingRuns := make([]*Run, 0)
	for _, run := range db.runs {
		if filter != nil {
			if run.Status != filter.Status {
				continue
			}
		}
		matchingRuns = append(matchingRuns, copyRun(run))
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	if sortCriteria != nil {
		sortRuns(matchingRuns, sortCriteria)
	}

	totalRuns := len(matchingRuns)
	totalPages := (totalRuns + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > totalRuns {
		end = totalRuns
	}
	var pageRuns []*Run
	if start < totalRuns {
		pageRuns = matchingRuns[start:end]
	} else {
		pageRuns = []*Run{}
	}

	return &PaginatedRuns{
		Runs:       pageRuns,
		TotalRuns:  totalRuns,
		TotalPages: totalPages,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func sortRuns(runs []*Run, sort *RunSort) {
	if sort == nil {
		return
	}

	sort.Field = strings.ToLower(sort.Field)

	sortFn := func(i, j *Run) int {
		var comparison int

		switch sort.Field {
		case "id":
			comparison = cmp.Compare(i.ID, j.ID)
		case "created_at":
			comparison = i.CreatedAt.Compare(j.CreatedAt)
		case "updated_at":
			comparison = i.UpdatedAt.Compare(j.UpdatedAt)
		case "status":
			comparison = strings.Compare(string(i.Status), string(j.Status))
		default:
			// Default sort by ID
			comparison = cmp.Compare(i.ID, j.ID)
		}

		if sort.Desc {
			return -comparison
		}
		return comparison
	}

	slices.SortFunc(runs, sortFn)
}

// Copy functions
func copyRun(run *Run) *Run {
	if run == nil {
		return nil
	}

	copy := &Run{
		ID:        run.ID,
		Status:    run.Status,
		CreatedAt: run.CreatedAt,
		UpdatedAt: run.UpdatedAt,
	}

	if run.Entities != nil {
		copy.Entities = make([]*WorkflowEntity, len(run.Entities))
		for i, entity := range run.Entities {
			copy.Entities[i] = copyWorkflowEntity(entity)
		}
	}

	if run.Hierarchies != nil {
		copy.Hierarchies = make([]*Hierarchy, len(run.Hierarchies))
		for i, hierarchy := range run.Hierarchies {
			copy.Hierarchies[i] = copyHierarchy(hierarchy)
		}
	}

	return copy
}

// func copyVersion(version *Version) *Version {
// 	if version == nil {
// 		return nil
// 	}

// 	copy := &Version{
// 		ID:       version.ID,
// 		EntityID: version.EntityID,
// 		ChangeID: version.ChangeID,
// 		Version:  version.Version,
// 	}

// 	if version.Data != nil {
// 		copy.Data = make(map[string]interface{}, len(version.Data))
// 		for k, v := range version.Data {
// 			copy.Data[k] = v
// 		}
// 	}

//		return copy
//	}
func copyVersion(v *Version) *Version {
	if v == nil {
		return nil
	}
	copied := *v
	if v.Data != nil {
		copied.Data = make(map[string]interface{})
		for k, v := range v.Data {
			copied.Data[k] = v
		}
	}
	return &copied
}

func copyHierarchy(hierarchy *Hierarchy) *Hierarchy {
	if hierarchy == nil {
		return nil
	}

	return &Hierarchy{
		ID:                hierarchy.ID,
		RunID:             hierarchy.RunID,
		ParentEntityID:    hierarchy.ParentEntityID,
		ChildEntityID:     hierarchy.ChildEntityID,
		ParentExecutionID: hierarchy.ParentExecutionID,
		ChildExecutionID:  hierarchy.ChildExecutionID,
		ParentStepID:      hierarchy.ParentStepID,
		ChildStepID:       hierarchy.ChildStepID,
		ParentType:        hierarchy.ParentType,
		ChildType:         hierarchy.ChildType,
	}
}

func copyQueue(queue *Queue) *Queue {
	if queue == nil {
		return nil
	}

	copy := &Queue{
		ID:        queue.ID,
		Name:      queue.Name,
		CreatedAt: queue.CreatedAt,
		UpdatedAt: queue.UpdatedAt,
	}

	if queue.Entities != nil {
		copy.Entities = make([]*WorkflowEntity, len(queue.Entities))
		for i, entity := range queue.Entities {
			copy.Entities[i] = copyWorkflowEntity(entity)
		}
	}

	return copy
}

func copyBaseEntity(base *BaseEntity) *BaseEntity {
	if base == nil {
		return nil
	}

	return &BaseEntity{
		HandlerName: base.HandlerName,
		Type:        base.Type,
		Status:      base.Status,
		QueueID:     base.QueueID,
		StepID:      base.StepID,
		CreatedAt:   base.CreatedAt,
		UpdatedAt:   base.UpdatedAt,
		RunID:       base.RunID,
		RetryPolicy: base.RetryPolicy,
		RetryState:  base.RetryState,
	}
}

func copyBaseExecution(base *BaseExecution) *BaseExecution {
	if base == nil {
		return nil
	}

	copy := *base

	if base.CompletedAt != nil {
		completedAtCopy := *base.CompletedAt
		copy.CompletedAt = &completedAtCopy
	}

	return &copy
}

// Entity Data copy functions
func copyWorkflowData(data *WorkflowData) *WorkflowData {
	if data == nil {
		return nil
	}

	c := *data

	// Deep copy versions map
	if data.Versions != nil {
		c.Versions = make(map[string]int)
		for k, v := range data.Versions {
			c.Versions[k] = v
		}
	}

	if data.Inputs != nil {
		c.Inputs = make([][]byte, len(data.Inputs))
		for i, input := range data.Inputs {
			inputCopy := make([]byte, len(input))
			copy(inputCopy, input)
			c.Inputs[i] = inputCopy
		}
	}

	return &c
}

func copyWorkflowEntityEdges(e *WorkflowEntityEdges) *WorkflowEntityEdges {
	if e == nil {
		return nil
	}
	copy := *e
	if e.Queue != nil {
		copy.Queue = copyQueue(e.Queue)
	}
	if e.Versions != nil {
		copy.Versions = make([]*Version, len(e.Versions))
		for i, v := range e.Versions {
			copy.Versions[i] = copyVersion(v)
		}
	}
	if e.ActivityChildren != nil {
		copy.ActivityChildren = make([]*ActivityEntity, len(e.ActivityChildren))
		for i, a := range e.ActivityChildren {
			copy.ActivityChildren[i] = copyActivityEntity(a)
		}
	}
	if e.SagaChildren != nil {
		copy.SagaChildren = make([]*SagaEntity, len(e.SagaChildren))
		for i, s := range e.SagaChildren {
			copy.SagaChildren[i] = copySagaEntity(s)
		}
	}
	if e.SideEffectChildren != nil {
		copy.SideEffectChildren = make([]*SideEffectEntity, len(e.SideEffectChildren))
		for i, se := range e.SideEffectChildren {
			copy.SideEffectChildren[i] = copySideEffectEntity(se)
		}
	}
	return &copy
}

func copyActivityData(data *ActivityData) *ActivityData {
	if data == nil {
		return nil
	}

	c := *data

	if data.Inputs != nil {
		c.Inputs = make([][]byte, len(data.Inputs))
		for i, input := range data.Inputs {
			inputCopy := make([]byte, len(input))
			copy(inputCopy, input)
			c.Inputs[i] = inputCopy
		}
	}

	if data.Output != nil {
		c.Output = make([][]byte, len(data.Output))
		for i, output := range data.Output {
			outputCopy := make([]byte, len(output))
			copy(outputCopy, output)
			c.Output[i] = outputCopy
		}
	}

	return &c
}

func copySagaData(data *SagaData) *SagaData {
	if data == nil {
		return nil
	}
	return &SagaData{
		ID:       data.ID,
		EntityID: data.EntityID,
	}
}

func copySagaExecutionData(data *SagaExecutionData) *SagaExecutionData {
	if data == nil {
		return nil
	}

	c := &SagaExecutionData{
		ID:          data.ID,
		ExecutionID: data.ExecutionID,
		StepIndex:   data.StepIndex,
	}

	if data.LastHeartbeat != nil {
		heartbeatCopy := *data.LastHeartbeat
		c.LastHeartbeat = &heartbeatCopy
	}

	return c
}

func copySagaExecution(exec *SagaExecution) *SagaExecution {
	if exec == nil {
		return nil
	}

	return &SagaExecution{
		BaseExecution:     *copyBaseExecution(&exec.BaseExecution),
		ExecutionType:     exec.ExecutionType,
		ID:                exec.ID,
		SagaEntityID:      exec.SagaEntityID,
		SagaExecutionData: copySagaExecutionData(exec.SagaExecutionData),
	}
}

func copySagaEntity(entity *SagaEntity) *SagaEntity {
	if entity == nil {
		return nil
	}

	return &SagaEntity{
		BaseEntity: *copyBaseEntity(&entity.BaseEntity),
		ID:         entity.ID,
		SagaData:   copySagaData(entity.SagaData),
	}
}

func copySideEffectData(data *SideEffectData) *SideEffectData {
	if data == nil {
		return nil
	}
	copy := *data
	return &copy
}

// Execution Data copy functions
func copyWorkflowExecutionData(data *WorkflowExecutionData) *WorkflowExecutionData {
	if data == nil {
		return nil
	}

	c := *data

	if data.LastHeartbeat != nil {
		heartbeatCopy := *data.LastHeartbeat
		c.LastHeartbeat = &heartbeatCopy
	}

	if data.Outputs != nil {
		c.Outputs = make([][]byte, len(data.Outputs))
		for i, output := range data.Outputs {
			outputCopy := make([]byte, len(output))
			copy(outputCopy, output)
			c.Outputs[i] = outputCopy
		}
	}

	return &c
}

func copyActivityExecutionData(data *ActivityExecutionData) *ActivityExecutionData {
	if data == nil {
		return nil
	}

	c := *data

	if data.LastHeartbeat != nil {
		heartbeatCopy := *data.LastHeartbeat
		c.LastHeartbeat = &heartbeatCopy
	}

	if data.Outputs != nil {
		c.Outputs = make([][]byte, len(data.Outputs))
		for i, output := range data.Outputs {
			outputCopy := make([]byte, len(output))
			copy(outputCopy, output)
			c.Outputs[i] = outputCopy
		}
	}

	return &c
}

func copySideEffectExecutionData(data *SideEffectExecutionData) *SideEffectExecutionData {
	if data == nil {
		return nil
	}

	c := *data

	if data.Outputs != nil {
		c.Outputs = make([][]byte, len(data.Outputs))
		for i, output := range data.Outputs {
			outputCopy := make([]byte, len(output))
			copy(outputCopy, output)
			c.Outputs[i] = outputCopy
		}
	}

	return &c
}

// Entity copy functions
func copyWorkflowEntity(entity *WorkflowEntity) *WorkflowEntity {
	if entity == nil {
		return nil
	}

	copy := WorkflowEntity{
		ID:           entity.ID,
		BaseEntity:   *copyBaseEntity(&entity.BaseEntity),
		WorkflowData: copyWorkflowData(entity.WorkflowData),
		Edges:        copyWorkflowEntityEdges(entity.Edges),
	}

	return &copy
}

func copyActivityEntity(entity *ActivityEntity) *ActivityEntity {
	if entity == nil {
		return nil
	}

	copy := ActivityEntity{
		ID:           entity.ID,
		BaseEntity:   *copyBaseEntity(&entity.BaseEntity),
		ActivityData: copyActivityData(entity.ActivityData),
	}

	return &copy
}

func copySideEffectEntity(entity *SideEffectEntity) *SideEffectEntity {
	if entity == nil {
		return nil
	}

	copy := SideEffectEntity{
		ID:             entity.ID,
		BaseEntity:     *copyBaseEntity(&entity.BaseEntity),
		SideEffectData: copySideEffectData(entity.SideEffectData),
	}

	return &copy
}

// Execution copy functions
func copyWorkflowExecution(exec *WorkflowExecution) *WorkflowExecution {
	if exec == nil {
		return nil
	}

	copy := WorkflowExecution{
		ID:                    exec.ID,
		WorkflowEntityID:      exec.WorkflowEntityID,
		BaseExecution:         *copyBaseExecution(&exec.BaseExecution),
		WorkflowExecutionData: copyWorkflowExecutionData(exec.WorkflowExecutionData),
	}

	return &copy
}

func copyActivityExecution(exec *ActivityExecution) *ActivityExecution {
	if exec == nil {
		return nil
	}

	copy := ActivityExecution{
		ID:                    exec.ID,
		ActivityEntityID:      exec.ActivityEntityID,
		BaseExecution:         *copyBaseExecution(&exec.BaseExecution),
		ActivityExecutionData: copyActivityExecutionData(exec.ActivityExecutionData),
	}

	return &copy
}

func copySideEffectExecution(exec *SideEffectExecution) *SideEffectExecution {
	if exec == nil {
		return nil
	}

	copy := SideEffectExecution{
		ID:                      exec.ID,
		SideEffectEntityID:      exec.SideEffectEntityID,
		BaseExecution:           *copyBaseExecution(&exec.BaseExecution),
		SideEffectExecutionData: copySideEffectExecutionData(exec.SideEffectExecutionData),
	}

	return &copy
}

func copySignalEntity(entity *SignalEntity) *SignalEntity {
	if entity == nil {
		return nil
	}
	copy := *entity

	if entity.SignalData != nil {
		copy.SignalData = copySignalData(entity.SignalData)
	}

	return &copy
}

func copySignalData(data *SignalData) *SignalData {
	if data == nil {
		return nil
	}
	copy := *data
	return &copy
}

func copySignalExecution(exec *SignalExecution) *SignalExecution {
	if exec == nil {
		return nil
	}
	copy := SignalExecution{
		BaseExecution: *copyBaseExecution(&exec.BaseExecution),
	}

	if exec.SignalExecutionData != nil {
		copy.SignalExecutionData = copySignalExecutionData(exec.SignalExecutionData)
	}

	return &copy
}

func copySignalExecutionData(data *SignalExecutionData) *SignalExecutionData {
	if data == nil {
		return nil
	}
	c := *data

	if data.Value != nil {
		c.Value = make([]byte, len(data.Value))
		copy(c.Value, data.Value)
	}

	return &c
}
