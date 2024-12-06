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
	sagaValues map[int]map[string]*SagaValue // executionID -> key -> value

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

	// Signal maps
	signalEntities      map[int]*SignalEntity
	signalExecutions    map[int]*SignalExecution
	signalData          map[int]*SignalData
	signalExecutionData map[int]*SignalExecutionData

	// Relationship maps
	entityToWorkflow   map[int]int                  // entity ID -> workflow ID
	workflowToChildren map[int]map[EntityType][]int // workflow ID -> type -> child IDs
	workflowToVersion  map[int][]int                // workflow ID -> version IDs
	workflowToQueue    map[int]int                  // workflow ID -> queue ID
	queueToWorkflows   map[int][]int                // queue ID -> workflow IDs
	runToWorkflows     map[int][]int                // run ID -> workflow IDs
	workflowVersions   map[int][]int                // WorkflowID -> []VersionID
	execToDataMap      map[int]int                  // execution ID -> data ID

	// Locks for each category of data
	runLock            deadlock.RWMutex
	versionLock        deadlock.RWMutex
	hierarchyLock      deadlock.RWMutex
	queueLock          deadlock.RWMutex
	workflowLock       deadlock.RWMutex
	activityLock       deadlock.RWMutex
	sagaLock           deadlock.RWMutex
	sideEffectLock     deadlock.RWMutex
	signalLock         deadlock.RWMutex
	workflowDataLock   deadlock.RWMutex
	activityDataLock   deadlock.RWMutex
	sagaDataLock       deadlock.RWMutex
	sideEffectDataLock deadlock.RWMutex
	signalDataLock     deadlock.RWMutex
	workflowExecLock   deadlock.RWMutex
	activityExecLock   deadlock.RWMutex
	sagaExecLock       deadlock.RWMutex
	sideEffectExecLock deadlock.RWMutex
	signalExecLock     deadlock.RWMutex
	relationshipLock   deadlock.RWMutex

	workflowExecDataLock   deadlock.RWMutex
	activityExecDataLock   deadlock.RWMutex
	sagaExecDataLock       deadlock.RWMutex
	sideEffectExecDataLock deadlock.RWMutex
	signalExecDataLock     deadlock.RWMutex
}

// NewMemoryDatabase initializes a new memory database with default queue.
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

		// Signal maps
		signalEntities:      make(map[int]*SignalEntity),
		signalExecutions:    make(map[int]*SignalExecution),
		signalData:          make(map[int]*SignalData),
		signalExecutionData: make(map[int]*SignalExecutionData),

		// Saga context
		sagaValues: make(map[int]map[string]*SagaValue),

		// Relationship maps
		entityToWorkflow:   make(map[int]int),
		workflowToChildren: make(map[int]map[EntityType][]int),
		workflowToVersion:  make(map[int][]int),
		workflowToQueue:    make(map[int]int),
		queueToWorkflows:   make(map[int][]int),
		runToWorkflows:     make(map[int][]int),
		workflowVersions:   make(map[int][]int),
		execToDataMap:      make(map[int]int),

		queueCounter: 1,
	}

	// Initialize default queue
	now := time.Now()
	db.queueLock.Lock()
	db.queues[1] = &Queue{
		ID:        1,
		Name:      DefaultQueue,
		CreatedAt: now,
		UpdatedAt: now,
		Entities:  make([]*WorkflowEntity, 0),
	}
	db.queueNames[DefaultQueue] = 1
	db.queueLock.Unlock()

	return db
}

func (db *MemoryDatabase) SaveAsJSON(path string) error {
	// We need to lock all relevant structures for a consistent snapshot.
	// We'll lock everything in alphabetical order to avoid deadlocks.
	db.activityDataLock.RLock()
	db.activityExecLock.RLock()
	db.activityLock.RLock()
	db.hierarchyLock.RLock()
	db.queueLock.RLock()
	db.relationshipLock.RLock()
	db.runLock.RLock()
	db.sagaDataLock.RLock()
	db.sagaExecLock.RLock()
	db.sagaLock.RLock()
	db.sideEffectDataLock.RLock()
	db.sideEffectExecLock.RLock()
	db.sideEffectLock.RLock()
	db.signalDataLock.RLock()
	db.signalExecLock.RLock()
	db.signalLock.RLock()
	db.versionLock.RLock()
	db.workflowDataLock.RLock()
	db.workflowExecLock.RLock()
	db.workflowLock.RLock()

	defer db.workflowLock.RUnlock()
	defer db.workflowExecLock.RUnlock()
	defer db.workflowDataLock.RUnlock()
	defer db.versionLock.RUnlock()
	defer db.signalLock.RUnlock()
	defer db.signalExecLock.RUnlock()
	defer db.signalDataLock.RUnlock()
	defer db.sideEffectLock.RUnlock()
	defer db.sideEffectExecLock.RUnlock()
	defer db.sideEffectDataLock.RUnlock()
	defer db.sagaLock.RUnlock()
	defer db.sagaExecLock.RUnlock()
	defer db.sagaDataLock.RUnlock()
	defer db.runLock.RUnlock()
	defer db.relationshipLock.RUnlock()
	defer db.queueLock.RUnlock()
	defer db.hierarchyLock.RUnlock()
	defer db.activityLock.RUnlock()
	defer db.activityExecLock.RUnlock()
	defer db.activityDataLock.RUnlock()

	data := struct {
		Runs                    map[int]*Run
		Versions                map[int]*Version
		Hierarchies             map[int]*Hierarchy
		Queues                  map[int]*Queue
		QueueNames              map[string]int
		WorkflowEntities        map[int]*WorkflowEntity
		ActivityEntities        map[int]*ActivityEntity
		SagaEntities            map[int]*SagaEntity
		SideEffectEntities      map[int]*SideEffectEntity
		WorkflowData            map[int]*WorkflowData
		ActivityData            map[int]*ActivityData
		SagaData                map[int]*SagaData
		SideEffectData          map[int]*SideEffectData
		WorkflowExecutions      map[int]*WorkflowExecution
		ActivityExecutions      map[int]*ActivityExecution
		SagaExecutions          map[int]*SagaExecution
		SideEffectExecutions    map[int]*SideEffectExecution
		WorkflowExecutionData   map[int]*WorkflowExecutionData
		ActivityExecutionData   map[int]*ActivityExecutionData
		SagaExecutionData       map[int]*SagaExecutionData
		SideEffectExecutionData map[int]*SideEffectExecutionData
		EntityToWorkflow        map[int]int
		WorkflowToChildren      map[int]map[EntityType][]int
		WorkflowToVersion       map[int][]int
		WorkflowToQueue         map[int]int
		QueueToWorkflows        map[int][]int
		RunToWorkflows          map[int][]int
		SagaValues              map[int]map[string]*SagaValue
		SignalEntities          map[int]*SignalEntity
		SignalExecutions        map[int]*SignalExecution
		SignalData              map[int]*SignalData
		SignalExecutionData     map[int]*SignalExecutionData
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
func (db *MemoryDatabase) AddRun(run *Run) (int, error) {
	db.runLock.Lock()
	defer db.runLock.Unlock()

	db.runCounter++
	run.ID = db.runCounter
	run.CreatedAt = time.Now()
	run.UpdatedAt = run.CreatedAt

	db.runs[run.ID] = copyRun(run)
	return run.ID, nil
}

// GetRun
func (db *MemoryDatabase) GetRun(id int, opts ...RunGetOption) (*Run, error) {
	db.runLock.RLock()
	r, exists := db.runs[id]
	if !exists {
		db.runLock.RUnlock()
		return nil, ErrRunNotFound
	}
	runCopy := copyRun(r) // Copy to avoid race after unlocking
	db.runLock.RUnlock()

	cfg := &RunGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeWorkflows {
		db.relationshipLock.RLock()
		workflowIDs := db.runToWorkflows[id]
		db.relationshipLock.RUnlock()

		if len(workflowIDs) > 0 {
			entities := make([]*WorkflowEntity, 0, len(workflowIDs))
			db.workflowLock.RLock()
			for _, wfID := range workflowIDs {
				if wf, ok := db.workflowEntities[wfID]; ok {
					entities = append(entities, copyWorkflowEntity(wf))
				}
			}
			db.workflowLock.RUnlock()
			runCopy.Entities = entities
		}
	}

	if cfg.IncludeHierarchies {
		db.hierarchyLock.RLock()
		hierarchies := make([]*Hierarchy, 0)
		for _, h := range db.hierarchies {
			if h.RunID == id {
				hierarchies = append(hierarchies, copyHierarchy(h))
			}
		}
		db.hierarchyLock.RUnlock()
		runCopy.Hierarchies = hierarchies
	}

	return runCopy, nil
}

func (db *MemoryDatabase) UpdateRun(run *Run) error {
	db.runLock.Lock()
	defer db.runLock.Unlock()

	if _, exists := db.runs[run.ID]; !exists {
		return ErrRunNotFound
	}

	run.UpdatedAt = time.Now()
	db.runs[run.ID] = copyRun(run)
	return nil
}

func (db *MemoryDatabase) GetRunProperties(id int, getters ...RunPropertyGetter) error {
	db.runLock.RLock()
	run, exists := db.runs[id]
	if !exists {
		db.runLock.RUnlock()
		return ErrRunNotFound
	}
	runCopy := copyRun(run)
	db.runLock.RUnlock()

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
		db.relationshipLock.RLock()
		workflowIDs := db.runToWorkflows[id]
		db.relationshipLock.RUnlock()

		if len(workflowIDs) > 0 {
			entities := make([]*WorkflowEntity, 0, len(workflowIDs))
			db.workflowLock.RLock()
			for _, wfID := range workflowIDs {
				if wf, ok := db.workflowEntities[wfID]; ok {
					entities = append(entities, copyWorkflowEntity(wf))
				}
			}
			db.workflowLock.RUnlock()
			runCopy.Entities = entities
		}
	}

	if opts.IncludeHierarchies {
		db.hierarchyLock.RLock()
		hierarchies := make([]*Hierarchy, 0)
		for _, h := range db.hierarchies {
			if h.RunID == id {
				hierarchies = append(hierarchies, copyHierarchy(h))
			}
		}
		db.hierarchyLock.RUnlock()
		runCopy.Hierarchies = hierarchies
	}

	return nil
}

func (db *MemoryDatabase) SetRunProperties(id int, setters ...RunPropertySetter) error {
	db.runLock.Lock()
	run, exists := db.runs[id]
	if !exists {
		db.runLock.Unlock()
		return ErrRunNotFound
	}

	opts := &RunSetterOptions{}
	for _, setter := range setters {
		opt, err := setter(run)
		if err != nil {
			db.runLock.Unlock()
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				db.runLock.Unlock()
				return err
			}
		}
	}

	db.runLock.Unlock()

	if opts.WorkflowID != nil {
		db.workflowLock.RLock()
		_, wfExists := db.workflowEntities[*opts.WorkflowID]
		db.workflowLock.RUnlock()
		if !wfExists {
			return ErrWorkflowEntityNotFound
		}

		db.relationshipLock.Lock()
		db.runToWorkflows[id] = append(db.runToWorkflows[id], *opts.WorkflowID)
		db.relationshipLock.Unlock()
	}

	db.runLock.Lock()
	run.UpdatedAt = time.Now()
	db.runs[id] = copyRun(run)
	db.runLock.Unlock()

	return nil
}

func (db *MemoryDatabase) AddQueue(queue *Queue) (int, error) {
	db.queueLock.Lock()
	defer db.queueLock.Unlock()

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
	db.queueLock.RLock()
	q, exists := db.queues[id]
	if !exists {
		db.queueLock.RUnlock()
		return nil, ErrQueueNotFound
	}
	queueCopy := copyQueue(q)
	db.queueLock.RUnlock()

	cfg := &QueueGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeWorkflows {
		db.relationshipLock.RLock()
		workflowIDs := db.queueToWorkflows[id]
		db.relationshipLock.RUnlock()

		if len(workflowIDs) > 0 {
			db.workflowLock.RLock()
			workflows := make([]*WorkflowEntity, 0, len(workflowIDs))
			for _, wfID := range workflowIDs {
				if wf, exists := db.workflowEntities[wfID]; exists {
					workflows = append(workflows, copyWorkflowEntity(wf))
				}
			}
			db.workflowLock.RUnlock()
			queueCopy.Entities = workflows
		}
	}

	return queueCopy, nil
}

func (db *MemoryDatabase) GetQueueByName(name string, opts ...QueueGetOption) (*Queue, error) {
	db.queueLock.RLock()
	id, exists := db.queueNames[name]
	if !exists {
		db.queueLock.RUnlock()
		return nil, ErrQueueNotFound
	}
	q, qexists := db.queues[id]
	if !qexists {
		db.queueLock.RUnlock()
		return nil, ErrQueueNotFound
	}
	queueCopy := copyQueue(q)
	db.queueLock.RUnlock()

	cfg := &QueueGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeWorkflows {
		db.relationshipLock.RLock()
		workflowIDs := db.queueToWorkflows[id]
		db.relationshipLock.RUnlock()

		if len(workflowIDs) > 0 {
			db.workflowLock.RLock()
			workflows := make([]*WorkflowEntity, 0, len(workflowIDs))
			for _, wfID := range workflowIDs {
				if wf, exists := db.workflowEntities[wfID]; exists {
					workflows = append(workflows, copyWorkflowEntity(wf))
				}
			}
			db.workflowLock.RUnlock()
			queueCopy.Entities = workflows
		}
	}

	return queueCopy, nil
}

