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
	rwMutexRuns        deadlock.RWMutex
	rwMutexVersions    deadlock.RWMutex
	rwMutexHierarchies deadlock.RWMutex
	rwMutexQueues      deadlock.RWMutex
	rwMutexWorkflows   deadlock.RWMutex
	rwMutexActivities  deadlock.RWMutex
	rwMutexSagas       deadlock.RWMutex
	rwMutexSideEffects deadlock.RWMutex
	rwMutexSignals     deadlock.RWMutex

	rwMutexWorkflowData   deadlock.RWMutex
	rwMutexActivityData   deadlock.RWMutex
	rwMutexSagaData       deadlock.RWMutex
	rwMutexSideEffectData deadlock.RWMutex
	rwMutexSignalData     deadlock.RWMutex

	rwMutexWorkflowExec   deadlock.RWMutex
	rwMutexActivityExec   deadlock.RWMutex
	rwMutexSagaExec       deadlock.RWMutex
	rwMutexSideEffectExec deadlock.RWMutex
	rwMutexSignalExec     deadlock.RWMutex

	rwMutexWorkflowExecData   deadlock.RWMutex
	rwMutexActivityExecData   deadlock.RWMutex
	rwMutexSagaExecData       deadlock.RWMutex
	rwMutexSideEffectExecData deadlock.RWMutex
	rwMutexSignalExecData     deadlock.RWMutex

	rwMutexRelationships deadlock.RWMutex
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
	db.rwMutexQueues.Lock()
	db.queues[1] = &Queue{
		ID:        1,
		Name:      DefaultQueue,
		CreatedAt: now,
		UpdatedAt: now,
		Entities:  make([]*WorkflowEntity, 0),
	}
	db.queueNames[DefaultQueue] = 1
	db.rwMutexQueues.Unlock()

	return db
}

func (db *MemoryDatabase) SaveAsJSON(path string) error {
	// We need to lock all relevant structures for a consistent snapshot.
	// We'll lock everything in alphabetical order to avoid deadlocks.
	db.rwMutexActivityData.RLock()
	db.rwMutexActivityExec.RLock()
	db.rwMutexActivities.RLock()
	db.rwMutexHierarchies.RLock()
	db.rwMutexQueues.RLock()
	db.rwMutexRelationships.RLock()
	db.rwMutexRuns.RLock()
	db.rwMutexSagaData.RLock()
	db.rwMutexSagaExec.RLock()
	db.rwMutexSagas.RLock()
	db.rwMutexSideEffectData.RLock()
	db.rwMutexSideEffectExec.RLock()
	db.rwMutexSideEffects.RLock()
	db.rwMutexSignalData.RLock()
	db.rwMutexSignalExec.RLock()
	db.rwMutexSignals.RLock()
	db.rwMutexVersions.RLock()
	db.rwMutexWorkflowData.RLock()
	db.rwMutexWorkflowExec.RLock()
	db.rwMutexWorkflows.RLock()

	defer db.rwMutexWorkflows.RUnlock()
	defer db.rwMutexWorkflowExec.RUnlock()
	defer db.rwMutexWorkflowData.RUnlock()
	defer db.rwMutexVersions.RUnlock()
	defer db.rwMutexSignals.RUnlock()
	defer db.rwMutexSignalExec.RUnlock()
	defer db.rwMutexSignalData.RUnlock()
	defer db.rwMutexSideEffects.RUnlock()
	defer db.rwMutexSideEffectExec.RUnlock()
	defer db.rwMutexSideEffectData.RUnlock()
	defer db.rwMutexSagas.RUnlock()
	defer db.rwMutexSagaExec.RUnlock()
	defer db.rwMutexSagaData.RUnlock()
	defer db.rwMutexRuns.RUnlock()
	defer db.rwMutexRelationships.RUnlock()
	defer db.rwMutexQueues.RUnlock()
	defer db.rwMutexHierarchies.RUnlock()
	defer db.rwMutexActivities.RUnlock()
	defer db.rwMutexActivityExec.RUnlock()
	defer db.rwMutexActivityData.RUnlock()

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
	db.rwMutexRuns.Lock()
	defer db.rwMutexRuns.Unlock()

	db.runCounter++
	run.ID = db.runCounter
	run.CreatedAt = time.Now()
	run.UpdatedAt = run.CreatedAt

	db.runs[run.ID] = copyRun(run)
	return run.ID, nil
}

// GetRun
func (db *MemoryDatabase) GetRun(id int, opts ...RunGetOption) (*Run, error) {
	db.rwMutexRuns.RLock()
	r, exists := db.runs[id]
	if !exists {
		db.rwMutexRuns.RUnlock()
		return nil, ErrRunNotFound
	}
	runCopy := copyRun(r) // Copy to avoid race after unlocking
	db.rwMutexRuns.RUnlock()

	cfg := &RunGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeWorkflows {
		db.rwMutexRelationships.RLock()
		workflowIDs := db.runToWorkflows[id]
		db.rwMutexRelationships.RUnlock()

		if len(workflowIDs) > 0 {
			entities := make([]*WorkflowEntity, 0, len(workflowIDs))
			db.rwMutexWorkflows.RLock()
			for _, wfID := range workflowIDs {
				if wf, ok := db.workflowEntities[wfID]; ok {
					entities = append(entities, copyWorkflowEntity(wf))
				}
			}
			db.rwMutexWorkflows.RUnlock()
			runCopy.Entities = entities
		}
	}

	if cfg.IncludeHierarchies {
		db.rwMutexHierarchies.RLock()
		hierarchies := make([]*Hierarchy, 0)
		for _, h := range db.hierarchies {
			if h.RunID == id {
				hierarchies = append(hierarchies, copyHierarchy(h))
			}
		}
		db.rwMutexHierarchies.RUnlock()
		runCopy.Hierarchies = hierarchies
	}

	return runCopy, nil
}

func (db *MemoryDatabase) UpdateRun(run *Run) error {
	db.rwMutexRuns.Lock()
	defer db.rwMutexRuns.Unlock()

	if _, exists := db.runs[run.ID]; !exists {
		return ErrRunNotFound
	}

	run.UpdatedAt = time.Now()
	db.runs[run.ID] = copyRun(run)
	return nil
}

func (db *MemoryDatabase) GetRunProperties(id int, getters ...RunPropertyGetter) error {
	db.rwMutexRuns.RLock()
	run, exists := db.runs[id]
	if !exists {
		db.rwMutexRuns.RUnlock()
		return ErrRunNotFound
	}
	runCopy := copyRun(run)
	db.rwMutexRuns.RUnlock()

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
		db.rwMutexRelationships.RLock()
		workflowIDs := db.runToWorkflows[id]
		db.rwMutexRelationships.RUnlock()

		if len(workflowIDs) > 0 {
			entities := make([]*WorkflowEntity, 0, len(workflowIDs))
			db.rwMutexWorkflows.RLock()
			for _, wfID := range workflowIDs {
				if wf, ok := db.workflowEntities[wfID]; ok {
					entities = append(entities, copyWorkflowEntity(wf))
				}
			}
			db.rwMutexWorkflows.RUnlock()
			runCopy.Entities = entities
		}
	}

	if opts.IncludeHierarchies {
		db.rwMutexHierarchies.RLock()
		hierarchies := make([]*Hierarchy, 0)
		for _, h := range db.hierarchies {
			if h.RunID == id {
				hierarchies = append(hierarchies, copyHierarchy(h))
			}
		}
		db.rwMutexHierarchies.RUnlock()
		runCopy.Hierarchies = hierarchies
	}

	return nil
}

func (db *MemoryDatabase) SetRunProperties(id int, setters ...RunPropertySetter) error {
	db.rwMutexRuns.Lock()
	run, exists := db.runs[id]
	if !exists {
		db.rwMutexRuns.Unlock()
		return ErrRunNotFound
	}

	opts := &RunSetterOptions{}
	for _, setter := range setters {
		opt, err := setter(run)
		if err != nil {
			db.rwMutexRuns.Unlock()
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				db.rwMutexRuns.Unlock()
				return err
			}
		}
	}

	db.rwMutexRuns.Unlock()

	if opts.WorkflowID != nil {
		db.rwMutexWorkflows.RLock()
		_, wfExists := db.workflowEntities[*opts.WorkflowID]
		db.rwMutexWorkflows.RUnlock()
		if !wfExists {
			return ErrWorkflowEntityNotFound
		}

		db.rwMutexRelationships.Lock()
		db.runToWorkflows[id] = append(db.runToWorkflows[id], *opts.WorkflowID)
		db.rwMutexRelationships.Unlock()
	}

	db.rwMutexRuns.Lock()
	run.UpdatedAt = time.Now()
	db.runs[id] = copyRun(run)
	db.rwMutexRuns.Unlock()

	return nil
}

func (db *MemoryDatabase) AddQueue(queue *Queue) (int, error) {
	db.rwMutexQueues.Lock()
	defer db.rwMutexQueues.Unlock()

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
	db.rwMutexQueues.RLock()
	q, exists := db.queues[id]
	if !exists {
		db.rwMutexQueues.RUnlock()
		return nil, ErrQueueNotFound
	}
	queueCopy := copyQueue(q)
	db.rwMutexQueues.RUnlock()

	cfg := &QueueGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeWorkflows {
		db.rwMutexRelationships.RLock()
		workflowIDs := db.queueToWorkflows[id]
		db.rwMutexRelationships.RUnlock()

		if len(workflowIDs) > 0 {
			db.rwMutexWorkflows.RLock()
			workflows := make([]*WorkflowEntity, 0, len(workflowIDs))
			for _, wfID := range workflowIDs {
				if wf, exists := db.workflowEntities[wfID]; exists {
					workflows = append(workflows, copyWorkflowEntity(wf))
				}
			}
			db.rwMutexWorkflows.RUnlock()
			queueCopy.Entities = workflows
		}
	}

	return queueCopy, nil
}

func (db *MemoryDatabase) GetQueueByName(name string, opts ...QueueGetOption) (*Queue, error) {
	db.rwMutexQueues.RLock()
	id, exists := db.queueNames[name]
	if !exists {
		db.rwMutexQueues.RUnlock()
		return nil, ErrQueueNotFound
	}
	q, qexists := db.queues[id]
	if !qexists {
		db.rwMutexQueues.RUnlock()
		return nil, ErrQueueNotFound
	}
	queueCopy := copyQueue(q)
	db.rwMutexQueues.RUnlock()

	cfg := &QueueGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeWorkflows {
		db.rwMutexRelationships.RLock()
		workflowIDs := db.queueToWorkflows[id]
		db.rwMutexRelationships.RUnlock()

		if len(workflowIDs) > 0 {
			db.rwMutexWorkflows.RLock()
			workflows := make([]*WorkflowEntity, 0, len(workflowIDs))
			for _, wfID := range workflowIDs {
				if wf, exists := db.workflowEntities[wfID]; exists {
					workflows = append(workflows, copyWorkflowEntity(wf))
				}
			}
			db.rwMutexWorkflows.RUnlock()
			queueCopy.Entities = workflows
		}
	}

	return queueCopy, nil
}

func (db *MemoryDatabase) AddVersion(version *Version) (int, error) {
	db.rwMutexVersions.Lock()
	defer db.rwMutexVersions.Unlock()

	db.versionCounter++
	version.ID = db.versionCounter
	version.CreatedAt = time.Now()
	version.UpdatedAt = version.CreatedAt

	db.versions[version.ID] = copyVersion(version)

	if version.EntityID != 0 {
		db.rwMutexRelationships.Lock()
		db.workflowToVersion[version.EntityID] = append(db.workflowToVersion[version.EntityID], version.ID)
		db.rwMutexRelationships.Unlock()
	}
	return version.ID, nil
}

func (db *MemoryDatabase) GetVersion(id int, opts ...VersionGetOption) (*Version, error) {
	db.rwMutexVersions.RLock()
	v, exists := db.versions[id]
	if !exists {
		db.rwMutexVersions.RUnlock()
		return nil, ErrVersionNotFound
	}
	versionCopy := copyVersion(v)
	db.rwMutexVersions.RUnlock()

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
	db.rwMutexHierarchies.Lock()
	defer db.rwMutexHierarchies.Unlock()

	db.hierarchyCounter++
	hierarchy.ID = db.hierarchyCounter

	db.hierarchies[hierarchy.ID] = copyHierarchy(hierarchy)
	return hierarchy.ID, nil
}

