package tempolite

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/sasha-s/go-deadlock"
)

// DefaultDatabase implementation
type DefaultDatabase struct {
	mu deadlock.RWMutex

	// Core maps
	runs        map[int]*Run
	versions    map[int]*Version
	hierarchies map[int]*Hierarchy
	queues      map[int]*Queue

	// Entity maps
	workflowEntities   map[int]*WorkflowEntity
	activityEntities   map[int]*ActivityEntity
	sagaEntities       map[int]*SagaEntity
	sideEffectEntities map[int]*SideEffectEntity

	// Execution maps
	workflowExecutions   map[int]*WorkflowExecution
	activityExecutions   map[int]*ActivityExecution
	sagaExecutions       map[int]*SagaExecution
	sideEffectExecutions map[int]*SideEffectExecution

	// Relationship maps
	entityToWorkflow   map[int]int                  // entity ID -> workflow ID (for activity/saga/sideeffect)
	workflowToChildren map[int]map[EntityType][]int // workflow ID -> type -> child entity IDs
	workflowToVersion  map[int][]int                // workflow ID -> version IDs
	workflowToQueue    map[int]int                  // workflow ID -> queue ID
	queueToWorkflows   map[int][]int                // queue ID -> workflow IDs
	runToWorkflows     map[int][]int                // run ID -> workflow IDs

	// Entity Data maps
	workflowData   map[int]*WorkflowData   // entity ID -> data
	activityData   map[int]*ActivityData   // entity ID -> data
	sagaData       map[int]*SagaData       // entity ID -> data
	sideEffectData map[int]*SideEffectData // entity ID -> data

	// Execution Data maps
	workflowExecutionData   map[int]*WorkflowExecutionData   // execution ID -> data
	activityExecutionData   map[int]*ActivityExecutionData   // execution ID -> data
	sagaExecutionData       map[int]*SagaExecutionData       // execution ID -> data
	sideEffectExecutionData map[int]*SideEffectExecutionData // execution ID -> data

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
		runs:        make(map[int]*Run),
		versions:    make(map[int]*Version),
		hierarchies: make(map[int]*Hierarchy),
		queues:      make(map[int]*Queue),

		// Entity maps
		workflowEntities:   make(map[int]*WorkflowEntity),
		activityEntities:   make(map[int]*ActivityEntity),
		sagaEntities:       make(map[int]*SagaEntity),
		sideEffectEntities: make(map[int]*SideEffectEntity),

		// Execution maps
		workflowExecutions:   make(map[int]*WorkflowExecution),
		activityExecutions:   make(map[int]*ActivityExecution),
		sagaExecutions:       make(map[int]*SagaExecution),
		sideEffectExecutions: make(map[int]*SideEffectExecution),

		// Relationship maps
		entityToWorkflow:   make(map[int]int),
		workflowToChildren: make(map[int]map[EntityType][]int),
		workflowToVersion:  make(map[int][]int),
		workflowToQueue:    make(map[int]int),
		queueToWorkflows:   make(map[int][]int),
		runToWorkflows:     make(map[int][]int),

		// Entity Data maps
		workflowData:   make(map[int]*WorkflowData),
		activityData:   make(map[int]*ActivityData),
		sagaData:       make(map[int]*SagaData),
		sideEffectData: make(map[int]*SideEffectData),

		// Execution Data maps
		workflowExecutionData:   make(map[int]*WorkflowExecutionData),
		activityExecutionData:   make(map[int]*ActivityExecutionData),
		sagaExecutionData:       make(map[int]*SagaExecutionData),
		sideEffectExecutionData: make(map[int]*SideEffectExecutionData),

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

// Helper function to remove an element from a slice of integers
func removeFromSlice[T comparable](slice *[]T, value T) {
	for i, v := range *slice {
		if v == value {
			*slice = append((*slice)[:i], (*slice)[i+1:]...)
			return
		}
	}
}

// Entity Data Operations
func (db *DefaultDatabase) GetWorkflowDataProperties(entityID int, getters ...WorkflowDataPropertyGetter) error {
	db.mu.RLock()
	data, exists := db.workflowData[entityID]
	if !exists {
		db.mu.RUnlock()
		return fmt.Errorf("workflow data for entity %d not found", entityID)
	}
	dataCopy := copyWorkflowData(data)
	db.mu.RUnlock()

	for _, getter := range getters {
		if option, err := getter(dataCopy); err != nil {
			return err
		} else if option != nil {
			opts := &WorkflowDataGetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) SetWorkflowDataProperties(entityID int, setters ...WorkflowDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.workflowData[entityID]
	if !exists {
		return fmt.Errorf("workflow data for entity %d not found", entityID)
	}

	dataCopy := copyWorkflowData(data)
	for _, setter := range setters {
		if option, err := setter(dataCopy); err != nil {
			return err
		} else if option != nil {
			opts := &WorkflowDataSetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
		}
	}

	db.workflowData[entityID] = dataCopy
	return nil
}

func (db *DefaultDatabase) GetActivityDataProperties(entityID int, getters ...ActivityDataPropertyGetter) error {
	db.mu.RLock()
	data, exists := db.activityData[entityID]
	if !exists {
		db.mu.RUnlock()
		return fmt.Errorf("activity data for entity %d not found", entityID)
	}
	dataCopy := copyActivityData(data)
	db.mu.RUnlock()

	for _, getter := range getters {
		if option, err := getter(dataCopy); err != nil {
			return err
		} else if option != nil {
			opts := &ActivityDataGetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) SetActivityDataProperties(entityID int, setters ...ActivityDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.activityData[entityID]
	if !exists {
		return fmt.Errorf("activity data for entity %d not found", entityID)
	}

	dataCopy := copyActivityData(data)
	for _, setter := range setters {
		if option, err := setter(dataCopy); err != nil {
			return err
		} else if option != nil {
			opts := &ActivityDataSetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
		}
	}

	db.activityData[entityID] = dataCopy
	return nil
}

func (db *DefaultDatabase) GetSagaDataProperties(entityID int, getters ...SagaDataPropertyGetter) error {
	db.mu.RLock()
	data, exists := db.sagaData[entityID]
	if !exists {
		db.mu.RUnlock()
		return fmt.Errorf("saga data for entity %d not found", entityID)
	}
	dataCopy := copySagaData(data)
	db.mu.RUnlock()

	for _, getter := range getters {
		if option, err := getter(dataCopy); err != nil {
			return err
		} else if option != nil {
			opts := &SagaDataGetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) SetSagaDataProperties(entityID int, setters ...SagaDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.sagaData[entityID]
	if !exists {
		return fmt.Errorf("saga data for entity %d not found", entityID)
	}

	dataCopy := copySagaData(data)
	for _, setter := range setters {
		if option, err := setter(dataCopy); err != nil {
			return err
		} else if option != nil {
			opts := &SagaDataSetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
		}
	}

	db.sagaData[entityID] = dataCopy
	return nil
}

func (db *DefaultDatabase) GetSideEffectDataProperties(entityID int, getters ...SideEffectDataPropertyGetter) error {
	db.mu.RLock()
	data, exists := db.sideEffectData[entityID]
	if !exists {
		db.mu.RUnlock()
		return fmt.Errorf("side effect data for entity %d not found", entityID)
	}
	dataCopy := copySideEffectData(data)
	db.mu.RUnlock()

	for _, getter := range getters {
		if option, err := getter(dataCopy); err != nil {
			return err
		} else if option != nil {
			opts := &SideEffectDataGetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) SetSideEffectDataProperties(entityID int, setters ...SideEffectDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.sideEffectData[entityID]
	if !exists {
		return fmt.Errorf("side effect data for entity %d not found", entityID)
	}

	dataCopy := copySideEffectData(data)
	for _, setter := range setters {
		if option, err := setter(dataCopy); err != nil {
			return err
		} else if option != nil {
			opts := &SideEffectDataSetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
		}
	}

	db.sideEffectData[entityID] = dataCopy
	return nil
}

// Execution Data Properties
func (db *DefaultDatabase) GetWorkflowExecutionDataProperties(entityID int, getters ...WorkflowExecutionDataPropertyGetter) error {
	db.mu.RLock()
	data, exists := db.workflowExecutionData[entityID]
	if !exists {
		db.mu.RUnlock()
		return fmt.Errorf("workflow execution data for entity %d not found", entityID)
	}
	dataCopy := copyWorkflowExecutionData(data)
	db.mu.RUnlock()

	for _, getter := range getters {
		if option, err := getter(dataCopy); err != nil {
			return err
		} else if option != nil {
			opts := &WorkflowExecutionDataGetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) SetWorkflowExecutionDataProperties(entityID int, setters ...WorkflowExecutionDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.workflowExecutionData[entityID]
	if !exists {
		return fmt.Errorf("workflow execution data for entity %d not found", entityID)
	}

	dataCopy := copyWorkflowExecutionData(data)
	for _, setter := range setters {
		if option, err := setter(dataCopy); err != nil {
			return err
		} else if option != nil {
			opts := &WorkflowExecutionDataSetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
		}
	}

	db.workflowExecutionData[entityID] = dataCopy
	return nil
}

func (db *DefaultDatabase) GetActivityExecutionDataProperties(entityID int, getters ...ActivityExecutionDataPropertyGetter) error {
	db.mu.RLock()
	data, exists := db.activityExecutionData[entityID]
	if !exists {
		db.mu.RUnlock()
		return fmt.Errorf("activity execution data for entity %d not found", entityID)
	}
	dataCopy := copyActivityExecutionData(data)
	db.mu.RUnlock()

	for _, getter := range getters {
		if option, err := getter(dataCopy); err != nil {
			return err
		} else if option != nil {
			opts := &ActivityExecutionDataGetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) SetActivityExecutionDataProperties(entityID int, setters ...ActivityExecutionDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.activityExecutionData[entityID]
	if !exists {
		return fmt.Errorf("activity execution data for entity %d not found", entityID)
	}

	dataCopy := copyActivityExecutionData(data)
	for _, setter := range setters {
		if option, err := setter(dataCopy); err != nil {
			return err
		} else if option != nil {
			opts := &ActivityExecutionDataSetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
		}
	}

	db.activityExecutionData[entityID] = dataCopy
	return nil
}

func (db *DefaultDatabase) GetSagaExecutionDataProperties(entityID int, getters ...SagaExecutionDataPropertyGetter) error {
	db.mu.RLock()
	data, exists := db.sagaExecutionData[entityID]
	if !exists {
		db.mu.RUnlock()
		return fmt.Errorf("saga execution data for entity %d not found", entityID)
	}
	dataCopy := copySagaExecutionData(data)
	db.mu.RUnlock()

	for _, getter := range getters {
		if option, err := getter(dataCopy); err != nil {
			return err
		} else if option != nil {
			opts := &SagaExecutionDataGetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) SetSagaExecutionDataProperties(entityID int, setters ...SagaExecutionDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.sagaExecutionData[entityID]
	if !exists {
		return fmt.Errorf("saga execution data for entity %d not found", entityID)
	}

	dataCopy := copySagaExecutionData(data)
	for _, setter := range setters {
		if option, err := setter(dataCopy); err != nil {
			return err
		} else if option != nil {
			opts := &SagaExecutionDataSetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
		}
	}

	db.sagaExecutionData[entityID] = dataCopy
	return nil
}

func (db *DefaultDatabase) GetSideEffectExecutionDataProperties(entityID int, getters ...SideEffectExecutionDataPropertyGetter) error {
	db.mu.RLock()
	data, exists := db.sideEffectExecutionData[entityID]
	if !exists {
		db.mu.RUnlock()
		return fmt.Errorf("side effect execution data for entity %d not found", entityID)
	}
	dataCopy := copySideEffectExecutionData(data)
	db.mu.RUnlock()

	for _, getter := range getters {
		if option, err := getter(dataCopy); err != nil {
			return err
		} else if option != nil {
			opts := &SideEffectExecutionDataGetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) SetSideEffectExecutionDataProperties(entityID int, setters ...SideEffectExecutionDataPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, exists := db.sideEffectExecutionData[entityID]
	if !exists {
		return fmt.Errorf("side effect execution data for entity %d not found", entityID)
	}

	dataCopy := copySideEffectExecutionData(data)
	for _, setter := range setters {
		if option, err := setter(dataCopy); err != nil {
			return err
		} else if option != nil {
			opts := &SideEffectExecutionDataSetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
		}
	}

	db.sideEffectExecutionData[entityID] = dataCopy
	return nil
}

// Run operations
func (db *DefaultDatabase) AddRun(run *Run) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	runCopy := copyRun(run)
	runCopy.ID = int(atomic.AddInt64(&db.runCounter, 1))
	runCopy.CreatedAt = time.Now()
	runCopy.UpdatedAt = time.Now()

	db.runs[runCopy.ID] = runCopy
	return runCopy.ID, nil
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

	runCopy := copyRun(run)
	runCopy.UpdatedAt = time.Now()
	db.runs[run.ID] = runCopy
	return nil
}

func (db *DefaultDatabase) GetRunProperties(id int, getters ...RunPropertyGetter) error {
	db.mu.RLock()
	run, exists := db.runs[id]
	if !exists {
		db.mu.RUnlock()
		return ErrRunNotFound
	}
	runCopy := copyRun(run)
	db.mu.RUnlock()

	for _, getter := range getters {
		if option, err := getter(runCopy); err != nil {
			return err
		} else if option != nil {
			opts := &RunPropertyGetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
			// Handle options
			if opts.IncludeWorkflows {
				if workflows, exists := db.runToWorkflows[id]; exists {
					for _, wfID := range workflows {
						if wf, exists := db.workflowEntities[wfID]; exists {
							runCopy.Entities = append(runCopy.Entities, copyWorkflowEntity(wf))
						}
					}
				}
			}
			if opts.IncludeHierarchies {
				// Load hierarchies if needed
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) SetRunProperties(id int, setters ...RunPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	run, exists := db.runs[id]
	if !exists {
		return ErrRunNotFound
	}

	runCopy := copyRun(run)
	for _, setter := range setters {
		if option, err := setter(runCopy); err != nil {
			return err
		} else if option != nil {
			opts := &RunPropertySetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
			// Handle relationship changes
			if opts.WorkflowID != nil {
				db.runToWorkflows[id] = append(db.runToWorkflows[id], *opts.WorkflowID)
			}
		}
	}

	runCopy.UpdatedAt = time.Now()
	db.runs[id] = runCopy
	return nil
}

// Version operations
func (db *DefaultDatabase) AddVersion(version *Version) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	versionCopy := copyVersion(version)
	versionCopy.ID = int(atomic.AddInt64(&db.versionCounter, 1))

	// Update workflow -> version relationship
	if versionCopy.EntityID != 0 {
		db.workflowToVersion[versionCopy.EntityID] = append(
			db.workflowToVersion[versionCopy.EntityID],
			versionCopy.ID,
		)
	}

	db.versions[versionCopy.ID] = versionCopy
	return versionCopy.ID, nil
}

func (db *DefaultDatabase) GetVersion(id int) (*Version, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	version, exists := db.versions[id]
	if !exists {
		return nil, ErrVersionNotFound
	}

	return copyVersion(version), nil
}

func (db *DefaultDatabase) UpdateVersion(version *Version) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.versions[version.ID]; !exists {
		return ErrVersionNotFound
	}

	versionCopy := copyVersion(version)
	db.versions[version.ID] = versionCopy
	return nil
}

func (db *DefaultDatabase) GetVersionProperties(id int, getters ...VersionPropertyGetter) error {
	db.mu.RLock()
	version, exists := db.versions[id]
	if !exists {
		db.mu.RUnlock()
		return ErrVersionNotFound
	}
	versionCopy := copyVersion(version)
	db.mu.RUnlock()

	for _, getter := range getters {
		if option, err := getter(versionCopy); err != nil {
			return err
		} else if option != nil {
			opts := &VersionGetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
			// Handle options if needed
		}
	}
	return nil
}

func (db *DefaultDatabase) SetVersionProperties(id int, setters ...VersionPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	version, exists := db.versions[id]
	if !exists {
		return ErrVersionNotFound
	}

	versionCopy := copyVersion(version)
	for _, setter := range setters {
		if option, err := setter(versionCopy); err != nil {
			return err
		} else if option != nil {
			opts := &VersionSetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
			// Handle options if needed
		}
	}

	db.versions[id] = versionCopy
	return nil
}

// Queue operations
func (db *DefaultDatabase) AddQueue(queue *Queue) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Check if queue with same name exists
	for _, q := range db.queues {
		if q.Name == queue.Name {
			return 0, ErrQueueExists
		}
	}

	queueCopy := copyQueue(queue)
	queueCopy.ID = int(atomic.AddInt64(&db.queueCounter, 1))
	queueCopy.CreatedAt = time.Now()
	queueCopy.UpdatedAt = time.Now()

	db.queues[queueCopy.ID] = queueCopy
	db.queueToWorkflows[queueCopy.ID] = make([]int, 0)
	return queueCopy.ID, nil
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

func (db *DefaultDatabase) UpdateQueue(queue *Queue) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.queues[queue.ID]; !exists {
		return ErrQueueNotFound
	}

	queueCopy := copyQueue(queue)
	queueCopy.UpdatedAt = time.Now()
	db.queues[queue.ID] = queueCopy
	return nil
}

func (db *DefaultDatabase) GetQueueProperties(id int, getters ...QueuePropertyGetter) error {
	db.mu.RLock()
	queue, exists := db.queues[id]
	if !exists {
		db.mu.RUnlock()
		return ErrQueueNotFound
	}
	queueCopy := copyQueue(queue)
	db.mu.RUnlock()

	for _, getter := range getters {
		if option, err := getter(queueCopy); err != nil {
			return err
		} else if option != nil {
			opts := &QueueGetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
			if opts.IncludeWorkflows {
				if workflowIDs, exists := db.queueToWorkflows[id]; exists {
					for _, wfID := range workflowIDs {
						if wf, exists := db.workflowEntities[wfID]; exists {
							queueCopy.Entities = append(queueCopy.Entities, copyWorkflowEntity(wf))
						}
					}
				}
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) SetQueueProperties(id int, setters ...QueuePropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	queue, exists := db.queues[id]
	if !exists {
		return ErrQueueNotFound
	}

	queueCopy := copyQueue(queue)
	for _, setter := range setters {
		if option, err := setter(queueCopy); err != nil {
			return err
		} else if option != nil {
			opts := &QueueSetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
			// Handle workflow relationships
			if opts.WorkflowIDs != nil {
				// Clear old relationships
				if oldWorkflows, exists := db.queueToWorkflows[id]; exists {
					for _, wfID := range oldWorkflows {
						delete(db.workflowToQueue, wfID)
					}
				}
				// Set new relationships
				db.queueToWorkflows[id] = opts.WorkflowIDs
				for _, wfID := range opts.WorkflowIDs {
					db.workflowToQueue[wfID] = id
				}
			}
		}
	}

	queueCopy.UpdatedAt = time.Now()
	db.queues[id] = queueCopy
	return nil
}

// Hierarchy operations
func (db *DefaultDatabase) AddHierarchy(hierarchy *Hierarchy) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	hierarchyCopy := copyHierarchy(hierarchy)
	hierarchyCopy.ID = int(atomic.AddInt64(&db.hierarchyCounter, 1))

	// Validate relationships
	if _, exists := db.runs[hierarchyCopy.RunID]; !exists {
		return 0, fmt.Errorf("run %d not found", hierarchyCopy.RunID)
	}

	if hierarchyCopy.ParentType == EntityWorkflow {
		if _, exists := db.workflowEntities[hierarchyCopy.ParentEntityID]; !exists {
			return 0, fmt.Errorf("parent workflow %d not found", hierarchyCopy.ParentEntityID)
		}
	}

	// Update parent-child relationship maps based on entity types
	switch hierarchyCopy.ChildType {
	case EntityActivity:
		if _, exists := db.activityEntities[hierarchyCopy.ChildEntityID]; !exists {
			return 0, fmt.Errorf("child activity %d not found", hierarchyCopy.ChildEntityID)
		}
		db.entityToWorkflow[hierarchyCopy.ChildEntityID] = hierarchyCopy.ParentEntityID
		if _, exists := db.workflowToChildren[hierarchyCopy.ParentEntityID]; !exists {
			db.workflowToChildren[hierarchyCopy.ParentEntityID] = make(map[EntityType][]int)
		}
		db.workflowToChildren[hierarchyCopy.ParentEntityID][EntityActivity] = append(
			db.workflowToChildren[hierarchyCopy.ParentEntityID][EntityActivity],
			hierarchyCopy.ChildEntityID,
		)

	case EntitySaga:
		if _, exists := db.sagaEntities[hierarchyCopy.ChildEntityID]; !exists {
			return 0, fmt.Errorf("child saga %d not found", hierarchyCopy.ChildEntityID)
		}
		db.entityToWorkflow[hierarchyCopy.ChildEntityID] = hierarchyCopy.ParentEntityID
		if _, exists := db.workflowToChildren[hierarchyCopy.ParentEntityID]; !exists {
			db.workflowToChildren[hierarchyCopy.ParentEntityID] = make(map[EntityType][]int)
		}
		db.workflowToChildren[hierarchyCopy.ParentEntityID][EntitySaga] = append(
			db.workflowToChildren[hierarchyCopy.ParentEntityID][EntitySaga],
			hierarchyCopy.ChildEntityID,
		)

	case EntitySideEffect:
		if _, exists := db.sideEffectEntities[hierarchyCopy.ChildEntityID]; !exists {
			return 0, fmt.Errorf("child side effect %d not found", hierarchyCopy.ChildEntityID)
		}
		db.entityToWorkflow[hierarchyCopy.ChildEntityID] = hierarchyCopy.ParentEntityID
		if _, exists := db.workflowToChildren[hierarchyCopy.ParentEntityID]; !exists {
			db.workflowToChildren[hierarchyCopy.ParentEntityID] = make(map[EntityType][]int)
		}
		db.workflowToChildren[hierarchyCopy.ParentEntityID][EntitySideEffect] = append(
			db.workflowToChildren[hierarchyCopy.ParentEntityID][EntitySideEffect],
			hierarchyCopy.ChildEntityID,
		)
	}

	db.hierarchies[hierarchyCopy.ID] = hierarchyCopy
	return hierarchyCopy.ID, nil
}

func (db *DefaultDatabase) GetHierarchy(id int) (*Hierarchy, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	hierarchy, exists := db.hierarchies[id]
	if !exists {
		return nil, ErrHierarchyNotFound
	}

	return copyHierarchy(hierarchy), nil
}

func (db *DefaultDatabase) UpdateHierarchy(hierarchy *Hierarchy) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.hierarchies[hierarchy.ID]; !exists {
		return ErrHierarchyNotFound
	}

	hierarchyCopy := copyHierarchy(hierarchy)
	db.hierarchies[hierarchy.ID] = hierarchyCopy
	return nil
}

func (db *DefaultDatabase) GetHierarchyProperties(id int, getters ...HierarchyPropertyGetter) error {
	db.mu.RLock()
	hierarchy, exists := db.hierarchies[id]
	if !exists {
		db.mu.RUnlock()
		return ErrHierarchyNotFound
	}
	hierarchyCopy := copyHierarchy(hierarchy)
	db.mu.RUnlock()

	for _, getter := range getters {
		if option, err := getter(hierarchyCopy); err != nil {
			return err
		} else if option != nil {
			opts := &HierarchyGetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) SetHierarchyProperties(id int, setters ...HierarchyPropertySetter) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	hierarchy, exists := db.hierarchies[id]
	if !exists {
		return ErrHierarchyNotFound
	}

	hierarchyCopy := copyHierarchy(hierarchy)
	for _, setter := range setters {
		if option, err := setter(hierarchyCopy); err != nil {
			return err
		} else if option != nil {
			opts := &HierarchySetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
		}
	}

	db.hierarchies[id] = hierarchyCopy
	return nil
}

// Entity operations
func (db *DefaultDatabase) AddWorkflowEntity(entity *WorkflowEntity) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	entityCopy := copyWorkflowEntity(entity)
	entityCopy.ID = int(atomic.AddInt64(&db.workflowEntityCounter, 1))
	entityCopy.Type = EntityWorkflow
	entityCopy.CreatedAt = time.Now()
	entityCopy.UpdatedAt = time.Now()

	// Store the entity
	db.workflowEntities[entityCopy.ID] = entityCopy

	// Initialize relationships
	db.workflowToChildren[entityCopy.ID] = make(map[EntityType][]int)

	// Store entity data if present
	if entityCopy.WorkflowData != nil {
		db.workflowData[entityCopy.ID] = copyWorkflowData(entityCopy.WorkflowData)
	}

	// Add to run relationship if RunID is set
	if entityCopy.RunID != 0 {
		db.runToWorkflows[entityCopy.RunID] = append(db.runToWorkflows[entityCopy.RunID], entityCopy.ID)
	}

	return entityCopy.ID, nil
}

func (db *DefaultDatabase) GetWorkflowEntity(id int) (*WorkflowEntity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	entity, exists := db.workflowEntities[id]
	if !exists {
		return nil, ErrWorkflowEntityNotFound
	}

	entityCopy := copyWorkflowEntity(entity)
	if data, exists := db.workflowData[id]; exists {
		entityCopy.WorkflowData = copyWorkflowData(data)
	}

	return entityCopy, nil
}

func (db *DefaultDatabase) UpdateWorkflowEntity(entity *WorkflowEntity) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.workflowEntities[entity.ID]; !exists {
		return ErrWorkflowEntityNotFound
	}

	entityCopy := copyWorkflowEntity(entity)
	entityCopy.UpdatedAt = time.Now()

	// Update entity
	db.workflowEntities[entity.ID] = entityCopy

	// Update data if present
	if entityCopy.WorkflowData != nil {
		db.workflowData[entity.ID] = copyWorkflowData(entityCopy.WorkflowData)
	}

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
		if option, err := getter(entityCopy); err != nil {
			return err
		} else if option != nil {
			opts := &WorkflowEntityGetterOptions{}
			if err := option(opts); err != nil {
				return err
			}

			// Handle options
			if opts.IncludeData {
				if data, exists := db.workflowData[id]; exists {
					entityCopy.WorkflowData = copyWorkflowData(data)
				}
			}

			if opts.IncludeChildren {
				if children, exists := db.workflowToChildren[id]; exists {
					// Load children by type
					if activities, ok := children[EntityActivity]; ok {
						for _, actID := range activities {
							if act, exists := db.activityEntities[actID]; exists {
								// Handle activity relationship
								_ = act // TODO: Add to appropriate collection
							}
						}
					}
					// Similar for sagas and side effects
				}
			}

			if opts.IncludeVersion {
				if versions, exists := db.workflowToVersion[id]; exists {
					for _, versionID := range versions {
						if version, exists := db.versions[versionID]; exists {
							// Handle version relationship
							_ = version // TODO: Add to appropriate collection
						}
					}
				}
			}
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
		if option, err := setter(entityCopy); err != nil {
			return err
		} else if option != nil {
			opts := &WorkflowEntitySetterOptions{}
			if err := option(opts); err != nil {
				return err
			}

			// Handle queue relationship
			if opts.QueueID != nil {
				// Remove from old queue if exists
				if oldQueueID, exists := db.workflowToQueue[id]; exists {
					if workflows, exists := db.queueToWorkflows[oldQueueID]; exists {
						newWorkflows := make([]int, len(workflows))
						copy(newWorkflows, workflows)
						removeFromSlice(&newWorkflows, id)
						db.queueToWorkflows[oldQueueID] = newWorkflows
					}
				}
				// Add to new queue
				db.workflowToQueue[id] = *opts.QueueID
				db.queueToWorkflows[*opts.QueueID] = append(db.queueToWorkflows[*opts.QueueID], id)
			}

			// Handle version relationship
			if opts.Version != nil {
				versionCopy := copyVersion(opts.Version)
				versionCopy.ID = int(atomic.AddInt64(&db.versionCounter, 1))
				versionCopy.EntityID = id
				db.versions[versionCopy.ID] = versionCopy
				db.workflowToVersion[id] = append(db.workflowToVersion[id], versionCopy.ID)
			}

			// Handle run relationship
			if opts.RunID != nil {
				if oldRunID := entityCopy.RunID; oldRunID != 0 {
					if workflows, exists := db.runToWorkflows[oldRunID]; exists {
						newWorkflows := make([]int, len(workflows))
						copy(newWorkflows, workflows)
						removeFromSlice(&newWorkflows, id)
						db.runToWorkflows[oldRunID] = newWorkflows
					}
				}
				entityCopy.RunID = *opts.RunID
				db.runToWorkflows[*opts.RunID] = append(db.runToWorkflows[*opts.RunID], id)
			}
		}
	}

	entityCopy.UpdatedAt = time.Now()
	db.workflowEntities[id] = entityCopy
	return nil
}

func (db *DefaultDatabase) AddActivityEntity(entity *ActivityEntity) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	entityCopy := copyActivityEntity(entity)
	entityCopy.ID = int(atomic.AddInt64(&db.activityEntityCounter, 1))
	entityCopy.Type = EntityActivity
	entityCopy.CreatedAt = time.Now()
	entityCopy.UpdatedAt = time.Now()

	// Store the entity
	db.activityEntities[entityCopy.ID] = entityCopy

	// Store entity data if present
	if entityCopy.ActivityData != nil {
		db.activityData[entityCopy.ID] = copyActivityData(entityCopy.ActivityData)
	}

	return entityCopy.ID, nil
}

func (db *DefaultDatabase) GetActivityEntity(id int) (*ActivityEntity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	entity, exists := db.activityEntities[id]
	if !exists {
		return nil, ErrActivityEntityNotFound
	}

	entityCopy := copyActivityEntity(entity)
	if data, exists := db.activityData[id]; exists {
		entityCopy.ActivityData = copyActivityData(data)
	}

	return entityCopy, nil
}

func (db *DefaultDatabase) UpdateActivityEntity(entity *ActivityEntity) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.activityEntities[entity.ID]; !exists {
		return ErrActivityEntityNotFound
	}

	entityCopy := copyActivityEntity(entity)
	entityCopy.UpdatedAt = time.Now()

	// Update entity
	db.activityEntities[entity.ID] = entityCopy

	// Update data if present
	if entityCopy.ActivityData != nil {
		db.activityData[entity.ID] = copyActivityData(entityCopy.ActivityData)
	}

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
		if option, err := getter(entityCopy); err != nil {
			return err
		} else if option != nil {
			opts := &ActivityEntityGetterOptions{}
			if err := option(opts); err != nil {
				return err
			}

			// Handle options
			if opts.IncludeData {
				if data, exists := db.activityData[id]; exists {
					entityCopy.ActivityData = copyActivityData(data)
				}
			}

			if opts.IncludeWorkflow {
				if workflowID, exists := db.entityToWorkflow[id]; exists {
					// Load parent workflow info if needed
					_ = workflowID // TODO: Handle parent workflow relationship
				}
			}
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
		if option, err := setter(entityCopy); err != nil {
			return err
		} else if option != nil {
			opts := &ActivityEntitySetterOptions{}
			if err := option(opts); err != nil {
				return err
			}

			// Handle workflow relationship
			if opts.ParentWorkflowID != nil {
				// Remove from old workflow if exists
				if oldWorkflowID, exists := db.entityToWorkflow[id]; exists {
					if children, exists := db.workflowToChildren[oldWorkflowID]; exists {
						if activities, exists := children[EntityActivity]; exists {
							newActivities := make([]int, len(activities))
							copy(newActivities, activities)
							removeFromSlice(&newActivities, id)
							db.workflowToChildren[oldWorkflowID][EntityActivity] = newActivities
						}
					}
				}

				// Add to new workflow
				db.entityToWorkflow[id] = *opts.ParentWorkflowID
				if _, exists := db.workflowToChildren[*opts.ParentWorkflowID]; !exists {
					db.workflowToChildren[*opts.ParentWorkflowID] = make(map[EntityType][]int)
				}
				db.workflowToChildren[*opts.ParentWorkflowID][EntityActivity] = append(
					db.workflowToChildren[*opts.ParentWorkflowID][EntityActivity],
					id,
				)
			}
		}
	}

	entityCopy.UpdatedAt = time.Now()
	db.activityEntities[id] = entityCopy
	return nil
}

func (db *DefaultDatabase) AddSagaEntity(entity *SagaEntity) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	entityCopy := copySagaEntity(entity)
	entityCopy.ID = int(atomic.AddInt64(&db.sagaEntityCounter, 1))
	entityCopy.Type = EntitySaga
	entityCopy.CreatedAt = time.Now()
	entityCopy.UpdatedAt = time.Now()

	// Store the entity
	db.sagaEntities[entityCopy.ID] = entityCopy

	// Store entity data if present
	if entityCopy.SagaData != nil {
		db.sagaData[entityCopy.ID] = copySagaData(entityCopy.SagaData)
	}

	return entityCopy.ID, nil
}

func (db *DefaultDatabase) GetSagaEntity(id int) (*SagaEntity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	entity, exists := db.sagaEntities[id]
	if !exists {
		return nil, ErrSagaEntityNotFound
	}

	entityCopy := copySagaEntity(entity)
	if data, exists := db.sagaData[id]; exists {
		entityCopy.SagaData = copySagaData(data)
	}

	return entityCopy, nil
}

func (db *DefaultDatabase) UpdateSagaEntity(entity *SagaEntity) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.sagaEntities[entity.ID]; !exists {
		return ErrSagaEntityNotFound
	}

	entityCopy := copySagaEntity(entity)
	entityCopy.UpdatedAt = time.Now()

	// Update entity
	db.sagaEntities[entity.ID] = entityCopy

	// Update data if present
	if entityCopy.SagaData != nil {
		db.sagaData[entity.ID] = copySagaData(entityCopy.SagaData)
	}

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
		if option, err := getter(entityCopy); err != nil {
			return err
		} else if option != nil {
			opts := &SagaEntityGetterOptions{}
			if err := option(opts); err != nil {
				return err
			}

			// Handle options
			if opts.IncludeData {
				if data, exists := db.sagaData[id]; exists {
					entityCopy.SagaData = copySagaData(data)
				}
			}

			if opts.IncludeWorkflow {
				if workflowID, exists := db.entityToWorkflow[id]; exists {
					_ = workflowID // Handled in the getter function
				}
			}
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
		if option, err := setter(entityCopy); err != nil {
			return err
		} else if option != nil {
			opts := &SagaEntitySetterOptions{}
			if err := option(opts); err != nil {
				return err
			}

			// Handle workflow relationship
			if opts.ParentWorkflowID != nil {
				// Remove from old workflow if exists
				if oldWorkflowID, exists := db.entityToWorkflow[id]; exists {
					if children, exists := db.workflowToChildren[oldWorkflowID]; exists {
						if sagas, exists := children[EntitySaga]; exists {
							newSagas := make([]int, len(sagas))
							copy(newSagas, sagas)
							removeFromSlice(&newSagas, id)
							db.workflowToChildren[oldWorkflowID][EntitySaga] = newSagas
						}
					}
				}

				// Add to new workflow
				db.entityToWorkflow[id] = *opts.ParentWorkflowID
				if _, exists := db.workflowToChildren[*opts.ParentWorkflowID]; !exists {
					db.workflowToChildren[*opts.ParentWorkflowID] = make(map[EntityType][]int)
				}
				db.workflowToChildren[*opts.ParentWorkflowID][EntitySaga] = append(
					db.workflowToChildren[*opts.ParentWorkflowID][EntitySaga],
					id,
				)
			}
		}
	}

	entityCopy.UpdatedAt = time.Now()
	db.sagaEntities[id] = entityCopy
	return nil
}

func (db *DefaultDatabase) AddSideEffectEntity(entity *SideEffectEntity) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	entityCopy := copySideEffectEntity(entity)
	entityCopy.ID = int(atomic.AddInt64(&db.sideEffectEntityCounter, 1))
	entityCopy.Type = EntitySideEffect
	entityCopy.CreatedAt = time.Now()
	entityCopy.UpdatedAt = time.Now()

	// Store the entity
	db.sideEffectEntities[entityCopy.ID] = entityCopy

	// Store entity data if present
	if entityCopy.SideEffectData != nil {
		db.sideEffectData[entityCopy.ID] = copySideEffectData(entityCopy.SideEffectData)
	}

	return entityCopy.ID, nil
}

func (db *DefaultDatabase) GetSideEffectEntity(id int) (*SideEffectEntity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	entity, exists := db.sideEffectEntities[id]
	if !exists {
		return nil, ErrSideEffectEntityNotFound
	}

	entityCopy := copySideEffectEntity(entity)
	if data, exists := db.sideEffectData[id]; exists {
		entityCopy.SideEffectData = copySideEffectData(data)
	}

	return entityCopy, nil
}

func (db *DefaultDatabase) UpdateSideEffectEntity(entity *SideEffectEntity) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.sideEffectEntities[entity.ID]; !exists {
		return ErrSideEffectEntityNotFound
	}

	entityCopy := copySideEffectEntity(entity)
	entityCopy.UpdatedAt = time.Now()

	// Update entity
	db.sideEffectEntities[entity.ID] = entityCopy

	// Update data if present
	if entityCopy.SideEffectData != nil {
		db.sideEffectData[entity.ID] = copySideEffectData(entityCopy.SideEffectData)
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
		if option, err := setter(entityCopy); err != nil {
			return err
		} else if option != nil {
			opts := &SideEffectEntitySetterOptions{}
			if err := option(opts); err != nil {
				return err
			}

			// Handle workflow relationship
			if opts.ParentWorkflowID != nil {
				// Remove from old workflow if exists
				if oldWorkflowID, exists := db.entityToWorkflow[id]; exists {
					if children, exists := db.workflowToChildren[oldWorkflowID]; exists {
						if sideEffects, exists := children[EntitySideEffect]; exists {
							newSideEffects := make([]int, len(sideEffects))
							copy(newSideEffects, sideEffects)
							removeFromSlice(&newSideEffects, id)
							db.workflowToChildren[oldWorkflowID][EntitySideEffect] = newSideEffects
						}
					}
				}

				// Add to new workflow
				db.entityToWorkflow[id] = *opts.ParentWorkflowID
				if _, exists := db.workflowToChildren[*opts.ParentWorkflowID]; !exists {
					db.workflowToChildren[*opts.ParentWorkflowID] = make(map[EntityType][]int)
				}
				db.workflowToChildren[*opts.ParentWorkflowID][EntitySideEffect] = append(
					db.workflowToChildren[*opts.ParentWorkflowID][EntitySideEffect],
					id,
				)
			}
		}
	}

	entityCopy.UpdatedAt = time.Now()
	db.sideEffectEntities[id] = entityCopy
	return nil
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
		if option, err := getter(execCopy); err != nil {
			return err
		} else if option != nil {
			opts := &WorkflowExecutionDataGetterOptions{}
			if err := option(opts); err != nil {
				return err
			}

			if opts.IncludeOutputs {
				if data, exists := db.workflowExecutionData[id]; exists {
					execCopy.WorkflowExecutionData = copyWorkflowExecutionData(data)
				}
			}
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
		if option, err := setter(execCopy); err != nil {
			return err
		} else if option != nil {
			opts := &WorkflowExecutionDataSetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
			// Apply any option-specific changes
		}
	}

	execCopy.UpdatedAt = time.Now()
	db.workflowExecutions[id] = execCopy
	return nil
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
		if option, err := getter(execCopy); err != nil {
			return err
		} else if option != nil {
			opts := &ActivityExecutionDataGetterOptions{}
			if err := option(opts); err != nil {
				return err
			}

			if opts.IncludeOutputs {
				if data, exists := db.activityExecutionData[id]; exists {
					execCopy.ActivityExecutionData = copyActivityExecutionData(data)
				}
			}
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
		if option, err := setter(execCopy); err != nil {
			return err
		} else if option != nil {
			opts := &ActivityExecutionDataSetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
			// Apply any option-specific changes
		}
	}

	execCopy.UpdatedAt = time.Now()
	db.activityExecutions[id] = execCopy
	return nil
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
		if option, err := getter(execCopy); err != nil {
			return err
		} else if option != nil {
			opts := &SagaExecutionDataGetterOptions{}
			if err := option(opts); err != nil {
				return err
			}

			if opts.IncludeOutput {
				if data, exists := db.sagaExecutionData[id]; exists {
					execCopy.SagaExecutionData = copySagaExecutionData(data)
				}
			}
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
		if option, err := setter(execCopy); err != nil {
			return err
		} else if option != nil {
			opts := &SagaExecutionDataSetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
			// Apply any option-specific changes
		}
	}

	execCopy.UpdatedAt = time.Now()
	db.sagaExecutions[id] = execCopy
	return nil
}

func (db *DefaultDatabase) AddSideEffectExecution(exec *SideEffectExecution) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	execCopy := copySideEffectExecution(exec)
	execCopy.ID = int(atomic.AddInt64(&db.sideEffectExecutionCounter, 1))
	execCopy.CreatedAt = time.Now()
	execCopy.UpdatedAt = time.Now()
	execCopy.StartedAt = time.Now()

	// Store the execution
	db.sideEffectExecutions[execCopy.ID] = execCopy

	// Store execution data if present
	if execCopy.SideEffectExecutionData != nil {
		db.sideEffectExecutionData[execCopy.ID] = copySideEffectExecutionData(execCopy.SideEffectExecutionData)
	} else {
		db.sideEffectExecutionData[execCopy.ID] = &SideEffectExecutionData{}
	}

	return execCopy.ID, nil
}

func (db *DefaultDatabase) GetSideEffectExecution(id int) (*SideEffectExecution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	exec, exists := db.sideEffectExecutions[id]
	if !exists {
		return nil, ErrSideEffectExecutionNotFound
	}

	execCopy := copySideEffectExecution(exec)
	if data, exists := db.sideEffectExecutionData[id]; exists {
		execCopy.SideEffectExecutionData = copySideEffectExecutionData(data)
	}

	return execCopy, nil
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
		if option, err := getter(execCopy); err != nil {
			return err
		} else if option != nil {
			opts := &SideEffectExecutionDataGetterOptions{}
			if err := option(opts); err != nil {
				return err
			}

			if opts.IncludeOutputs {
				if data, exists := db.sideEffectExecutionData[id]; exists {
					execCopy.SideEffectExecutionData = copySideEffectExecutionData(data)
				}
			}
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
		if option, err := setter(execCopy); err != nil {
			return err
		} else if option != nil {
			opts := &SideEffectExecutionDataSetterOptions{}
			if err := option(opts); err != nil {
				return err
			}
			// Apply any option-specific changes
		}
	}

	execCopy.UpdatedAt = time.Now()
	db.sideEffectExecutions[id] = execCopy
	return nil
}