func (db *MemoryDatabase) AddVersion(version *Version) (int, error) {
	db.versionLock.Lock()
	defer db.versionLock.Unlock()

	db.versionCounter++
	version.ID = db.versionCounter
	version.CreatedAt = time.Now()
	version.UpdatedAt = version.CreatedAt

	db.versions[version.ID] = copyVersion(version)

	if version.EntityID != 0 {
		db.relationshipLock.Lock()
		db.workflowToVersion[version.EntityID] = append(db.workflowToVersion[version.EntityID], version.ID)
		db.relationshipLock.Unlock()
	}
	return version.ID, nil
}

func (db *MemoryDatabase) GetVersion(id int, opts ...VersionGetOption) (*Version, error) {
	db.versionLock.RLock()
	v, exists := db.versions[id]
	if !exists {
		db.versionLock.RUnlock()
		return nil, ErrVersionNotFound
	}
	versionCopy := copyVersion(v)
	db.versionLock.RUnlock()

	// Options
	cfg := &VersionGetterOptions{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	return versionCopy, nil
}

func (db *MemoryDatabase) AddHierarchy(hierarchy *Hierarchy) (int, error) {
	db.hierarchyLock.Lock()
	defer db.hierarchyLock.Unlock()

	db.hierarchyCounter++
	hierarchy.ID = db.hierarchyCounter

	db.hierarchies[hierarchy.ID] = copyHierarchy(hierarchy)
	return hierarchy.ID, nil
}

func (db *MemoryDatabase) GetHierarchy(id int, opts ...HierarchyGetOption) (*Hierarchy, error) {
	db.hierarchyLock.RLock()
	h, exists := db.hierarchies[id]
	if !exists {
		db.hierarchyLock.RUnlock()
		return nil, ErrHierarchyNotFound
	}
	hCopy := copyHierarchy(h)
	db.hierarchyLock.RUnlock()

	cfg := &HierarchyGetterOptions{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	return hCopy, nil
}

func (db *MemoryDatabase) AddWorkflowEntity(entity *WorkflowEntity) (int, error) {
	db.workflowLock.Lock()
	db.workflowEntityCounter++
	entity.ID = db.workflowEntityCounter

	if entity.WorkflowData != nil {
		db.workflowDataLock.Lock()
		db.workflowDataCounter++
		entity.WorkflowData.ID = db.workflowDataCounter
		entity.WorkflowData.EntityID = entity.ID
		db.workflowData[entity.WorkflowData.ID] = copyWorkflowData(entity.WorkflowData)
		db.workflowDataLock.Unlock()
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = entity.CreatedAt

	// Default queue if none
	db.queueLock.RLock()
	defaultQueueID := 1
	db.queueLock.RUnlock()
	if entity.QueueID == 0 {
		entity.QueueID = defaultQueueID
		db.relationshipLock.Lock()
		db.queueToWorkflows[defaultQueueID] = append(db.queueToWorkflows[defaultQueueID], entity.ID)
		db.relationshipLock.Unlock()
	} else {
		db.relationshipLock.Lock()
		db.workflowToQueue[entity.ID] = entity.QueueID
		db.queueToWorkflows[entity.QueueID] = append(db.queueToWorkflows[entity.QueueID], entity.ID)
		db.relationshipLock.Unlock()
	}

	db.workflowEntities[entity.ID] = copyWorkflowEntity(entity)
	db.workflowLock.Unlock()
	return entity.ID, nil
}

func (db *MemoryDatabase) AddWorkflowExecution(exec *WorkflowExecution) (int, error) {
	db.workflowExecLock.Lock()
	db.workflowExecDataLock.Lock()
	defer db.workflowExecLock.Unlock()
	defer db.workflowExecDataLock.Unlock()

	db.workflowExecutionCounter++
	exec.ID = db.workflowExecutionCounter

	if exec.WorkflowExecutionData != nil {
		db.workflowExecutionDataCounter++
		exec.WorkflowExecutionData.ID = db.workflowExecutionDataCounter
		exec.WorkflowExecutionData.ExecutionID = exec.ID
		db.workflowExecutionData[exec.WorkflowExecutionData.ID] = copyWorkflowExecutionData(exec.WorkflowExecutionData)

		db.relationshipLock.Lock()
		db.execToDataMap[exec.ID] = exec.WorkflowExecutionData.ID
		db.relationshipLock.Unlock()
	}
	exec.CreatedAt = time.Now()
	exec.UpdatedAt = exec.CreatedAt

	db.workflowExecutions[exec.ID] = copyWorkflowExecution(exec)
	return exec.ID, nil
}

func (db *MemoryDatabase) GetWorkflowEntity(id int, opts ...WorkflowEntityGetOption) (*WorkflowEntity, error) {
	db.workflowLock.RLock()
	e, exists := db.workflowEntities[id]
	if !exists {
		db.workflowLock.RUnlock()
		return nil, ErrWorkflowEntityNotFound
	}
	entityCopy := copyWorkflowEntity(e)
	db.workflowLock.RUnlock()

	cfg := &WorkflowEntityGetterOptions{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeQueue {
		db.relationshipLock.RLock()
		queueID, hasQueue := db.workflowToQueue[id]
		db.relationshipLock.RUnlock()
		if hasQueue {
			db.queueLock.RLock()
			if q, qexists := db.queues[queueID]; qexists {
				if entityCopy.Edges == nil {
					entityCopy.Edges = &WorkflowEntityEdges{}
				}
				entityCopy.Edges.Queue = copyQueue(q)
			}
			db.queueLock.RUnlock()
		}
	}

	if cfg.IncludeData {
		db.workflowDataLock.RLock()
		for _, d := range db.workflowData {
			if d.EntityID == id {
				entityCopy.WorkflowData = copyWorkflowData(d)
				break
			}
		}
		db.workflowDataLock.RUnlock()
	}

	return entityCopy, nil
}

func (db *MemoryDatabase) GetWorkflowEntityProperties(id int, getters ...WorkflowEntityPropertyGetter) error {
	db.workflowLock.RLock()
	entity, exists := db.workflowEntities[id]
	if !exists {
		db.workflowLock.RUnlock()
		return ErrWorkflowEntityNotFound
	}
	entityCopy := copyWorkflowEntity(entity)
	db.workflowLock.RUnlock()

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
		db.relationshipLock.RLock()
		versionIDs := db.workflowToVersion[id]
		db.relationshipLock.RUnlock()

		if len(versionIDs) > 0 {
			db.versionLock.RLock()
			versions := make([]*Version, 0, len(versionIDs))
			for _, vID := range versionIDs {
				if v, vexists := db.versions[vID]; vexists {
					versions = append(versions, copyVersion(v))
				}
			}
			db.versionLock.RUnlock()

			if entityCopy.Edges == nil {
				entityCopy.Edges = &WorkflowEntityEdges{}
			}
			entityCopy.Edges.Versions = versions
		}
	}

	if opts.IncludeQueue {
		db.relationshipLock.RLock()
		queueID, qexists := db.workflowToQueue[id]
		db.relationshipLock.RUnlock()
		if qexists {
			db.queueLock.RLock()
			if q, qfound := db.queues[queueID]; qfound {
				if entityCopy.Edges == nil {
					entityCopy.Edges = &WorkflowEntityEdges{}
				}
				entityCopy.Edges.Queue = copyQueue(q)
			}
			db.queueLock.RUnlock()
		}
	}

	if opts.IncludeChildren {
		db.relationshipLock.RLock()
		childMap, cexists := db.workflowToChildren[id]
		db.relationshipLock.RUnlock()
		if cexists {
			if entityCopy.Edges == nil {
				entityCopy.Edges = &WorkflowEntityEdges{}
			}

			if activityIDs, ok := childMap[EntityActivity]; ok {
				db.activityLock.RLock()
				activities := make([]*ActivityEntity, 0, len(activityIDs))
				for _, aID := range activityIDs {
					if a, aexists := db.activityEntities[aID]; aexists {
						activities = append(activities, copyActivityEntity(a))
					}
				}
				db.activityLock.RUnlock()
				entityCopy.Edges.ActivityChildren = activities
			}

			if sagaIDs, ok := childMap[EntitySaga]; ok {
				db.sagaLock.RLock()
				sagas := make([]*SagaEntity, 0, len(sagaIDs))
				for _, sID := range sagaIDs {
					if s, sexists := db.sagaEntities[sID]; sexists {
						sagas = append(sagas, copySagaEntity(s))
					}
				}
				db.sagaLock.RUnlock()
				entityCopy.Edges.SagaChildren = sagas
			}

			if sideEffectIDs, ok := childMap[EntitySideEffect]; ok {
				db.sideEffectLock.RLock()
				sideEffects := make([]*SideEffectEntity, 0, len(sideEffectIDs))
				for _, seID := range sideEffectIDs {
					if se, seexists := db.sideEffectEntities[seID]; seexists {
						sideEffects = append(sideEffects, copySideEffectEntity(se))
					}
				}
				db.sideEffectLock.RUnlock()
				entityCopy.Edges.SideEffectChildren = sideEffects
			}
		}
	}

	if opts.IncludeData {
		db.workflowDataLock.RLock()
		for _, d := range db.workflowData {
			if d.EntityID == id {
				entityCopy.WorkflowData = copyWorkflowData(d)
				break
			}
		}
		db.workflowDataLock.RUnlock()
	}

	return nil
}

func (db *MemoryDatabase) SetWorkflowEntityProperties(id int, setters ...WorkflowEntityPropertySetter) error {
	db.workflowLock.Lock()
	entity, exists := db.workflowEntities[id]
	if !exists {
		db.workflowLock.Unlock()
		return ErrWorkflowEntityNotFound
	}

	opts := &WorkflowEntitySetterOptions{}
	for _, setter := range setters {
		opt, err := setter(entity)
		if err != nil {
			db.workflowLock.Unlock()
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				db.workflowLock.Unlock()
				return err
			}
		}
	}
	db.workflowLock.Unlock()

	if opts.QueueID != nil {
		db.queueLock.RLock()
		_, qexists := db.queues[*opts.QueueID]
		db.queueLock.RUnlock()
		if !qexists {
			return ErrQueueNotFound
		}

		db.relationshipLock.Lock()
		// Remove from old queue
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
		db.relationshipLock.Unlock()
	}

	if opts.Version != nil {
		db.relationshipLock.Lock()
		db.workflowToVersion[id] = append(db.workflowToVersion[id], opts.Version.ID)
		db.relationshipLock.Unlock()
	}

	if opts.ChildID != nil && opts.ChildType != nil {
		db.relationshipLock.Lock()
		if db.workflowToChildren[id] == nil {
			db.workflowToChildren[id] = make(map[EntityType][]int)
		}
		db.workflowToChildren[id][*opts.ChildType] = append(db.workflowToChildren[id][*opts.ChildType], *opts.ChildID)
		db.entityToWorkflow[*opts.ChildID] = id
		db.relationshipLock.Unlock()
	}

	db.workflowLock.Lock()
	entity.UpdatedAt = time.Now()
	db.workflowEntities[id] = copyWorkflowEntity(entity)
	db.workflowLock.Unlock()

	return nil
}

func (db *MemoryDatabase) AddActivityEntity(entity *ActivityEntity, parentWorkflowID int) (int, error) {
	db.activityLock.Lock()
	db.activityEntityCounter++
	entity.ID = db.activityEntityCounter

	if entity.ActivityData != nil {
		db.activityDataLock.Lock()
		db.activityDataCounter++
		entity.ActivityData.ID = db.activityDataCounter
		entity.ActivityData.EntityID = entity.ID
		db.activityData[entity.ActivityData.ID] = copyActivityData(entity.ActivityData)
		db.activityDataLock.Unlock()
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = entity.CreatedAt

	db.activityEntities[entity.ID] = copyActivityEntity(entity)
	db.activityLock.Unlock()

	db.relationshipLock.Lock()
	if _, ok := db.workflowToChildren[parentWorkflowID]; !ok {
		db.workflowToChildren[parentWorkflowID] = make(map[EntityType][]int)
	}
	db.workflowToChildren[parentWorkflowID][EntityActivity] = append(db.workflowToChildren[parentWorkflowID][EntityActivity], entity.ID)
	db.relationshipLock.Unlock()

	return entity.ID, nil
}

func (db *MemoryDatabase) AddActivityExecution(exec *ActivityExecution) (int, error) {
	db.activityExecLock.Lock()
	db.activityExecutionCounter++
	exec.ID = db.activityExecutionCounter

	if exec.ActivityExecutionData != nil {
		db.activityExecDataLock.Lock()
		db.activityExecutionDataCounter++
		exec.ActivityExecutionData.ID = db.activityExecutionDataCounter
		exec.ActivityExecutionData.ExecutionID = exec.ID
		db.activityExecutionData[exec.ActivityExecutionData.ID] = copyActivityExecutionData(exec.ActivityExecutionData)
		db.activityExecDataLock.Unlock()
	}

	exec.CreatedAt = time.Now()
	exec.UpdatedAt = exec.CreatedAt

	db.activityExecutions[exec.ID] = copyActivityExecution(exec)
	db.activityExecLock.Unlock()
	return exec.ID, nil
}

func (db *MemoryDatabase) GetActivityEntity(id int, opts ...ActivityEntityGetOption) (*ActivityEntity, error) {
	db.activityLock.RLock()
	entity, exists := db.activityEntities[id]
	if !exists {
		db.activityLock.RUnlock()
		return nil, ErrActivityEntityNotFound
	}
	eCopy := copyActivityEntity(entity)
	db.activityLock.RUnlock()

	cfg := &ActivityEntityGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		db.activityDataLock.RLock()
		for _, d := range db.activityData {
			if d.EntityID == id {
				eCopy.ActivityData = copyActivityData(d)
				break
			}
		}
		db.activityDataLock.RUnlock()
	}

	return eCopy, nil
}

func (db *MemoryDatabase) GetActivityEntities(workflowID int, opts ...ActivityEntityGetOption) ([]*ActivityEntity, error) {
	db.relationshipLock.RLock()
	activityIDs := db.workflowToChildren[workflowID][EntityActivity]
	db.relationshipLock.RUnlock()

	entities := make([]*ActivityEntity, 0, len(activityIDs))
	for _, aID := range activityIDs {
		e, err := db.GetActivityEntity(aID, opts...)
		if err != nil && err != ErrActivityEntityNotFound {
			return nil, err
		}
		if e != nil {
			entities = append(entities, e)
		}
	}

	return entities, nil
}

func (db *MemoryDatabase) GetActivityEntityProperties(id int, getters ...ActivityEntityPropertyGetter) error {
	db.activityLock.RLock()
	entity, exists := db.activityEntities[id]
	if !exists {
		db.activityLock.RUnlock()
		return ErrActivityEntityNotFound
	}
	eCopy := copyActivityEntity(entity)
	db.activityLock.RUnlock()

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
		db.activityDataLock.RLock()
		for _, d := range db.activityData {
			if d.EntityID == id {
				eCopy.ActivityData = copyActivityData(d)
				break
			}
		}
		db.activityDataLock.RUnlock()
	}

	return nil
}

func (db *MemoryDatabase) SetActivityEntityProperties(id int, setters ...ActivityEntityPropertySetter) error {
	db.activityLock.Lock()
	entity, exists := db.activityEntities[id]
	if !exists {
		db.activityLock.Unlock()
		return ErrActivityEntityNotFound
	}

	opts := &ActivityEntitySetterOptions{}
	for _, setter := range setters {
		opt, err := setter(entity)
		if err != nil {
			db.activityLock.Unlock()
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				db.activityLock.Unlock()
				return err
			}
		}
	}
	db.activityLock.Unlock()

	if opts.ParentWorkflowID != nil {
		db.workflowLock.RLock()
		_, wExists := db.workflowEntities[*opts.ParentWorkflowID]
		db.workflowLock.RUnlock()
		if !wExists {
			return ErrWorkflowEntityNotFound
		}

		db.relationshipLock.Lock()
		db.entityToWorkflow[id] = *opts.ParentWorkflowID
		if db.workflowToChildren[*opts.ParentWorkflowID] == nil {
			db.workflowToChildren[*opts.ParentWorkflowID] = make(map[EntityType][]int)
		}
		db.workflowToChildren[*opts.ParentWorkflowID][EntityActivity] = append(db.workflowToChildren[*opts.ParentWorkflowID][EntityActivity], id)
		db.relationshipLock.Unlock()
	}

	db.activityLock.Lock()
	entity.UpdatedAt = time.Now()
	db.activityEntities[id] = copyActivityEntity(entity)
	db.activityLock.Unlock()

	return nil
}