func (db *MemoryDatabase) GetHierarchy(id int, opts ...HierarchyGetOption) (*Hierarchy, error) {
	db.rwMutexHierarchies.RLock()
	h, exists := db.hierarchies[id]
	if !exists {
		db.rwMutexHierarchies.RUnlock()
		return nil, ErrHierarchyNotFound
	}
	hCopy := copyHierarchy(h)
	db.rwMutexHierarchies.RUnlock()

	cfg := &HierarchyGetterOptions{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	return hCopy, nil
}

func (db *MemoryDatabase) AddWorkflowEntity(entity *WorkflowEntity) (int, error) {
	db.rwMutexWorkflows.Lock()
	db.workflowEntityCounter++
	entity.ID = db.workflowEntityCounter

	if entity.WorkflowData != nil {
		db.rwMutexWorkflowData.Lock()
		db.workflowDataCounter++
		entity.WorkflowData.ID = db.workflowDataCounter
		entity.WorkflowData.EntityID = entity.ID
		db.workflowData[entity.WorkflowData.ID] = copyWorkflowData(entity.WorkflowData)
		db.rwMutexWorkflowData.Unlock()
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = entity.CreatedAt

	// Default queue if none
	db.rwMutexQueues.RLock()
	defaultQueueID := 1
	db.rwMutexQueues.RUnlock()
	if entity.QueueID == 0 {
		entity.QueueID = defaultQueueID
		db.rwMutexRelationships.Lock()
		db.queueToWorkflows[defaultQueueID] = append(db.queueToWorkflows[defaultQueueID], entity.ID)
		db.rwMutexRelationships.Unlock()
	} else {
		db.rwMutexRelationships.Lock()
		db.workflowToQueue[entity.ID] = entity.QueueID
		db.queueToWorkflows[entity.QueueID] = append(db.queueToWorkflows[entity.QueueID], entity.ID)
		db.rwMutexRelationships.Unlock()
	}

	db.workflowEntities[entity.ID] = copyWorkflowEntity(entity)
	db.rwMutexWorkflows.Unlock()
	return entity.ID, nil
}

func (db *MemoryDatabase) AddWorkflowExecution(exec *WorkflowExecution) (int, error) {
	db.rwMutexWorkflowExec.Lock()
	db.rwMutexWorkflowExecData.Lock()
	defer db.rwMutexWorkflowExec.Unlock()
	defer db.rwMutexWorkflowExecData.Unlock()

	db.workflowExecutionCounter++
	exec.ID = db.workflowExecutionCounter

	if exec.WorkflowExecutionData != nil {
		db.workflowExecutionDataCounter++
		exec.WorkflowExecutionData.ID = db.workflowExecutionDataCounter
		exec.WorkflowExecutionData.ExecutionID = exec.ID
		db.workflowExecutionData[exec.WorkflowExecutionData.ID] = copyWorkflowExecutionData(exec.WorkflowExecutionData)

		db.rwMutexRelationships.Lock()
		db.execToDataMap[exec.ID] = exec.WorkflowExecutionData.ID
		db.rwMutexRelationships.Unlock()
	}
	exec.CreatedAt = time.Now()
	exec.UpdatedAt = exec.CreatedAt

	db.workflowExecutions[exec.ID] = copyWorkflowExecution(exec)
	return exec.ID, nil
}

func (db *MemoryDatabase) GetWorkflowEntity(id int, opts ...WorkflowEntityGetOption) (*WorkflowEntity, error) {
	db.rwMutexWorkflows.RLock()
	e, exists := db.workflowEntities[id]
	if !exists {
		db.rwMutexWorkflows.RUnlock()
		return nil, ErrWorkflowEntityNotFound
	}
	entityCopy := copyWorkflowEntity(e)
	db.rwMutexWorkflows.RUnlock()

	cfg := &WorkflowEntityGetterOptions{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeQueue {
		db.rwMutexRelationships.RLock()
		queueID, hasQueue := db.workflowToQueue[id]
		db.rwMutexRelationships.RUnlock()
		if hasQueue {
			db.rwMutexQueues.RLock()
			if q, qexists := db.queues[queueID]; qexists {
				if entityCopy.Edges == nil {
					entityCopy.Edges = &WorkflowEntityEdges{}
				}
				entityCopy.Edges.Queue = copyQueue(q)
			}
			db.rwMutexQueues.RUnlock()
		}
	}

	if cfg.IncludeData {
		db.rwMutexWorkflowData.RLock()
		for _, d := range db.workflowData {
			if d.EntityID == id {
				entityCopy.WorkflowData = copyWorkflowData(d)
				break
			}
		}
		db.rwMutexWorkflowData.RUnlock()
	}

	return entityCopy, nil
}

func (db *MemoryDatabase) GetWorkflowEntityProperties(id int, getters ...WorkflowEntityPropertyGetter) error {
	db.rwMutexWorkflows.RLock()
	entity, exists := db.workflowEntities[id]
	if !exists {
		db.rwMutexWorkflows.RUnlock()
		return ErrWorkflowEntityNotFound
	}
	entityCopy := copyWorkflowEntity(entity)
	db.rwMutexWorkflows.RUnlock()

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
		db.rwMutexRelationships.RLock()
		versionIDs := db.workflowToVersion[id]
		db.rwMutexRelationships.RUnlock()

		if len(versionIDs) > 0 {
			db.rwMutexVersions.RLock()
			versions := make([]*Version, 0, len(versionIDs))
			for _, vID := range versionIDs {
				if v, vexists := db.versions[vID]; vexists {
					versions = append(versions, copyVersion(v))
				}
			}
			db.rwMutexVersions.RUnlock()

			if entityCopy.Edges == nil {
				entityCopy.Edges = &WorkflowEntityEdges{}
			}
			entityCopy.Edges.Versions = versions
		}
	}

	if opts.IncludeQueue {
		db.rwMutexRelationships.RLock()
		queueID, qexists := db.workflowToQueue[id]
		db.rwMutexRelationships.RUnlock()
		if qexists {
			db.rwMutexQueues.RLock()
			if q, qfound := db.queues[queueID]; qfound {
				if entityCopy.Edges == nil {
					entityCopy.Edges = &WorkflowEntityEdges{}
				}
				entityCopy.Edges.Queue = copyQueue(q)
			}
			db.rwMutexQueues.RUnlock()
		}
	}

	if opts.IncludeChildren {
		db.rwMutexRelationships.RLock()
		childMap, cexists := db.workflowToChildren[id]
		db.rwMutexRelationships.RUnlock()
		if cexists {
			if entityCopy.Edges == nil {
				entityCopy.Edges = &WorkflowEntityEdges{}
			}

			if activityIDs, ok := childMap[EntityActivity]; ok {
				db.rwMutexActivities.RLock()
				activities := make([]*ActivityEntity, 0, len(activityIDs))
				for _, aID := range activityIDs {
					if a, aexists := db.activityEntities[aID]; aexists {
						activities = append(activities, copyActivityEntity(a))
					}
				}
				db.rwMutexActivities.RUnlock()
				entityCopy.Edges.ActivityChildren = activities
			}

			if sagaIDs, ok := childMap[EntitySaga]; ok {
				db.rwMutexSagas.RLock()
				sagas := make([]*SagaEntity, 0, len(sagaIDs))
				for _, sID := range sagaIDs {
					if s, sexists := db.sagaEntities[sID]; sexists {
						sagas = append(sagas, copySagaEntity(s))
					}
				}
				db.rwMutexSagas.RUnlock()
				entityCopy.Edges.SagaChildren = sagas
			}

			if sideEffectIDs, ok := childMap[EntitySideEffect]; ok {
				db.rwMutexSideEffects.RLock()
				sideEffects := make([]*SideEffectEntity, 0, len(sideEffectIDs))
				for _, seID := range sideEffectIDs {
					if se, seexists := db.sideEffectEntities[seID]; seexists {
						sideEffects = append(sideEffects, copySideEffectEntity(se))
					}
				}
				db.rwMutexSideEffects.RUnlock()
				entityCopy.Edges.SideEffectChildren = sideEffects
			}
		}
	}

	if opts.IncludeData {
		db.rwMutexWorkflowData.RLock()
		for _, d := range db.workflowData {
			if d.EntityID == id {
				entityCopy.WorkflowData = copyWorkflowData(d)
				break
			}
		}
		db.rwMutexWorkflowData.RUnlock()
	}

	return nil
}

func (db *MemoryDatabase) SetWorkflowEntityProperties(id int, setters ...WorkflowEntityPropertySetter) error {
	db.rwMutexWorkflows.Lock()
	entity, exists := db.workflowEntities[id]
	if !exists {
		db.rwMutexWorkflows.Unlock()
		return ErrWorkflowEntityNotFound
	}

	opts := &WorkflowEntitySetterOptions{}
	for _, setter := range setters {
		opt, err := setter(entity)
		if err != nil {
			db.rwMutexWorkflows.Unlock()
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				db.rwMutexWorkflows.Unlock()
				return err
			}
		}
	}
	db.rwMutexWorkflows.Unlock()

	if opts.QueueID != nil {
		db.rwMutexQueues.RLock()
		_, qexists := db.queues[*opts.QueueID]
		db.rwMutexQueues.RUnlock()
		if !qexists {
			return ErrQueueNotFound
		}

		db.rwMutexRelationships.Lock()
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
		db.rwMutexRelationships.Unlock()
	}

	if opts.Version != nil {
		db.rwMutexRelationships.Lock()
		db.workflowToVersion[id] = append(db.workflowToVersion[id], opts.Version.ID)
		db.rwMutexRelationships.Unlock()
	}

	if opts.ChildID != nil && opts.ChildType != nil {
		db.rwMutexRelationships.Lock()
		if db.workflowToChildren[id] == nil {
			db.workflowToChildren[id] = make(map[EntityType][]int)
		}
		db.workflowToChildren[id][*opts.ChildType] = append(db.workflowToChildren[id][*opts.ChildType], *opts.ChildID)
		db.entityToWorkflow[*opts.ChildID] = id
		db.rwMutexRelationships.Unlock()
	}

	db.rwMutexWorkflows.Lock()
	entity.UpdatedAt = time.Now()
	db.workflowEntities[id] = copyWorkflowEntity(entity)
	db.rwMutexWorkflows.Unlock()

	return nil
}

func (db *MemoryDatabase) AddActivityEntity(entity *ActivityEntity, parentWorkflowID int) (int, error) {
	db.rwMutexActivityData.Lock()
	db.rwMutexActivities.Lock()
	db.rwMutexRelationships.Lock()
	defer db.rwMutexRelationships.Unlock()
	defer db.rwMutexActivities.Unlock()
	defer db.rwMutexActivityData.Unlock()
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

	if _, ok := db.workflowToChildren[parentWorkflowID]; !ok {
		db.workflowToChildren[parentWorkflowID] = make(map[EntityType][]int)
	}
	db.workflowToChildren[parentWorkflowID][EntityActivity] = append(db.workflowToChildren[parentWorkflowID][EntityActivity], entity.ID)

	return entity.ID, nil
}

func (db *MemoryDatabase) AddActivityExecution(exec *ActivityExecution) (int, error) {
	db.rwMutexActivityExec.Lock()
	db.activityExecutionCounter++
	exec.ID = db.activityExecutionCounter

	if exec.ActivityExecutionData != nil {
		db.rwMutexActivityExecData.Lock()
		db.activityExecutionDataCounter++
		exec.ActivityExecutionData.ID = db.activityExecutionDataCounter
		exec.ActivityExecutionData.ExecutionID = exec.ID
		db.activityExecutionData[exec.ActivityExecutionData.ID] = copyActivityExecutionData(exec.ActivityExecutionData)
		db.rwMutexActivityExecData.Unlock()
	}

	exec.CreatedAt = time.Now()
	exec.UpdatedAt = exec.CreatedAt

	db.activityExecutions[exec.ID] = copyActivityExecution(exec)
	db.rwMutexActivityExec.Unlock()
	return exec.ID, nil
}

func (db *MemoryDatabase) GetActivityEntity(id int, opts ...ActivityEntityGetOption) (*ActivityEntity, error) {
	db.rwMutexActivities.RLock()
	entity, exists := db.activityEntities[id]
	if !exists {
		db.rwMutexActivities.RUnlock()
		return nil, ErrActivityEntityNotFound
	}
	eCopy := copyActivityEntity(entity)
	db.rwMutexActivities.RUnlock()

	cfg := &ActivityEntityGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		db.rwMutexActivityData.RLock()
		for _, d := range db.activityData {
			if d.EntityID == id {
				eCopy.ActivityData = copyActivityData(d)
				break
			}
		}
		db.rwMutexActivityData.RUnlock()
	}

	return eCopy, nil
}

