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
	db.rwMutexRuns.RLock()
	db.rwMutexVersions.RLock()
	db.rwMutexHierarchies.RLock()
	db.rwMutexQueues.RLock()
	db.rwMutexWorkflows.RLock()
	db.rwMutexActivities.RLock()
	db.rwMutexSagas.RLock()
	db.rwMutexSideEffects.RLock()
	db.rwMutexSignals.RLock()
	db.rwMutexWorkflowData.RLock()
	db.rwMutexActivityData.RLock()
	db.rwMutexSagaData.RLock()
	db.rwMutexSideEffectData.RLock()
	db.rwMutexSignalData.RLock()
	db.rwMutexWorkflowExec.RLock()
	db.rwMutexActivityExec.RLock()
	db.rwMutexSagaExec.RLock()
	db.rwMutexSideEffectExec.RLock()
	db.rwMutexSignalExec.RLock()
	db.rwMutexWorkflowExecData.RLock()
	db.rwMutexActivityExecData.RLock()
	db.rwMutexSagaExecData.RLock()
	db.rwMutexSideEffectExecData.RLock()
	db.rwMutexSignalExecData.RLock()
	db.rwMutexRelationships.RLock()

	defer db.rwMutexRuns.RUnlock()
	defer db.rwMutexVersions.RUnlock()
	defer db.rwMutexHierarchies.RUnlock()
	defer db.rwMutexQueues.RUnlock()
	defer db.rwMutexWorkflows.RUnlock()
	defer db.rwMutexActivities.RUnlock()
	defer db.rwMutexSagas.RUnlock()
	defer db.rwMutexSideEffects.RUnlock()
	defer db.rwMutexSignals.RUnlock()
	defer db.rwMutexWorkflowData.RUnlock()
	defer db.rwMutexActivityData.RUnlock()
	defer db.rwMutexSagaData.RUnlock()
	defer db.rwMutexSideEffectData.RUnlock()
	defer db.rwMutexSignalData.RUnlock()
	defer db.rwMutexWorkflowExec.RUnlock()
	defer db.rwMutexActivityExec.RUnlock()
	defer db.rwMutexSagaExec.RUnlock()
	defer db.rwMutexSideEffectExec.RUnlock()
	defer db.rwMutexSignalExec.RUnlock()
	defer db.rwMutexWorkflowExecData.RUnlock()
	defer db.rwMutexActivityExecData.RUnlock()
	defer db.rwMutexSagaExecData.RUnlock()
	defer db.rwMutexSideEffectExecData.RUnlock()
	defer db.rwMutexSignalExecData.RUnlock()
	defer db.rwMutexRelationships.RUnlock()

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
func (db *MemoryDatabase) AddRun(run *Run) (RunID, error) {
	db.rwMutexRuns.Lock()
	defer db.rwMutexRuns.Unlock()

	db.runCounter++
	run.ID = RunID(db.runCounter)
	run.CreatedAt = time.Now()
	run.UpdatedAt = run.CreatedAt

	db.runs[run.ID] = copyRun(run)
	return run.ID, nil
}