func (db *MemoryDatabase) AddSagaEntity(entity *SagaEntity, parentWorkflowID int) (int, error) {
	db.sagaLock.Lock()
	db.sagaEntityCounter++
	entity.ID = db.sagaEntityCounter

	if entity.SagaData != nil {
		db.sagaDataLock.Lock()
		db.sagaDataCounter++
		entity.SagaData.ID = db.sagaDataCounter
		entity.SagaData.EntityID = entity.ID
		db.sagaData[entity.SagaData.ID] = copySagaData(entity.SagaData)
		db.sagaDataLock.Unlock()
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = entity.CreatedAt

	db.sagaEntities[entity.ID] = copySagaEntity(entity)
	db.sagaLock.Unlock()

	db.relationshipLock.Lock()
	if _, ok := db.workflowToChildren[parentWorkflowID]; !ok {
		db.workflowToChildren[parentWorkflowID] = make(map[EntityType][]int)
	}
	db.workflowToChildren[parentWorkflowID][EntitySaga] = append(db.workflowToChildren[parentWorkflowID][EntitySaga], entity.ID)
	db.relationshipLock.Unlock()

	return entity.ID, nil
}

func (db *MemoryDatabase) GetSagaEntities(workflowID int, opts ...SagaEntityGetOption) ([]*SagaEntity, error) {
	db.relationshipLock.RLock()
	sagaIDs := db.workflowToChildren[workflowID][EntitySaga]
	db.relationshipLock.RUnlock()

	entities := make([]*SagaEntity, 0, len(sagaIDs))
	for _, sID := range sagaIDs {
		e, err := db.GetSagaEntity(sID, opts...)
		if err != nil && err != ErrSagaEntityNotFound {
			return nil, err
		}
		if e != nil {
			entities = append(entities, e)
		}
	}

	return entities, nil
}

func (db *MemoryDatabase) AddSagaExecution(exec *SagaExecution) (int, error) {
	db.sagaExecLock.Lock()
	db.sagaExecutionCounter++
	exec.ID = db.sagaExecutionCounter

	if exec.SagaExecutionData != nil {
		db.sagaExecDataLock.Lock()
		db.sagaExecutionDataCounter++
		exec.SagaExecutionData.ID = db.sagaExecutionDataCounter
		exec.SagaExecutionData.ExecutionID = exec.ID
		db.sagaExecutionData[exec.SagaExecutionData.ID] = copySagaExecutionData(exec.SagaExecutionData)
		db.sagaExecDataLock.Unlock()
	}

	exec.CreatedAt = time.Now()
	exec.UpdatedAt = exec.CreatedAt

	db.sagaExecutions[exec.ID] = copySagaExecution(exec)
	db.sagaExecLock.Unlock()
	return exec.ID, nil
}

func (db *MemoryDatabase) GetSagaEntity(id int, opts ...SagaEntityGetOption) (*SagaEntity, error) {
	db.sagaLock.RLock()
	entity, exists := db.sagaEntities[id]
	if !exists {
		db.sagaLock.RUnlock()
		return nil, ErrSagaEntityNotFound
	}
	eCopy := copySagaEntity(entity)
	db.sagaLock.RUnlock()

	cfg := &SagaEntityGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		db.sagaDataLock.RLock()
		for _, d := range db.sagaData {
			if d.EntityID == id {
				eCopy.SagaData = copySagaData(d)
				break
			}
		}
		db.sagaDataLock.RUnlock()
	}

	return eCopy, nil
}

func (db *MemoryDatabase) GetSagaEntityProperties(id int, getters ...SagaEntityPropertyGetter) error {
	db.sagaLock.RLock()
	entity, exists := db.sagaEntities[id]
	if !exists {
		db.sagaLock.RUnlock()
		return ErrSagaEntityNotFound
	}
	eCopy := copySagaEntity(entity)
	db.sagaLock.RUnlock()

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
		db.sagaDataLock.RLock()
		for _, d := range db.sagaData {
			if d.EntityID == id {
				eCopy.SagaData = copySagaData(d)
				break
			}
		}
		db.sagaDataLock.RUnlock()
	}

	return nil
}

func (db *MemoryDatabase) SetSagaEntityProperties(id int, setters ...SagaEntityPropertySetter) error {
	db.sagaLock.Lock()
	entity, exists := db.sagaEntities[id]
	if !exists {
		db.sagaLock.Unlock()
		return ErrSagaEntityNotFound
	}

	opts := &SagaEntitySetterOptions{}
	for _, setter := range setters {
		opt, err := setter(entity)
		if err != nil {
			db.sagaLock.Unlock()
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				db.sagaLock.Unlock()
				return err
			}
		}
	}
	db.sagaLock.Unlock()

	if opts.ParentWorkflowID != nil {
		db.workflowLock.RLock()
		_, wExists := db.workflowEntities[*opts.ParentWorkflowID]
		db.workflowLock.RUnlock()
		if !wExists {
			return ErrWorkflowEntityNotFound
		}

		db.relationshipLock.Lock()
		db.entityToWorkflow[id] = *opts.ParentWorkflowID
		if db.workflowToChildren[*opts.ParentWorkflowID] == nil {
			db.workflowToChildren[*opts.ParentWorkflowID] = make(map[EntityType][]int)
		}
		db.workflowToChildren[*opts.ParentWorkflowID][EntitySaga] = append(db.workflowToChildren[*opts.ParentWorkflowID][EntitySaga], id)
		db.relationshipLock.Unlock()
	}

	db.sagaLock.Lock()
	entity.UpdatedAt = time.Now()
	db.sagaEntities[id] = copySagaEntity(entity)
	db.sagaLock.Unlock()

	return nil
}

func (db *MemoryDatabase) AddSideEffectEntity(entity *SideEffectEntity, parentWorkflowID int) (int, error) {
	db.sideEffectLock.Lock()
	db.sideEffectEntityCounter++
	entity.ID = db.sideEffectEntityCounter

	if entity.SideEffectData != nil {
		db.sideEffectDataLock.Lock()
		db.sideEffectDataCounter++
		entity.SideEffectData.ID = db.sideEffectDataCounter
		entity.SideEffectData.EntityID = entity.ID
		db.sideEffectData[entity.SideEffectData.ID] = copySideEffectData(entity.SideEffectData)
		db.sideEffectDataLock.Unlock()
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = entity.CreatedAt

	db.sideEffectEntities[entity.ID] = copySideEffectEntity(entity)
	db.sideEffectLock.Unlock()

	db.relationshipLock.Lock()
	if _, ok := db.workflowToChildren[parentWorkflowID]; !ok {
		db.workflowToChildren[parentWorkflowID] = make(map[EntityType][]int)
	}
	db.workflowToChildren[parentWorkflowID][EntitySideEffect] = append(db.workflowToChildren[parentWorkflowID][EntitySideEffect], entity.ID)
	db.relationshipLock.Unlock()

	return entity.ID, nil
}

func (db *MemoryDatabase) GetSideEffectEntities(workflowID int, opts ...SideEffectEntityGetOption) ([]*SideEffectEntity, error) {
	db.relationshipLock.RLock()
	sideEffectIDs := db.workflowToChildren[workflowID][EntitySideEffect]
	db.relationshipLock.RUnlock()

	entities := make([]*SideEffectEntity, 0, len(sideEffectIDs))
	for _, seID := range sideEffectIDs {
		e, err := db.GetSideEffectEntity(seID, opts...)
		if err != nil && err != ErrSideEffectEntityNotFound {
			return nil, err
		}
		if e != nil {
			entities = append(entities, e)
		}
	}

	return entities, nil
}

func (db *MemoryDatabase) AddSideEffectExecution(exec *SideEffectExecution) (int, error) {
	db.sideEffectExecLock.Lock()
	db.sideEffectExecutionCounter++
	exec.ID = db.sideEffectExecutionCounter

	if exec.SideEffectExecutionData != nil {
		db.sideEffectExecDataLock.Lock()
		db.sideEffectExecutionDataCounter++
		exec.SideEffectExecutionData.ID = db.sideEffectExecutionDataCounter
		exec.SideEffectExecutionData.ExecutionID = exec.ID
		db.sideEffectExecutionData[exec.SideEffectExecutionData.ID] = copySideEffectExecutionData(exec.SideEffectExecutionData)
		db.sideEffectExecDataLock.Unlock()
	}

	exec.CreatedAt = time.Now()
	exec.UpdatedAt = exec.CreatedAt

	db.sideEffectExecutions[exec.ID] = copySideEffectExecution(exec)
	db.sideEffectExecLock.Unlock()
	return exec.ID, nil
}

func (db *MemoryDatabase) GetSideEffectEntity(id int, opts ...SideEffectEntityGetOption) (*SideEffectEntity, error) {
	db.sideEffectLock.RLock()
	entity, exists := db.sideEffectEntities[id]
	if !exists {
		db.sideEffectLock.RUnlock()
		return nil, ErrSideEffectEntityNotFound
	}
	eCopy := copySideEffectEntity(entity)
	db.sideEffectLock.RUnlock()

	cfg := &SideEffectEntityGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		db.sideEffectDataLock.RLock()
		for _, d := range db.sideEffectData {
			if d.EntityID == id {
				eCopy.SideEffectData = copySideEffectData(d)
				break
			}
		}
		db.sideEffectDataLock.RUnlock()
	}

	return eCopy, nil
}