func (db *MemoryDatabase) GetActivityEntities(workflowID int, opts ...ActivityEntityGetOption) ([]*ActivityEntity, error) {
	db.rwMutexRelationships.RLock()
	activityIDs := db.workflowToChildren[workflowID][EntityActivity]
	db.rwMutexRelationships.RUnlock()

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
	db.rwMutexActivities.RLock()
	entity, exists := db.activityEntities[id]
	if !exists {
		db.rwMutexActivities.RUnlock()
		return ErrActivityEntityNotFound
	}
	eCopy := copyActivityEntity(entity)
	db.rwMutexActivities.RUnlock()

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
		db.rwMutexActivityData.RLock()
		for _, d := range db.activityData {
			if d.EntityID == id {
				eCopy.ActivityData = copyActivityData(d)
				break
			}
		}
		db.rwMutexActivityData.RUnlock()
	}

	return nil
}

func (db *MemoryDatabase) SetActivityEntityProperties(id int, setters ...ActivityEntityPropertySetter) error {
	db.rwMutexActivities.Lock()
	entity, exists := db.activityEntities[id]
	if !exists {
		db.rwMutexActivities.Unlock()
		return ErrActivityEntityNotFound
	}

	opts := &ActivityEntitySetterOptions{}
	for _, setter := range setters {
		opt, err := setter(entity)
		if err != nil {
			db.rwMutexActivities.Unlock()
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				db.rwMutexActivities.Unlock()
				return err
			}
		}
	}
	db.rwMutexActivities.Unlock()

	if opts.ParentWorkflowID != nil {
		db.rwMutexWorkflows.RLock()
		_, wExists := db.workflowEntities[*opts.ParentWorkflowID]
		db.rwMutexWorkflows.RUnlock()
		if !wExists {
			return ErrWorkflowEntityNotFound
		}

		db.rwMutexRelationships.Lock()
		db.entityToWorkflow[id] = *opts.ParentWorkflowID
		if db.workflowToChildren[*opts.ParentWorkflowID] == nil {
			db.workflowToChildren[*opts.ParentWorkflowID] = make(map[EntityType][]int)
		}
		db.workflowToChildren[*opts.ParentWorkflowID][EntityActivity] = append(db.workflowToChildren[*opts.ParentWorkflowID][EntityActivity], id)
		db.rwMutexRelationships.Unlock()
	}

	db.rwMutexActivities.Lock()
	entity.UpdatedAt = time.Now()
	db.activityEntities[id] = copyActivityEntity(entity)
	db.rwMutexActivities.Unlock()

	return nil
}

func (db *MemoryDatabase) AddSagaEntity(entity *SagaEntity, parentWorkflowID int) (int, error) {
	db.rwMutexSagas.Lock()
	db.sagaEntityCounter++
	entity.ID = db.sagaEntityCounter

	if entity.SagaData != nil {
		db.rwMutexSagaData.Lock()
		db.sagaDataCounter++
		entity.SagaData.ID = db.sagaDataCounter
		entity.SagaData.EntityID = entity.ID
		db.sagaData[entity.SagaData.ID] = copySagaData(entity.SagaData)
		db.rwMutexSagaData.Unlock()
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = entity.CreatedAt

	db.sagaEntities[entity.ID] = copySagaEntity(entity)
	db.rwMutexSagas.Unlock()

	db.rwMutexRelationships.Lock()
	if _, ok := db.workflowToChildren[parentWorkflowID]; !ok {
		db.workflowToChildren[parentWorkflowID] = make(map[EntityType][]int)
	}
	db.workflowToChildren[parentWorkflowID][EntitySaga] = append(db.workflowToChildren[parentWorkflowID][EntitySaga], entity.ID)
	db.rwMutexRelationships.Unlock()

	return entity.ID, nil
}

func (db *MemoryDatabase) GetSagaEntities(workflowID int, opts ...SagaEntityGetOption) ([]*SagaEntity, error) {
	db.rwMutexRelationships.RLock()
	sagaIDs := db.workflowToChildren[workflowID][EntitySaga]
	db.rwMutexRelationships.RUnlock()

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
	db.rwMutexSagaExec.Lock()
	db.sagaExecutionCounter++
	exec.ID = db.sagaExecutionCounter

	if exec.SagaExecutionData != nil {
		db.rwMutexSagaExecData.Lock()
		db.sagaExecutionDataCounter++
		exec.SagaExecutionData.ID = db.sagaExecutionDataCounter
		exec.SagaExecutionData.ExecutionID = exec.ID
		db.sagaExecutionData[exec.SagaExecutionData.ID] = copySagaExecutionData(exec.SagaExecutionData)
		db.rwMutexSagaExecData.Unlock()
	}

	exec.CreatedAt = time.Now()
	exec.UpdatedAt = exec.CreatedAt

	db.sagaExecutions[exec.ID] = copySagaExecution(exec)
	db.rwMutexSagaExec.Unlock()
	return exec.ID, nil
}

func (db *MemoryDatabase) GetSagaEntity(id int, opts ...SagaEntityGetOption) (*SagaEntity, error) {
	db.rwMutexSagas.RLock()
	entity, exists := db.sagaEntities[id]
	if !exists {
		db.rwMutexSagas.RUnlock()
		return nil, ErrSagaEntityNotFound
	}
	eCopy := copySagaEntity(entity)
	db.rwMutexSagas.RUnlock()

	cfg := &SagaEntityGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		db.rwMutexSagaData.RLock()
		for _, d := range db.sagaData {
			if d.EntityID == id {
				eCopy.SagaData = copySagaData(d)
				break
			}
		}
		db.rwMutexSagaData.RUnlock()
	}

	return eCopy, nil
}

func (db *MemoryDatabase) GetSagaEntityProperties(id int, getters ...SagaEntityPropertyGetter) error {
	db.rwMutexSagas.RLock()
	entity, exists := db.sagaEntities[id]
	if !exists {
		db.rwMutexSagas.RUnlock()
		return ErrSagaEntityNotFound
	}
	eCopy := copySagaEntity(entity)
	db.rwMutexSagas.RUnlock()

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
		db.rwMutexSagaData.RLock()
		for _, d := range db.sagaData {
			if d.EntityID == id {
				eCopy.SagaData = copySagaData(d)
				break
			}
		}
		db.rwMutexSagaData.RUnlock()
	}

	return nil
}

func (db *MemoryDatabase) SetSagaEntityProperties(id int, setters ...SagaEntityPropertySetter) error {
	db.rwMutexSagas.Lock()
	entity, exists := db.sagaEntities[id]
	if !exists {
		db.rwMutexSagas.Unlock()
		return ErrSagaEntityNotFound
	}

	opts := &SagaEntitySetterOptions{}
	for _, setter := range setters {
		opt, err := setter(entity)
		if err != nil {
			db.rwMutexSagas.Unlock()
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				db.rwMutexSagas.Unlock()
				return err
			}
		}
	}
	db.rwMutexSagas.Unlock()

	if opts.ParentWorkflowID != nil {
		db.rwMutexWorkflows.RLock()
		_, wExists := db.workflowEntities[*opts.ParentWorkflowID]
		db.rwMutexWorkflows.RUnlock()
		if !wExists {
			return ErrWorkflowEntityNotFound
		}

		db.rwMutexRelationships.Lock()
		db.entityToWorkflow[id] = *opts.ParentWorkflowID
		if db.workflowToChildren[*opts.ParentWorkflowID] == nil {
			db.workflowToChildren[*opts.ParentWorkflowID] = make(map[EntityType][]int)
		}
		db.workflowToChildren[*opts.ParentWorkflowID][EntitySaga] = append(db.workflowToChildren[*opts.ParentWorkflowID][EntitySaga], id)
		db.rwMutexRelationships.Unlock()
	}

	db.rwMutexSagas.Lock()
	entity.UpdatedAt = time.Now()
	db.sagaEntities[id] = copySagaEntity(entity)
	db.rwMutexSagas.Unlock()

	return nil
}

func (db *MemoryDatabase) AddSideEffectEntity(entity *SideEffectEntity, parentWorkflowID int) (int, error) {
	db.rwMutexSideEffects.Lock()
	db.sideEffectEntityCounter++
	entity.ID = db.sideEffectEntityCounter

	if entity.SideEffectData != nil {
		db.rwMutexSideEffectData.Lock()
		db.sideEffectDataCounter++
		entity.SideEffectData.ID = db.sideEffectDataCounter
		entity.SideEffectData.EntityID = entity.ID
		db.sideEffectData[entity.SideEffectData.ID] = copySideEffectData(entity.SideEffectData)
		db.rwMutexSideEffectData.Unlock()
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = entity.CreatedAt

	db.sideEffectEntities[entity.ID] = copySideEffectEntity(entity)
	db.rwMutexSideEffects.Unlock()

	db.rwMutexRelationships.Lock()
	if _, ok := db.workflowToChildren[parentWorkflowID]; !ok {
		db.workflowToChildren[parentWorkflowID] = make(map[EntityType][]int)
	}
	db.workflowToChildren[parentWorkflowID][EntitySideEffect] = append(db.workflowToChildren[parentWorkflowID][EntitySideEffect], entity.ID)
	db.rwMutexRelationships.Unlock()

	return entity.ID, nil
}

func (db *MemoryDatabase) GetSideEffectEntities(workflowID int, opts ...SideEffectEntityGetOption) ([]*SideEffectEntity, error) {
	db.rwMutexRelationships.RLock()
	sideEffectIDs := db.workflowToChildren[workflowID][EntitySideEffect]
	db.rwMutexRelationships.RUnlock()

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
	db.rwMutexSideEffectExec.Lock()
	db.sideEffectExecutionCounter++
	exec.ID = db.sideEffectExecutionCounter

	if exec.SideEffectExecutionData != nil {
		db.rwMutexSideEffectExecData.Lock()
		db.sideEffectExecutionDataCounter++
		exec.SideEffectExecutionData.ID = db.sideEffectExecutionDataCounter
		exec.SideEffectExecutionData.ExecutionID = exec.ID
		db.sideEffectExecutionData[exec.SideEffectExecutionData.ID] = copySideEffectExecutionData(exec.SideEffectExecutionData)
		db.rwMutexSideEffectExecData.Unlock()
	}

	exec.CreatedAt = time.Now()
	exec.UpdatedAt = exec.CreatedAt

	db.sideEffectExecutions[exec.ID] = copySideEffectExecution(exec)
	db.rwMutexSideEffectExec.Unlock()
	return exec.ID, nil
}

func (db *MemoryDatabase) GetSideEffectEntity(id int, opts ...SideEffectEntityGetOption) (*SideEffectEntity, error) {
	db.rwMutexSideEffects.RLock()
	entity, exists := db.sideEffectEntities[id]
	if !exists {
		db.rwMutexSideEffects.RUnlock()
		return nil, ErrSideEffectEntityNotFound
	}
	eCopy := copySideEffectEntity(entity)
	db.rwMutexSideEffects.RUnlock()

	cfg := &SideEffectEntityGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		db.rwMutexSideEffectData.RLock()
		for _, d := range db.sideEffectData {
			if d.EntityID == id {
				eCopy.SideEffectData = copySideEffectData(d)
				break
			}
		}
		db.rwMutexSideEffectData.RUnlock()
	}

	return eCopy, nil
}

func (db *MemoryDatabase) GetSideEffectEntityProperties(id int, getters ...SideEffectEntityPropertyGetter) error {
	db.rwMutexSideEffects.RLock()
	entity, exists := db.sideEffectEntities[id]
	if !exists {
		db.rwMutexSideEffects.RUnlock()
		return ErrSideEffectEntityNotFound
	}
	eCopy := copySideEffectEntity(entity)
	db.rwMutexSideEffects.RUnlock()

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
		db.rwMutexSideEffectData.RLock()
		for _, d := range db.sideEffectData {
			if d.EntityID == id {
				eCopy.SideEffectData = copySideEffectData(d)
				break
			}
		}
		db.rwMutexSideEffectData.RUnlock()
	}

	return nil
}