// GetRun
func (db *MemoryDatabase) GetRun(id RunID, opts ...RunGetOption) (*Run, error) {
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

func (db *MemoryDatabase) GetRunProperties(id RunID, getters ...RunPropertyGetter) error {
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

func (db *MemoryDatabase) SetRunProperties(id RunID, setters ...RunPropertySetter) error {
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

func (db *MemoryDatabase) AddQueue(queue *Queue) (QueueID, error) {
	db.rwMutexQueues.Lock()
	defer db.rwMutexQueues.Unlock()

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

func (db *MemoryDatabase) AddVersion(version *Version) (VersionID, error) {
	db.rwMutexVersions.Lock()
	defer db.rwMutexVersions.Unlock()

	db.versionCounter++
	version.ID = VersionID(db.versionCounter)
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

func (db *MemoryDatabase) GetVersion(id VersionID, opts ...VersionGetOption) (*Version, error) {
	db.rwMutexVersions.RLock()
	v, exists := db.versions[id]
	if !exists {
		db.rwMutexVersions.RUnlock()
		return nil, ErrVersionNotFound
	}
	versionCopy := copyVersion(v)
	db.rwMutexVersions.RUnlock()

	cfg := &VersionGetterOptions{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	return versionCopy, nil
}

func (db *MemoryDatabase) AddHierarchy(hierarchy *Hierarchy) (HierarchyID, error) {
	db.rwMutexHierarchies.Lock()
	defer db.rwMutexHierarchies.Unlock()

	db.hierarchyCounter++
	hierarchy.ID = HierarchyID(db.hierarchyCounter)

	db.hierarchies[hierarchy.ID] = copyHierarchy(hierarchy)
	return hierarchy.ID, nil
}

func (db *MemoryDatabase) GetHierarchy(id HierarchyID, opts ...HierarchyGetOption) (*Hierarchy, error) {
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

func (db *MemoryDatabase) AddWorkflowEntity(entity *WorkflowEntity) (WorkflowEntityID, error) {
	db.rwMutexQueues.Lock()
	db.rwMutexRelationships.Lock()
	db.rwMutexWorkflowData.Lock()
	db.rwMutexWorkflows.Lock()

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

	db.rwMutexQueues.Unlock()
	db.rwMutexRelationships.Unlock()
	db.rwMutexWorkflowData.Unlock()
	db.rwMutexWorkflows.Unlock()

	return entity.ID, nil
}

func (db *MemoryDatabase) AddWorkflowExecution(exec *WorkflowExecution) (WorkflowExecutionID, error) {
	db.rwMutexWorkflowExec.Lock()
	db.rwMutexWorkflowExecData.Lock()
	db.rwMutexRelationships.Lock()
	defer db.rwMutexRelationships.Unlock()
	defer db.rwMutexWorkflowExec.Unlock()
	defer db.rwMutexWorkflowExecData.Unlock()

	db.workflowExecutionCounter++
	exec.ID = WorkflowExecutionID(db.workflowExecutionCounter)

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

func (db *MemoryDatabase) GetWorkflowEntityProperties(id WorkflowEntityID, getters ...WorkflowEntityPropertyGetter) error {
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
					if a, aexists := db.activityEntities[ActivityEntityID(aID)]; aexists {
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
					if s, sexists := db.sagaEntities[SagaEntityID(sID)]; sexists {
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
					if se, seexists := db.sideEffectEntities[SideEffectEntityID(seID)]; seexists {
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

func (db *MemoryDatabase) SetWorkflowEntityProperties(id WorkflowEntityID, setters ...WorkflowEntityPropertySetter) error {
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

	if opts.RunID != nil {
		db.rwMutexRuns.RLock()
		_, exists := db.runs[*opts.RunID]
		db.rwMutexRuns.RUnlock()
		if !exists {
			return ErrRunNotFound
		}

		db.rwMutexRelationships.Lock()
		db.runToWorkflows[*opts.RunID] = append(db.runToWorkflows[*opts.RunID], id)
		db.rwMutexRelationships.Unlock()
	}

	db.rwMutexWorkflows.Lock()
	entity.UpdatedAt = time.Now()
	db.workflowEntities[id] = copyWorkflowEntity(entity)
	db.rwMutexWorkflows.Unlock()

	return nil
}

func (db *MemoryDatabase) AddActivityEntity(entity *ActivityEntity, parentWorkflowID WorkflowEntityID) (ActivityEntityID, error) {
	db.rwMutexActivityData.Lock()
	db.rwMutexActivities.Lock()
	db.rwMutexRelationships.Lock()
	defer db.rwMutexRelationships.Unlock()
	defer db.rwMutexActivities.Unlock()
	defer db.rwMutexActivityData.Unlock()

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
	db.rwMutexActivityExec.Lock()
	db.activityExecutionCounter++
	exec.ID = ActivityExecutionID(db.activityExecutionCounter)

	if exec.ActivityExecutionData != nil {
		db.rwMutexActivityExecData.Lock()
		db.activityExecutionDataCounter++
		exec.ActivityExecutionData.ID = ActivityExecutionDataID(db.activityExecutionDataCounter)
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

func (db *MemoryDatabase) GetActivityEntity(id ActivityEntityID, opts ...ActivityEntityGetOption) (*ActivityEntity, error) {
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

func (db *MemoryDatabase) GetActivityEntities(workflowID WorkflowEntityID, opts ...ActivityEntityGetOption) ([]*ActivityEntity, error) {
	db.rwMutexRelationships.RLock()
	activityIDs := db.workflowToChildren[workflowID][EntityActivity]
	db.rwMutexRelationships.RUnlock()

	entities := make([]*ActivityEntity, 0, len(activityIDs))
	for _, aID := range activityIDs {
		e, err := db.GetActivityEntity(ActivityEntityID(aID), opts...)
		if err != nil && err != ErrActivityEntityNotFound {
			return nil, err
		}
		if e != nil {
			entities = append(entities, e)
		}
	}

	return entities, nil
}

func (db *MemoryDatabase) GetActivityEntityProperties(id ActivityEntityID, getters ...ActivityEntityPropertyGetter) error {
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

func (db *MemoryDatabase) SetActivityEntityProperties(id ActivityEntityID, setters ...ActivityEntityPropertySetter) error {
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
		db.entityToWorkflow[int(id)] = *opts.ParentWorkflowID
		if db.workflowToChildren[*opts.ParentWorkflowID] == nil {
			db.workflowToChildren[*opts.ParentWorkflowID] = make(map[EntityType][]int)
		}
		db.workflowToChildren[*opts.ParentWorkflowID][EntityActivity] = append(
			db.workflowToChildren[*opts.ParentWorkflowID][EntityActivity],
			int(id),
		)
		db.rwMutexRelationships.Unlock()
	}

	if opts.ParentRunID != nil {
		db.rwMutexRuns.RLock()
		_, rExists := db.runs[*opts.ParentRunID]
		db.rwMutexRuns.RUnlock()
		if !rExists {
			return ErrRunNotFound
		}
	}

	db.rwMutexActivities.Lock()
	entity.UpdatedAt = time.Now()
	db.activityEntities[id] = copyActivityEntity(entity)
	db.rwMutexActivities.Unlock()

	return nil
}

func (db *MemoryDatabase) AddSagaEntity(entity *SagaEntity, parentWorkflowID WorkflowEntityID) (SagaEntityID, error) {
	db.rwMutexSagas.Lock()
	db.sagaEntityCounter++
	entity.ID = SagaEntityID(db.sagaEntityCounter)

	if entity.SagaData != nil {
		db.rwMutexSagaData.Lock()
		db.sagaDataCounter++
		entity.SagaData.ID = SagaDataID(db.sagaDataCounter)
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
	db.workflowToChildren[parentWorkflowID][EntitySaga] = append(
		db.workflowToChildren[parentWorkflowID][EntitySaga],
		int(entity.ID),
	)
	db.rwMutexRelationships.Unlock()

	return entity.ID, nil
}

func (db *MemoryDatabase) GetSagaEntities(workflowID WorkflowEntityID, opts ...SagaEntityGetOption) ([]*SagaEntity, error) {
	db.rwMutexRelationships.RLock()
	sagaIDs := db.workflowToChildren[workflowID][EntitySaga]
	db.rwMutexRelationships.RUnlock()

	entities := make([]*SagaEntity, 0, len(sagaIDs))
	for _, sID := range sagaIDs {
		e, err := db.GetSagaEntity(SagaEntityID(sID), opts...)
		if err != nil && err != ErrSagaEntityNotFound {
			return nil, err
		}
		if e != nil {
			entities = append(entities, e)
		}
	}

	return entities, nil
}

func (db *MemoryDatabase) AddSagaExecution(exec *SagaExecution) (SagaExecutionID, error) {
	db.rwMutexSagaExec.Lock()
	db.sagaExecutionCounter++
	exec.ID = SagaExecutionID(db.sagaExecutionCounter)

	if exec.SagaExecutionData != nil {
		db.rwMutexSagaExecData.Lock()
		db.sagaExecutionDataCounter++
		exec.SagaExecutionData.ID = SagaExecutionDataID(db.sagaExecutionDataCounter)
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

func (db *MemoryDatabase) GetSagaEntity(id SagaEntityID, opts ...SagaEntityGetOption) (*SagaEntity, error) {
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

func (db *MemoryDatabase) GetSagaEntityProperties(id SagaEntityID, getters ...SagaEntityPropertyGetter) error {
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

func (db *MemoryDatabase) SetSagaEntityProperties(id SagaEntityID, setters ...SagaEntityPropertySetter) error {
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
		db.entityToWorkflow[int(id)] = *opts.ParentWorkflowID
		if db.workflowToChildren[*opts.ParentWorkflowID] == nil {
			db.workflowToChildren[*opts.ParentWorkflowID] = make(map[EntityType][]int)
		}
		db.workflowToChildren[*opts.ParentWorkflowID][EntitySaga] = append(
			db.workflowToChildren[*opts.ParentWorkflowID][EntitySaga],
			int(id),
		)
		db.rwMutexRelationships.Unlock()
	}

	db.rwMutexSagas.Lock()
	entity.UpdatedAt = time.Now()
	db.sagaEntities[id] = copySagaEntity(entity)
	db.rwMutexSagas.Unlock()

	return nil
}

func (db *MemoryDatabase) AddSideEffectEntity(entity *SideEffectEntity, parentWorkflowID WorkflowEntityID) (SideEffectEntityID, error) {
	db.rwMutexSideEffects.Lock()
	db.sideEffectEntityCounter++
	entity.ID = SideEffectEntityID(db.sideEffectEntityCounter)

	if entity.SideEffectData != nil {
		db.rwMutexSideEffectData.Lock()
		db.sideEffectDataCounter++
		entity.SideEffectData.ID = SideEffectDataID(db.sideEffectDataCounter)
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
	db.workflowToChildren[parentWorkflowID][EntitySideEffect] = append(
		db.workflowToChildren[parentWorkflowID][EntitySideEffect],
		int(entity.ID),
	)
	db.rwMutexRelationships.Unlock()

	return entity.ID, nil
}

func (db *MemoryDatabase) GetSideEffectEntities(workflowID WorkflowEntityID, opts ...SideEffectEntityGetOption) ([]*SideEffectEntity, error) {
	db.rwMutexRelationships.RLock()
	sideEffectIDs := db.workflowToChildren[workflowID][EntitySideEffect]
	db.rwMutexRelationships.RUnlock()

	entities := make([]*SideEffectEntity, 0, len(sideEffectIDs))
	for _, seID := range sideEffectIDs {
		e, err := db.GetSideEffectEntity(SideEffectEntityID(seID), opts...)
		if err != nil && err != ErrSideEffectEntityNotFound {
			return nil, err
		}
		if e != nil {
			entities = append(entities, e)
		}
	}

	return entities, nil
}

func (db *MemoryDatabase) AddSideEffectExecution(exec *SideEffectExecution) (SideEffectExecutionID, error) {
	db.rwMutexSideEffectExec.Lock()
	db.sideEffectExecutionCounter++
	exec.ID = SideEffectExecutionID(db.sideEffectExecutionCounter)

	if exec.SideEffectExecutionData != nil {
		db.rwMutexSideEffectExecData.Lock()
		db.sideEffectExecutionDataCounter++
		exec.SideEffectExecutionData.ID = SideEffectExecutionDataID(db.sideEffectExecutionDataCounter)
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

func (db *MemoryDatabase) GetSideEffectEntity(id SideEffectEntityID, opts ...SideEffectEntityGetOption) (*SideEffectEntity, error) {
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

func (db *MemoryDatabase) GetSideEffectEntityProperties(id SideEffectEntityID, getters ...SideEffectEntityPropertyGetter) error {
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
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSideEffectEntityProperties(id SideEffectEntityID, setters ...SideEffectEntityPropertySetter) error {
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

	if opts.ParentWorkflowID != nil {
		db.rwMutexWorkflows.RLock()
		_, wExists := db.workflowEntities[*opts.ParentWorkflowID]
		db.rwMutexWorkflows.RUnlock()
		if !wExists {
			return ErrWorkflowEntityNotFound
		}

		db.rwMutexRelationships.Lock()
		db.entityToWorkflow[int(id)] = *opts.ParentWorkflowID
		if db.workflowToChildren[*opts.ParentWorkflowID] == nil {
			db.workflowToChildren[*opts.ParentWorkflowID] = make(map[EntityType][]int)
		}
		db.workflowToChildren[*opts.ParentWorkflowID][EntitySideEffect] = append(
			db.workflowToChildren[*opts.ParentWorkflowID][EntitySideEffect],
			int(id),
		)
		db.rwMutexRelationships.Unlock()
	}

	entity.UpdatedAt = time.Now()
	db.sideEffectEntities[id] = copySideEffectEntity(entity)
	db.rwMutexSideEffects.Unlock()

	return nil
}

func (db *MemoryDatabase) GetWorkflowExecution(id WorkflowExecutionID, opts ...WorkflowExecutionGetOption) (*WorkflowExecution, error) {
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

func (db *MemoryDatabase) GetWorkflowExecutions(entityID WorkflowEntityID) ([]*WorkflowExecution, error) {
	db.rwMutexWorkflowExec.RLock()
	results := make([]*WorkflowExecution, 0)
	for _, exec := range db.workflowExecutions {
		if exec.WorkflowEntityID == entityID {
			results = append(results, copyWorkflowExecution(exec))
		}
	}
	db.rwMutexWorkflowExec.RUnlock()
	return results, nil
}

func (db *MemoryDatabase) GetActivityExecution(id ActivityExecutionID, opts ...ActivityExecutionGetOption) (*ActivityExecution, error) {
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

func (db *MemoryDatabase) GetActivityExecutions(entityID ActivityEntityID) ([]*ActivityExecution, error) {
	db.rwMutexActivityExec.RLock()
	results := make([]*ActivityExecution, 0)
	for _, exec := range db.activityExecutions {
		if exec.ActivityEntityID == entityID {
			results = append(results, copyActivityExecution(exec))
		}
	}
	db.rwMutexActivityExec.RUnlock()
	return results, nil
}

func (db *MemoryDatabase) GetSagaExecution(id SagaExecutionID, opts ...SagaExecutionGetOption) (*SagaExecution, error) {
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

func (db *MemoryDatabase) GetSagaExecutions(entityID SagaEntityID) ([]*SagaExecution, error) {
	db.rwMutexSagaExec.RLock()
	results := make([]*SagaExecution, 0)
	for _, exec := range db.sagaExecutions {
		if exec.SagaEntityID == entityID {
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

func (db *MemoryDatabase) GetSideEffectExecution(id SideEffectExecutionID, opts ...SideEffectExecutionGetOption) (*SideEffectExecution, error) {
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

func (db *MemoryDatabase) GetSideEffectExecutions(entityID SideEffectEntityID) ([]*SideEffectExecution, error) {
	db.rwMutexSideEffectExec.RLock()
	results := make([]*SideEffectExecution, 0)
	for _, exec := range db.sideEffectExecutions {
		if exec.SideEffectEntityID == entityID {
			results = append(results, copySideEffectExecution(exec))
		}
	}
	db.rwMutexSideEffectExec.RUnlock()
	return results, nil
}

func (db *MemoryDatabase) GetActivityDataProperties(entityID ActivityEntityID, getters ...ActivityDataPropertyGetter) error {
	db.rwMutexActivityData.RLock()
	var data *ActivityData
	for _, d := range db.activityData {
		if d.EntityID == entityID {
			data = d
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

func (db *MemoryDatabase) SetActivityDataProperties(entityID ActivityEntityID, setters ...ActivityDataPropertySetter) error {
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

func (db *MemoryDatabase) GetSagaDataProperties(entityID SagaEntityID, getters ...SagaDataPropertyGetter) error {
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

func (db *MemoryDatabase) SetSagaDataProperties(entityID SagaEntityID, setters ...SagaDataPropertySetter) error {
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

func (db *MemoryDatabase) GetSideEffectDataProperties(entityID SideEffectEntityID, getters ...SideEffectDataPropertyGetter) error {
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

func (db *MemoryDatabase) SetSideEffectDataProperties(entityID SideEffectEntityID, setters ...SideEffectDataPropertySetter) error {
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

func (db *MemoryDatabase) GetWorkflowDataProperties(id WorkflowDataID, getters ...WorkflowDataPropertyGetter) error {
	db.rwMutexWorkflowData.RLock()
	data, exists := db.workflowData[id]
	if !exists {
		db.rwMutexWorkflowData.RUnlock()
		return ErrWorkflowEntityNotFound
	}
	dataCopy := copyWorkflowData(data)
	db.rwMutexWorkflowData.RUnlock()

	for _, getter := range getters {
		if _, err := getter(dataCopy); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetWorkflowDataProperties(id WorkflowDataID, setters ...WorkflowDataPropertySetter) error {
	db.rwMutexWorkflowData.Lock()
	data, exists := db.workflowData[id]
	if !exists {
		db.rwMutexWorkflowData.Unlock()
		return ErrWorkflowEntityNotFound
	}

	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			db.rwMutexWorkflowData.Unlock()
			return err
		}
	}

	db.workflowData[id] = copyWorkflowData(data)
	db.rwMutexWorkflowData.Unlock()
	return nil
}

func (db *MemoryDatabase) GetWorkflowExecutionDataProperties(entityID WorkflowExecutionID, getters ...WorkflowExecutionDataPropertyGetter) error {
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

func (db *MemoryDatabase) SetWorkflowExecutionDataProperties(entityID WorkflowExecutionID, setters ...WorkflowExecutionDataPropertySetter) error {
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

func (db *MemoryDatabase) GetActivityExecutionDataProperties(entityID ActivityExecutionID, getters ...ActivityExecutionDataPropertyGetter) error {
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

func (db *MemoryDatabase) SetActivityExecutionDataProperties(entityID ActivityExecutionID, setters ...ActivityExecutionDataPropertySetter) error {
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

func (db *MemoryDatabase) GetSagaExecutionDataProperties(entityID SagaExecutionID, getters ...SagaExecutionDataPropertyGetter) error {
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

func (db *MemoryDatabase) SetSagaExecutionDataProperties(entityID SagaExecutionID, setters ...SagaExecutionDataPropertySetter) error {
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

func (db *MemoryDatabase) GetSideEffectExecutionDataProperties(id SideEffectExecutionDataID, getters ...SideEffectExecutionDataPropertyGetter) error {
	db.rwMutexSideEffectExecData.RLock()
	data, exists := db.sideEffectExecutionData[id]
	if !exists {
		db.rwMutexSideEffectExecData.RUnlock()
		return ErrSideEffectExecutionNotFound
	}
	dataCopy := copySideEffectExecutionData(data)
	db.rwMutexSideEffectExecData.RUnlock()

	for _, getter := range getters {
		if _, err := getter(dataCopy); err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDatabase) SetSideEffectExecutionDataProperties(id SideEffectExecutionDataID, setters ...SideEffectExecutionDataPropertySetter) error {
	db.rwMutexSideEffectExecData.Lock()
	data, exists := db.sideEffectExecutionData[id]
	if !exists {
		db.rwMutexSideEffectExecData.Unlock()
		return ErrSideEffectExecutionNotFound
	}

	for _, setter := range setters {
		if _, err := setter(data); err != nil {
			db.rwMutexSideEffectExecData.Unlock()
			return err
		}
	}

	db.sideEffectExecutionData[id] = copySideEffectExecutionData(data)
	db.rwMutexSideEffectExecData.Unlock()
	return nil
}

func (db *MemoryDatabase) GetWorkflowExecutionProperties(id WorkflowExecutionID, getters ...WorkflowExecutionPropertyGetter) error {
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

func (db *MemoryDatabase) SetWorkflowExecutionProperties(id WorkflowExecutionID, setters ...WorkflowExecutionPropertySetter) error {
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

func (db *MemoryDatabase) GetActivityExecutionProperties(id ActivityExecutionID, getters ...ActivityExecutionPropertyGetter) error {
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

func (db *MemoryDatabase) SetActivityExecutionProperties(id ActivityExecutionID, setters ...ActivityExecutionPropertySetter) error {
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

func (db *MemoryDatabase) GetSagaExecutionProperties(id SagaExecutionID, getters ...SagaExecutionPropertyGetter) error {
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

func (db *MemoryDatabase) SetSagaExecutionProperties(id SagaExecutionID, setters ...SagaExecutionPropertySetter) error {
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

func (db *MemoryDatabase) GetSideEffectExecutionProperties(id SideEffectExecutionID, getters ...SideEffectExecutionPropertyGetter) error {
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

func (db *MemoryDatabase) SetSideEffectExecutionProperties(id SideEffectExecutionID, setters ...SideEffectExecutionPropertySetter) error {
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

func (db *MemoryDatabase) GetHierarchyProperties(id HierarchyID, getters ...HierarchyPropertyGetter) error {
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

func (db *MemoryDatabase) SetHierarchyProperties(id HierarchyID, setters ...HierarchyPropertySetter) error {
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

func (db *MemoryDatabase) GetQueueProperties(id QueueID, getters ...QueuePropertyGetter) error {
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

func (db *MemoryDatabase) SetQueueProperties(id QueueID, setters ...QueuePropertySetter) error {
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
func (db *MemoryDatabase) GetVersionByWorkflowAndChangeID(workflowID WorkflowEntityID, changeID string) (*Version, error) {
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

func (db *MemoryDatabase) GetVersionsByWorkflowID(workflowID WorkflowEntityID) ([]*Version, error) {
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
		version.ID = VersionID(db.versionCounter)
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

func (db *MemoryDatabase) DeleteVersionsForWorkflow(workflowID WorkflowEntityID) error {
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

func (db *MemoryDatabase) GetVersionProperties(id VersionID, getters ...VersionPropertyGetter) error {
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

func (db *MemoryDatabase) SetVersionProperties(id VersionID, setters ...VersionPropertySetter) error {
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

func (db *MemoryDatabase) AddWorkflowData(entityID WorkflowEntityID, data *WorkflowData) (WorkflowDataID, error) {
	db.rwMutexWorkflowData.Lock()
	db.workflowDataCounter++
	data.ID = WorkflowDataID(db.workflowDataCounter)
	data.EntityID = entityID
	db.workflowData[data.ID] = copyWorkflowData(data)
	db.rwMutexWorkflowData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) AddActivityData(entityID ActivityEntityID, data *ActivityData) (ActivityDataID, error) {
	db.rwMutexActivityData.Lock()
	db.activityDataCounter++
	data.ID = ActivityDataID(db.activityDataCounter)
	data.EntityID = entityID
	db.activityData[data.ID] = copyActivityData(data)
	db.rwMutexActivityData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) AddSagaData(entityID SagaEntityID, data *SagaData) (SagaDataID, error) {
	db.rwMutexSagaData.Lock()
	db.sagaDataCounter++
	data.ID = SagaDataID(db.sagaDataCounter)
	data.EntityID = entityID
	db.sagaData[data.ID] = copySagaData(data)
	db.rwMutexSagaData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) AddSideEffectData(entityID SideEffectEntityID, data *SideEffectData) (SideEffectDataID, error) {
	db.rwMutexSideEffectData.Lock()
	db.sideEffectDataCounter++
	data.ID = SideEffectDataID(db.sideEffectDataCounter)
	data.EntityID = entityID
	db.sideEffectData[data.ID] = copySideEffectData(data)
	db.rwMutexSideEffectData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) GetWorkflowData(id WorkflowDataID) (*WorkflowData, error) {
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

func (db *MemoryDatabase) GetActivityData(id ActivityDataID) (*ActivityData, error) {
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

func (db *MemoryDatabase) GetActivityDataByEntityID(entityID ActivityEntityID) (*ActivityData, error) {
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

func (db *MemoryDatabase) GetSagaData(id SagaDataID) (*SagaData, error) {
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

func (db *MemoryDatabase) GetSideEffectData(id SideEffectDataID) (*SideEffectData, error) {
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

func (db *MemoryDatabase) GetWorkflowDataByEntityID(entityID WorkflowEntityID) (*WorkflowData, error) {
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

func (db *MemoryDatabase) GetSagaDataByEntityID(entityID SagaEntityID) (*SagaData, error) {
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

func (db *MemoryDatabase) GetSideEffectDataByEntityID(entityID SideEffectEntityID) (*SideEffectData, error) {
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

func (db *MemoryDatabase) AddWorkflowExecutionData(executionID WorkflowExecutionID, data *WorkflowExecutionData) (WorkflowExecutionDataID, error) {
	db.rwMutexWorkflowExecData.Lock()
	db.workflowExecutionDataCounter++
	data.ID = WorkflowExecutionDataID(db.workflowExecutionDataCounter)
	data.ExecutionID = executionID
	db.workflowExecutionData[data.ID] = copyWorkflowExecutionData(data)
	db.rwMutexWorkflowExecData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) AddActivityExecutionData(executionID ActivityExecutionID, data *ActivityExecutionData) (ActivityExecutionDataID, error) {
	db.rwMutexActivityExecData.Lock()
	db.activityExecutionDataCounter++
	data.ID = ActivityExecutionDataID(db.activityExecutionDataCounter)
	data.ExecutionID = executionID
	db.activityExecutionData[data.ID] = copyActivityExecutionData(data)
	db.rwMutexActivityExecData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) AddSagaExecutionData(executionID SagaExecutionID, data *SagaExecutionData) (SagaExecutionDataID, error) {
	db.rwMutexSagaExecData.Lock()
	db.sagaExecutionDataCounter++
	data.ID = SagaExecutionDataID(db.sagaExecutionDataCounter)
	data.ExecutionID = executionID
	db.sagaExecutionData[data.ID] = copySagaExecutionData(data)
	db.rwMutexSagaExecData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) AddSideEffectExecutionData(executionID SideEffectExecutionID, data *SideEffectExecutionData) (SideEffectExecutionDataID, error) {
	db.rwMutexSideEffectExecData.Lock()
	db.sideEffectExecutionDataCounter++
	data.ID = SideEffectExecutionDataID(db.sideEffectExecutionDataCounter)
	data.ExecutionID = executionID
	db.sideEffectExecutionData[data.ID] = copySideEffectExecutionData(data)
	db.rwMutexSideEffectExecData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) GetWorkflowExecutionData(id WorkflowExecutionDataID) (*WorkflowExecutionData, error) {
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

func (db *MemoryDatabase) GetActivityExecutionData(id ActivityExecutionDataID) (*ActivityExecutionData, error) {
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

func (db *MemoryDatabase) GetSagaExecutionData(id SagaExecutionDataID) (*SagaExecutionData, error) {
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

func (db *MemoryDatabase) GetSideEffectExecutionData(id SideEffectExecutionDataID) (*SideEffectExecutionData, error) {
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

func (db *MemoryDatabase) GetWorkflowExecutionDataByExecutionID(executionID WorkflowExecutionID) (*WorkflowExecutionData, error) {
	db.rwMutexRelationships.RLock()
	dataID, exists := db.workflowExecToDataMap[executionID]
	db.rwMutexRelationships.RUnlock()

	if !exists {
		return nil, ErrWorkflowExecutionNotFound
	}

	db.rwMutexWorkflowExecData.RLock()
	defer db.rwMutexWorkflowExecData.RUnlock()

	if data, ok := db.workflowExecutionData[dataID]; ok {
		return copyWorkflowExecutionData(data), nil
	}

	return nil, ErrWorkflowExecutionNotFound
}

func (db *MemoryDatabase) GetActivityExecutionDataByExecutionID(executionID ActivityExecutionID) (*ActivityExecutionData, error) {
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

func (db *MemoryDatabase) GetSagaExecutionDataByExecutionID(executionID SagaExecutionID) (*SagaExecutionData, error) {
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

func (db *MemoryDatabase) GetSideEffectExecutionDataByExecutionID(executionID SideEffectExecutionID) (*SideEffectExecutionData, error) {
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

func (db *MemoryDatabase) HasRun(id RunID) (bool, error) {
	db.rwMutexRuns.RLock()
	_, exists := db.runs[id]
	db.rwMutexRuns.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasVersion(id VersionID) (bool, error) {
	db.rwMutexVersions.RLock()
	_, exists := db.versions[id]
	db.rwMutexVersions.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasHierarchy(id HierarchyID) (bool, error) {
	db.rwMutexHierarchies.RLock()
	_, exists := db.hierarchies[id]
	db.rwMutexHierarchies.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasQueue(id QueueID) (bool, error) {
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

func (db *MemoryDatabase) HasWorkflowEntity(id WorkflowEntityID) (bool, error) {
	db.rwMutexWorkflows.RLock()
	_, exists := db.workflowEntities[id]
	db.rwMutexWorkflows.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasActivityEntity(id ActivityEntityID) (bool, error) {
	db.rwMutexActivities.RLock()
	_, exists := db.activityEntities[id]
	db.rwMutexActivities.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSagaEntity(id SagaEntityID) (bool, error) {
	db.rwMutexSagas.RLock()
	_, exists := db.sagaEntities[id]
	db.rwMutexSagas.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectEntity(id SideEffectEntityID) (bool, error) {
	db.rwMutexSideEffects.RLock()
	_, exists := db.sideEffectEntities[id]
	db.rwMutexSideEffects.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowExecution(id WorkflowExecutionID) (bool, error) {
	db.rwMutexWorkflowExec.RLock()
	_, exists := db.workflowExecutions[id]
	db.rwMutexWorkflowExec.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasActivityExecution(id ActivityExecutionID) (bool, error) {
	db.rwMutexActivityExec.RLock()
	_, exists := db.activityExecutions[id]
	db.rwMutexActivityExec.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSagaExecution(id SagaExecutionID) (bool, error) {
	db.rwMutexSagaExec.RLock()
	_, exists := db.sagaExecutions[id]
	db.rwMutexSagaExec.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectExecution(id SideEffectExecutionID) (bool, error) {
	db.rwMutexSideEffectExec.RLock()
	_, exists := db.sideEffectExecutions[id]
	db.rwMutexSideEffectExec.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowData(id WorkflowDataID) (bool, error) {
	db.rwMutexWorkflowData.RLock()
	_, exists := db.workflowData[id]
	db.rwMutexWorkflowData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasActivityData(id ActivityDataID) (bool, error) {
	db.rwMutexActivityData.RLock()
	_, exists := db.activityData[id]
	db.rwMutexActivityData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSagaData(id SagaDataID) (bool, error) {
	db.rwMutexSagaData.RLock()
	_, exists := db.sagaData[id]
	db.rwMutexSagaData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectData(id SideEffectDataID) (bool, error) {
	db.rwMutexSideEffectData.RLock()
	_, exists := db.sideEffectData[id]
	db.rwMutexSideEffectData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowDataByEntityID(entityID WorkflowEntityID) (bool, error) {
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

func (db *MemoryDatabase) HasActivityDataByEntityID(entityID ActivityEntityID) (bool, error) {
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

func (db *MemoryDatabase) HasSagaDataByEntityID(entityID SagaEntityID) (bool, error) {
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

func (db *MemoryDatabase) HasSideEffectDataByEntityID(entityID SideEffectEntityID) (bool, error) {
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

func (db *MemoryDatabase) HasWorkflowExecutionData(id WorkflowExecutionDataID) (bool, error) {
	db.rwMutexWorkflowExecData.RLock()
	_, exists := db.workflowExecutionData[id]
	db.rwMutexWorkflowExecData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasActivityExecutionData(id ActivityExecutionDataID) (bool, error) {
	db.rwMutexActivityExecData.RLock()
	_, exists := db.activityExecutionData[id]
	db.rwMutexActivityExecData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSagaExecutionData(id SagaExecutionDataID) (bool, error) {
	db.rwMutexSagaExecData.RLock()
	_, exists := db.sagaExecutionData[id]
	db.rwMutexSagaExecData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSideEffectExecutionData(id SideEffectExecutionDataID) (bool, error) {
	db.rwMutexSideEffectExecData.RLock()
	_, exists := db.sideEffectExecutionData[id]
	db.rwMutexSideEffectExecData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasWorkflowExecutionDataByExecutionID(executionID WorkflowExecutionID) (bool, error) {
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

func (db *MemoryDatabase) HasActivityExecutionDataByExecutionID(executionID ActivityExecutionID) (bool, error) {
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

func (db *MemoryDatabase) HasSagaExecutionDataByExecutionID(executionID SagaExecutionID) (bool, error) {
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

func (db *MemoryDatabase) HasSideEffectExecutionDataByExecutionID(executionID SideEffectExecutionID) (bool, error) {
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

func (db *MemoryDatabase) GetWorkflowExecutionLatestByEntityID(entityID WorkflowEntityID) (*WorkflowExecution, error) {
	db.rwMutexWorkflowExec.RLock()
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
	db.rwMutexWorkflowExec.RUnlock()

	if latestExec == nil {
		return nil, ErrWorkflowExecutionNotFound
	}
	return copyWorkflowExecution(latestExec), nil
}

func (db *MemoryDatabase) GetActivityExecutionLatestByEntityID(entityID ActivityEntityID) (*ActivityExecution, error) {
	db.rwMutexActivityExec.RLock()
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
	db.rwMutexActivityExec.RUnlock()

	if latestExec == nil {
		return nil, ErrActivityExecutionNotFound
	}
	return copyActivityExecution(latestExec), nil
}

func (db *MemoryDatabase) SetSagaValue(executionID SagaExecutionID, key string, value []byte) (SagaValueID, error) {
	db.rwMutexSagas.Lock()
	defer db.rwMutexSagas.Unlock()

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
	db.rwMutexSagas.RLock()
	defer db.rwMutexSagas.RUnlock()

	for _, keyMap := range db.sagaValues {
		if val, exists := keyMap[key]; exists && val.ID == id {
			return val.Value, nil
		}
	}
	return nil, ErrSagaValueNotFound
}

func (db *MemoryDatabase) GetSagaValueByExecutionID(executionID SagaExecutionID, key string) ([]byte, error) {
	db.rwMutexSagas.RLock()
	defer db.rwMutexSagas.RUnlock()

	if keyMap, exists := db.sagaValues[executionID]; exists {
		if val, vexists := keyMap[key]; vexists {
			return val.Value, nil
		}
	}
	return nil, ErrSagaValueNotFound
}

func (db *MemoryDatabase) AddSignalEntity(entity *SignalEntity, parentWorkflowID WorkflowEntityID) (SignalEntityID, error) {
	db.rwMutexSignals.Lock()
	db.signalEntityCounter++
	entity.ID = SignalEntityID(db.signalEntityCounter)

	if entity.SignalData != nil {
		db.rwMutexSignalData.Lock()
		db.signalDataCounter++
		entity.SignalData.ID = SignalDataID(db.signalDataCounter)
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

func (db *MemoryDatabase) GetSignalEntity(id SignalEntityID, opts ...SignalEntityGetOption) (*SignalEntity, error) {
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

func (db *MemoryDatabase) GetSignalEntities(workflowID WorkflowEntityID, opts ...SignalEntityGetOption) ([]*SignalEntity, error) {
	db.rwMutexHierarchies.RLock()
	var entities []*SignalEntity
	for _, h := range db.hierarchies {
		if h.ParentEntityID == int(workflowID) && h.ChildType == EntitySignal {
			db.rwMutexSignals.RLock()
			if e, ex := db.signalEntities[SignalEntityID(h.ChildEntityID)]; ex {
				entities = append(entities, copySignalEntity(e))
			}
			db.rwMutexSignals.RUnlock()
		}
	}
	db.rwMutexHierarchies.RUnlock()

	return entities, nil
}

func (db *MemoryDatabase) AddSignalExecution(exec *SignalExecution) (SignalExecutionID, error) {
	db.rwMutexSignalExec.Lock()
	db.signalExecutionCounter++
	exec.ID = SignalExecutionID(db.signalExecutionCounter)

	if exec.SignalExecutionData != nil {
		db.rwMutexSignalExecData.Lock()
		db.signalExecutionDataCounter++
		exec.SignalExecutionData.ID = SignalExecutionDataID(db.signalExecutionDataCounter)
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

func (db *MemoryDatabase) GetSignalExecution(id SignalExecutionID, opts ...SignalExecutionGetOption) (*SignalExecution, error) {
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

func (db *MemoryDatabase) GetSignalExecutions(entityID SignalEntityID) ([]*SignalExecution, error) {
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

func (db *MemoryDatabase) GetSignalExecutionLatestByEntityID(entityID SignalEntityID) (*SignalExecution, error) {
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

func (db *MemoryDatabase) HasSignalEntity(id SignalEntityID) (bool, error) {
	db.rwMutexSignals.RLock()
	_, exists := db.signalEntities[id]
	db.rwMutexSignals.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSignalExecution(id SignalExecutionID) (bool, error) {
	db.rwMutexSignalExec.RLock()
	_, exists := db.signalExecutions[id]
	db.rwMutexSignalExec.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSignalData(id SignalDataID) (bool, error) {
	db.rwMutexSignalData.RLock()
	_, exists := db.signalData[id]
	db.rwMutexSignalData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) HasSignalExecutionData(id SignalExecutionDataID) (bool, error) {
	db.rwMutexSignalExecData.RLock()
	_, exists := db.signalExecutionData[id]
	db.rwMutexSignalExecData.RUnlock()
	return exists, nil
}

func (db *MemoryDatabase) GetSignalEntityProperties(id SignalEntityID, getters ...SignalEntityPropertyGetter) error {
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

func (db *MemoryDatabase) SetSignalEntityProperties(id SignalEntityID, setters ...SignalEntityPropertySetter) error {
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

func (db *MemoryDatabase) GetSignalExecutionProperties(id SignalExecutionID, getters ...SignalExecutionPropertyGetter) error {
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

func (db *MemoryDatabase) SetSignalExecutionProperties(id SignalExecutionID, setters ...SignalExecutionPropertySetter) error {
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

func (db *MemoryDatabase) AddSignalData(entityID SignalEntityID, data *SignalData) (SignalDataID, error) {
	db.rwMutexSignalData.Lock()
	db.signalDataCounter++
	data.ID = SignalDataID(db.signalDataCounter)
	data.EntityID = entityID
	db.signalData[data.ID] = copySignalData(data)
	db.rwMutexSignalData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) GetSignalData(id SignalDataID) (*SignalData, error) {
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

func (db *MemoryDatabase) GetSignalDataByEntityID(entityID SignalEntityID) (*SignalData, error) {
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

func (db *MemoryDatabase) HasSignalDataByEntityID(entityID SignalEntityID) (bool, error) {
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

func (db *MemoryDatabase) GetSignalDataProperties(entityID SignalEntityID, getters ...SignalDataPropertyGetter) error {
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

func (db *MemoryDatabase) SetSignalDataProperties(entityID SignalEntityID, setters ...SignalDataPropertySetter) error {
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

func (db *MemoryDatabase) AddSignalExecutionData(executionID SignalExecutionID, data *SignalExecutionData) (SignalExecutionDataID, error) {
	db.rwMutexSignalExecData.Lock()
	db.signalExecutionDataCounter++
	data.ID = SignalExecutionDataID(db.signalExecutionDataCounter)
	data.ExecutionID = executionID
	db.signalExecutionData[data.ID] = copySignalExecutionData(data)
	db.rwMutexSignalExecData.Unlock()
	return data.ID, nil
}

func (db *MemoryDatabase) GetSignalExecutionData(id SignalExecutionDataID) (*SignalExecutionData, error) {
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

func (db *MemoryDatabase) GetSignalExecutionDataByExecutionID(executionID SignalExecutionID) (*SignalExecutionData, error) {
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

func (db *MemoryDatabase) HasSignalExecutionDataByExecutionID(executionID SignalExecutionID) (bool, error) {
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

func (db *MemoryDatabase) GetSignalExecutionDataProperties(entityID SignalExecutionID, getters ...SignalExecutionDataPropertyGetter) error {
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

func (db *MemoryDatabase) SetSignalExecutionDataProperties(entityID SignalExecutionID, setters ...SignalExecutionDataPropertySetter) error {
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

func (db *MemoryDatabase) DeleteRunsByStatus(status RunStatus) error {
	db.rwMutexRuns.RLock()
	runsToDelete := make([]RunID, 0)
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
		workflowIDs := make([]WorkflowEntityID, len(db.runToWorkflows[runID]))
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

func (db *MemoryDatabase) deleteWorkflowAndChildren(workflowID WorkflowEntityID) error {
	db.rwMutexHierarchies.RLock()
	var hierarchies []*Hierarchy
	for _, h := range db.hierarchies {
		if h.ParentEntityID == int(workflowID) {
			hierarchies = append(hierarchies, copyHierarchy(h))
		}
	}
	db.rwMutexHierarchies.RUnlock()

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

	db.rwMutexHierarchies.Lock()
	for id, hh := range db.hierarchies {
		if hh.ParentEntityID == int(workflowID) || hh.ChildEntityID == int(workflowID) {
			delete(db.hierarchies, id)
		}
	}
	db.rwMutexHierarchies.Unlock()

	return nil
}

func (db *MemoryDatabase) deleteWorkflowEntity(workflowID WorkflowEntityID) error {
	db.rwMutexWorkflowExec.Lock()
	for id, exec := range db.workflowExecutions {
		if exec.WorkflowEntityID == workflowID {
			db.rwMutexRelationships.Lock()
			if dataID, exists := db.workflowExecToDataMap[id]; exists {
				db.rwMutexWorkflowExecData.Lock()
				delete(db.workflowExecutionData, dataID)
				db.rwMutexWorkflowExecData.Unlock()
				delete(db.workflowExecToDataMap, id)
			}
			db.rwMutexRelationships.Unlock()
		}
	}
	db.rwMutexWorkflowExec.Unlock()

	db.rwMutexWorkflowData.Lock()
	for id, data := range db.workflowData {
		if data.EntityID == workflowID {
			delete(db.workflowData, id)
		}
	}
	db.rwMutexWorkflowData.Unlock()

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

	db.rwMutexRelationships.Lock()
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
	db.rwMutexRelationships.Unlock()

	db.rwMutexWorkflows.Lock()
	delete(db.workflowEntities, workflowID)
	db.rwMutexWorkflows.Unlock()

	return nil
}

func (db *MemoryDatabase) deleteActivityEntity(entityID ActivityEntityID) error {
	db.rwMutexActivityExec.Lock()
	for id, exec := range db.activityExecutions {
		if exec.ActivityEntityID == entityID {
			db.rwMutexRelationships.Lock()
			if dataID, exists := db.activityExecToDataMap[id]; exists {
				db.rwMutexActivityExecData.Lock()
				delete(db.activityExecutionData, dataID)
				db.rwMutexActivityExecData.Unlock()
				delete(db.activityExecToDataMap, id)
			}
			db.rwMutexRelationships.Unlock()
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

func (db *MemoryDatabase) deleteSagaEntity(entityID SagaEntityID) error {
	db.rwMutexSagaExec.Lock()
	for id, exec := range db.sagaExecutions {
		if exec.SagaEntityID == entityID {
			db.rwMutexRelationships.Lock()
			if dataID, exists := db.sagaExecToDataMap[id]; exists {
				db.rwMutexSagaExecData.Lock()
				delete(db.sagaExecutionData, dataID)
				db.rwMutexSagaExecData.Unlock()
				delete(db.sagaExecToDataMap, id)
			}
			db.rwMutexRelationships.Unlock()

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

func (db *MemoryDatabase) deleteSideEffectEntity(entityID SideEffectEntityID) error {
	db.rwMutexSideEffectExec.Lock()
	for id, exec := range db.sideEffectExecutions {
		if exec.SideEffectEntityID == entityID {
			db.rwMutexRelationships.Lock()
			if dataID, exists := db.sideEffectExecToDataMap[id]; exists {
				db.rwMutexSideEffectExecData.Lock()
				delete(db.sideEffectExecutionData, dataID)
				db.rwMutexSideEffectExecData.Unlock()
				delete(db.sideEffectExecToDataMap, id)
			}
			db.rwMutexRelationships.Unlock()
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

func (db *MemoryDatabase) deleteSignalEntity(entityID SignalEntityID) error {
	db.rwMutexSignalExec.Lock()
	for id, exec := range db.signalExecutions {
		if exec.EntityID == entityID {
			db.rwMutexRelationships.Lock()
			if dataID, exists := db.signalExecToDataMap[id]; exists {
				db.rwMutexSignalExecData.Lock()
				delete(db.signalExecutionData, dataID)
				db.rwMutexSignalExecData.Unlock()
				delete(db.signalExecToDataMap, id)
			}
			db.rwMutexRelationships.Unlock()
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
		SagaExecutionData: copySagaExecutionData(exec.SagaExecutionData),
	}
}

func copySagaEntity(entity *SagaEntity) *SagaEntity {
	if entity == nil {
		return nil
	}

	return &SagaEntity{
		BaseEntity: *copyBaseEntity(&entity.BaseEntity),
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