func (db *MemoryDatabase) GetSideEffectEntityProperties(id int, getters ...SideEffectEntityPropertyGetter) error {
	db.sideEffectLock.RLock()
	entity, exists := db.sideEffectEntities[id]
	if !exists {
		db.sideEffectLock.RUnlock()
		return ErrSideEffectEntityNotFound
	}
	eCopy := copySideEffectEntity(entity)
	db.sideEffectLock.RUnlock()

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
		}
	}

	if opts.IncludeData {
		db.sideEffectDataLock.RLock()
		for _, d := range db.sideEffectData {
			if d.EntityID == id {
				eCopy.SideEffectData = copySideEffectData(d)
				break
			}
		}
		db.sideEffectDataLock.RUnlock()
	}

	return nil
}

func (db *MemoryDatabase) SetSideEffectEntityProperties(id int, setters ...SideEffectEntityPropertySetter) error {
	db.sideEffectLock.Lock()
	entity, exists := db.sideEffectEntities[id]
	if !exists {
		db.sideEffectLock.Unlock()
		return ErrSideEffectEntityNotFound
	}

	opts := &SideEffectEntitySetterOptions{}
	for _, setter := range setters {
		opt, err := setter(entity)
		if err != nil {
			db.sideEffectLock.Unlock()
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				db.sideEffectLock.Unlock()
				return err
			}
		}
	}
	db.sideEffectLock.Unlock()

	if opts.ParentWorkflowID != nil {
		db.workflowLock.RLock()
		_, wExists := db.workflowEntities[*opts.ParentWorkflowID]
		db.workflowLock.RUnlock()
		if !wExists {
			return ErrWorkflowEntityNotFound
		}

		db.relationshipLock.Lock()
		db.entityToWorkflow[id] = *opts.ParentWorkflowID
		if db.workflowToChildren[*opts.ParentWorkflowID] == nil {
			db.workflowToChildren[*opts.ParentWorkflowID] = make(map[EntityType][]int)
		}
		db.workflowToChildren[*opts.ParentWorkflowID][EntitySideEffect] = append(db.workflowToChildren[*opts.ParentWorkflowID][EntitySideEffect], id)
		db.relationshipLock.Unlock()
	}

	db.sideEffectLock.Lock()
	entity.UpdatedAt = time.Now()
	db.sideEffectEntities[id] = copySideEffectEntity(entity)
	db.sideEffectLock.Unlock()

	return nil
}

func (db *MemoryDatabase) GetWorkflowExecution(id int, opts ...WorkflowExecutionGetOption) (*WorkflowExecution, error) {
	db.workflowExecLock.RLock()
	exec, exists := db.workflowExecutions[id]
	if !exists {
		db.workflowExecLock.RUnlock()
		return nil, ErrWorkflowExecutionNotFound
	}
	execCopy := copyWorkflowExecution(exec)
	db.workflowExecLock.RUnlock()

	cfg := &WorkflowExecutionGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		db.workflowExecDataLock.RLock()
		for _, d := range db.workflowExecutionData {
			if d.ExecutionID == id {
				execCopy.WorkflowExecutionData = copyWorkflowExecutionData(d)
				break
			}
		}
		db.workflowExecDataLock.RUnlock()
	}

	return execCopy, nil
}

func (db *MemoryDatabase) GetWorkflowExecutions(entityID int) ([]*WorkflowExecution, error) {
	db.workflowExecLock.RLock()
	results := make([]*WorkflowExecution, 0)
	for _, exec := range db.workflowExecutions {
		if exec.EntityID == entityID {
			results = append(results, copyWorkflowExecution(exec))
		}
	}
	db.workflowExecLock.RUnlock()
	return results, nil
}

func (db *MemoryDatabase) GetActivityExecution(id int, opts ...ActivityExecutionGetOption) (*ActivityExecution, error) {
	db.activityExecLock.RLock()
	exec, exists := db.activityExecutions[id]
	if !exists {
		db.activityExecLock.RUnlock()
		return nil, ErrActivityExecutionNotFound
	}
	execCopy := copyActivityExecution(exec)
	db.activityExecLock.RUnlock()

	cfg := &ActivityExecutionGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		db.activityExecDataLock.RLock()
		for _, d := range db.activityExecutionData {
			if d.ExecutionID == id {
				execCopy.ActivityExecutionData = copyActivityExecutionData(d)
				break
			}
		}
		db.activityExecDataLock.RUnlock()
	}

	return execCopy, nil
}

func (db *MemoryDatabase) GetActivityExecutions(entityID int) ([]*ActivityExecution, error) {
	db.activityExecLock.RLock()
	results := make([]*ActivityExecution, 0)
	for _, exec := range db.activityExecutions {
		if exec.EntityID == entityID {
			results = append(results, copyActivityExecution(exec))
		}
	}
	db.activityExecLock.RUnlock()
	return results, nil
}

func (db *MemoryDatabase) GetSagaExecution(id int, opts ...SagaExecutionGetOption) (*SagaExecution, error) {
	db.sagaExecLock.RLock()
	exec, exists := db.sagaExecutions[id]
	if !exists {
		db.sagaExecLock.RUnlock()
		return nil, ErrSagaExecutionNotFound
	}
	execCopy := copySagaExecution(exec)
	db.sagaExecLock.RUnlock()

	cfg := &SagaExecutionGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		db.sagaExecDataLock.RLock()
		for _, d := range db.sagaExecutionData {
			if d.ExecutionID == id {
				execCopy.SagaExecutionData = copySagaExecutionData(d)
				break
			}
		}
		db.sagaExecDataLock.RUnlock()
	}

	return execCopy, nil
}

func (db *MemoryDatabase) GetSagaExecutions(entityID int) ([]*SagaExecution, error) {
	db.sagaExecLock.RLock()
	results := make([]*SagaExecution, 0)
	for _, exec := range db.sagaExecutions {
		if exec.EntityID == entityID {
			results = append(results, copySagaExecution(exec))
		}
	}
	db.sagaExecLock.RUnlock()

	if len(results) == 0 {
		return nil, ErrSagaExecutionNotFound
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.Before(results[j].CreatedAt)
	})

	return results, nil
}

func (db *MemoryDatabase) GetSideEffectExecution(id int, opts ...SideEffectExecutionGetOption) (*SideEffectExecution, error) {
	db.sideEffectExecLock.RLock()
	exec, exists := db.sideEffectExecutions[id]
	if !exists {
		db.sideEffectExecLock.RUnlock()
		return nil, ErrSideEffectExecutionNotFound
	}
	execCopy := copySideEffectExecution(exec)
	db.sideEffectExecLock.RUnlock()

	cfg := &SideEffectExecutionGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		db.sideEffectExecDataLock.RLock()
		for _, d := range db.sideEffectExecutionData {
			if d.ExecutionID == id {
				execCopy.SideEffectExecutionData = copySideEffectExecutionData(d)
				break
			}
		}
		db.sideEffectExecDataLock.RUnlock()
	}

	return execCopy, nil
}

func (db *MemoryDatabase) GetSideEffectExecutions(entityID int) ([]*SideEffectExecution, error) {
	db.sideEffectExecLock.RLock()
	results := make([]*SideEffectExecution, 0)
	for _, exec := range db.sideEffectExecutions {
		if exec.EntityID == entityID {
			results = append(results, copySideEffectExecution(exec))
		}
	}
	db.sideEffectExecLock.RUnlock()
	return results, nil
}