func (db *MemoryDatabase) SetSideEffectEntityProperties(id int, setters ...SideEffectEntityPropertySetter) error {
	db.rwMutexSideEffects.Lock()
	entity, exists := db.sideEffectEntities[id]
	if !exists {
		db.rwMutexSideEffects.Unlock()
		return ErrSideEffectEntityNotFound
	}

	opts := &SideEffectEntitySetterOptions{}
	for _, setter := range setters {
		opt, err := setter(entity)
		if err != nil {
			db.rwMutexSideEffects.Unlock()
			return err
		}
		if opt != nil {
			if err := opt(opts); err != nil {
				db.rwMutexSideEffects.Unlock()
				return err
			}
		}
	}
	db.rwMutexSideEffects.Unlock()

	if opts.ParentWorkflowID != nil {
		db.rwMutexWorkflows.RLock()
		_, wExists := db.workflowEntities[*opts.ParentWorkflowID]
		db.rwMutexWorkflows.RUnlock()
		if !wExists {
			return ErrWorkflowEntityNotFound
		}

		db.rwMutexRelationships.Lock()
		db.entityToWorkflow[id] = *opts.ParentWorkflowID
		if db.workflowToChildren[*opts.ParentWorkflowID] == nil {
			db.workflowToChildren[*opts.ParentWorkflowID] = make(map[EntityType][]int)
		}
		db.workflowToChildren[*opts.ParentWorkflowID][EntitySideEffect] = append(db.workflowToChildren[*opts.ParentWorkflowID][EntitySideEffect], id)
		db.rwMutexRelationships.Unlock()
	}

	db.rwMutexSideEffects.Lock()
	entity.UpdatedAt = time.Now()
	db.sideEffectEntities[id] = copySideEffectEntity(entity)
	db.rwMutexSideEffects.Unlock()

	return nil
}

func (db *MemoryDatabase) GetWorkflowExecution(id int, opts ...WorkflowExecutionGetOption) (*WorkflowExecution, error) {
	db.rwMutexWorkflowExec.RLock()
	exec, exists := db.workflowExecutions[id]
	if !exists {
		db.rwMutexWorkflowExec.RUnlock()
		return nil, ErrWorkflowExecutionNotFound
	}
	execCopy := copyWorkflowExecution(exec)
	db.rwMutexWorkflowExec.RUnlock()

	cfg := &WorkflowExecutionGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		db.rwMutexWorkflowExecData.RLock()
		for _, d := range db.workflowExecutionData {
			if d.ExecutionID == id {
				execCopy.WorkflowExecutionData = copyWorkflowExecutionData(d)
				break
			}
		}
		db.rwMutexWorkflowExecData.RUnlock()
	}

	return execCopy, nil
}

func (db *MemoryDatabase) GetWorkflowExecutions(entityID int) ([]*WorkflowExecution, error) {
	db.rwMutexWorkflowExec.RLock()
	results := make([]*WorkflowExecution, 0)
	for _, exec := range db.workflowExecutions {
		if exec.EntityID == entityID {
			results = append(results, copyWorkflowExecution(exec))
		}
	}
	db.rwMutexWorkflowExec.RUnlock()
	return results, nil
}

func (db *MemoryDatabase) GetActivityExecution(id int, opts ...ActivityExecutionGetOption) (*ActivityExecution, error) {
	db.rwMutexActivityExec.RLock()
	exec, exists := db.activityExecutions[id]
	if !exists {
		db.rwMutexActivityExec.RUnlock()
		return nil, ErrActivityExecutionNotFound
	}
	execCopy := copyActivityExecution(exec)
	db.rwMutexActivityExec.RUnlock()

	cfg := &ActivityExecutionGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		db.rwMutexActivityExecData.RLock()
		for _, d := range db.activityExecutionData {
			if d.ExecutionID == id {
				execCopy.ActivityExecutionData = copyActivityExecutionData(d)
				break
			}
		}
		db.rwMutexActivityExecData.RUnlock()
	}

	return execCopy, nil
}

func (db *MemoryDatabase) GetActivityExecutions(entityID int) ([]*ActivityExecution, error) {
	db.rwMutexActivityExec.RLock()
	results := make([]*ActivityExecution, 0)
	for _, exec := range db.activityExecutions {
		if exec.EntityID == entityID {
			results = append(results, copyActivityExecution(exec))
		}
	}
	db.rwMutexActivityExec.RUnlock()
	return results, nil
}

func (db *MemoryDatabase) GetSagaExecution(id int, opts ...SagaExecutionGetOption) (*SagaExecution, error) {
	db.rwMutexSagaExec.RLock()
	exec, exists := db.sagaExecutions[id]
	if !exists {
		db.rwMutexSagaExec.RUnlock()
		return nil, ErrSagaExecutionNotFound
	}
	execCopy := copySagaExecution(exec)
	db.rwMutexSagaExec.RUnlock()

	cfg := &SagaExecutionGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		db.rwMutexSagaExecData.RLock()
		for _, d := range db.sagaExecutionData {
			if d.ExecutionID == id {
				execCopy.SagaExecutionData = copySagaExecutionData(d)
				break
			}
		}
		db.rwMutexSagaExecData.RUnlock()
	}

	return execCopy, nil
}

func (db *MemoryDatabase) GetSagaExecutions(entityID int) ([]*SagaExecution, error) {
	db.rwMutexSagaExec.RLock()
	results := make([]*SagaExecution, 0)
	for _, exec := range db.sagaExecutions {
		if exec.EntityID == entityID {
			results = append(results, copySagaExecution(exec))
		}
	}
	db.rwMutexSagaExec.RUnlock()

	if len(results) == 0 {
		return nil, ErrSagaExecutionNotFound
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.Before(results[j].CreatedAt)
	})

	return results, nil
}

func (db *MemoryDatabase) GetSideEffectExecution(id int, opts ...SideEffectExecutionGetOption) (*SideEffectExecution, error) {
	db.rwMutexSideEffectExec.RLock()
	exec, exists := db.sideEffectExecutions[id]
	if !exists {
		db.rwMutexSideEffectExec.RUnlock()
		return nil, ErrSideEffectExecutionNotFound
	}
	execCopy := copySideEffectExecution(exec)
	db.rwMutexSideEffectExec.RUnlock()

	cfg := &SideEffectExecutionGetterOptions{}
	for _, v := range opts {
		if err := v(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		db.rwMutexSideEffectExecData.RLock()
		for _, d := range db.sideEffectExecutionData {
			if d.ExecutionID == id {
				execCopy.SideEffectExecutionData = copySideEffectExecutionData(d)
				break
			}
		}
		db.rwMutexSideEffectExecData.RUnlock()
	}

	return execCopy, nil
}

func (db *MemoryDatabase) GetSideEffectExecutions(entityID int) ([]*SideEffectExecution, error) {
	db.rwMutexSideEffectExec.RLock()
	results := make([]*SideEffectExecution, 0)
	for _, exec := range db.sideEffectExecutions {
		if exec.EntityID == entityID {
			results = append(results, copySideEffectExecution(exec))
		}
	}
	db.rwMutexSideEffectExec.RUnlock()
	return results, nil
}

// Activity Data properties
func (db *MemoryDatabase) GetActivityDataProperties(entityID int, getters ...ActivityDataPropertyGetter) error {
	db.rwMutexActivityData.RLock()
	var data *ActivityData
	for _, d := range db.activityData {
		if d.EntityID == entityID {
			tmp := d
			data = tmp
			break
		}
	}
	db.rwMutexActivityData.RUnlock()

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
	db.rwMutexActivityData.Lock()
	defer db.rwMutexActivityData.Unlock()

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
	db.rwMutexSagaData.RLock()
	var data *SagaData
	for _, d := range db.sagaData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}
	db.rwMutexSagaData.RUnlock()

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
	db.rwMutexSagaData.Lock()
	defer db.rwMutexSagaData.Unlock()

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
	db.rwMutexSideEffectData.RLock()
	var data *SideEffectData
	for _, d := range db.sideEffectData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}
	db.rwMutexSideEffectData.RUnlock()

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
	db.rwMutexSideEffectData.Lock()
	defer db.rwMutexSideEffectData.Unlock()

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
	db.rwMutexWorkflowData.RLock()
	var data *WorkflowData
	for _, d := range db.workflowData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}
	db.rwMutexWorkflowData.RUnlock()

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
	db.rwMutexWorkflowData.Lock()
	defer db.rwMutexWorkflowData.Unlock()

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
	db.rwMutexWorkflowExecData.RLock()
	var data *WorkflowExecutionData
	for _, d := range db.workflowExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}
	db.rwMutexWorkflowExecData.RUnlock()

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
	db.rwMutexWorkflowExecData.Lock()
	defer db.rwMutexWorkflowExecData.Unlock()

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
	db.rwMutexActivityExecData.RLock()
	var data *ActivityExecutionData
	for _, d := range db.activityExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}
	db.rwMutexActivityExecData.RUnlock()

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
	db.rwMutexActivityExecData.Lock()
	defer db.rwMutexActivityExecData.Unlock()

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
	db.rwMutexSagaExecData.RLock()
	var data *SagaExecutionData
	for _, d := range db.sagaExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}
	db.rwMutexSagaExecData.RUnlock()

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
	db.rwMutexSagaExecData.Lock()
	defer db.rwMutexSagaExecData.Unlock()

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
	db.rwMutexSideEffectExecData.RLock()
	var data *SideEffectExecutionData
	for _, d := range db.sideEffectExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}
	db.rwMutexSideEffectExecData.RUnlock()

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
	db.rwMutexSideEffectExecData.Lock()
	defer db.rwMutexSideEffectExecData.Unlock()

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
	db.rwMutexWorkflowExec.RLock()
	exec, exists := db.workflowExecutions[id]
	if !exists {
		db.rwMutexWorkflowExec.RUnlock()
		return ErrWorkflowExecutionNotFound
	}
	execCopy := copyWorkflowExecution(exec)
	db.rwMutexWorkflowExec.RUnlock()

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
				db.rwMutexWorkflowExecData.RLock()
				for _, d := range db.workflowExecutionData {
					if d.ExecutionID == id {
						execCopy.WorkflowExecutionData = copyWorkflowExecutionData(d)
						break
					}
				}
				db.rwMutexWorkflowExecData.RUnlock()
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetWorkflowExecutionProperties(id int, setters ...WorkflowExecutionPropertySetter) error {
	db.rwMutexWorkflowExec.Lock()
	exec, exists := db.workflowExecutions[id]
	if !exists {
		db.rwMutexWorkflowExec.Unlock()
		return ErrWorkflowExecutionNotFound
	}
	for _, setter := range setters {
		if _, err := setter(exec); err != nil {
			db.rwMutexWorkflowExec.Unlock()
			return err
		}
	}
	db.workflowExecutions[id] = copyWorkflowExecution(exec)
	db.rwMutexWorkflowExec.Unlock()
	return nil
}

// Activity Execution properties
func (db *MemoryDatabase) GetActivityExecutionProperties(id int, getters ...ActivityExecutionPropertyGetter) error {
	db.rwMutexActivityExec.RLock()
	exec, exists := db.activityExecutions[id]
	if !exists {
		db.rwMutexActivityExec.RUnlock()
		return ErrActivityExecutionNotFound
	}
	execCopy := copyActivityExecution(exec)
	db.rwMutexActivityExec.RUnlock()

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
				db.rwMutexActivityExecData.RLock()
				for _, d := range db.activityExecutionData {
					if d.ExecutionID == id {
						execCopy.ActivityExecutionData = copyActivityExecutionData(d)
						break
					}
				}
				db.rwMutexActivityExecData.RUnlock()
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetActivityExecutionProperties(id int, setters ...ActivityExecutionPropertySetter) error {
	db.rwMutexActivityExec.Lock()
	exec, exists := db.activityExecutions[id]
	if !exists {
		db.rwMutexActivityExec.Unlock()
		return ErrActivityExecutionNotFound
	}
	for _, setter := range setters {
		if _, err := setter(exec); err != nil {
			db.rwMutexActivityExec.Unlock()
			return err
		}
	}
	db.activityExecutions[id] = copyActivityExecution(exec)
	db.rwMutexActivityExec.Unlock()
	return nil
}

// Saga Execution properties
func (db *MemoryDatabase) GetSagaExecutionProperties(id int, getters ...SagaExecutionPropertyGetter) error {
	db.rwMutexSagaExec.RLock()
	exec, exists := db.sagaExecutions[id]
	if !exists {
		db.rwMutexSagaExec.RUnlock()
		return ErrSagaExecutionNotFound
	}
	execCopy := copySagaExecution(exec)
	db.rwMutexSagaExec.RUnlock()

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
				db.rwMutexSagaExecData.RLock()
				for _, d := range db.sagaExecutionData {
					if d.ExecutionID == id {
						execCopy.SagaExecutionData = copySagaExecutionData(d)
						break
					}
				}
				db.rwMutexSagaExecData.RUnlock()
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSagaExecutionProperties(id int, setters ...SagaExecutionPropertySetter) error {
	db.rwMutexSagaExec.Lock()
	exec, exists := db.sagaExecutions[id]
	if !exists {
		db.rwMutexSagaExec.Unlock()
		return ErrSagaExecutionNotFound
	}
	for _, setter := range setters {
		if _, err := setter(exec); err != nil {
			db.rwMutexSagaExec.Unlock()
			return err
		}
	}
	db.sagaExecutions[id] = copySagaExecution(exec)
	db.rwMutexSagaExec.Unlock()
	return nil
}

// SideEffect Execution properties
func (db *MemoryDatabase) GetSideEffectExecutionProperties(id int, getters ...SideEffectExecutionPropertyGetter) error {
	db.rwMutexSideEffectExec.RLock()
	exec, exists := db.sideEffectExecutions[id]
	if !exists {
		db.rwMutexSideEffectExec.RUnlock()
		return ErrSideEffectExecutionNotFound
	}
	execCopy := copySideEffectExecution(exec)
	db.rwMutexSideEffectExec.RUnlock()

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
				db.rwMutexSideEffectExecData.RLock()
				for _, d := range db.sideEffectExecutionData {
					if d.ExecutionID == id {
						execCopy.SideEffectExecutionData = copySideEffectExecutionData(d)
						break
					}
				}
				db.rwMutexSideEffectExecData.RUnlock()
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSideEffectExecutionProperties(id int, setters ...SideEffectExecutionPropertySetter) error {
	db.rwMutexSideEffectExec.Lock()
	exec, exists := db.sideEffectExecutions[id]
	if !exists {
		db.rwMutexSideEffectExec.Unlock()
		return ErrSideEffectExecutionNotFound
	}
	for _, setter := range setters {
		if _, err := setter(exec); err != nil {
			db.rwMutexSideEffectExec.Unlock()
			return err
		}
	}
	db.sideEffectExecutions[id] = copySideEffectExecution(exec)
	db.rwMutexSideEffectExec.Unlock()
	return nil
}

// Hierarchy properties
func (db *MemoryDatabase) GetHierarchyProperties(id int, getters ...HierarchyPropertyGetter) error {
	db.rwMutexHierarchies.RLock()
	hierarchy, exists := db.hierarchies[id]
	if !exists {
		db.rwMutexHierarchies.RUnlock()
		return ErrHierarchyNotFound
	}
	hCopy := copyHierarchy(hierarchy)
	db.rwMutexHierarchies.RUnlock()

	for _, getter := range getters {
		if _, err := getter(hCopy); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) GetHierarchyByParentEntity(parentEntityID int, childStepID string, specificType EntityType) (*Hierarchy, error) {
	db.rwMutexHierarchies.RLock()
	defer db.rwMutexHierarchies.RUnlock()

	for _, h := range db.hierarchies {
		if h.ParentEntityID == parentEntityID && h.ChildStepID == childStepID && h.ChildType == specificType {
			return copyHierarchy(h), nil
		}
	}

	return nil, ErrHierarchyNotFound
}

func (db *MemoryDatabase) GetHierarchiesByParentEntityAndStep(parentEntityID int, childStepID string, specificType EntityType) ([]*Hierarchy, error) {
	db.rwMutexHierarchies.RLock()
	defer db.rwMutexHierarchies.RUnlock()

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
	db.rwMutexHierarchies.Lock()
	hierarchy, exists := db.hierarchies[id]
	if !exists {
		db.rwMutexHierarchies.Unlock()
		return ErrHierarchyNotFound
	}
	for _, setter := range setters {
		if _, err := setter(hierarchy); err != nil {
			db.rwMutexHierarchies.Unlock()
			return err
		}
	}
	db.hierarchies[id] = copyHierarchy(hierarchy)
	db.rwMutexHierarchies.Unlock()
	return nil
}

func (db *MemoryDatabase) GetHierarchiesByParentEntity(parentEntityID int) ([]*Hierarchy, error) {
	db.rwMutexHierarchies.RLock()
	defer db.rwMutexHierarchies.RUnlock()

	var results []*Hierarchy
	for _, h := range db.hierarchies {
		if h.ParentEntityID == parentEntityID {
			results = append(results, copyHierarchy(h))
		}
	}
	return results, nil
}

func (db *MemoryDatabase) GetHierarchiesByChildEntity(childEntityID int) ([]*Hierarchy, error) {
	db.rwMutexHierarchies.RLock()
	defer db.rwMutexHierarchies.RUnlock()

	var results []*Hierarchy
	for _, h := range db.hierarchies {
		if h.ChildEntityID == childEntityID {
			results = append(results, copyHierarchy(h))
		}
	}
	return results, nil
}

func (db *MemoryDatabase) UpdateHierarchy(hierarchy *Hierarchy) error {
	db.rwMutexHierarchies.Lock()
	defer db.rwMutexHierarchies.Unlock()

	if _, exists := db.hierarchies[hierarchy.ID]; !exists {
		return ErrHierarchyNotFound
	}

	db.hierarchies[hierarchy.ID] = copyHierarchy(hierarchy)
	return nil
}

func (db *MemoryDatabase) GetQueueProperties(id int, getters ...QueuePropertyGetter) error {
	db.rwMutexQueues.RLock()
	queue, exists := db.queues[id]
	if !exists {
		db.rwMutexQueues.RUnlock()
		return ErrQueueNotFound
	}
	qCopy := copyQueue(queue)
	db.rwMutexQueues.RUnlock()

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
				db.rwMutexRelationships.RLock()
				workflowIDs := db.queueToWorkflows[id]
				db.rwMutexRelationships.RUnlock()

				if len(workflowIDs) > 0 {
					db.rwMutexWorkflows.RLock()
					workflows := make([]*WorkflowEntity, 0, len(workflowIDs))
					for _, wfID := range workflowIDs {
						if wf, wexists := db.workflowEntities[wfID]; wexists {
							workflows = append(workflows, copyWorkflowEntity(wf))
						}
					}
					db.rwMutexWorkflows.RUnlock()
					qCopy.Entities = workflows
				}
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetQueueProperties(id int, setters ...QueuePropertySetter) error {
	db.rwMutexQueues.Lock()
	queue, exists := db.queues[id]
	if !exists {
		db.rwMutexQueues.Unlock()
		return ErrQueueNotFound
	}
	for _, setter := range setters {
		opt, err := setter(queue)
		if err != nil {
			db.rwMutexQueues.Unlock()
			return err
		}
		if opt != nil {
			opts := &QueueSetterOptions{}
			if err := opt(opts); err != nil {
				db.rwMutexQueues.Unlock()
				return err
			}
			if opts.WorkflowIDs != nil {
				db.rwMutexRelationships.Lock()
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
				db.rwMutexRelationships.Unlock()
			}
		}
	}
	db.queues[id] = copyQueue(queue)
	db.rwMutexQueues.Unlock()
	return nil
}

func (db *MemoryDatabase) UpdateQueue(queue *Queue) error {
	db.rwMutexQueues.Lock()
	defer db.rwMutexQueues.Unlock()

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
	db.rwMutexRelationships.RLock()
	versionIDs := db.workflowVersions[workflowID]
	db.rwMutexRelationships.RUnlock()

	db.rwMutexVersions.RLock()
	defer db.rwMutexVersions.RUnlock()

	for _, vID := range versionIDs {
		if v, vexists := db.versions[vID]; vexists && v.ChangeID == changeID {
			return copyVersion(v), nil
		}
	}
	return nil, ErrVersionNotFound
}

func (db *MemoryDatabase) GetVersionsByWorkflowID(workflowID int) ([]*Version, error) {
	db.rwMutexRelationships.RLock()
	versionIDs := db.workflowVersions[workflowID]
	db.rwMutexRelationships.RUnlock()

	db.rwMutexVersions.RLock()
	defer db.rwMutexVersions.RUnlock()

	versions := make([]*Version, 0, len(versionIDs))
	for _, vID := range versionIDs {
		if version, exists := db.versions[vID]; exists {
			versions = append(versions, copyVersion(version))
		}
	}

	return versions, nil
}

func (db *MemoryDatabase) SetVersion(version *Version) error {
	db.rwMutexVersions.Lock()
	defer db.rwMutexVersions.Unlock()

	if version.ID == 0 {
		db.versionCounter++
		version.ID = db.versionCounter
		version.CreatedAt = time.Now()
	}
	version.UpdatedAt = time.Now()

	db.versions[version.ID] = copyVersion(version)

	db.rwMutexRelationships.Lock()
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
	db.rwMutexRelationships.Unlock()

	return nil
}

func (db *MemoryDatabase) DeleteVersionsForWorkflow(workflowID int) error {
	db.rwMutexVersions.Lock()
	db.rwMutexRelationships.Lock()
	versionIDs, ok := db.workflowVersions[workflowID]
	if ok {
		for _, vID := range versionIDs {
			delete(db.versions, vID)
		}
		delete(db.workflowVersions, workflowID)
	}
	db.rwMutexRelationships.Unlock()
	db.rwMutexVersions.Unlock()
	return nil
}

func (db *MemoryDatabase) GetVersionProperties(id int, getters ...VersionPropertyGetter) error {
	db.rwMutexVersions.RLock()
	version, exists := db.versions[id]
	if !exists {
		db.rwMutexVersions.RUnlock()
		return ErrVersionNotFound
	}
	vCopy := copyVersion(version)
	db.rwMutexVersions.RUnlock()

	for _, getter := range getters {
		if _, err := getter(vCopy); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetVersionProperties(id int, setters ...VersionPropertySetter) error {
	db.rwMutexVersions.Lock()
	version, exists := db.versions[id]
	if !exists {
		db.rwMutexVersions.Unlock()
		return ErrVersionNotFound
	}

	for _, setter := range setters {
		if _, err := setter(version); err != nil {
			db.rwMutexVersions.Unlock()
			return err
		}
	}
	version.UpdatedAt = time.Now()
	db.versions[id] = copyVersion(version)
	db.rwMutexVersions.Unlock()
	return nil
}

func (db *MemoryDatabase) UpdateVersion(version *Version) error {
	db.rwMutexVersions.Lock()
	defer db.rwMutexVersions.Unlock()

	if _, exists := db.versions[version.ID]; !exists {
		return ErrVersionNotFound
	}

	version.UpdatedAt = time.Now()
	db.versions[version.ID] = copyVersion(version)

	return nil
}

func (db *MemoryDatabase) UpdateActivityEntity(entity *ActivityEntity) error {
	db.rwMutexActivities.Lock()
	defer db.rwMutexActivities.Unlock()

	if _, exists := db.activityEntities[entity.ID]; !exists {
		return ErrActivityEntityNotFound
	}

	entity.UpdatedAt = time.Now()
	db.activityEntities[entity.ID] = copyActivityEntity(entity)
	return nil
}

func (db *MemoryDatabase) UpdateSagaEntity(entity *SagaEntity) error {
	db.rwMutexSagas.Lock()
	defer db.rwMutexSagas.Unlock()

	if _, exists := db.sagaEntities[entity.ID]; !exists {
		return ErrSagaEntityNotFound
	}

	entity.UpdatedAt = time.Now()
	db.sagaEntities[entity.ID] = copySagaEntity(entity)
	return nil
}

func (db *MemoryDatabase) UpdateSideEffectEntity(entity *SideEffectEntity) error {
	db.rwMutexSideEffects.Lock()
	defer db.rwMutexSideEffects.Unlock()

	if _, exists := db.sideEffectEntities[entity.ID]; !exists {
		return ErrSideEffectEntityNotFound
	}

	entity.UpdatedAt = time.Now()
	db.sideEffectEntities[entity.ID] = copySideEffectEntity(entity)
	return nil
}

func (db *MemoryDatabase) UpdateWorkflowEntity(entity *WorkflowEntity) error {
	db.rwMutexWorkflows.Lock()
	defer db.rwMutexWorkflows.Unlock()

	if _, exists := db.workflowEntities[entity.ID]; !exists {
		return ErrWorkflowEntityNotFound
	}

	entity.UpdatedAt = time.Now()
	db.workflowEntities[entity.ID] = copyWorkflowEntity(entity)
	return nil
}

func (db *MemoryDatabase) AddWorkflowData(entityID int, data *WorkflowData) (int, error) {
	db.rwMutexWorkflowData.Lock()
	db.workflowDataCounter++
	data.ID = db.workflowDataCounter
	data.EntityID = entityID
	db.workflowData[data.ID] = copyWorkflowData(data)
	db.rwMutexWorkflowData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) AddActivityData(entityID int, data *ActivityData) (int, error) {
	db.rwMutexActivityData.Lock()
	db.activityDataCounter++
	data.ID = db.activityDataCounter
	data.EntityID = entityID
	db.activityData[data.ID] = copyActivityData(data)
	db.rwMutexActivityData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) AddSagaData(entityID int, data *SagaData) (int, error) {
	db.rwMutexSagaData.Lock()
	db.sagaDataCounter++
	data.ID = db.sagaDataCounter
	data.EntityID = entityID
	db.sagaData[data.ID] = copySagaData(data)
	db.rwMutexSagaData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) AddSideEffectData(entityID int, data *SideEffectData) (int, error) {
	db.rwMutexSideEffectData.Lock()
	db.sideEffectDataCounter++
	data.ID = db.sideEffectDataCounter
	data.EntityID = entityID
	db.sideEffectData[data.ID] = copySideEffectData(data)
	db.rwMutexSideEffectData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) GetWorkflowData(id int) (*WorkflowData, error) {
	db.rwMutexWorkflowData.RLock()
	d, exists := db.workflowData[id]
	if !exists {
		db.rwMutexWorkflowData.RUnlock()
		return nil, ErrWorkflowEntityNotFound
	}
	dCopy := copyWorkflowData(d)
	db.rwMutexWorkflowData.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetActivityData(id int) (*ActivityData, error) {
	db.rwMutexActivityData.RLock()
	d, exists := db.activityData[id]
	if !exists {
		db.rwMutexActivityData.RUnlock()
		return nil, ErrActivityEntityNotFound
	}
	dCopy := copyActivityData(d)
	db.rwMutexActivityData.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetSagaData(id int) (*SagaData, error) {
	db.rwMutexSagaData.RLock()
	d, exists := db.sagaData[id]
	if !exists {
		db.rwMutexSagaData.RUnlock()
		return nil, ErrSagaEntityNotFound
	}
	dCopy := copySagaData(d)
	db.rwMutexSagaData.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetSideEffectData(id int) (*SideEffectData, error) {
	db.rwMutexSideEffectData.RLock()
	d, exists := db.sideEffectData[id]
	if !exists {
		db.rwMutexSideEffectData.RUnlock()
		return nil, ErrSideEffectEntityNotFound
	}
	dCopy := copySideEffectData(d)
	db.rwMutexSideEffectData.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetWorkflowDataByEntityID(entityID int) (*WorkflowData, error) {
	db.rwMutexWorkflowData.RLock()
	for _, d := range db.workflowData {
		if d.EntityID == entityID {
			dCopy := copyWorkflowData(d)
			db.rwMutexWorkflowData.RUnlock()
			return dCopy, nil
		}
	}
	db.rwMutexWorkflowData.RUnlock()
	return nil, ErrWorkflowEntityNotFound
}

func (db *MemoryDatabase) GetActivityDataByEntityID(entityID int) (*ActivityData, error) {
	db.rwMutexActivityData.RLock()
	for _, d := range db.activityData {
		if d.EntityID == entityID {
			dCopy := copyActivityData(d)
			db.rwMutexActivityData.RUnlock()
			return dCopy, nil
		}
	}
	db.rwMutexActivityData.RUnlock()
	return nil, ErrActivityEntityNotFound
}

func (db *MemoryDatabase) GetSagaDataByEntityID(entityID int) (*SagaData, error) {
	db.rwMutexSagaData.RLock()
	for _, d := range db.sagaData {
		if d.EntityID == entityID {
			dCopy := copySagaData(d)
			db.rwMutexSagaData.RUnlock()
			return dCopy, nil
		}
	}
	db.rwMutexSagaData.RUnlock()
	return nil, ErrSagaEntityNotFound
}

func (db *MemoryDatabase) GetSideEffectDataByEntityID(entityID int) (*SideEffectData, error) {
	db.rwMutexSideEffectData.RLock()
	for _, d := range db.sideEffectData {
		if d.EntityID == entityID {
			dCopy := copySideEffectData(d)
			db.rwMutexSideEffectData.RUnlock()
			return dCopy, nil
		}
	}
	db.rwMutexSideEffectData.RUnlock()
	return nil, ErrSideEffectEntityNotFound
}

func (db *MemoryDatabase) AddWorkflowExecutionData(executionID int, data *WorkflowExecutionData) (int, error) {
	db.rwMutexWorkflowExecData.Lock()
	db.workflowExecutionDataCounter++
	data.ID = db.workflowExecutionDataCounter
	data.ExecutionID = executionID
	db.workflowExecutionData[data.ID] = copyWorkflowExecutionData(data)
	db.rwMutexWorkflowExecData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) AddActivityExecutionData(executionID int, data *ActivityExecutionData) (int, error) {
	db.rwMutexActivityExecData.Lock()
	db.activityExecutionDataCounter++
	data.ID = db.activityExecutionDataCounter
	data.ExecutionID = executionID
	db.activityExecutionData[data.ID] = copyActivityExecutionData(data)
	db.rwMutexActivityExecData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) AddSagaExecutionData(executionID int, data *SagaExecutionData) (int, error) {
	db.rwMutexSagaExecData.Lock()
	db.sagaExecutionDataCounter++
	data.ID = db.sagaExecutionDataCounter
	data.ExecutionID = executionID
	db.sagaExecutionData[data.ID] = copySagaExecutionData(data)
	db.rwMutexSagaExecData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) AddSideEffectExecutionData(executionID int, data *SideEffectExecutionData) (int, error) {
	db.rwMutexSideEffectExecData.Lock()
	db.sideEffectExecutionDataCounter++
	data.ID = db.sideEffectExecutionDataCounter
	data.ExecutionID = executionID
	db.sideEffectExecutionData[data.ID] = copySideEffectExecutionData(data)
	db.rwMutexSideEffectExecData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) GetWorkflowExecutionData(id int) (*WorkflowExecutionData, error) {
	db.rwMutexWorkflowExecData.RLock()
	d, exists := db.workflowExecutionData[id]
	if !exists {
		db.rwMutexWorkflowExecData.RUnlock()
		return nil, ErrWorkflowExecutionNotFound
	}
	dCopy := copyWorkflowExecutionData(d)
	db.rwMutexWorkflowExecData.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetActivityExecutionData(id int) (*ActivityExecutionData, error) {
	db.rwMutexActivityExecData.RLock()
	d, exists := db.activityExecutionData[id]
	if !exists {
		db.rwMutexActivityExecData.RUnlock()
		return nil, ErrActivityExecutionNotFound
	}
	dCopy := copyActivityExecutionData(d)
	db.rwMutexActivityExecData.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetSagaExecutionData(id int) (*SagaExecutionData, error) {
	db.rwMutexSagaExecData.RLock()
	d, exists := db.sagaExecutionData[id]
	if !exists {
		db.rwMutexSagaExecData.RUnlock()
		return nil, ErrSagaExecutionNotFound
	}
	dCopy := copySagaExecutionData(d)
	db.rwMutexSagaExecData.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetSideEffectExecutionData(id int) (*SideEffectExecutionData, error) {
	db.rwMutexSideEffectExecData.RLock()
	d, exists := db.sideEffectExecutionData[id]
	if !exists {
		db.rwMutexSideEffectExecData.RUnlock()
		return nil, ErrSideEffectExecutionNotFound
	}
	dCopy := copySideEffectExecutionData(d)
	db.rwMutexSideEffectExecData.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetWorkflowExecutionDataByExecutionID(executionID int) (*WorkflowExecutionData, error) {
	db.rwMutexWorkflowExecData.RLock()
	for _, d := range db.workflowExecutionData {
		if d.ExecutionID == executionID {
			dCopy := copyWorkflowExecutionData(d)
			db.rwMutexWorkflowExecData.RUnlock()
			return dCopy, nil
		}
	}
	db.rwMutexWorkflowExecData.RUnlock()
	return nil, ErrWorkflowExecutionNotFound
}

func (db *MemoryDatabase) GetActivityExecutionDataByExecutionID(executionID int) (*ActivityExecutionData, error) {
	db.rwMutexActivityExecData.RLock()
	for _, d := range db.activityExecutionData {
		if d.ExecutionID == executionID {
			dCopy := copyActivityExecutionData(d)
			db.rwMutexActivityExecData.RUnlock()
			return dCopy, nil
		}
	}
	db.rwMutexActivityExecData.RUnlock()
	return nil, ErrActivityExecutionNotFound
}

func (db *MemoryDatabase) GetSagaExecutionDataByExecutionID(executionID int) (*SagaExecutionData, error) {
	db.rwMutexSagaExecData.RLock()
	for _, d := range db.sagaExecutionData {
		if d.ExecutionID == executionID {
			dCopy := copySagaExecutionData(d)
			db.rwMutexSagaExecData.RUnlock()
			return dCopy, nil
		}
	}
	db.rwMutexSagaExecData.RUnlock()
	return nil, ErrSagaExecutionNotFound
}

func (db *MemoryDatabase) GetSideEffectExecutionDataByExecutionID(executionID int) (*SideEffectExecutionData, error) {
	db.rwMutexSideEffectExecData.RLock()
	for _, d := range db.sideEffectExecutionData {
		if d.ExecutionID == executionID {
			dCopy := copySideEffectExecutionData(d)
			db.rwMutexSideEffectExecData.RUnlock()
			return dCopy, nil
		}
	}
	db.rwMutexSideEffectExecData.RUnlock()
	return nil, ErrSideEffectExecutionNotFound
}

func (db *MemoryDatabase) HasRun(id int) (bool, error) {
	db.rwMutexRuns.RLock()
	_, exists := db.runs[id]
	db.rwMutexRuns.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasVersion(id int) (bool, error) {
	db.rwMutexVersions.RLock()
	_, exists := db.versions[id]
	db.rwMutexVersions.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasHierarchy(id int) (bool, error) {
	db.rwMutexHierarchies.RLock()
	_, exists := db.hierarchies[id]
	db.rwMutexHierarchies.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasQueue(id int) (bool, error) {
	db.rwMutexQueues.RLock()
	_, exists := db.queues[id]
	db.rwMutexQueues.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasQueueName(name string) (bool, error) {
	db.rwMutexQueues.RLock()
	_, exists := db.queueNames[name]
	db.rwMutexQueues.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowEntity(id int) (bool, error) {
	db.rwMutexWorkflows.RLock()
	_, exists := db.workflowEntities[id]
	db.rwMutexWorkflows.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasActivityEntity(id int) (bool, error) {
	db.rwMutexActivities.RLock()
	_, exists := db.activityEntities[id]
	db.rwMutexActivities.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSagaEntity(id int) (bool, error) {
	db.rwMutexSagas.RLock()
	_, exists := db.sagaEntities[id]
	db.rwMutexSagas.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectEntity(id int) (bool, error) {
	db.rwMutexSideEffects.RLock()
	_, exists := db.sideEffectEntities[id]
	db.rwMutexSideEffects.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowExecution(id int) (bool, error) {
	db.rwMutexWorkflowExec.RLock()
	_, exists := db.workflowExecutions[id]
	db.rwMutexWorkflowExec.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasActivityExecution(id int) (bool, error) {
	db.rwMutexActivityExec.RLock()
	_, exists := db.activityExecutions[id]
	db.rwMutexActivityExec.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSagaExecution(id int) (bool, error) {
	db.rwMutexSagaExec.RLock()
	_, exists := db.sagaExecutions[id]
	db.rwMutexSagaExec.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectExecution(id int) (bool, error) {
	db.rwMutexSideEffectExec.RLock()
	_, exists := db.sideEffectExecutions[id]
	db.rwMutexSideEffectExec.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowData(id int) (bool, error) {
	db.rwMutexWorkflowData.RLock()
	_, exists := db.workflowData[id]
	db.rwMutexWorkflowData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasActivityData(id int) (bool, error) {
	db.rwMutexActivityData.RLock()
	_, exists := db.activityData[id]
	db.rwMutexActivityData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSagaData(id int) (bool, error) {
	db.rwMutexSagaData.RLock()
	_, exists := db.sagaData[id]
	db.rwMutexSagaData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectData(id int) (bool, error) {
	db.rwMutexSideEffectData.RLock()
	_, exists := db.sideEffectData[id]
	db.rwMutexSideEffectData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowDataByEntityID(entityID int) (bool, error) {
	db.rwMutexWorkflowData.RLock()
	for _, d := range db.workflowData {
		if d.EntityID == entityID {
			db.rwMutexWorkflowData.RUnlock()
			return true, nil
		}
	}
	db.rwMutexWorkflowData.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) HasActivityDataByEntityID(entityID int) (bool, error) {
	db.rwMutexActivityData.RLock()
	for _, d := range db.activityData {
		if d.EntityID == entityID {
			db.rwMutexActivityData.RUnlock()
			return true, nil
		}
	}
	db.rwMutexActivityData.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) HasSagaDataByEntityID(entityID int) (bool, error) {
	db.rwMutexSagaData.RLock()
	for _, d := range db.sagaData {
		if d.EntityID == entityID {
			db.rwMutexSagaData.RUnlock()
			return true, nil
		}
	}
	db.rwMutexSagaData.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) HasSideEffectDataByEntityID(entityID int) (bool, error) {
	db.rwMutexSideEffectData.RLock()
	for _, d := range db.sideEffectData {
		if d.EntityID == entityID {
			db.rwMutexSideEffectData.RUnlock()
			return true, nil
		}
	}
	db.rwMutexSideEffectData.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) HasWorkflowExecutionData(id int) (bool, error) {
	db.rwMutexWorkflowExecData.RLock()
	_, exists := db.workflowExecutionData[id]
	db.rwMutexWorkflowExecData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasActivityExecutionData(id int) (bool, error) {
	db.rwMutexActivityExecData.RLock()
	_, exists := db.activityExecutionData[id]
	db.rwMutexActivityExecData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSagaExecutionData(id int) (bool, error) {
	db.rwMutexSagaExecData.RLock()
	_, exists := db.sagaExecutionData[id]
	db.rwMutexSagaExecData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectExecutionData(id int) (bool, error) {
	db.rwMutexSideEffectExecData.RLock()
	_, exists := db.sideEffectExecutionData[id]
	db.rwMutexSideEffectExecData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowExecutionDataByExecutionID(executionID int) (bool, error) {
	db.rwMutexWorkflowExecData.RLock()
	for _, d := range db.workflowExecutionData {
		if d.ExecutionID == executionID {
			db.rwMutexWorkflowExecData.RUnlock()
			return true, nil
		}
	}
	db.rwMutexWorkflowExecData.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) HasActivityExecutionDataByExecutionID(executionID int) (bool, error) {
	db.rwMutexActivityExecData.RLock()
	for _, d := range db.activityExecutionData {
		if d.ExecutionID == executionID {
			db.rwMutexActivityExecData.RUnlock()
			return true, nil
		}
	}
	db.rwMutexActivityExecData.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) HasSagaExecutionDataByExecutionID(executionID int) (bool, error) {
	db.rwMutexSagaExecData.RLock()
	for _, d := range db.sagaExecutionData {
		if d.ExecutionID == executionID {
			db.rwMutexSagaExecData.RUnlock()
			return true, nil
		}
	}
	db.rwMutexSagaExecData.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) HasSideEffectExecutionDataByExecutionID(executionID int) (bool, error) {
	db.rwMutexSideEffectExecData.RLock()
	for _, d := range db.sideEffectExecutionData {
		if d.ExecutionID == executionID {
			db.rwMutexSideEffectExecData.RUnlock()
			return true, nil
		}
	}
	db.rwMutexSideEffectExecData.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) GetWorkflowExecutionLatestByEntityID(entityID int) (*WorkflowExecution, error) {
	db.rwMutexWorkflowExec.RLock()
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
	db.rwMutexWorkflowExec.RUnlock()

	if latestExec == nil {
		return nil, ErrWorkflowExecutionNotFound
	}
	return copyWorkflowExecution(latestExec), nil
}

func (db *MemoryDatabase) GetActivityExecutionLatestByEntityID(entityID int) (*ActivityExecution, error) {
	db.rwMutexActivityExec.RLock()
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
	db.rwMutexActivityExec.RUnlock()

	if latestExec == nil {
		return nil, ErrActivityExecutionNotFound
	}
	return copyActivityExecution(latestExec), nil
}

func (db *MemoryDatabase) SetSagaValue(executionID int, key string, value []byte) (int, error) {
	db.rwMutexSagas.Lock()
	defer db.rwMutexSagas.Unlock()

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
	db.rwMutexSagas.RLock()
	defer db.rwMutexSagas.RUnlock()

	for _, keyMap := range db.sagaValues {
		if val, exists := keyMap[key]; exists && val.ID == id {
			return val.Value, nil
		}
	}
	return nil, ErrSagaValueNotFound
}

func (db *MemoryDatabase) GetSagaValueByExecutionID(executionID int, key string) ([]byte, error) {
	db.rwMutexSagas.RLock()
	defer db.rwMutexSagas.RUnlock()

	if keyMap, exists := db.sagaValues[executionID]; exists {
		if val, vexists := keyMap[key]; vexists {
			return val.Value, nil
		}
	}
	return nil, ErrSagaValueNotFound
}

func (db *MemoryDatabase) AddSignalEntity(entity *SignalEntity, parentWorkflowID int) (int, error) {
	db.rwMutexSignals.Lock()
	db.signalEntityCounter++
	entity.ID = db.signalEntityCounter

	if entity.SignalData != nil {
		db.rwMutexSignalData.Lock()
		db.signalDataCounter++
		entity.SignalData.ID = db.signalDataCounter
		entity.SignalData.EntityID = entity.ID
		db.signalData[entity.SignalData.ID] = copySignalData(entity.SignalData)
		db.rwMutexSignalData.Unlock()
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = entity.CreatedAt

	db.signalEntities[entity.ID] = copySignalEntity(entity)
	db.rwMutexSignals.Unlock()

	// Signal relationships to workflow are through hierarchies,
	// but the code provided doesn't explicitly add them here.
	return entity.ID, nil
}

func (db *MemoryDatabase) GetSignalEntity(id int, opts ...SignalEntityGetOption) (*SignalEntity, error) {
	db.rwMutexSignals.RLock()
	entity, exists := db.signalEntities[id]
	if !exists {
		db.rwMutexSignals.RUnlock()
		return nil, fmt.Errorf("signal entity not found")
	}
	eCopy := copySignalEntity(entity)
	db.rwMutexSignals.RUnlock()

	cfg := &SignalEntityGetterOptions{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		db.rwMutexSignalData.RLock()
		for _, d := range db.signalData {
			if d.EntityID == id {
				eCopy.SignalData = copySignalData(d)
				break
			}
		}
		db.rwMutexSignalData.RUnlock()
	}

	return eCopy, nil
}

func (db *MemoryDatabase) GetSignalEntities(workflowID int, opts ...SignalEntityGetOption) ([]*SignalEntity, error) {
	// Signals linked via hierarchies
	db.rwMutexHierarchies.RLock()
	var entities []*SignalEntity
	for _, h := range db.hierarchies {
		if h.ParentEntityID == workflowID && h.ChildType == EntitySignal {
			db.rwMutexSignals.RLock()
			if e, ex := db.signalEntities[h.ChildEntityID]; ex {
				entities = append(entities, copySignalEntity(e))
			}
			db.rwMutexSignals.RUnlock()
		}
	}
	db.rwMutexHierarchies.RUnlock()

	return entities, nil
}

func (db *MemoryDatabase) AddSignalExecution(exec *SignalExecution) (int, error) {
	db.rwMutexSignalExec.Lock()
	db.signalExecutionCounter++
	exec.ID = db.signalExecutionCounter

	if exec.SignalExecutionData != nil {
		db.rwMutexSignalExecData.Lock()
		db.signalExecutionDataCounter++
		exec.SignalExecutionData.ID = db.signalExecutionDataCounter
		exec.SignalExecutionData.ExecutionID = exec.ID
		db.signalExecutionData[exec.SignalExecutionData.ID] = copySignalExecutionData(exec.SignalExecutionData)
		db.rwMutexSignalExecData.Unlock()
	}

	exec.CreatedAt = time.Now()
	exec.UpdatedAt = exec.CreatedAt

	db.signalExecutions[exec.ID] = copySignalExecution(exec)
	db.rwMutexSignalExec.Unlock()
	return exec.ID, nil
}

func (db *MemoryDatabase) GetSignalExecution(id int, opts ...SignalExecutionGetOption) (*SignalExecution, error) {
	db.rwMutexSignalExec.RLock()
	exec, exists := db.signalExecutions[id]
	if !exists {
		db.rwMutexSignalExec.RUnlock()
		return nil, fmt.Errorf("signal execution not found")
	}
	execCopy := copySignalExecution(exec)
	db.rwMutexSignalExec.RUnlock()

	cfg := &SignalExecutionGetterOptions{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.IncludeData {
		db.rwMutexSignalExecData.RLock()
		for _, d := range db.signalExecutionData {
			if d.ExecutionID == id {
				execCopy.SignalExecutionData = copySignalExecutionData(d)
				break
			}
		}
		db.rwMutexSignalExecData.RUnlock()
	}

	return execCopy, nil
}

func (db *MemoryDatabase) GetSignalExecutions(entityID int) ([]*SignalExecution, error) {
	db.rwMutexSignalExec.RLock()
	var executions []*SignalExecution
	for _, exec := range db.signalExecutions {
		if exec.EntityID == entityID {
			executions = append(executions, copySignalExecution(exec))
		}
	}
	db.rwMutexSignalExec.RUnlock()
	return executions, nil
}

func (db *MemoryDatabase) GetSignalExecutionLatestByEntityID(entityID int) (*SignalExecution, error) {
	db.rwMutexSignalExec.RLock()
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
	db.rwMutexSignalExec.RUnlock()

	if latest == nil {
		return nil, fmt.Errorf("no signal execution found")
	}

	return copySignalExecution(latest), nil
}

func (db *MemoryDatabase) HasSignalEntity(id int) (bool, error) {
	db.rwMutexSignals.RLock()
	_, exists := db.signalEntities[id]
	db.rwMutexSignals.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSignalExecution(id int) (bool, error) {
	db.rwMutexSignalExec.RLock()
	_, exists := db.signalExecutions[id]
	db.rwMutexSignalExec.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSignalData(id int) (bool, error) {
	db.rwMutexSignalData.RLock()
	_, exists := db.signalData[id]
	db.rwMutexSignalData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSignalExecutionData(id int) (bool, error) {
	db.rwMutexSignalExecData.RLock()
	_, exists := db.signalExecutionData[id]
	db.rwMutexSignalExecData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) GetSignalEntityProperties(id int, getters ...SignalEntityPropertyGetter) error {
	db.rwMutexSignals.RLock()
	entity, exists := db.signalEntities[id]
	if !exists {
		db.rwMutexSignals.RUnlock()
		return fmt.Errorf("signal entity not found")
	}
	eCopy := copySignalEntity(entity)
	db.rwMutexSignals.RUnlock()

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
				db.rwMutexSignalData.RLock()
				for _, d := range db.signalData {
					if d.EntityID == id {
						eCopy.SignalData = copySignalData(d)
						break
					}
				}
				db.rwMutexSignalData.RUnlock()
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSignalEntityProperties(id int, setters ...SignalEntityPropertySetter) error {
	db.rwMutexSignals.Lock()
	entity, exists := db.signalEntities[id]
	if !exists {
		db.rwMutexSignals.Unlock()
		return fmt.Errorf("signal entity not found")
	}

	for _, setter := range setters {
		if _, err := setter(entity); err != nil {
			db.rwMutexSignals.Unlock()
			return err
		}
	}
	entity.UpdatedAt = time.Now()
	db.signalEntities[id] = copySignalEntity(entity)
	db.rwMutexSignals.Unlock()

	return nil
}

func (db *MemoryDatabase) GetSignalExecutionProperties(id int, getters ...SignalExecutionPropertyGetter) error {
	db.rwMutexSignalExec.RLock()
	exec, exists := db.signalExecutions[id]
	if !exists {
		db.rwMutexSignalExec.RUnlock()
		return fmt.Errorf("signal execution not found")
	}
	execCopy := copySignalExecution(exec)
	db.rwMutexSignalExec.RUnlock()

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
				db.rwMutexSignalExecData.RLock()
				for _, d := range db.signalExecutionData {
					if d.ExecutionID == id {
						execCopy.SignalExecutionData = copySignalExecutionData(d)
						break
					}
				}
				db.rwMutexSignalExecData.RUnlock()
			}
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSignalExecutionProperties(id int, setters ...SignalExecutionPropertySetter) error {
	db.rwMutexSignalExec.Lock()
	exec, exists := db.signalExecutions[id]
	if !exists {
		db.rwMutexSignalExec.Unlock()
		return fmt.Errorf("signal execution not found")
	}

	for _, setter := range setters {
		if _, err := setter(exec); err != nil {
			db.rwMutexSignalExec.Unlock()
			return err
		}
	}
	exec.UpdatedAt = time.Now()
	db.signalExecutions[id] = copySignalExecution(exec)
	db.rwMutexSignalExec.Unlock()
	return nil
}

func (db *MemoryDatabase) AddSignalData(entityID int, data *SignalData) (int, error) {
	db.rwMutexSignalData.Lock()
	db.signalDataCounter++
	data.ID = db.signalDataCounter
	data.EntityID = entityID
	db.signalData[data.ID] = copySignalData(data)
	db.rwMutexSignalData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) GetSignalData(id int) (*SignalData, error) {
	db.rwMutexSignalData.RLock()
	d, exists := db.signalData[id]
	if !exists {
		db.rwMutexSignalData.RUnlock()
		return nil, fmt.Errorf("signal data not found")
	}
	dCopy := copySignalData(d)
	db.rwMutexSignalData.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetSignalDataByEntityID(entityID int) (*SignalData, error) {
	db.rwMutexSignalData.RLock()
	for _, d := range db.signalData {
		if d.EntityID == entityID {
			dCopy := copySignalData(d)
			db.rwMutexSignalData.RUnlock()
			return dCopy, nil
		}
	}
	db.rwMutexSignalData.RUnlock()
	return nil, fmt.Errorf("signal data not found for entity")
}

func (db *MemoryDatabase) HasSignalDataByEntityID(entityID int) (bool, error) {
	db.rwMutexSignalData.RLock()
	for _, d := range db.signalData {
		if d.EntityID == entityID {
			db.rwMutexSignalData.RUnlock()
			return true, nil
		}
	}
	db.rwMutexSignalData.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) GetSignalDataProperties(entityID int, getters ...SignalDataPropertyGetter) error {
	db.rwMutexSignalData.RLock()
	var data *SignalData
	for _, d := range db.signalData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}
	db.rwMutexSignalData.RUnlock()

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
	db.rwMutexSignalData.Lock()
	var data *SignalData
	for _, d := range db.signalData {
		if d.EntityID == entityID {
			data = d
			break
		}
	}
	if data == nil {
		db.rwMutexSignalData.Unlock()
		return fmt.Errorf("signal data not found")
	}

	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			db.rwMutexSignalData.Unlock()
			return err
		}
	}
	db.rwMutexSignalData.Unlock()
	return nil
}

func (db *MemoryDatabase) AddSignalExecutionData(executionID int, data *SignalExecutionData) (int, error) {
	db.rwMutexSignalExecData.Lock()
	db.signalExecutionDataCounter++
	data.ID = db.signalExecutionDataCounter
	data.ExecutionID = executionID
	db.signalExecutionData[data.ID] = copySignalExecutionData(data)
	db.rwMutexSignalExecData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) GetSignalExecutionData(id int) (*SignalExecutionData, error) {
	db.rwMutexSignalExecData.RLock()
	d, exists := db.signalExecutionData[id]
	if !exists {
		db.rwMutexSignalExecData.RUnlock()
		return nil, fmt.Errorf("signal execution data not found")
	}
	dCopy := copySignalExecutionData(d)
	db.rwMutexSignalExecData.RUnlock()
	return dCopy, nil
}

func (db *MemoryDatabase) GetSignalExecutionDataByExecutionID(executionID int) (*SignalExecutionData, error) {
	db.rwMutexSignalExecData.RLock()
	for _, d := range db.signalExecutionData {
		if d.ExecutionID == executionID {
			dCopy := copySignalExecutionData(d)
			db.rwMutexSignalExecData.RUnlock()
			return dCopy, nil
		}
	}
	db.rwMutexSignalExecData.RUnlock()
	return nil, fmt.Errorf("signal execution data not found for execution")
}

func (db *MemoryDatabase) HasSignalExecutionDataByExecutionID(executionID int) (bool, error) {
	db.rwMutexSignalExecData.RLock()
	for _, d := range db.signalExecutionData {
		if d.ExecutionID == executionID {
			db.rwMutexSignalExecData.RUnlock()
			return true, nil
		}
	}
	db.rwMutexSignalExecData.RUnlock()
	return false, nil
}

func (db *MemoryDatabase) GetSignalExecutionDataProperties(entityID int, getters ...SignalExecutionDataPropertyGetter) error {
	db.rwMutexSignalExecData.RLock()
	var data *SignalExecutionData
	for _, d := range db.signalExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}
	db.rwMutexSignalExecData.RUnlock()

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
	db.rwMutexSignalExecData.Lock()
	var data *SignalExecutionData
	for _, d := range db.signalExecutionData {
		if d.ExecutionID == entityID {
			data = d
			break
		}
	}
	if data == nil {
		db.rwMutexSignalExecData.Unlock()
		return fmt.Errorf("signal execution data not found")
	}

	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			db.rwMutexSignalExecData.Unlock()
			return err
		}
	}
	db.rwMutexSignalExecData.Unlock()
	return nil
}

func (db *MemoryDatabase) UpdateSignalEntity(entity *SignalEntity) error {
	db.rwMutexSignals.Lock()
	defer db.rwMutexSignals.Unlock()

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
	db.rwMutexRuns.RLock()
	runsToDelete := make([]int, 0)
	for _, run := range db.runs {
		if run.Status == status {
			runsToDelete = append(runsToDelete, run.ID)
		}
	}
	db.rwMutexRuns.RUnlock()

	if len(runsToDelete) == 0 {
		return nil
	}

	for _, runID := range runsToDelete {
		db.rwMutexRelationships.RLock()
		workflowIDs := make([]int, len(db.runToWorkflows[runID]))
		copy(workflowIDs, db.runToWorkflows[runID])
		db.rwMutexRelationships.RUnlock()

		for _, wfID := range workflowIDs {
			if err := db.deleteWorkflowAndChildren(wfID); err != nil {
				return fmt.Errorf("failed to delete workflow %d: %w", wfID, err)
			}
		}

		db.rwMutexRuns.Lock()
		delete(db.runs, runID)
		db.rwMutexRuns.Unlock()

		db.rwMutexRelationships.Lock()
		delete(db.runToWorkflows, runID)
		db.rwMutexRelationships.Unlock()
	}

	return nil
}

func (db *MemoryDatabase) deleteWorkflowAndChildren(workflowID int) error {
	db.rwMutexHierarchies.RLock()
	var hierarchies []*Hierarchy
	for _, h := range db.hierarchies {
		if h.ParentEntityID == workflowID {
			hierarchies = append(hierarchies, copyHierarchy(h))
		}
	}
	db.rwMutexHierarchies.RUnlock()

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

	db.rwMutexHierarchies.Lock()
	for id, hh := range db.hierarchies {
		if hh.ParentEntityID == workflowID || hh.ChildEntityID == workflowID {
			delete(db.hierarchies, id)
		}
	}
	db.rwMutexHierarchies.Unlock()

	return nil
}

// deleteWorkflowEntity
func (db *MemoryDatabase) deleteWorkflowEntity(workflowID int) error {
	// Delete workflow execution data
	db.rwMutexWorkflowExec.Lock()
	for id, exec := range db.workflowExecutions {
		if exec.EntityID == workflowID {
			db.rwMutexWorkflowExecData.Lock()
			delete(db.workflowExecutionData, db.execToDataMap[exec.ID])
			db.rwMutexWorkflowExecData.Unlock()
			delete(db.workflowExecutions, id)
			db.rwMutexRelationships.Lock()
			delete(db.execToDataMap, exec.ID)
			db.rwMutexRelationships.Unlock()
		}
	}
	db.rwMutexWorkflowExec.Unlock()

	// Delete workflow data
	db.rwMutexWorkflowData.Lock()
	for id, data := range db.workflowData {
		if data.EntityID == workflowID {
			delete(db.workflowData, id)
		}
	}
	db.rwMutexWorkflowData.Unlock()

	// Delete versions
	db.rwMutexRelationships.Lock()
	versionIDs := db.workflowToVersion[workflowID]
	delete(db.workflowToVersion, workflowID)
	delete(db.workflowVersions, workflowID)
	db.rwMutexRelationships.Unlock()

	db.rwMutexVersions.Lock()
	for _, vID := range versionIDs {
		delete(db.versions, vID)
	}
	db.rwMutexVersions.Unlock()

	// Clean up queue mappings
	db.rwMutexRelationships.Lock()
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
	db.rwMutexRelationships.Unlock()

	db.rwMutexWorkflows.Lock()
	delete(db.workflowEntities, workflowID)
	db.rwMutexWorkflows.Unlock()

	return nil
}

func (db *MemoryDatabase) deleteActivityEntity(entityID int) error {
	db.rwMutexActivityExec.Lock()
	for id, exec := range db.activityExecutions {
		if exec.EntityID == entityID {
			db.rwMutexActivityExecData.Lock()
			delete(db.activityExecutionData, exec.ID)
			db.rwMutexActivityExecData.Unlock()
			delete(db.activityExecutions, id)
		}
	}
	db.rwMutexActivityExec.Unlock()

	db.rwMutexActivityData.Lock()
	for id, data := range db.activityData {
		if data.EntityID == entityID {
			delete(db.activityData, id)
		}
	}
	db.rwMutexActivityData.Unlock()

	db.rwMutexActivities.Lock()
	delete(db.activityEntities, entityID)
	db.rwMutexActivities.Unlock()

	return nil
}

func (db *MemoryDatabase) deleteSagaEntity(entityID int) error {
	db.rwMutexSagaExec.Lock()
	for id, exec := range db.sagaExecutions {
		if exec.EntityID == entityID {
			db.rwMutexSagaExecData.Lock()
			delete(db.sagaExecutionData, exec.ID)
			db.rwMutexSagaExecData.Unlock()

			db.rwMutexSagas.Lock()
			delete(db.sagaValues, exec.ID)
			db.rwMutexSagas.Unlock()

			delete(db.sagaExecutions, id)
		}
	}
	db.rwMutexSagaExec.Unlock()

	db.rwMutexSagaData.Lock()
	for id, data := range db.sagaData {
		if data.EntityID == entityID {
			delete(db.sagaData, id)
		}
	}
	db.rwMutexSagaData.Unlock()

	db.rwMutexSagas.Lock()
	delete(db.sagaEntities, entityID)
	db.rwMutexSagas.Unlock()

	return nil
}

func (db *MemoryDatabase) deleteSideEffectEntity(entityID int) error {
	db.rwMutexSideEffectExec.Lock()
	for id, exec := range db.sideEffectExecutions {
		if exec.EntityID == entityID {
			db.rwMutexSideEffectExecData.Lock()
			delete(db.sideEffectExecutionData, exec.ID)
			db.rwMutexSideEffectExecData.Unlock()
			delete(db.sideEffectExecutions, id)
		}
	}
	db.rwMutexSideEffectExec.Unlock()

	db.rwMutexSideEffectData.Lock()
	for id, data := range db.sideEffectData {
		if data.EntityID == entityID {
			delete(db.sideEffectData, id)
		}
	}
	db.rwMutexSideEffectData.Unlock()

	db.rwMutexSideEffects.Lock()
	delete(db.sideEffectEntities, entityID)
	db.rwMutexSideEffects.Unlock()

	return nil
}

func (db *MemoryDatabase) deleteSignalEntity(entityID int) error {
	db.rwMutexSignalExec.Lock()
	for id, exec := range db.signalExecutions {
		if exec.EntityID == entityID {
			db.rwMutexSignalExecData.Lock()
			delete(db.signalExecutionData, exec.ID)
			db.rwMutexSignalExecData.Unlock()
			delete(db.signalExecutions, id)
		}
	}
	db.rwMutexSignalExec.Unlock()

	db.rwMutexSignalData.Lock()
	for id, data := range db.signalData {
		if data.EntityID == entityID {
			delete(db.signalData, id)
		}
	}
	db.rwMutexSignalData.Unlock()

	db.rwMutexSignals.Lock()
	delete(db.signalEntities, entityID)
	db.rwMutexSignals.Unlock()

	return nil
}

func (db *MemoryDatabase) GetRunsPaginated(page, pageSize int, filter *RunFilter, sortCriteria *RunSort) (*PaginatedRuns, error) {
	db.rwMutexRuns.RLock()
	matchingRuns := make([]*Run, 0)
	for _, run := range db.runs {
		if filter != nil {
			if run.Status != filter.Status {
				continue
			}
		}
		matchingRuns = append(matchingRuns, copyRun(run))
	}
	db.rwMutexRuns.RUnlock()

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