// Activity Data properties
func (db *MemoryDatabase) GetActivityDataProperties(entityID int, getters ...ActivityDataPropertyGetter) error {
	db.activityDataLock.RLock()
	var data *ActivityData
	for _, d := range db.activityData {
		if d.EntityID == entityID {
			tmp := d
			data = tmp
			break
		}
	}
	db.activityDataLock.RUnlock()

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

func (db *MemoryDatabase) SetActivityDataProperties(entityID int, setters ...ActivityDataPropertySetter) error {
	db.activityDataLock.Lock()
	defer db.activityDataLock.Unlock()

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

// Saga Data properties
func (db *MemoryDatabase) GetSagaDataProperties(entityID int, getters ...SagaDataPropertyGetter) error {
	db.sagaDataLock.RLock()
	var data *SagaData
	for _, d := range db.sagaData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}
	db.sagaDataLock.RUnlock()

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

func (db *MemoryDatabase) SetSagaDataProperties(entityID int, setters ...SagaDataPropertySetter) error {
	db.sagaDataLock.Lock()
	defer db.sagaDataLock.Unlock()

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

// SideEffect Data properties
func (db *MemoryDatabase) GetSideEffectDataProperties(entityID int, getters ...SideEffectDataPropertyGetter) error {
	db.sideEffectDataLock.RLock()
	var data *SideEffectData
	for _, d := range db.sideEffectData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}
	db.sideEffectDataLock.RUnlock()

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

func (db *MemoryDatabase) SetSideEffectDataProperties(entityID int, setters ...SideEffectDataPropertySetter) error {
	db.sideEffectDataLock.Lock()
	defer db.sideEffectDataLock.Unlock()

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

// Workflow Data properties
func (db *MemoryDatabase) GetWorkflowDataProperties(entityID int, getters ...WorkflowDataPropertyGetter) error {
	db.workflowDataLock.RLock()
	var data *WorkflowData
	for _, d := range db.workflowData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}
	db.workflowDataLock.RUnlock()

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

func (db *MemoryDatabase) SetWorkflowDataProperties(entityID int, setters ...WorkflowDataPropertySetter) error {
	db.workflowDataLock.Lock()
	defer db.workflowDataLock.Unlock()

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

// Execution Data properties
func (db *MemoryDatabase) GetWorkflowExecutionDataProperties(entityID int, getters ...WorkflowExecutionDataPropertyGetter) error {
	db.workflowExecDataLock.RLock()
	var data *WorkflowExecutionData
	for _, d := range db.workflowExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}
	db.workflowExecDataLock.RUnlock()

	if data == nil {
		return ErrWorkflowExecutionNotFound
	}

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetWorkflowExecutionDataProperties(entityID int, setters ...WorkflowExecutionDataPropertySetter) error {
	db.workflowExecDataLock.Lock()
	defer db.workflowExecDataLock.Unlock()

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
		if _, err := setter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) GetActivityExecutionDataProperties(entityID int, getters ...ActivityExecutionDataPropertyGetter) error {
	db.activityExecDataLock.RLock()
	var data *ActivityExecutionData
	for _, d := range db.activityExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}
	db.activityExecDataLock.RUnlock()

	if data == nil {
		return ErrActivityExecutionNotFound
	}

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetActivityExecutionDataProperties(entityID int, setters ...ActivityExecutionDataPropertySetter) error {
	db.activityExecDataLock.Lock()
	defer db.activityExecDataLock.Unlock()

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
		if _, err := setter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) GetSagaExecutionDataProperties(entityID int, getters ...SagaExecutionDataPropertyGetter) error {
	db.sagaExecDataLock.RLock()
	var data *SagaExecutionData
	for _, d := range db.sagaExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}
	db.sagaExecDataLock.RUnlock()

	if data == nil {
		return ErrSagaExecutionNotFound
	}

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSagaExecutionDataProperties(entityID int, setters ...SagaExecutionDataPropertySetter) error {
	db.sagaExecDataLock.Lock()
	defer db.sagaExecDataLock.Unlock()

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
		if _, err := setter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) GetSideEffectExecutionDataProperties(entityID int, getters ...SideEffectExecutionDataPropertyGetter) error {
	db.sideEffectExecDataLock.RLock()
	var data *SideEffectExecutionData
	for _, d := range db.sideEffectExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}
	db.sideEffectExecDataLock.RUnlock()

	if data == nil {
		return ErrSideEffectExecutionNotFound
	}

	for _, getter := range getters {
		if _, err := getter(data); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSideEffectExecutionDataProperties(entityID int, setters ...SideEffectExecutionDataPropertySetter) error {
	db.sideEffectExecDataLock.Lock()
	defer db.sideEffectExecDataLock.Unlock()

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
		if _, err := setter(data); err != nil {
			return err
		}
	}

	return nil
}

// Workflow Execution properties
func (db *MemoryDatabase) GetWorkflowExecutionProperties(id int, getters ...WorkflowExecutionPropertyGetter) error {
	db.workflowExecLock.RLock()
	exec, exists := db.workflowExecutions[id]
	if !exists {
		db.workflowExecLock.RUnlock()
		return ErrWorkflowExecutionNotFound
	}
	execCopy := copyWorkflowExecution(exec)
	db.workflowExecLock.RUnlock()

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
				db.workflowExecDataLock.RLock()
				for _, d := range db.workflowExecutionData {
					if d.ExecutionID == id {
						execCopy.WorkflowExecutionData = copyWorkflowExecutionData(d)
						break
					}
				}
				db.workflowExecDataLock.RUnlock()
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetWorkflowExecutionProperties(id int, setters ...WorkflowExecutionPropertySetter) error {
	db.workflowExecLock.Lock()
	exec, exists := db.workflowExecutions[id]
	if !exists {
		db.workflowExecLock.Unlock()
		return ErrWorkflowExecutionNotFound
	}
	for _, setter := range setters {
		if _, err := setter(exec); err != nil {
			db.workflowExecLock.Unlock()
			return err
		}
	}
	db.workflowExecutions[id] = copyWorkflowExecution(exec)
	db.workflowExecLock.Unlock()
	return nil
}

// Activity Execution properties
func (db *MemoryDatabase) GetActivityExecutionProperties(id int, getters ...ActivityExecutionPropertyGetter) error {
	db.activityExecLock.RLock()
	exec, exists := db.activityExecutions[id]
	if !exists {
		db.activityExecLock.RUnlock()
		return ErrActivityExecutionNotFound
	}
	execCopy := copyActivityExecution(exec)
	db.activityExecLock.RUnlock()

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
				db.activityExecDataLock.RLock()
				for _, d := range db.activityExecutionData {
					if d.ExecutionID == id {
						execCopy.ActivityExecutionData = copyActivityExecutionData(d)
						break
					}
				}
				db.activityExecDataLock.RUnlock()
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetActivityExecutionProperties(id int, setters ...ActivityExecutionPropertySetter) error {
	db.activityExecLock.Lock()
	exec, exists := db.activityExecutions[id]
	if !exists {
		db.activityExecLock.Unlock()
		return ErrActivityExecutionNotFound
	}
	for _, setter := range setters {
		if _, err := setter(exec); err != nil {
			db.activityExecLock.Unlock()
			return err
		}
	}
	db.activityExecutions[id] = copyActivityExecution(exec)
	db.activityExecLock.Unlock()
	return nil
}

// Saga Execution properties
func (db *MemoryDatabase) GetSagaExecutionProperties(id int, getters ...SagaExecutionPropertyGetter) error {
	db.sagaExecLock.RLock()
	exec, exists := db.sagaExecutions[id]
	if !exists {
		db.sagaExecLock.RUnlock()
		return ErrSagaExecutionNotFound
	}
	execCopy := copySagaExecution(exec)
	db.sagaExecLock.RUnlock()

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
				db.sagaExecDataLock.RLock()
				for _, d := range db.sagaExecutionData {
					if d.ExecutionID == id {
						execCopy.SagaExecutionData = copySagaExecutionData(d)
						break
					}
				}
				db.sagaExecDataLock.RUnlock()
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSagaExecutionProperties(id int, setters ...SagaExecutionPropertySetter) error {
	db.sagaExecLock.Lock()
	exec, exists := db.sagaExecutions[id]
	if !exists {
		db.sagaExecLock.Unlock()
		return ErrSagaExecutionNotFound
	}
	for _, setter := range setters {
		if _, err := setter(exec); err != nil {
			db.sagaExecLock.Unlock()
			return err
		}
	}
	db.sagaExecutions[id] = copySagaExecution(exec)
	db.sagaExecLock.Unlock()
	return nil
}

// SideEffect Execution properties
func (db *MemoryDatabase) GetSideEffectExecutionProperties(id int, getters ...SideEffectExecutionPropertyGetter) error {
	db.sideEffectExecLock.RLock()
	exec, exists := db.sideEffectExecutions[id]
	if !exists {
		db.sideEffectExecLock.RUnlock()
		return ErrSideEffectExecutionNotFound
	}
	execCopy := copySideEffectExecution(exec)
	db.sideEffectExecLock.RUnlock()

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
				db.sideEffectExecDataLock.RLock()
				for _, d := range db.sideEffectExecutionData {
					if d.ExecutionID == id {
						execCopy.SideEffectExecutionData = copySideEffectExecutionData(d)
						break
					}
				}
				db.sideEffectExecDataLock.RUnlock()
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSideEffectExecutionProperties(id int, setters ...SideEffectExecutionPropertySetter) error {
	db.sideEffectExecLock.Lock()
	exec, exists := db.sideEffectExecutions[id]
	if !exists {
		db.sideEffectExecLock.Unlock()
		return ErrSideEffectExecutionNotFound
	}
	for _, setter := range setters {
		if _, err := setter(exec); err != nil {
			db.sideEffectExecLock.Unlock()
			return err
		}
	}
	db.sideEffectExecutions[id] = copySideEffectExecution(exec)
	db.sideEffectExecLock.Unlock()
	return nil
}

// Hierarchy properties
func (db *MemoryDatabase) GetHierarchyProperties(id int, getters ...HierarchyPropertyGetter) error {
	db.hierarchyLock.RLock()
	hierarchy, exists := db.hierarchies[id]
	if !exists {
		db.hierarchyLock.RUnlock()
		return ErrHierarchyNotFound
	}
	hCopy := copyHierarchy(hierarchy)
	db.hierarchyLock.RUnlock()

	for _, getter := range getters {
		if _, err := getter(hCopy); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) GetHierarchyByParentEntity(parentEntityID int, childStepID string, specificType EntityType) (*Hierarchy, error) {
	db.hierarchyLock.RLock()
	defer db.hierarchyLock.RUnlock()

	for _, h := range db.hierarchies {
		if h.ParentEntityID == parentEntityID && h.ChildStepID == childStepID && h.ChildType == specificType {
			return copyHierarchy(h), nil
		}
	}

	return nil, ErrHierarchyNotFound
}

func (db *MemoryDatabase) GetHierarchiesByParentEntityAndStep(parentEntityID int, childStepID string, specificType EntityType) ([]*Hierarchy, error) {
	db.hierarchyLock.RLock()
	defer db.hierarchyLock.RUnlock()

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

func (db *MemoryDatabase) SetHierarchyProperties(id int, setters ...HierarchyPropertySetter) error {
	db.hierarchyLock.Lock()
	hierarchy, exists := db.hierarchies[id]
	if !exists {
		db.hierarchyLock.Unlock()
		return ErrHierarchyNotFound
	}
	for _, setter := range setters {
		if _, err := setter(hierarchy); err != nil {
			db.hierarchyLock.Unlock()
			return err
		}
	}
	db.hierarchies[id] = copyHierarchy(hierarchy)
	db.hierarchyLock.Unlock()
	return nil
}

func (db *MemoryDatabase) GetHierarchiesByParentEntity(parentEntityID int) ([]*Hierarchy, error) {
	db.hierarchyLock.RLock()
	defer db.hierarchyLock.RUnlock()

	var results []*Hierarchy
	for _, h := range db.hierarchies {
		if h.ParentEntityID == parentEntityID {
			results = append(results, copyHierarchy(h))
		}
	}
	return results, nil
}

func (db *MemoryDatabase) GetHierarchiesByChildEntity(childEntityID int) ([]*Hierarchy, error) {
	db.hierarchyLock.RLock()
	defer db.hierarchyLock.RUnlock()

	var results []*Hierarchy
	for _, h := range db.hierarchies {
		if h.ChildEntityID == childEntityID {
			results = append(results, copyHierarchy(h))
		}
	}
	return results, nil
}

func (db *MemoryDatabase) UpdateHierarchy(hierarchy *Hierarchy) error {
	db.hierarchyLock.Lock()
	defer db.hierarchyLock.Unlock()

	if _, exists := db.hierarchies[hierarchy.ID]; !exists {
		return ErrHierarchyNotFound
	}

	db.hierarchies[hierarchy.ID] = copyHierarchy(hierarchy)
	return nil
}

func (db *MemoryDatabase) GetQueueProperties(id int, getters ...QueuePropertyGetter) error {
	db.queueLock.RLock()
	queue, exists := db.queues[id]
	if !exists {
		db.queueLock.RUnlock()
		return ErrQueueNotFound
	}
	qCopy := copyQueue(queue)
	db.queueLock.RUnlock()

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
				db.relationshipLock.RLock()
				workflowIDs := db.queueToWorkflows[id]
				db.relationshipLock.RUnlock()

				if len(workflowIDs) > 0 {
					db.workflowLock.RLock()
					workflows := make([]*WorkflowEntity, 0, len(workflowIDs))
					for _, wfID := range workflowIDs {
						if wf, wexists := db.workflowEntities[wfID]; wexists {
							workflows = append(workflows, copyWorkflowEntity(wf))
						}
					}
					db.workflowLock.RUnlock()
					qCopy.Entities = workflows
				}
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetQueueProperties(id int, setters ...QueuePropertySetter) error {
	db.queueLock.Lock()
	queue, exists := db.queues[id]
	if !exists {
		db.queueLock.Unlock()
		return ErrQueueNotFound
	}
	for _, setter := range setters {
		opt, err := setter(queue)
		if err != nil {
			db.queueLock.Unlock()
			return err
		}
		if opt != nil {
			opts := &QueueSetterOptions{}
			if err := opt(opts); err != nil {
				db.queueLock.Unlock()
				return err
			}
			if opts.WorkflowIDs != nil {
				db.relationshipLock.Lock()
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
				db.relationshipLock.Unlock()
			}
		}
	}
	db.queues[id] = copyQueue(queue)
	db.queueLock.Unlock()
	return nil
}

func (db *MemoryDatabase) UpdateQueue(queue *Queue) error {
	db.queueLock.Lock()
	defer db.queueLock.Unlock()

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
func (db *MemoryDatabase) GetVersionByWorkflowAndChangeID(workflowID int, changeID string) (*Version, error) {
	db.relationshipLock.RLock()
	versionIDs := db.workflowVersions[workflowID]
	db.relationshipLock.RUnlock()

	db.versionLock.RLock()
	defer db.versionLock.RUnlock()

	for _, vID := range versionIDs {
		if v, vexists := db.versions[vID]; vexists && v.ChangeID == changeID {
			return copyVersion(v), nil
		}
	}
	return nil, ErrVersionNotFound
}

func (db *MemoryDatabase) GetVersionsByWorkflowID(workflowID int) ([]*Version, error) {
	db.relationshipLock.RLock()
	versionIDs := db.workflowVersions[workflowID]
	db.relationshipLock.RUnlock()

	db.versionLock.RLock()
	defer db.versionLock.RUnlock()

	versions := make([]*Version, 0, len(versionIDs))
	for _, vID := range versionIDs {
		if version, exists := db.versions[vID]; exists {
			versions = append(versions, copyVersion(version))
		}
	}

	return versions, nil
}

func (db *MemoryDatabase) SetVersion(version *Version) error {
	db.versionLock.Lock()
	defer db.versionLock.Unlock()

	if version.ID == 0 {
		db.versionCounter++
		version.ID = db.versionCounter
		version.CreatedAt = time.Now()
	}
	version.UpdatedAt = time.Now()

	db.versions[version.ID] = copyVersion(version)

	db.relationshipLock.Lock()
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
	db.relationshipLock.Unlock()

	return nil
}

func (db *MemoryDatabase) DeleteVersionsForWorkflow(workflowID int) error {
	db.versionLock.Lock()
	db.relationshipLock.Lock()
	versionIDs, ok := db.workflowVersions[workflowID]
	if ok {
		for _, vID := range versionIDs {
			delete(db.versions, vID)
		}
		delete(db.workflowVersions, workflowID)
	}
	db.relationshipLock.Unlock()
	db.versionLock.Unlock()
	return nil
}

func (db *MemoryDatabase) GetVersionProperties(id int, getters ...VersionPropertyGetter) error {
	db.versionLock.RLock()
	version, exists := db.versions[id]
	if !exists {
		db.versionLock.RUnlock()
		return ErrVersionNotFound
	}
	vCopy := copyVersion(version)
	db.versionLock.RUnlock()

	for _, getter := range getters {
		if _, err := getter(vCopy); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetVersionProperties(id int, setters ...VersionPropertySetter) error {
	db.versionLock.Lock()
	version, exists := db.versions[id]
	if !exists {
		db.versionLock.Unlock()
		return ErrVersionNotFound
	}

	for _, setter := range setters {
		if _, err := setter(version); err != nil {
			db.versionLock.Unlock()
			return err
		}
	}
	version.UpdatedAt = time.Now()
	db.versions[id] = copyVersion(version)
	db.versionLock.Unlock()
	return nil
}

func (db *MemoryDatabase) UpdateVersion(version *Version) error {
	db.versionLock.Lock()
	defer db.versionLock.Unlock()

	if _, exists := db.versions[version.ID]; !exists {
		return ErrVersionNotFound
	}

	version.UpdatedAt = time.Now()
	db.versions[version.ID] = copyVersion(version)

	return nil
}

func (db *MemoryDatabase) UpdateActivityEntity(entity *ActivityEntity) error {
	db.activityLock.Lock()
	defer db.activityLock.Unlock()

	if _, exists := db.activityEntities[entity.ID]; !exists {
		return ErrActivityEntityNotFound
	}

	entity.UpdatedAt = time.Now()
	db.activityEntities[entity.ID] = copyActivityEntity(entity)
	return nil
}

func (db *MemoryDatabase) UpdateSagaEntity(entity *SagaEntity) error {
	db.sagaLock.Lock()
	defer db.sagaLock.Unlock()

	if _, exists := db.sagaEntities[entity.ID]; !exists {
		return ErrSagaEntityNotFound
	}

	entity.UpdatedAt = time.Now()
	db.sagaEntities[entity.ID] = copySagaEntity(entity)
	return nil
}

func (db *MemoryDatabase) UpdateSideEffectEntity(entity *SideEffectEntity) error {
	db.sideEffectLock.Lock()
	defer db.sideEffectLock.Unlock()

	if _, exists := db.sideEffectEntities[entity.ID]; !exists {
		return ErrSideEffectEntityNotFound
	}

	entity.UpdatedAt = time.Now()
	db.sideEffectEntities[entity.ID] = copySideEffectEntity(entity)
	return nil
}

func (db *MemoryDatabase) UpdateWorkflowEntity(entity *WorkflowEntity) error {
	db.workflowLock.Lock()
	defer db.workflowLock.Unlock()

	if _, exists := db.workflowEntities[entity.ID]; !exists {
		return ErrWorkflowEntityNotFound
	}

	entity.UpdatedAt = time.Now()
	db.workflowEntities[entity.ID] = copyWorkflowEntity(entity)
	return nil
}

func (db *MemoryDatabase) AddWorkflowData(entityID int, data *WorkflowData) (int, error) {
	db.workflowDataLock.Lock()
	db.workflowDataCounter++
	data.ID = db.workflowDataCounter
	data.EntityID = entityID
	db.workflowData[data.ID] = copyWorkflowData(data)
	db.workflowDataLock.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) AddActivityData(entityID int, data *ActivityData) (int, error) {
	db.activityDataLock.Lock()
	db.activityDataCounter++
	data.ID = db.activityDataCounter
	data.EntityID = entityID
	db.activityData[data.ID] = copyActivityData(data)
	db.activityDataLock.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) AddSagaData(entityID int, data *SagaData) (int, error) {
	db.sagaDataLock.Lock()
	db.sagaDataCounter++
	data.ID = db.sagaDataCounter
	data.EntityID = entityID
	db.sagaData[data.ID] = copySagaData(data)
	db.sagaDataLock.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) AddSideEffectData(entityID int, data *SideEffectData) (int, error) {
	db.sideEffectDataLock.Lock()
	db.sideEffectDataCounter++
	data.ID = db.sideEffectDataCounter
	data.EntityID = entityID
	db.sideEffectData[data.ID] = copySideEffectData(data)
	db.sideEffectDataLock.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) GetWorkflowData(id int) (*WorkflowData, error) {
	db.workflowDataLock.RLock()
	d, exists := db.workflowData[id]
	if !exists {
		db.workflowDataLock.RUnlock()
		return nil, ErrWorkflowEntityNotFound
	}
	dCopy := copyWorkflowData(d)
	db.workflowDataLock.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetActivityData(id int) (*ActivityData, error) {
	db.activityDataLock.RLock()
	d, exists := db.activityData[id]
	if !exists {
		db.activityDataLock.RUnlock()
		return nil, ErrActivityEntityNotFound
	}
	dCopy := copyActivityData(d)
	db.activityDataLock.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetSagaData(id int) (*SagaData, error) {
	db.sagaDataLock.RLock()
	d, exists := db.sagaData[id]
	if !exists {
		db.sagaDataLock.RUnlock()
		return nil, ErrSagaEntityNotFound
	}
	dCopy := copySagaData(d)
	db.sagaDataLock.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetSideEffectData(id int) (*SideEffectData, error) {
	db.sideEffectDataLock.RLock()
	d, exists := db.sideEffectData[id]
	if !exists {
		db.sideEffectDataLock.RUnlock()
		return nil, ErrSideEffectEntityNotFound
	}
	dCopy := copySideEffectData(d)
	db.sideEffectDataLock.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetWorkflowDataByEntityID(entityID int) (*WorkflowData, error) {
	db.workflowDataLock.RLock()
	for _, d := range db.workflowData {
		if d.EntityID == entityID {
			dCopy := copyWorkflowData(d)
			db.workflowDataLock.RUnlock()
			return dCopy, nil
		}
	}
	db.workflowDataLock.RUnlock()
	return nil, ErrWorkflowEntityNotFound
}

func (db *MemoryDatabase) GetActivityDataByEntityID(entityID int) (*ActivityData, error) {
	db.activityDataLock.RLock()
	for _, d := range db.activityData {
		if d.EntityID == entityID {
			dCopy := copyActivityData(d)
			db.activityDataLock.RUnlock()
			return dCopy, nil
		}
	}
	db.activityDataLock.RUnlock()
	return nil, ErrActivityEntityNotFound
}

func (db *MemoryDatabase) GetSagaDataByEntityID(entityID int) (*SagaData, error) {
	db.sagaDataLock.RLock()
	for _, d := range db.sagaData {
		if d.EntityID == entityID {
			dCopy := copySagaData(d)
			db.sagaDataLock.RUnlock()
			return dCopy, nil
		}
	}
	db.sagaDataLock.RUnlock()
	return nil, ErrSagaEntityNotFound
}

func (db *MemoryDatabase) GetSideEffectDataByEntityID(entityID int) (*SideEffectData, error) {
	db.sideEffectDataLock.RLock()
	for _, d := range db.sideEffectData {
		if d.EntityID == entityID {
			dCopy := copySideEffectData(d)
			db.sideEffectDataLock.RUnlock()
			return dCopy, nil
		}
	}
	db.sideEffectDataLock.RUnlock()
	return nil, ErrSideEffectEntityNotFound
}

func (db *MemoryDatabase) AddWorkflowExecutionData(executionID int, data *WorkflowExecutionData) (int, error) {
	db.workflowExecDataLock.Lock()
	db.workflowExecutionDataCounter++
	data.ID = db.workflowExecutionDataCounter
	data.ExecutionID = executionID
	db.workflowExecutionData[data.ID] = copyWorkflowExecutionData(data)
	db.workflowExecDataLock.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) AddActivityExecutionData(executionID int, data *ActivityExecutionData) (int, error) {
	db.activityExecDataLock.Lock()
	db.activityExecutionDataCounter++
	data.ID = db.activityExecutionDataCounter
	data.ExecutionID = executionID
	db.activityExecutionData[data.ID] = copyActivityExecutionData(data)
	db.activityExecDataLock.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) AddSagaExecutionData(executionID int, data *SagaExecutionData) (int, error) {
	db.sagaExecDataLock.Lock()
	db.sagaExecutionDataCounter++
	data.ID = db.sagaExecutionDataCounter
	data.ExecutionID = executionID
	db.sagaExecutionData[data.ID] = copySagaExecutionData(data)
	db.sagaExecDataLock.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) AddSideEffectExecutionData(executionID int, data *SideEffectExecutionData) (int, error) {
	db.sideEffectExecDataLock.Lock()
	db.sideEffectExecutionDataCounter++
	data.ID = db.sideEffectExecutionDataCounter
	data.ExecutionID = executionID
	db.sideEffectExecutionData[data.ID] = copySideEffectExecutionData(data)
	db.sideEffectExecDataLock.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) GetWorkflowExecutionData(id int) (*WorkflowExecutionData, error) {
	db.workflowExecDataLock.RLock()
	d, exists := db.workflowExecutionData[id]
	if !exists {
		db.workflowExecDataLock.RUnlock()
		return nil, ErrWorkflowExecutionNotFound
	}
	dCopy := copyWorkflowExecutionData(d)
	db.workflowExecDataLock.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetActivityExecutionData(id int) (*ActivityExecutionData, error) {
	db.activityExecDataLock.RLock()
	d, exists := db.activityExecutionData[id]
	if !exists {
		db.activityExecDataLock.RUnlock()
		return nil, ErrActivityExecutionNotFound
	}
	dCopy := copyActivityExecutionData(d)
	db.activityExecDataLock.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetSagaExecutionData(id int) (*SagaExecutionData, error) {
	db.sagaExecDataLock.RLock()
	d, exists := db.sagaExecutionData[id]
	if !exists {
		db.sagaExecDataLock.RUnlock()
		return nil, ErrSagaExecutionNotFound
	}
	dCopy := copySagaExecutionData(d)
	db.sagaExecDataLock.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetSideEffectExecutionData(id int) (*SideEffectExecutionData, error) {
	db.sideEffectExecDataLock.RLock()
	d, exists := db.sideEffectExecutionData[id]
	if !exists {
		db.sideEffectExecDataLock.RUnlock()
		return nil, ErrSideEffectExecutionNotFound
	}
	dCopy := copySideEffectExecutionData(d)
	db.sideEffectExecDataLock.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetWorkflowExecutionDataByExecutionID(executionID int) (*WorkflowExecutionData, error) {
	db.workflowExecDataLock.RLock()
	for _, d := range db.workflowExecutionData {
		if d.ExecutionID == executionID {
			dCopy := copyWorkflowExecutionData(d)
			db.workflowExecDataLock.RUnlock()
			return dCopy, nil
		}
	}
	db.workflowExecDataLock.RUnlock()
	return nil, ErrWorkflowExecutionNotFound
}

func (db *MemoryDatabase) GetActivityExecutionDataByExecutionID(executionID int) (*ActivityExecutionData, error) {
	db.activityExecDataLock.RLock()
	for _, d := range db.activityExecutionData {
		if d.ExecutionID == executionID {
			dCopy := copyActivityExecutionData(d)
			db.activityExecDataLock.RUnlock()
			return dCopy, nil
		}
	}
	db.activityExecDataLock.RUnlock()
	return nil, ErrActivityExecutionNotFound
}

func (db *MemoryDatabase) GetSagaExecutionDataByExecutionID(executionID int) (*SagaExecutionData, error) {
	db.sagaExecDataLock.RLock()
	for _, d := range db.sagaExecutionData {
		if d.ExecutionID == executionID {
			dCopy := copySagaExecutionData(d)
			db.sagaExecDataLock.RUnlock()
			return dCopy, nil
		}
	}
	db.sagaExecDataLock.RUnlock()
	return nil, ErrSagaExecutionNotFound
}

func (db *MemoryDatabase) GetSideEffectExecutionDataByExecutionID(executionID int) (*SideEffectExecutionData, error) {
	db.sideEffectExecDataLock.RLock()
	for _, d := range db.sideEffectExecutionData {
		if d.ExecutionID == executionID {
			dCopy := copySideEffectExecutionData(d)
			db.sideEffectExecDataLock.RUnlock()
			return dCopy, nil
		}
	}
	db.sideEffectExecDataLock.RUnlock()
	return nil, ErrSideEffectExecutionNotFound
}

func (db *MemoryDatabase) HasRun(id int) (bool, error) {
	db.runLock.RLock()
	_, exists := db.runs[id]
	db.runLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasVersion(id int) (bool, error) {
	db.versionLock.RLock()
	_, exists := db.versions[id]
	db.versionLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasHierarchy(id int) (bool, error) {
	db.hierarchyLock.RLock()
	_, exists := db.hierarchies[id]
	db.hierarchyLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasQueue(id int) (bool, error) {
	db.queueLock.RLock()
	_, exists := db.queues[id]
	db.queueLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasQueueName(name string) (bool, error) {
	db.queueLock.RLock()
	_, exists := db.queueNames[name]
	db.queueLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowEntity(id int) (bool, error) {
	db.workflowLock.RLock()
	_, exists := db.workflowEntities[id]
	db.workflowLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasActivityEntity(id int) (bool, error) {
	db.activityLock.RLock()
	_, exists := db.activityEntities[id]
	db.activityLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSagaEntity(id int) (bool, error) {
	db.sagaLock.RLock()
	_, exists := db.sagaEntities[id]
	db.sagaLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectEntity(id int) (bool, error) {
	db.sideEffectLock.RLock()
	_, exists := db.sideEffectEntities[id]
	db.sideEffectLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowExecution(id int) (bool, error) {
	db.workflowExecLock.RLock()
	_, exists := db.workflowExecutions[id]
	db.workflowExecLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasActivityExecution(id int) (bool, error) {
	db.activityExecLock.RLock()
	_, exists := db.activityExecutions[id]
	db.activityExecLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSagaExecution(id int) (bool, error) {
	db.sagaExecLock.RLock()
	_, exists := db.sagaExecutions[id]
	db.sagaExecLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectExecution(id int) (bool, error) {
	db.sideEffectExecLock.RLock()
	_, exists := db.sideEffectExecutions[id]
	db.sideEffectExecLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowData(id int) (bool, error) {
	db.workflowDataLock.RLock()
	_, exists := db.workflowData[id]
	db.workflowDataLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasActivityData(id int) (bool, error) {
	db.activityDataLock.RLock()
	_, exists := db.activityData[id]
	db.activityDataLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSagaData(id int) (bool, error) {
	db.sagaDataLock.RLock()
	_, exists := db.sagaData[id]
	db.sagaDataLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectData(id int) (bool, error) {
	db.sideEffectDataLock.RLock()
	_, exists := db.sideEffectData[id]
	db.sideEffectDataLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowDataByEntityID(entityID int) (bool, error) {
	db.workflowDataLock.RLock()
	for _, d := range db.workflowData {
		if d.EntityID == entityID {
			db.workflowDataLock.RUnlock()
			return true, nil
		}
	}
	db.workflowDataLock.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) HasActivityDataByEntityID(entityID int) (bool, error) {
	db.activityDataLock.RLock()
	for _, d := range db.activityData {
		if d.EntityID == entityID {
			db.activityDataLock.RUnlock()
			return true, nil
		}
	}
	db.activityDataLock.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) HasSagaDataByEntityID(entityID int) (bool, error) {
	db.sagaDataLock.RLock()
	for _, d := range db.sagaData {
		if d.EntityID == entityID {
			db.sagaDataLock.RUnlock()
			return true, nil
		}
	}
	db.sagaDataLock.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) HasSideEffectDataByEntityID(entityID int) (bool, error) {
	db.sideEffectDataLock.RLock()
	for _, d := range db.sideEffectData {
		if d.EntityID == entityID {
			db.sideEffectDataLock.RUnlock()
			return true, nil
		}
	}
	db.sideEffectDataLock.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) HasWorkflowExecutionData(id int) (bool, error) {
	db.workflowExecDataLock.RLock()
	_, exists := db.workflowExecutionData[id]
	db.workflowExecDataLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasActivityExecutionData(id int) (bool, error) {
	db.activityExecDataLock.RLock()
	_, exists := db.activityExecutionData[id]
	db.activityExecDataLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSagaExecutionData(id int) (bool, error) {
	db.sagaExecDataLock.RLock()
	_, exists := db.sagaExecutionData[id]
	db.sagaExecDataLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectExecutionData(id int) (bool, error) {
	db.sideEffectExecDataLock.RLock()
	_, exists := db.sideEffectExecutionData[id]
	db.sideEffectExecDataLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowExecutionDataByExecutionID(executionID int) (bool, error) {
	db.workflowExecDataLock.RLock()
	for _, d := range db.workflowExecutionData {
		if d.ExecutionID == executionID {
			db.workflowExecDataLock.RUnlock()
			return true, nil
		}
	}
	db.workflowExecDataLock.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) HasActivityExecutionDataByExecutionID(executionID int) (bool, error) {
	db.activityExecDataLock.RLock()
	for _, d := range db.activityExecutionData {
		if d.ExecutionID == executionID {
			db.activityExecDataLock.RUnlock()
			return true, nil
		}
	}
	db.activityExecDataLock.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) HasSagaExecutionDataByExecutionID(executionID int) (bool, error) {
	db.sagaExecDataLock.RLock()
	for _, d := range db.sagaExecutionData {
		if d.ExecutionID == executionID {
			db.sagaExecDataLock.RUnlock()
			return true, nil
		}
	}
	db.sagaExecDataLock.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) HasSideEffectExecutionDataByExecutionID(executionID int) (bool, error) {
	db.sideEffectExecDataLock.RLock()
	for _, d := range db.sideEffectExecutionData {
		if d.ExecutionID == executionID {
			db.sideEffectExecDataLock.RUnlock()
			return true, nil
		}
	}
	db.sideEffectExecDataLock.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) GetWorkflowExecutionLatestByEntityID(entityID int) (*WorkflowExecution, error) {
	db.workflowExecLock.RLock()
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
	db.workflowExecLock.RUnlock()

	if latestExec == nil {
		return nil, ErrWorkflowExecutionNotFound
	}
	return copyWorkflowExecution(latestExec), nil
}

func (db *MemoryDatabase) GetActivityExecutionLatestByEntityID(entityID int) (*ActivityExecution, error) {
	db.activityExecLock.RLock()
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
	db.activityExecLock.RUnlock()

	if latestExec == nil {
		return nil, ErrActivityExecutionNotFound
	}
	return copyActivityExecution(latestExec), nil
}

func (db *MemoryDatabase) SetSagaValue(executionID int, key string, value []byte) (int, error) {
	db.sagaLock.Lock()
	defer db.sagaLock.Unlock()

	if _, exists := db.sagaValues[executionID]; !exists {
		db.sagaValues[executionID] = make(map[string]*SagaValue)
	}

	sv := &SagaValue{
		ID:          len(db.sagaValues[executionID]) + 1,
		ExecutionID: executionID,
		Key:         key,
		Value:       value,
	}
	db.sagaValues[executionID][key] = sv
	return sv.ID, nil
}

func (db *MemoryDatabase) GetSagaValue(id int, key string) ([]byte, error) {
	db.sagaLock.RLock()
	defer db.sagaLock.RUnlock()

	for _, keyMap := range db.sagaValues {
		if val, exists := keyMap[key]; exists && val.ID == id {
			return val.Value, nil
		}
	}
	return nil, ErrSagaValueNotFound
}

func (db *MemoryDatabase) GetSagaValueByExecutionID(executionID int, key string) ([]byte, error) {
	db.sagaLock.RLock()
	defer db.sagaLock.RUnlock()

	if keyMap, exists := db.sagaValues[executionID]; exists {
		if val, vexists := keyMap[key]; vexists {
			return val.Value, nil
		}
	}
	return nil, ErrSagaValueNotFound
}

func (db *MemoryDatabase) AddSignalEntity(entity *SignalEntity, parentWorkflowID int) (int, error) {
	db.signalLock.Lock()
	db.signalEntityCounter++
	entity.ID = db.signalEntityCounter

	if entity.SignalData != nil {
		db.signalDataLock.Lock()
		db.signalDataCounter++
		entity.SignalData.ID = db.signalDataCounter
		entity.SignalData.EntityID = entity.ID
		db.signalData[entity.SignalData.ID] = copySignalData(entity.SignalData)
		db.signalDataLock.Unlock()
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = entity.CreatedAt

	db.signalEntities[entity.ID] = copySignalEntity(entity)
	db.signalLock.Unlock()

	// Signal relationships to workflow are through hierarchies,
	// but the code provided doesn't explicitly add them here.
	return entity.ID, nil
}

func (db *MemoryDatabase) GetSignalEntity(id int, opts ...SignalEntityGetOption) (*SignalEntity, error) {
	db.signalLock.RLock()
	entity, exists := db.signalEntities[id]
	if !exists {
		db.signalLock.RUnlock()
		return nil, fmt.Errorf("signal entity not found")
	}
	eCopy := copySignalEntity(entity)
	db.signalLock.RUnlock()

	cfg := &SignalEntityGetterOptions{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		db.signalDataLock.RLock()
		for _, d := range db.signalData {
			if d.EntityID == id {
				eCopy.SignalData = copySignalData(d)
				break
			}
		}
		db.signalDataLock.RUnlock()
	}

	return eCopy, nil
}

func (db *MemoryDatabase) GetSignalEntities(workflowID int, opts ...SignalEntityGetOption) ([]*SignalEntity, error) {
	// Signals linked via hierarchies
	db.hierarchyLock.RLock()
	var entities []*SignalEntity
	for _, h := range db.hierarchies {
		if h.ParentEntityID == workflowID && h.ChildType == EntitySignal {
			db.signalLock.RLock()
			if e, ex := db.signalEntities[h.ChildEntityID]; ex {
				entities = append(entities, copySignalEntity(e))
			}
			db.signalLock.RUnlock()
		}
	}
	db.hierarchyLock.RUnlock()

	return entities, nil
}

func (db *MemoryDatabase) AddSignalExecution(exec *SignalExecution) (int, error) {
	db.signalExecLock.Lock()
	db.signalExecutionCounter++
	exec.ID = db.signalExecutionCounter

	if exec.SignalExecutionData != nil {
		db.signalExecDataLock.Lock()
		db.signalExecutionDataCounter++
		exec.SignalExecutionData.ID = db.signalExecutionDataCounter
		exec.SignalExecutionData.ExecutionID = exec.ID
		db.signalExecutionData[exec.SignalExecutionData.ID] = copySignalExecutionData(exec.SignalExecutionData)
		db.signalExecDataLock.Unlock()
	}

	exec.CreatedAt = time.Now()
	exec.UpdatedAt = exec.CreatedAt

	db.signalExecutions[exec.ID] = copySignalExecution(exec)
	db.signalExecLock.Unlock()
	return exec.ID, nil
}

func (db *MemoryDatabase) GetSignalExecution(id int, opts ...SignalExecutionGetOption) (*SignalExecution, error) {
	db.signalExecLock.RLock()
	exec, exists := db.signalExecutions[id]
	if !exists {
		db.signalExecLock.RUnlock()
		return nil, fmt.Errorf("signal execution not found")
	}
	execCopy := copySignalExecution(exec)
	db.signalExecLock.RUnlock()

	cfg := &SignalExecutionGetterOptions{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		db.signalExecDataLock.RLock()
		for _, d := range db.signalExecutionData {
			if d.ExecutionID == id {
				execCopy.SignalExecutionData = copySignalExecutionData(d)
				break
			}
		}
		db.signalExecDataLock.RUnlock()
	}

	return execCopy, nil
}

func (db *MemoryDatabase) GetSignalExecutions(entityID int) ([]*SignalExecution, error) {
	db.signalExecLock.RLock()
	var executions []*SignalExecution
	for _, exec := range db.signalExecutions {
		if exec.EntityID == entityID {
			executions = append(executions, copySignalExecution(exec))
		}
	}
	db.signalExecLock.RUnlock()
	return executions, nil
}

func (db *MemoryDatabase) GetSignalExecutionLatestByEntityID(entityID int) (*SignalExecution, error) {
	db.signalExecLock.RLock()
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
	db.signalExecLock.RUnlock()

	if latest == nil {
		return nil, fmt.Errorf("no signal execution found")
	}

	return copySignalExecution(latest), nil
}

func (db *MemoryDatabase) HasSignalEntity(id int) (bool, error) {
	db.signalLock.RLock()
	_, exists := db.signalEntities[id]
	db.signalLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSignalExecution(id int) (bool, error) {
	db.signalExecLock.RLock()
	_, exists := db.signalExecutions[id]
	db.signalExecLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSignalData(id int) (bool, error) {
	db.signalDataLock.RLock()
	_, exists := db.signalData[id]
	db.signalDataLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSignalExecutionData(id int) (bool, error) {
	db.signalExecDataLock.RLock()
	_, exists := db.signalExecutionData[id]
	db.signalExecDataLock.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) GetSignalEntityProperties(id int, getters ...SignalEntityPropertyGetter) error {
	db.signalLock.RLock()
	entity, exists := db.signalEntities[id]
	if !exists {
		db.signalLock.RUnlock()
		return fmt.Errorf("signal entity not found")
	}
	eCopy := copySignalEntity(entity)
	db.signalLock.RUnlock()

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
				db.signalDataLock.RLock()
				for _, d := range db.signalData {
					if d.EntityID == id {
						eCopy.SignalData = copySignalData(d)
						break
					}
				}
				db.signalDataLock.RUnlock()
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSignalEntityProperties(id int, setters ...SignalEntityPropertySetter) error {
	db.signalLock.Lock()
	entity, exists := db.signalEntities[id]
	if !exists {
		db.signalLock.Unlock()
		return fmt.Errorf("signal entity not found")
	}

	for _, setter := range setters {
		if _, err := setter(entity); err != nil {
			db.signalLock.Unlock()
			return err
		}
	}
	entity.UpdatedAt = time.Now()
	db.signalEntities[id] = copySignalEntity(entity)
	db.signalLock.Unlock()

	return nil
}

func (db *MemoryDatabase) GetSignalExecutionProperties(id int, getters ...SignalExecutionPropertyGetter) error {
	db.signalExecLock.RLock()
	exec, exists := db.signalExecutions[id]
	if !exists {
		db.signalExecLock.RUnlock()
		return fmt.Errorf("signal execution not found")
	}
	execCopy := copySignalExecution(exec)
	db.signalExecLock.RUnlock()

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
				db.signalExecDataLock.RLock()
				for _, d := range db.signalExecutionData {
					if d.ExecutionID == id {
						execCopy.SignalExecutionData = copySignalExecutionData(d)
						break
					}
				}
				db.signalExecDataLock.RUnlock()
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSignalExecutionProperties(id int, setters ...SignalExecutionPropertySetter) error {
	db.signalExecLock.Lock()
	exec, exists := db.signalExecutions[id]
	if !exists {
		db.signalExecLock.Unlock()
		return fmt.Errorf("signal execution not found")
	}

	for _, setter := range setters {
		if _, err := setter(exec); err != nil {
			db.signalExecLock.Unlock()
			return err
		}
	}
	exec.UpdatedAt = time.Now()
	db.signalExecutions[id] = copySignalExecution(exec)
	db.signalExecLock.Unlock()
	return nil
}

func (db *MemoryDatabase) AddSignalData(entityID int, data *SignalData) (int, error) {
	db.signalDataLock.Lock()
	db.signalDataCounter++
	data.ID = db.signalDataCounter
	data.EntityID = entityID
	db.signalData[data.ID] = copySignalData(data)
	db.signalDataLock.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) GetSignalData(id int) (*SignalData, error) {
	db.signalDataLock.RLock()
	d, exists := db.signalData[id]
	if !exists {
		db.signalDataLock.RUnlock()
		return nil, fmt.Errorf("signal data not found")
	}
	dCopy := copySignalData(d)
	db.signalDataLock.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetSignalDataByEntityID(entityID int) (*SignalData, error) {
	db.signalDataLock.RLock()
	for _, d := range db.signalData {
		if d.EntityID == entityID {
			dCopy := copySignalData(d)
			db.signalDataLock.RUnlock()
			return dCopy, nil
		}
	}
	db.signalDataLock.RUnlock()
	return nil, fmt.Errorf("signal data not found for entity")
}

func (db *MemoryDatabase) HasSignalDataByEntityID(entityID int) (bool, error) {
	db.signalDataLock.RLock()
	for _, d := range db.signalData {
		if d.EntityID == entityID {
			db.signalDataLock.RUnlock()
			return true, nil
		}
	}
	db.signalDataLock.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) GetSignalDataProperties(entityID int, getters ...SignalDataPropertyGetter) error {
	db.signalDataLock.RLock()
	var data *SignalData
	for _, d := range db.signalData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}
	db.signalDataLock.RUnlock()

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

func (db *MemoryDatabase) SetSignalDataProperties(entityID int, setters ...SignalDataPropertySetter) error {
	db.signalDataLock.Lock()
	var data *SignalData
	for _, d := range db.signalData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}
	if data == nil {
		db.signalDataLock.Unlock()
		return fmt.Errorf("signal data not found")
	}

	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			db.signalDataLock.Unlock()
			return err
		}
	}
	db.signalDataLock.Unlock()
	return nil
}

func (db *MemoryDatabase) AddSignalExecutionData(executionID int, data *SignalExecutionData) (int, error) {
	db.signalExecDataLock.Lock()
	db.signalExecutionDataCounter++
	data.ID = db.signalExecutionDataCounter
	data.ExecutionID = executionID
	db.signalExecutionData[data.ID] = copySignalExecutionData(data)
	db.signalExecDataLock.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) GetSignalExecutionData(id int) (*SignalExecutionData, error) {
	db.signalExecDataLock.RLock()
	d, exists := db.signalExecutionData[id]
	if !exists {
		db.signalExecDataLock.RUnlock()
		return nil, fmt.Errorf("signal execution data not found")
	}
	dCopy := copySignalExecutionData(d)
	db.signalExecDataLock.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetSignalExecutionDataByExecutionID(executionID int) (*SignalExecutionData, error) {
	db.signalExecDataLock.RLock()
	for _, d := range db.signalExecutionData {
		if d.ExecutionID == executionID {
			dCopy := copySignalExecutionData(d)
			db.signalExecDataLock.RUnlock()
			return dCopy, nil
		}
	}
	db.signalExecDataLock.RUnlock()
	return nil, fmt.Errorf("signal execution data not found for execution")
}

func (db *MemoryDatabase) HasSignalExecutionDataByExecutionID(executionID int) (bool, error) {
	db.signalExecDataLock.RLock()
	for _, d := range db.signalExecutionData {
		if d.ExecutionID == executionID {
			db.signalExecDataLock.RUnlock()
			return true, nil
		}
	}
	db.signalExecDataLock.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) GetSignalExecutionDataProperties(entityID int, getters ...SignalExecutionDataPropertyGetter) error {
	db.signalExecDataLock.RLock()
	var data *SignalExecutionData
	for _, d := range db.signalExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}
	db.signalExecDataLock.RUnlock()

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

func (db *MemoryDatabase) SetSignalExecutionDataProperties(entityID int, setters ...SignalExecutionDataPropertySetter) error {
	db.signalExecDataLock.Lock()
	var data *SignalExecutionData
	for _, d := range db.signalExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}
	if data == nil {
		db.signalExecDataLock.Unlock()
		return fmt.Errorf("signal execution data not found")
	}

	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			db.signalExecDataLock.Unlock()
			return err
		}
	}
	db.signalExecDataLock.Unlock()
	return nil
}

func (db *MemoryDatabase) UpdateSignalEntity(entity *SignalEntity) error {
	db.signalLock.Lock()
	defer db.signalLock.Unlock()

	if _, exists := db.signalEntities[entity.ID]; !exists {
		return fmt.Errorf("signal entity not found")
	}

	entity.UpdatedAt = time.Now()
	db.signalEntities[entity.ID] = copySignalEntity(entity)
	return nil
}

// DeleteRunsByStatus - lock and delete in a safe manner
func (db *MemoryDatabase) DeleteRunsByStatus(status RunStatus) error {
	// Identify runs
	db.runLock.RLock()
	runsToDelete := make([]int, 0)
	for _, run := range db.runs {
		if run.Status == status {
			runsToDelete = append(runsToDelete, run.ID)
		}
	}
	db.runLock.RUnlock()

	if len(runsToDelete) == 0 {
		return nil
	}

	for _, runID := range runsToDelete {
		db.relationshipLock.RLock()
		workflowIDs := make([]int, len(db.runToWorkflows[runID]))
		copy(workflowIDs, db.runToWorkflows[runID])
		db.relationshipLock.RUnlock()

		for _, wfID := range workflowIDs {
			if err := db.deleteWorkflowAndChildren(wfID); err != nil {
				return fmt.Errorf("failed to delete workflow %d: %w", wfID, err)
			}
		}

		db.runLock.Lock()
		delete(db.runs, runID)
		db.runLock.Unlock()

		db.relationshipLock.Lock()
		delete(db.runToWorkflows, runID)
		db.relationshipLock.Unlock()
	}

	return nil
}

func (db *MemoryDatabase) deleteWorkflowAndChildren(workflowID int) error {
	db.hierarchyLock.RLock()
	var hierarchies []*Hierarchy
	for _, h := range db.hierarchies {
		if h.ParentEntityID == workflowID {
			hierarchies = append(hierarchies, copyHierarchy(h))
		}
	}
	db.hierarchyLock.RUnlock()

	for _, h := range hierarchies {
		switch h.ChildType {
		case EntityActivity:
			if err := db.deleteActivityEntity(h.ChildEntityID); err != nil {
				return err
			}
		case EntitySaga:
			if err := db.deleteSagaEntity(h.ChildEntityID); err != nil {
				return err
			}
		case EntitySideEffect:
			if err := db.deleteSideEffectEntity(h.ChildEntityID); err != nil {
				return err
			}
		case EntitySignal:
			if err := db.deleteSignalEntity(h.ChildEntityID); err != nil {
				return err
			}
		}
	}

	if err := db.deleteWorkflowEntity(workflowID); err != nil {
		return err
	}

	db.hierarchyLock.Lock()
	for id, hh := range db.hierarchies {
		if hh.ParentEntityID == workflowID || hh.ChildEntityID == workflowID {
			delete(db.hierarchies, id)
		}
	}
	db.hierarchyLock.Unlock()

	return nil
}

// deleteWorkflowEntity
func (db *MemoryDatabase) deleteWorkflowEntity(workflowID int) error {
	// Delete workflow execution data
	db.workflowExecLock.Lock()
	for id, exec := range db.workflowExecutions {
		if exec.EntityID == workflowID {
			db.workflowExecDataLock.Lock()
			delete(db.workflowExecutionData, db.execToDataMap[exec.ID])
			db.workflowExecDataLock.Unlock()
			delete(db.workflowExecutions, id)
			db.relationshipLock.Lock()
			delete(db.execToDataMap, exec.ID)
			db.relationshipLock.Unlock()
		}
	}
	db.workflowExecLock.Unlock()

	// Delete workflow data
	db.workflowDataLock.Lock()
	for id, data := range db.workflowData {
		if data.EntityID == workflowID {
			delete(db.workflowData, id)
		}
	}
	db.workflowDataLock.Unlock()

	// Delete versions
	db.relationshipLock.Lock()
	versionIDs := db.workflowToVersion[workflowID]
	delete(db.workflowToVersion, workflowID)
	delete(db.workflowVersions, workflowID)
	db.relationshipLock.Unlock()

	db.versionLock.Lock()
	for _, vID := range versionIDs {
		delete(db.versions, vID)
	}
	db.versionLock.Unlock()

	// Clean up queue mappings
	db.relationshipLock.Lock()
	if queueID, ok := db.workflowToQueue[workflowID]; ok {
		if workflows, exists := db.queueToWorkflows[queueID]; exists {
			newWorkflows := make([]int, 0, len(workflows)-1)
			for _, wID := range workflows {
				if wID != workflowID {
					newWorkflows = append(newWorkflows, wID)
				}
			}
			db.queueToWorkflows[queueID] = newWorkflows
		}
		delete(db.workflowToQueue, workflowID)
	}
	db.relationshipLock.Unlock()

	db.workflowLock.Lock()
	delete(db.workflowEntities, workflowID)
	db.workflowLock.Unlock()

	return nil
}

func (db *MemoryDatabase) deleteActivityEntity(entityID int) error {
	db.activityExecLock.Lock()
	for id, exec := range db.activityExecutions {
		if exec.EntityID == entityID {
			db.activityExecDataLock.Lock()
			delete(db.activityExecutionData, exec.ID)
			db.activityExecDataLock.Unlock()
			delete(db.activityExecutions, id)
		}
	}
	db.activityExecLock.Unlock()

	db.activityDataLock.Lock()
	for id, data := range db.activityData {
		if data.EntityID == entityID {
			delete(db.activityData, id)
		}
	}
	db.activityDataLock.Unlock()

	db.activityLock.Lock()
	delete(db.activityEntities, entityID)
	db.activityLock.Unlock()

	return nil
}

func (db *MemoryDatabase) deleteSagaEntity(entityID int) error {
	db.sagaExecLock.Lock()
	for id, exec := range db.sagaExecutions {
		if exec.EntityID == entityID {
			db.sagaExecDataLock.Lock()
			delete(db.sagaExecutionData, exec.ID)
			db.sagaExecDataLock.Unlock()

			db.sagaLock.Lock()
			delete(db.sagaValues, exec.ID)
			db.sagaLock.Unlock()

			delete(db.sagaExecutions, id)
		}
	}
	db.sagaExecLock.Unlock()

	db.sagaDataLock.Lock()
	for id, data := range db.sagaData {
		if data.EntityID == entityID {
			delete(db.sagaData, id)
		}
	}
	db.sagaDataLock.Unlock()

	db.sagaLock.Lock()
	delete(db.sagaEntities, entityID)
	db.sagaLock.Unlock()

	return nil
}

func (db *MemoryDatabase) deleteSideEffectEntity(entityID int) error {
	db.sideEffectExecLock.Lock()
	for id, exec := range db.sideEffectExecutions {
		if exec.EntityID == entityID {
			db.sideEffectExecDataLock.Lock()
			delete(db.sideEffectExecutionData, exec.ID)
			db.sideEffectExecDataLock.Unlock()
			delete(db.sideEffectExecutions, id)
		}
	}
	db.sideEffectExecLock.Unlock()

	db.sideEffectDataLock.Lock()
	for id, data := range db.sideEffectData {
		if data.EntityID == entityID {
			delete(db.sideEffectData, id)
		}
	}
	db.sideEffectDataLock.Unlock()

	db.sideEffectLock.Lock()
	delete(db.sideEffectEntities, entityID)
	db.sideEffectLock.Unlock()

	return nil
}

func (db *MemoryDatabase) deleteSignalEntity(entityID int) error {
	db.signalExecLock.Lock()
	for id, exec := range db.signalExecutions {
		if exec.EntityID == entityID {
			db.signalExecDataLock.Lock()
			delete(db.signalExecutionData, exec.ID)
			db.signalExecDataLock.Unlock()
			delete(db.signalExecutions, id)
		}
	}
	db.signalExecLock.Unlock()

	db.signalDataLock.Lock()
	for id, data := range db.signalData {
		if data.EntityID == entityID {
			delete(db.signalData, id)
		}
	}
	db.signalDataLock.Unlock()

	db.signalLock.Lock()
	delete(db.signalEntities, entityID)
	db.signalLock.Unlock()

	return nil
}

func (db *MemoryDatabase) GetRunsPaginated(page, pageSize int, filter *RunFilter, sortCriteria *RunSort) (*PaginatedRuns, error) {
	db.runLock.RLock()
	matchingRuns := make([]*Run, 0)
	for _, run := range db.runs {
		if filter != nil {
			if run.Status != filter.Status {
				continue
			}
		}
		matchingRuns = append(matchingRuns, copyRun(run))
	}
	db.runLock.RUnlock()

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
