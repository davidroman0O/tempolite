package tempolite

import (
	"errors"
	"time"
)

// Run property getters/setters
func GetRunID(id *RunID) RunPropertyGetter {
	return func(r *Run) (RunPropertyGetterOption, error) {
		if r == nil {
			return nil, errors.New("run is nil")
		}
		*id = r.ID
		return nil, nil
	}
}

func GetRunStatus(status *RunStatus) RunPropertyGetter {
	return func(r *Run) (RunPropertyGetterOption, error) {
		if r == nil {
			return nil, errors.New("run is nil")
		}
		*status = r.Status
		return nil, nil
	}
}

func GetRunCreatedAt(createdAt *time.Time) RunPropertyGetter {
	return func(r *Run) (RunPropertyGetterOption, error) {
		if r == nil {
			return nil, errors.New("run is nil")
		}
		*createdAt = r.CreatedAt
		return nil, nil
	}
}

func GetRunUpdatedAt(updatedAt *time.Time) RunPropertyGetter {
	return func(r *Run) (RunPropertyGetterOption, error) {
		if r == nil {
			return nil, errors.New("run is nil")
		}
		*updatedAt = r.UpdatedAt
		return nil, nil
	}
}

func GetRunEntities(entities *[]*WorkflowEntity) RunPropertyGetter {
	return func(r *Run) (RunPropertyGetterOption, error) {
		if r == nil {
			return nil, errors.New("run is nil")
		}
		*entities = make([]*WorkflowEntity, len(r.Entities))
		copy(*entities, r.Entities)
		return func(opts *RunGetterOptions) error {
			opts.IncludeWorkflows = true
			return nil
		}, nil
	}
}

func GetRunHierarchies(hierarchies *[]*Hierarchy) RunPropertyGetter {
	return func(r *Run) (RunPropertyGetterOption, error) {
		if r == nil {
			return nil, errors.New("run is nil")
		}
		*hierarchies = make([]*Hierarchy, len(r.Hierarchies))
		copy(*hierarchies, r.Hierarchies)
		return func(opts *RunGetterOptions) error {
			opts.IncludeHierarchies = true
			return nil
		}, nil
	}
}

func SetRunStatus(status RunStatus) RunPropertySetter {
	return func(r *Run) (RunPropertySetterOption, error) {
		if r == nil {
			return nil, errors.New("run is nil")
		}
		r.Status = status
		return nil, nil
	}
}

func SetRunCreatedAt(createdAt time.Time) RunPropertySetter {
	return func(r *Run) (RunPropertySetterOption, error) {
		if r == nil {
			return nil, errors.New("run is nil")
		}
		r.CreatedAt = createdAt
		return nil, nil
	}
}

func SetRunUpdatedAt(updatedAt time.Time) RunPropertySetter {
	return func(r *Run) (RunPropertySetterOption, error) {
		if r == nil {
			return nil, errors.New("run is nil")
		}
		r.UpdatedAt = updatedAt
		return nil, nil
	}
}

func SetRunWorkflowEntity(workflowID WorkflowEntityID) RunPropertySetter {
	return func(r *Run) (RunPropertySetterOption, error) {
		if r == nil {
			return nil, errors.New("run is nil")
		}
		return func(opts *RunSetterOptions) error {
			opts.WorkflowID = &workflowID
			return nil
		}, nil
	}
}

// Version property getters/setters
func GetVersionID(id *VersionID) VersionPropertyGetter {
	return func(v *Version) (VersionPropertyGetterOption, error) {
		if v == nil {
			return nil, errors.New("version is nil")
		}
		*id = v.ID
		return nil, nil
	}
}

func GetVersionEntityID(entityID *WorkflowEntityID) VersionPropertyGetter {
	return func(v *Version) (VersionPropertyGetterOption, error) {
		if v == nil {
			return nil, errors.New("version is nil")
		}
		*entityID = v.EntityID
		return nil, nil
	}
}

func GetVersionChangeID(changeID *string) VersionPropertyGetter {
	return func(v *Version) (VersionPropertyGetterOption, error) {
		if v == nil {
			return nil, errors.New("version is nil")
		}
		*changeID = v.ChangeID
		return nil, nil
	}
}

func GetVersionValue(version *int) VersionPropertyGetter {
	return func(v *Version) (VersionPropertyGetterOption, error) {
		if v == nil {
			return nil, errors.New("version is nil")
		}
		*version = v.Version
		return nil, nil
	}
}

func GetVersionData(data *map[string]interface{}) VersionPropertyGetter {
	return func(v *Version) (VersionPropertyGetterOption, error) {
		if v == nil {
			return nil, errors.New("version is nil")
		}
		*data = make(map[string]interface{})
		for k, val := range v.Data {
			(*data)[k] = val
		}
		return func(opts *VersionGetterOptions) error {
			opts.IncludeData = true
			return nil
		}, nil
	}
}

func GetVersionCreatedAt(createdAt *time.Time) VersionPropertyGetter {
	return func(v *Version) (VersionPropertyGetterOption, error) {
		if v == nil {
			return nil, errors.New("version is nil")
		}
		*createdAt = v.CreatedAt
		return nil, nil
	}
}

func GetVersionUpdatedAt(updatedAt *time.Time) VersionPropertyGetter {
	return func(v *Version) (VersionPropertyGetterOption, error) {
		if v == nil {
			return nil, errors.New("version is nil")
		}
		*updatedAt = v.UpdatedAt
		return nil, nil
	}
}

func SetVersionEntityID(entityID WorkflowEntityID) VersionPropertySetter {
	return func(v *Version) (VersionPropertySetterOption, error) {
		if v == nil {
			return nil, errors.New("version is nil")
		}
		v.EntityID = entityID
		return nil, nil
	}
}

func SetVersionChangeID(changeID string) VersionPropertySetter {
	return func(v *Version) (VersionPropertySetterOption, error) {
		if v == nil {
			return nil, errors.New("version is nil")
		}
		v.ChangeID = changeID
		return nil, nil
	}
}

func SetVersionValue(version int) VersionPropertySetter {
	return func(v *Version) (VersionPropertySetterOption, error) {
		if v == nil {
			return nil, errors.New("version is nil")
		}
		v.Version = version
		return nil, nil
	}
}

func SetVersionData(data map[string]interface{}) VersionPropertySetter {
	return func(v *Version) (VersionPropertySetterOption, error) {
		if v == nil {
			return nil, errors.New("version is nil")
		}
		v.Data = make(map[string]interface{})
		for k, val := range data {
			v.Data[k] = val
		}
		return nil, nil
	}
}

func SetVersionCreatedAt(createdAt time.Time) VersionPropertySetter {
	return func(v *Version) (VersionPropertySetterOption, error) {
		if v == nil {
			return nil, errors.New("version is nil")
		}
		v.CreatedAt = createdAt
		return nil, nil
	}
}

func SetVersionUpdatedAt(updatedAt time.Time) VersionPropertySetter {
	return func(v *Version) (VersionPropertySetterOption, error) {
		if v == nil {
			return nil, errors.New("version is nil")
		}
		v.UpdatedAt = updatedAt
		return nil, nil
	}
}

// WorkflowData property getters/setters
func GetWorkflowDataID(id *WorkflowDataID) WorkflowDataPropertyGetter {
	return func(d *WorkflowData) (WorkflowDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow data is nil")
		}
		*id = d.ID
		return nil, nil
	}
}

func GetWorkflowDataContinuedFrom(continuedFrom *WorkflowEntityID) WorkflowDataPropertyGetter {
	return func(d *WorkflowData) (WorkflowDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow data is nil")
		}
		if d.ContinuedFrom == nil {
			return nil, nil
		}
		*continuedFrom = *d.ContinuedFrom
		return nil, nil
	}
}

func GetWorkflowDataEntityID(entityID *WorkflowEntityID) WorkflowDataPropertyGetter {
	return func(d *WorkflowData) (WorkflowDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow data is nil")
		}
		*entityID = d.EntityID
		return nil, nil
	}
}

func GetWorkflowDataDuration(duration *string) WorkflowDataPropertyGetter {
	return func(d *WorkflowData) (WorkflowDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow data is nil")
		}
		*duration = d.Duration
		return nil, nil
	}
}

func GetWorkflowDataPaused(paused *bool) WorkflowDataPropertyGetter {
	return func(d *WorkflowData) (WorkflowDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow data is nil")
		}
		*paused = d.Paused
		return nil, nil
	}
}

func GetWorkflowDataResumable(resumable *bool) WorkflowDataPropertyGetter {
	return func(d *WorkflowData) (WorkflowDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow data is nil")
		}
		*resumable = d.Resumable
		return nil, nil
	}
}

func GetWorkflowDataIsRoot(isroot *bool) WorkflowDataPropertyGetter {
	return func(d *WorkflowData) (WorkflowDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow data is nil")
		}
		*isroot = d.IsRoot
		return nil, nil
	}
}

func GetWorkflowDataInputs(inputs *[][]byte) WorkflowDataPropertyGetter {
	return func(d *WorkflowData) (WorkflowDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow data is nil")
		}
		*inputs = make([][]byte, len(d.Inputs))
		for i, input := range d.Inputs {
			inputCopy := make([]byte, len(input))
			copy(inputCopy, input)
			(*inputs)[i] = inputCopy
		}
		return func(opts *WorkflowDataGetterOptions) error {
			opts.IncludeInputs = true
			return nil
		}, nil
	}
}

func SetWorkflowDataEntityID(entityID WorkflowEntityID) WorkflowDataPropertySetter {
	return func(d *WorkflowData) (WorkflowDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow data is nil")
		}
		d.EntityID = entityID
		return nil, nil
	}
}

func SetWorkflowDataContinuedFrom(continuedFrom *WorkflowEntityID) WorkflowDataPropertySetter {
	return func(d *WorkflowData) (WorkflowDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow data is nil")
		}
		d.ContinuedFrom = continuedFrom
		return nil, nil
	}
}

func SetWorkflowDataDuration(duration string) WorkflowDataPropertySetter {
	return func(d *WorkflowData) (WorkflowDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow data is nil")
		}
		d.Duration = duration
		return nil, nil
	}
}

func SetWorkflowDataPaused(paused bool) WorkflowDataPropertySetter {
	return func(d *WorkflowData) (WorkflowDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow data is nil")
		}
		d.Paused = paused
		return nil, nil
	}
}

func SetWorkflowDataResumable(resumable bool) WorkflowDataPropertySetter {
	return func(d *WorkflowData) (WorkflowDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow data is nil")
		}
		d.Resumable = resumable
		return nil, nil
	}
}

func SetWorkflowDataInputs(inputs [][]byte) WorkflowDataPropertySetter {
	return func(d *WorkflowData) (WorkflowDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow data is nil")
		}
		d.Inputs = make([][]byte, len(inputs))
		for i, input := range inputs {
			inputCopy := make([]byte, len(input))
			copy(inputCopy, input)
			d.Inputs[i] = inputCopy
		}
		return nil, nil
	}
}

// WorkflowEntityEdges getters/setters
func GetWorkflowEntityEdges(edges **WorkflowEntityEdges) WorkflowEntityPropertyGetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		*edges = e.Edges
		return func(opts *WorkflowEntityGetterOptions) error {
			opts.IncludeChildren = true
			return nil
		}, nil
	}
}

func GetWorkflowEntityVersions(versions *[]*Version) WorkflowEntityPropertyGetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		if e.Edges == nil || e.Edges.Versions == nil {
			*versions = make([]*Version, 0)
		} else {
			*versions = make([]*Version, len(e.Edges.Versions))
			copy(*versions, e.Edges.Versions)
		}
		return func(opts *WorkflowEntityGetterOptions) error {
			opts.IncludeVersion = true
			return nil
		}, nil
	}
}

func GetWorkflowEntityActivityChildren(activities *[]*ActivityEntity) WorkflowEntityPropertyGetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		if e.Edges == nil || e.Edges.ActivityChildren == nil {
			*activities = make([]*ActivityEntity, 0)
		} else {
			*activities = make([]*ActivityEntity, len(e.Edges.ActivityChildren))
			copy(*activities, e.Edges.ActivityChildren)
		}
		return func(opts *WorkflowEntityGetterOptions) error {
			opts.IncludeChildren = true
			return nil
		}, nil
	}
}

func GetWorkflowEntitySagaChildren(sagas *[]*SagaEntity) WorkflowEntityPropertyGetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		if e.Edges == nil || e.Edges.SagaChildren == nil {
			*sagas = make([]*SagaEntity, 0)
		} else {
			*sagas = make([]*SagaEntity, len(e.Edges.SagaChildren))
			copy(*sagas, e.Edges.SagaChildren)
		}
		return func(opts *WorkflowEntityGetterOptions) error {
			opts.IncludeChildren = true
			return nil
		}, nil
	}
}

func GetWorkflowEntitySideEffectChildren(sideEffects *[]*SideEffectEntity) WorkflowEntityPropertyGetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		if e.Edges == nil || e.Edges.SideEffectChildren == nil {
			*sideEffects = make([]*SideEffectEntity, 0)
		} else {
			*sideEffects = make([]*SideEffectEntity, len(e.Edges.SideEffectChildren))
			copy(*sideEffects, e.Edges.SideEffectChildren)
		}
		return func(opts *WorkflowEntityGetterOptions) error {
			opts.IncludeChildren = true
			return nil
		}, nil
	}
}

func SetWorkflowEntityVersion(version *Version) WorkflowEntityPropertySetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		if e.Edges == nil {
			e.Edges = &WorkflowEntityEdges{}
		}
		if e.Edges.Versions == nil {
			e.Edges.Versions = make([]*Version, 0)
		}
		e.Edges.Versions = append(e.Edges.Versions, version)
		return func(opts *WorkflowEntitySetterOptions) error {
			opts.Version = version
			return nil
		}, nil
	}
}

func AddWorkflowEntityActivityChild(activity *ActivityEntity) WorkflowEntityPropertySetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		if e.Edges == nil {
			e.Edges = &WorkflowEntityEdges{}
		}
		if e.Edges.ActivityChildren == nil {
			e.Edges.ActivityChildren = make([]*ActivityEntity, 0)
		}
		e.Edges.ActivityChildren = append(e.Edges.ActivityChildren, activity)
		return func(opts *WorkflowEntitySetterOptions) error {
			ptr := int(activity.ID)
			opts.ChildID = &ptr
			childType := EntityActivity
			opts.ChildType = &childType
			return nil
		}, nil
	}
}

func AddWorkflowEntitySagaChild(saga *SagaEntity) WorkflowEntityPropertySetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		if e.Edges == nil {
			e.Edges = &WorkflowEntityEdges{}
		}
		if e.Edges.SagaChildren == nil {
			e.Edges.SagaChildren = make([]*SagaEntity, 0)
		}
		e.Edges.SagaChildren = append(e.Edges.SagaChildren, saga)
		return func(opts *WorkflowEntitySetterOptions) error {
			ptr := int(saga.ID)
			opts.ChildID = &ptr
			childType := EntitySaga
			opts.ChildType = &childType
			return nil
		}, nil
	}
}

func AddWorkflowEntitySideEffectChild(sideEffect *SideEffectEntity) WorkflowEntityPropertySetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		if e.Edges == nil {
			e.Edges = &WorkflowEntityEdges{}
		}
		if e.Edges.SideEffectChildren == nil {
			e.Edges.SideEffectChildren = make([]*SideEffectEntity, 0)
		}
		e.Edges.SideEffectChildren = append(e.Edges.SideEffectChildren, sideEffect)
		return func(opts *WorkflowEntitySetterOptions) error {
			ptr := int(sideEffect.ID)
			opts.ChildID = &ptr
			childType := EntitySideEffect
			opts.ChildType = &childType
			return nil
		}, nil
	}
}

// WorkflowEntityData getters/setters
func GetWorkflowEntityData(data **WorkflowData) WorkflowEntityPropertyGetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		*data = e.WorkflowData
		return func(opts *WorkflowEntityGetterOptions) error {
			opts.IncludeData = true
			return nil
		}, nil
	}
}

func SetWorkflowEntityData(data *WorkflowData) WorkflowEntityPropertySetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		e.WorkflowData = data
		return nil, nil
	}
}

// WorkflowEntity base property getters/setters
func GetWorkflowEntityStatus(status *EntityStatus) WorkflowEntityPropertyGetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		*status = e.Status
		return nil, nil
	}
}

func GetWorkflowEntityHandlerName(name *string) WorkflowEntityPropertyGetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		*name = e.HandlerName
		return nil, nil
	}
}

func GetWorkflowEntityStepID(stepID *string) WorkflowEntityPropertyGetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		*stepID = e.StepID
		return nil, nil
	}
}

func GetWorkflowEntityRunID(runID *RunID) WorkflowEntityPropertyGetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		*runID = e.RunID
		return func(opts *WorkflowEntityGetterOptions) error {
			opts.IncludeRun = true
			return nil
		}, nil
	}
}
func GetWorkflowEntityRetryPolicy(policy *RetryPolicy) WorkflowEntityPropertyGetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		*policy = RetryPolicy{
			MaxAttempts: e.RetryPolicy.MaxAttempts,
			MaxInterval: time.Duration(e.RetryPolicy.MaxInterval),
		}
		return nil, nil
	}
}

func GetWorkflowEntityRetryState(state *RetryState) WorkflowEntityPropertyGetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		*state = e.RetryState
		return nil, nil
	}
}

func SetWorkflowEntityStatus(status EntityStatus) WorkflowEntityPropertySetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		e.Status = status
		return nil, nil
	}
}

func SetWorkflowEntityHandlerName(name string) WorkflowEntityPropertySetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		e.HandlerName = name
		return nil, nil
	}
}

func SetWorkflowEntityStepID(stepID string) WorkflowEntityPropertySetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		e.StepID = stepID
		return nil, nil
	}
}

func SetWorkflowEntityRunID(runID RunID) WorkflowEntityPropertySetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		e.RunID = runID
		return func(opts *WorkflowEntitySetterOptions) error {
			opts.RunID = &runID
			return nil
		}, nil
	}
}

func SetWorkflowEntityRetryPolicy(policy RetryPolicy) WorkflowEntityPropertySetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		e.RetryPolicy = *ToInternalRetryPolicy(&policy)
		return nil, nil
	}
}

func SetWorkflowEntityRetryState(state RetryState) WorkflowEntityPropertySetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		e.RetryState = state
		return nil, nil
	}
}

// WorkflowExecution property getters/setters
func GetWorkflowExecutionID(id *WorkflowExecutionID) WorkflowExecutionPropertyGetter {
	return func(e *WorkflowExecution) (WorkflowExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		*id = e.ID
		return nil, nil
	}
}

func GetWorkflowExecutionEntityID(entityID *WorkflowEntityID) WorkflowExecutionPropertyGetter {
	return func(e *WorkflowExecution) (WorkflowExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		*entityID = e.WorkflowEntityID
		return nil, nil
	}
}

func GetWorkflowExecutionStatus(status *ExecutionStatus) WorkflowExecutionPropertyGetter {
	return func(e *WorkflowExecution) (WorkflowExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		*status = e.Status
		return nil, nil
	}
}

func GetWorkflowExecutionStartedAt(startedAt *time.Time) WorkflowExecutionPropertyGetter {
	return func(e *WorkflowExecution) (WorkflowExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		*startedAt = e.StartedAt
		return nil, nil
	}
}

func GetWorkflowExecutionCompletedAt(completedAt *time.Time) WorkflowExecutionPropertyGetter {
	return func(e *WorkflowExecution) (WorkflowExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		if e.CompletedAt != nil {
			*completedAt = *e.CompletedAt
		}
		return nil, nil
	}
}

func GetWorkflowExecutionData(data **WorkflowExecutionData) WorkflowExecutionPropertyGetter {
	return func(e *WorkflowExecution) (WorkflowExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		*data = e.WorkflowExecutionData
		return func(opts *WorkflowExecutionGetterOptions) error {
			opts.IncludeData = true
			return nil
		}, nil
	}
}

func GetWorkflowExecutionError(errStr *string) WorkflowExecutionPropertyGetter {
	return func(e *WorkflowExecution) (WorkflowExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		*errStr = e.Error
		return nil, nil
	}
}

func SetWorkflowExecutionEntityID(entityID WorkflowEntityID) WorkflowExecutionPropertySetter {
	return func(e *WorkflowExecution) (WorkflowExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		e.WorkflowEntityID = entityID
		return nil, nil
	}
}

func SetWorkflowExecutionStatus(status ExecutionStatus) WorkflowExecutionPropertySetter {
	return func(e *WorkflowExecution) (WorkflowExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		e.Status = status
		return nil, nil
	}
}

func SetWorkflowExecutionStartedAt(startedAt time.Time) WorkflowExecutionPropertySetter {
	return func(e *WorkflowExecution) (WorkflowExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		e.StartedAt = startedAt
		return nil, nil
	}
}

func SetWorkflowExecutionCompletedAt(completedAt time.Time) WorkflowExecutionPropertySetter {
	return func(e *WorkflowExecution) (WorkflowExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		e.CompletedAt = &completedAt
		return nil, nil
	}
}

func SetWorkflowExecutionData(data *WorkflowExecutionData) WorkflowExecutionPropertySetter {
	return func(e *WorkflowExecution) (WorkflowExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		e.WorkflowExecutionData = data
		return nil, nil
	}
}

func SetWorkflowExecutionError(err string) WorkflowExecutionPropertySetter {
	return func(e *WorkflowExecution) (WorkflowExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		e.Error = err
		return nil, nil
	}
}

// ActivityData property getters/setters
func GetActivityDataID(id *ActivityDataID) ActivityDataPropertyGetter {
	return func(d *ActivityData) (ActivityDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("activity data is nil")
		}
		*id = d.ID
		return nil, nil
	}
}

func GetActivityDataEntityID(entityID *ActivityEntityID) ActivityDataPropertyGetter {
	return func(d *ActivityData) (ActivityDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("activity data is nil")
		}
		*entityID = d.EntityID
		return nil, nil
	}
}

func GetActivityDataInputs(inputs *[][]byte) ActivityDataPropertyGetter {
	return func(d *ActivityData) (ActivityDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("activity data is nil")
		}
		*inputs = make([][]byte, len(d.Inputs))
		for i, input := range d.Inputs {
			inputCopy := make([]byte, len(input))
			copy(inputCopy, input)
			(*inputs)[i] = inputCopy
		}
		return func(opts *ActivityDataGetterOptions) error {
			opts.IncludeInputs = true
			return nil
		}, nil
	}
}

func SetActivityDataEntityID(entityID ActivityEntityID) ActivityDataPropertySetter {
	return func(d *ActivityData) (ActivityDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("activity data is nil")
		}
		d.EntityID = entityID
		return nil, nil
	}
}

func SetActivityDataInputs(inputs [][]byte) ActivityDataPropertySetter {
	return func(d *ActivityData) (ActivityDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("activity data is nil")
		}
		d.Inputs = make([][]byte, len(inputs))
		for i, input := range inputs {
			inputCopy := make([]byte, len(input))
			copy(inputCopy, input)
			d.Inputs[i] = inputCopy
		}
		return nil, nil
	}
}

// ActivityEntity base property getters/setters
func GetActivityEntityStatus(status *EntityStatus) ActivityEntityPropertyGetter {
	return func(e *ActivityEntity) (ActivityEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		*status = e.Status
		return nil, nil
	}
}

func GetActivityEntityHandlerName(name *string) ActivityEntityPropertyGetter {
	return func(e *ActivityEntity) (ActivityEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		*name = e.HandlerName
		return nil, nil
	}
}

func GetActivityEntityStepID(stepID *string) ActivityEntityPropertyGetter {
	return func(e *ActivityEntity) (ActivityEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		*stepID = e.StepID
		return nil, nil
	}
}

func GetActivityEntityRunID(runID *RunID) ActivityEntityPropertyGetter {
	return func(e *ActivityEntity) (ActivityEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		*runID = e.RunID
		return func(opts *ActivityEntityGetterOptions) error {
			// opts.IncludeRun = true
			return nil
		}, nil
	}
}

func GetActivityEntityRetryPolicy(policy *RetryPolicy) ActivityEntityPropertyGetter {
	return func(e *ActivityEntity) (ActivityEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		*policy = RetryPolicy{
			MaxAttempts: e.RetryPolicy.MaxAttempts,
			MaxInterval: time.Duration(e.RetryPolicy.MaxInterval),
		}
		return nil, nil
	}
}

func GetActivityEntityRetryState(state *RetryState) ActivityEntityPropertyGetter {
	return func(e *ActivityEntity) (ActivityEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		*state = e.RetryState
		return nil, nil
	}
}

func SetActivityEntityStatus(status EntityStatus) ActivityEntityPropertySetter {
	return func(e *ActivityEntity) (ActivityEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		e.Status = status
		return nil, nil
	}
}

func SetActivityEntityHandlerName(name string) ActivityEntityPropertySetter {
	return func(e *ActivityEntity) (ActivityEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		e.HandlerName = name
		return nil, nil
	}
}

func SetActivityEntityStepID(stepID string) ActivityEntityPropertySetter {
	return func(e *ActivityEntity) (ActivityEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		e.StepID = stepID
		return nil, nil
	}
}

func SetActivityEntityRunID(runID RunID) ActivityEntityPropertySetter {
	return func(e *ActivityEntity) (ActivityEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		e.RunID = runID
		return func(opts *ActivityEntitySetterOptions) error {
			opts.ParentRunID = &runID
			return nil
		}, nil
	}
}

func SetActivityEntityRetryPolicy(policy RetryPolicy) ActivityEntityPropertySetter {
	return func(e *ActivityEntity) (ActivityEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		e.RetryPolicy = *ToInternalRetryPolicy(&policy)
		return nil, nil
	}
}

func SetActivityEntityRetryState(state RetryState) ActivityEntityPropertySetter {
	return func(e *ActivityEntity) (ActivityEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		e.RetryState = state
		return nil, nil
	}
}

// ActivityEntity data getters/setters
func GetActivityEntityData(data **ActivityData) ActivityEntityPropertyGetter {
	return func(e *ActivityEntity) (ActivityEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		*data = e.ActivityData
		return func(opts *ActivityEntityGetterOptions) error {
			opts.IncludeData = true
			return nil
		}, nil
	}
}

func SetActivityEntityData(data *ActivityData) ActivityEntityPropertySetter {
	return func(e *ActivityEntity) (ActivityEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		e.ActivityData = data
		return nil, nil
	}
}

// ActivityExecution property getters/setters
func GetActivityExecutionID(id *ActivityExecutionID) ActivityExecutionPropertyGetter {
	return func(e *ActivityExecution) (ActivityExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("activity execution is nil")
		}
		*id = e.ID
		return nil, nil
	}
}

func GetActivityExecutionEntityID(entityID *ActivityEntityID) ActivityExecutionPropertyGetter {
	return func(e *ActivityExecution) (ActivityExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("activity execution is nil")
		}
		*entityID = e.ActivityEntityID
		return nil, nil
	}
}

func GetActivityExecutionStatus(status *ExecutionStatus) ActivityExecutionPropertyGetter {
	return func(e *ActivityExecution) (ActivityExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("activity execution is nil")
		}
		*status = e.Status
		return nil, nil
	}
}

func GetActivityExecutionStartedAt(startedAt *time.Time) ActivityExecutionPropertyGetter {
	return func(e *ActivityExecution) (ActivityExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("activity execution is nil")
		}
		*startedAt = e.StartedAt
		return nil, nil
	}
}

func GetActivityExecutionCompletedAt(completedAt *time.Time) ActivityExecutionPropertyGetter {
	return func(e *ActivityExecution) (ActivityExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("activity execution is nil")
		}
		if e.CompletedAt != nil {
			*completedAt = *e.CompletedAt
		}
		return nil, nil
	}
}

func GetActivityExecutionData(data **ActivityExecutionData) ActivityExecutionPropertyGetter {
	return func(e *ActivityExecution) (ActivityExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("activity execution is nil")
		}
		*data = e.ActivityExecutionData
		return func(opts *ActivityExecutionGetterOptions) error {
			opts.IncludeData = true
			return nil
		}, nil
	}
}

func GetActivityExecutionError(errStr *string) ActivityExecutionPropertyGetter {
	return func(e *ActivityExecution) (ActivityExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("activity execution is nil")
		}
		*errStr = e.Error
		return nil, nil
	}
}

func SetActivityExecutionEntityID(entityID ActivityEntityID) ActivityExecutionPropertySetter {
	return func(e *ActivityExecution) (ActivityExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("activity execution is nil")
		}
		e.ActivityEntityID = entityID
		return nil, nil
	}
}

func SetActivityExecutionStatus(status ExecutionStatus) ActivityExecutionPropertySetter {
	return func(e *ActivityExecution) (ActivityExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("activity execution is nil")
		}
		e.Status = status
		return nil, nil
	}
}

func SetActivityExecutionStartedAt(startedAt time.Time) ActivityExecutionPropertySetter {
	return func(e *ActivityExecution) (ActivityExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("activity execution is nil")
		}
		e.StartedAt = startedAt
		return nil, nil
	}
}

func SetActivityExecutionCompletedAt(completedAt time.Time) ActivityExecutionPropertySetter {
	return func(e *ActivityExecution) (ActivityExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("activity execution is nil")
		}
		e.CompletedAt = &completedAt
		return nil, nil
	}
}

func SetActivityExecutionData(data *ActivityExecutionData) ActivityExecutionPropertySetter {
	return func(e *ActivityExecution) (ActivityExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("activity execution is nil")
		}
		e.ActivityExecutionData = data
		return nil, nil
	}
}

func SetActivityExecutionError(err string) ActivityExecutionPropertySetter {
	return func(e *ActivityExecution) (ActivityExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("activity execution is nil")
		}
		e.Error = err
		return nil, nil
	}
}

func GetActivityExecutionDataID(id *ActivityExecutionDataID) ActivityExecutionDataPropertyGetter {
	return func(d *ActivityExecutionData) (ActivityExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("activity execution data is nil")
		}
		*id = d.ID
		return nil, nil
	}
}

func GetActivityExecutionDataOutputs(outputs *[][]byte) ActivityExecutionDataPropertyGetter {
	return func(d *ActivityExecutionData) (ActivityExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("activity execution data is nil")
		}
		*outputs = make([][]byte, len(d.Outputs))
		for i, output := range d.Outputs {
			outputCopy := make([]byte, len(output))
			copy(outputCopy, output)
			(*outputs)[i] = outputCopy
		}
		return func(opts *ActivityExecutionDataGetterOptions) error {
			opts.IncludeOutputs = true
			return nil
		}, nil
	}
}

func SetActivityExecutionDataOutputs(outputs [][]byte) ActivityExecutionDataPropertySetter {
	return func(d *ActivityExecutionData) (ActivityExecutionDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("activity execution data is nil")
		}
		d.Outputs = make([][]byte, len(outputs))
		for i, output := range outputs {
			outputCopy := make([]byte, len(output))
			copy(outputCopy, output)
			d.Outputs[i] = outputCopy
		}
		return nil, nil
	}
}

// SagaData property getters/setters
func GetSagaDataID(id *SagaDataID) SagaDataPropertyGetter {
	return func(d *SagaData) (SagaDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("saga data is nil")
		}
		*id = d.ID
		return nil, nil
	}
}

func GetSagaDataEntityID(entityID *SagaEntityID) SagaDataPropertyGetter {
	return func(d *SagaData) (SagaDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("saga data is nil")
		}
		*entityID = d.EntityID
		return nil, nil
	}
}

func SetSagaDataEntityID(entityID SagaEntityID) SagaDataPropertySetter {
	return func(d *SagaData) (SagaDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("saga data is nil")
		}
		d.EntityID = entityID
		return nil, nil
	}
}

// SagaEntity base property getters/setters
func GetSagaEntityStatus(status *EntityStatus) SagaEntityPropertyGetter {
	return func(e *SagaEntity) (SagaEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("saga entity is nil")
		}
		*status = e.Status
		return nil, nil
	}
}

func GetSagaEntityHandlerName(name *string) SagaEntityPropertyGetter {
	return func(e *SagaEntity) (SagaEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("saga entity is nil")
		}
		*name = e.HandlerName
		return nil, nil
	}
}

func GetSagaEntityStepID(stepID *string) SagaEntityPropertyGetter {
	return func(e *SagaEntity) (SagaEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("saga entity is nil")
		}
		*stepID = e.StepID
		return nil, nil
	}
}

func GetSagaEntityRunID(runID *RunID) SagaEntityPropertyGetter {
	return func(e *SagaEntity) (SagaEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("saga entity is nil")
		}
		*runID = e.RunID
		return func(opts *SagaEntityGetterOptions) error {
			return nil
		}, nil
	}
}

func GetSagaEntityRetryPolicy(policy *RetryPolicy) SagaEntityPropertyGetter {
	return func(e *SagaEntity) (SagaEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("saga entity is nil")
		}
		*policy = RetryPolicy{
			MaxAttempts: e.RetryPolicy.MaxAttempts,
			MaxInterval: time.Duration(e.RetryPolicy.MaxInterval),
		}
		return nil, nil
	}
}

func GetSagaEntityRetryState(state *RetryState) SagaEntityPropertyGetter {
	return func(e *SagaEntity) (SagaEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("saga entity is nil")
		}
		*state = e.RetryState
		return nil, nil
	}
}

func SetSagaEntityStatus(status EntityStatus) SagaEntityPropertySetter {
	return func(e *SagaEntity) (SagaEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("saga entity is nil")
		}
		e.Status = status
		return nil, nil
	}
}

func SetSagaEntityHandlerName(name string) SagaEntityPropertySetter {
	return func(e *SagaEntity) (SagaEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("saga entity is nil")
		}
		e.HandlerName = name
		return nil, nil
	}
}

func SetSagaEntityStepID(stepID string) SagaEntityPropertySetter {
	return func(e *SagaEntity) (SagaEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("saga entity is nil")
		}
		e.StepID = stepID
		return nil, nil
	}
}

func SetSagaEntityRunID(runID RunID) SagaEntityPropertySetter {
	return func(e *SagaEntity) (SagaEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("saga entity is nil")
		}
		e.RunID = runID
		return nil, nil
	}
}

func SetSagaEntityRetryPolicy(policy RetryPolicy) SagaEntityPropertySetter {
	return func(e *SagaEntity) (SagaEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("saga entity is nil")
		}
		e.RetryPolicy = *ToInternalRetryPolicy(&policy)
		return nil, nil
	}
}

func SetSagaEntityRetryState(state RetryState) SagaEntityPropertySetter {
	return func(e *SagaEntity) (SagaEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("saga entity is nil")
		}
		e.RetryState = state
		return nil, nil
	}
}

// SagaEntity data getters/setters
func GetSagaEntityData(data **SagaData) SagaEntityPropertyGetter {
	return func(e *SagaEntity) (SagaEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("saga entity is nil")
		}
		*data = e.SagaData
		return func(opts *SagaEntityGetterOptions) error {
			opts.IncludeData = true
			return nil
		}, nil
	}
}

func SetSagaEntityData(data *SagaData) SagaEntityPropertySetter {
	return func(e *SagaEntity) (SagaEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("saga entity is nil")
		}
		e.SagaData = data
		return nil, nil
	}
}

func GetSagaExecutionID(id *SagaExecutionID) SagaExecutionPropertyGetter {
	return func(e *SagaExecution) (SagaExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("saga execution is nil")
		}
		*id = e.ID
		return nil, nil
	}
}

func GetSagaExecutionEntityID(entityID *SagaEntityID) SagaExecutionPropertyGetter {
	return func(e *SagaExecution) (SagaExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("saga execution is nil")
		}
		*entityID = e.SagaEntityID
		return nil, nil
	}
}

func GetSagaExecutionStatus(status *ExecutionStatus) SagaExecutionPropertyGetter {
	return func(e *SagaExecution) (SagaExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("saga execution is nil")
		}
		*status = e.Status
		return nil, nil
	}
}

func GetSagaExecutionStartedAt(startedAt *time.Time) SagaExecutionPropertyGetter {
	return func(e *SagaExecution) (SagaExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("saga execution is nil")
		}
		*startedAt = e.StartedAt
		return nil, nil
	}
}

func GetSagaExecutionCompletedAt(completedAt *time.Time) SagaExecutionPropertyGetter {
	return func(e *SagaExecution) (SagaExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("saga execution is nil")
		}
		if e.CompletedAt != nil {
			*completedAt = *e.CompletedAt
		}
		return nil, nil
	}
}

func GetSagaExecutionData(data **SagaExecutionData) SagaExecutionPropertyGetter {
	return func(e *SagaExecution) (SagaExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("saga execution is nil")
		}
		*data = e.SagaExecutionData
		return func(opts *SagaExecutionGetterOptions) error {
			opts.IncludeData = true
			return nil
		}, nil
	}
}

func GetSagaExecutionError(errStr *string) SagaExecutionPropertyGetter {
	return func(e *SagaExecution) (SagaExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("saga execution is nil")
		}
		*errStr = e.Error
		return nil, nil
	}
}

func SetSagaExecutionEntityID(entityID SagaEntityID) SagaExecutionPropertySetter {
	return func(e *SagaExecution) (SagaExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("saga execution is nil")
		}
		e.SagaEntityID = entityID
		return nil, nil
	}
}

func SetSagaExecutionStatus(status ExecutionStatus) SagaExecutionPropertySetter {
	return func(e *SagaExecution) (SagaExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("saga execution is nil")
		}
		e.Status = status
		return nil, nil
	}
}

func SetSagaExecutionStartedAt(startedAt time.Time) SagaExecutionPropertySetter {
	return func(e *SagaExecution) (SagaExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("saga execution is nil")
		}
		e.StartedAt = startedAt
		return nil, nil
	}
}

func SetSagaExecutionCompletedAt(completedAt time.Time) SagaExecutionPropertySetter {
	return func(e *SagaExecution) (SagaExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("saga execution is nil")
		}
		e.CompletedAt = &completedAt
		return nil, nil
	}
}

func SetSagaExecutionData(data *SagaExecutionData) SagaExecutionPropertySetter {
	return func(e *SagaExecution) (SagaExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("saga execution is nil")
		}
		e.SagaExecutionData = data
		return nil, nil
	}
}

func SetSagaExecutionError(err string) SagaExecutionPropertySetter {
	return func(e *SagaExecution) (SagaExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("saga execution is nil")
		}
		e.Error = err
		return nil, nil
	}
}

// SagaExecutionData property getters/setters
func GetSagaExecutionDataID(id *SagaExecutionDataID) SagaExecutionDataPropertyGetter {
	return func(d *SagaExecutionData) (SagaExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("saga execution data is nil")
		}
		*id = d.ID
		return nil, nil
	}
}

func GetSagaExecutionDataExecutionID(executionID *SagaExecutionID) SagaExecutionDataPropertyGetter {
	return func(d *SagaExecutionData) (SagaExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("saga execution data is nil")
		}
		*executionID = d.ExecutionID
		return nil, nil
	}
}

func GetSagaExecutionDataLastHeartbeat(lastHeartbeat *time.Time) SagaExecutionDataPropertyGetter {
	return func(d *SagaExecutionData) (SagaExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("saga execution data is nil")
		}
		if d.LastHeartbeat != nil {
			*lastHeartbeat = *d.LastHeartbeat
		}
		return nil, nil
	}
}

func SetSagaExecutionDataExecutionID(executionID SagaExecutionID) SagaExecutionDataPropertySetter {
	return func(d *SagaExecutionData) (SagaExecutionDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("saga execution data is nil")
		}
		d.ExecutionID = executionID
		return nil, nil
	}
}

func SetSagaExecutionDataLastHeartbeat(lastHeartbeat time.Time) SagaExecutionDataPropertySetter {
	return func(d *SagaExecutionData) (SagaExecutionDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("saga execution data is nil")
		}
		d.LastHeartbeat = &lastHeartbeat
		return nil, nil
	}
}

// SideEffectData property getters/setters
func GetSideEffectDataID(id *SideEffectDataID) SideEffectDataPropertyGetter {
	return func(d *SideEffectData) (SideEffectDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("side effect data is nil")
		}
		*id = d.ID
		return nil, nil
	}
}

func GetSideEffectDataEntityID(entityID *SideEffectEntityID) SideEffectDataPropertyGetter {
	return func(d *SideEffectData) (SideEffectDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("side effect data is nil")
		}
		*entityID = d.EntityID
		return nil, nil
	}
}

func SetSideEffectDataEntityID(entityID SideEffectEntityID) SideEffectDataPropertySetter {
	return func(d *SideEffectData) (SideEffectDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("side effect data is nil")
		}
		d.EntityID = entityID
		return nil, nil
	}
}

// SideEffectEntity property getters/setters
func GetSideEffectEntityStatus(status *EntityStatus) SideEffectEntityPropertyGetter {
	return func(e *SideEffectEntity) (SideEffectEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect entity is nil")
		}
		*status = e.Status
		return nil, nil
	}
}

func GetSideEffectEntityHandlerName(name *string) SideEffectEntityPropertyGetter {
	return func(e *SideEffectEntity) (SideEffectEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect entity is nil")
		}
		*name = e.HandlerName
		return nil, nil
	}
}

func GetSideEffectEntityStepID(stepID *string) SideEffectEntityPropertyGetter {
	return func(e *SideEffectEntity) (SideEffectEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect entity is nil")
		}
		*stepID = e.StepID
		return nil, nil
	}
}

func GetSideEffectEntityRunID(runID *RunID) SideEffectEntityPropertyGetter {
	return func(e *SideEffectEntity) (SideEffectEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect entity is nil")
		}
		*runID = e.RunID
		return nil, nil
	}
}

func GetSideEffectEntityRetryPolicy(policy *RetryPolicy) SideEffectEntityPropertyGetter {
	return func(e *SideEffectEntity) (SideEffectEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect entity is nil")
		}
		*policy = RetryPolicy{
			MaxAttempts: e.RetryPolicy.MaxAttempts,
			MaxInterval: time.Duration(e.RetryPolicy.MaxInterval),
		}
		return nil, nil
	}
}

func GetSideEffectEntityRetryState(state *RetryState) SideEffectEntityPropertyGetter {
	return func(e *SideEffectEntity) (SideEffectEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect entity is nil")
		}
		*state = e.RetryState
		return nil, nil
	}
}

func SetSideEffectEntityStatus(status EntityStatus) SideEffectEntityPropertySetter {
	return func(e *SideEffectEntity) (SideEffectEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect entity is nil")
		}
		e.Status = status
		return nil, nil
	}
}

func SetSideEffectEntityHandlerName(name string) SideEffectEntityPropertySetter {
	return func(e *SideEffectEntity) (SideEffectEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect entity is nil")
		}
		e.HandlerName = name
		return nil, nil
	}
}

func SetSideEffectEntityStepID(stepID string) SideEffectEntityPropertySetter {
	return func(e *SideEffectEntity) (SideEffectEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect entity is nil")
		}
		e.StepID = stepID
		return nil, nil
	}
}

func SetSideEffectEntityRunID(runID RunID) SideEffectEntityPropertySetter {
	return func(e *SideEffectEntity) (SideEffectEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect entity is nil")
		}
		e.RunID = runID
		return nil, nil
	}
}

func SetSideEffectEntityRetryPolicy(policy RetryPolicy) SideEffectEntityPropertySetter {
	return func(e *SideEffectEntity) (SideEffectEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect entity is nil")
		}
		e.RetryPolicy = *ToInternalRetryPolicy(&policy)
		return nil, nil
	}
}

func SetSideEffectEntityRetryState(state RetryState) SideEffectEntityPropertySetter {
	return func(e *SideEffectEntity) (SideEffectEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect entity is nil")
		}
		e.RetryState = state
		return nil, nil
	}
}

// SideEffectEntity data getters/setters
func GetSideEffectEntityData(data **SideEffectData) SideEffectEntityPropertyGetter {
	return func(e *SideEffectEntity) (SideEffectEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect entity is nil")
		}
		*data = e.SideEffectData
		return func(opts *SideEffectEntityGetterOptions) error {
			opts.IncludeData = true
			return nil
		}, nil
	}
}

func SetSideEffectEntityData(data *SideEffectData) SideEffectEntityPropertySetter {
	return func(e *SideEffectEntity) (SideEffectEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect entity is nil")
		}
		e.SideEffectData = data
		return nil, nil
	}
}

// SideEffectExecution property getters/setters
func GetSideEffectExecutionID(id *SideEffectExecutionID) SideEffectExecutionPropertyGetter {
	return func(e *SideEffectExecution) (SideEffectExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect execution is nil")
		}
		*id = e.ID
		return nil, nil
	}
}

func GetSideEffectExecutionEntityID(entityID *SideEffectEntityID) SideEffectExecutionPropertyGetter {
	return func(e *SideEffectExecution) (SideEffectExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect execution is nil")
		}
		*entityID = e.SideEffectEntityID
		return nil, nil
	}
}

func GetSideEffectExecutionStatus(status *ExecutionStatus) SideEffectExecutionPropertyGetter {
	return func(e *SideEffectExecution) (SideEffectExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect execution is nil")
		}
		*status = e.Status
		return nil, nil
	}
}

func GetSideEffectExecutionStartedAt(startedAt *time.Time) SideEffectExecutionPropertyGetter {
	return func(e *SideEffectExecution) (SideEffectExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect execution is nil")
		}
		*startedAt = e.StartedAt
		return nil, nil
	}
}

func GetSideEffectExecutionCompletedAt(completedAt *time.Time) SideEffectExecutionPropertyGetter {
	return func(e *SideEffectExecution) (SideEffectExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect execution is nil")
		}
		if e.CompletedAt != nil {
			*completedAt = *e.CompletedAt
		}
		return nil, nil
	}
}

func GetSideEffectExecutionData(data **SideEffectExecutionData) SideEffectExecutionPropertyGetter {
	return func(e *SideEffectExecution) (SideEffectExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect execution is nil")
		}
		*data = e.SideEffectExecutionData
		return func(opts *SideEffectExecutionGetterOptions) error {
			opts.IncludeData = true
			return nil
		}, nil
	}
}

func GetSideEffectExecutionError(errStr *string) SideEffectExecutionPropertyGetter {
	return func(e *SideEffectExecution) (SideEffectExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect execution is nil")
		}
		*errStr = e.Error
		return nil, nil
	}
}

func SetSideEffectExecutionEntityID(entityID SideEffectEntityID) SideEffectExecutionPropertySetter {
	return func(e *SideEffectExecution) (SideEffectExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect execution is nil")
		}
		e.SideEffectEntityID = entityID
		return nil, nil
	}
}

func SetSignalExecutionEntityID(entityID SignalEntityID) SignalExecutionPropertySetter {
	return func(e *SignalExecution) (SignalExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("signal execution is nil")
		}
		e.EntityID = entityID
		return nil, nil
	}
}

func SetSideEffectExecutionStatus(status ExecutionStatus) SideEffectExecutionPropertySetter {
	return func(e *SideEffectExecution) (SideEffectExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect execution is nil")
		}
		e.Status = status
		return nil, nil
	}
}

func SetSideEffectExecutionStartedAt(startedAt time.Time) SideEffectExecutionPropertySetter {
	return func(e *SideEffectExecution) (SideEffectExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect execution is nil")
		}
		e.StartedAt = startedAt
		return nil, nil
	}
}

func SetSideEffectExecutionCompletedAt(completedAt time.Time) SideEffectExecutionPropertySetter {
	return func(e *SideEffectExecution) (SideEffectExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect execution is nil")
		}
		e.CompletedAt = &completedAt
		return nil, nil
	}
}

func SetSideEffectExecutionData(data *SideEffectExecutionData) SideEffectExecutionPropertySetter {
	return func(e *SideEffectExecution) (SideEffectExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect execution is nil")
		}
		e.SideEffectExecutionData = data
		return nil, nil
	}
}

func SetSideEffectExecutionError(err string) SideEffectExecutionPropertySetter {
	return func(e *SideEffectExecution) (SideEffectExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect execution is nil")
		}
		e.Error = err
		return nil, nil
	}
}

// SideEffectExecutionData property getters/setters
func GetSideEffectExecutionDataID(id *SideEffectExecutionDataID) SideEffectExecutionDataPropertyGetter {
	return func(d *SideEffectExecutionData) (SideEffectExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("side effect execution data is nil")
		}
		*id = d.ID
		return nil, nil
	}
}

func GetSideEffectExecutionDataExecutionID(executionID *SideEffectExecutionID) SideEffectExecutionDataPropertyGetter {
	return func(d *SideEffectExecutionData) (SideEffectExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("side effect execution data is nil")
		}
		*executionID = d.ExecutionID
		return nil, nil
	}
}

func GetSideEffectExecutionDataOutputs(outputs *[][]byte) SideEffectExecutionDataPropertyGetter {
	return func(d *SideEffectExecutionData) (SideEffectExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("side effect execution data is nil")
		}
		*outputs = make([][]byte, len(d.Outputs))
		for i, output := range d.Outputs {
			outputCopy := make([]byte, len(output))
			copy(outputCopy, output)
			(*outputs)[i] = outputCopy
		}
		return func(opts *SideEffectExecutionDataGetterOptions) error {
			opts.IncludeOutputs = true
			return nil
		}, nil
	}
}

func SetSideEffectExecutionDataExecutionID(executionID SideEffectExecutionID) SideEffectExecutionDataPropertySetter {
	return func(d *SideEffectExecutionData) (SideEffectExecutionDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("side effect execution data is nil")
		}
		d.ExecutionID = executionID
		return nil, nil
	}
}

func SetSideEffectExecutionDataOutputs(outputs [][]byte) SideEffectExecutionDataPropertySetter {
	return func(d *SideEffectExecutionData) (SideEffectExecutionDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("side effect execution data is nil")
		}
		d.Outputs = make([][]byte, len(outputs))
		for i, output := range outputs {
			outputCopy := make([]byte, len(output))
			copy(outputCopy, output)
			d.Outputs[i] = outputCopy
		}
		return nil, nil
	}
}

// Queue property getters/setters
func GetQueueID(id *QueueID) QueuePropertyGetter {
	return func(q *Queue) (QueuePropertyGetterOption, error) {
		if q == nil {
			return nil, errors.New("queue is nil")
		}
		*id = q.ID
		return nil, nil
	}
}

func GetQueueName(name *string) QueuePropertyGetter {
	return func(q *Queue) (QueuePropertyGetterOption, error) {
		if q == nil {
			return nil, errors.New("queue is nil")
		}
		*name = q.Name
		return nil, nil
	}
}

func GetQueueCreatedAt(createdAt *time.Time) QueuePropertyGetter {
	return func(q *Queue) (QueuePropertyGetterOption, error) {
		if q == nil {
			return nil, errors.New("queue is nil")
		}
		*createdAt = q.CreatedAt
		return nil, nil
	}
}

func GetQueueUpdatedAt(updatedAt *time.Time) QueuePropertyGetter {
	return func(q *Queue) (QueuePropertyGetterOption, error) {
		if q == nil {
			return nil, errors.New("queue is nil")
		}
		*updatedAt = q.UpdatedAt
		return nil, nil
	}
}

func GetQueueEntities(entities *[]*WorkflowEntity) QueuePropertyGetter {
	return func(q *Queue) (QueuePropertyGetterOption, error) {
		if q == nil {
			return nil, errors.New("queue is nil")
		}
		*entities = make([]*WorkflowEntity, len(q.Entities))
		copy(*entities, q.Entities)
		return func(opts *QueueGetterOptions) error {
			opts.IncludeWorkflows = true
			return nil
		}, nil
	}
}

func SetQueueName(name string) QueuePropertySetter {
	return func(q *Queue) (QueuePropertySetterOption, error) {
		if q == nil {
			return nil, errors.New("queue is nil")
		}
		q.Name = name
		return nil, nil
	}
}

func SetQueueCreatedAt(createdAt time.Time) QueuePropertySetter {
	return func(q *Queue) (QueuePropertySetterOption, error) {
		if q == nil {
			return nil, errors.New("queue is nil")
		}
		q.CreatedAt = createdAt
		return nil, nil
	}
}

func SetQueueUpdatedAt(updatedAt time.Time) QueuePropertySetter {
	return func(q *Queue) (QueuePropertySetterOption, error) {
		if q == nil {
			return nil, errors.New("queue is nil")
		}
		q.UpdatedAt = updatedAt
		return nil, nil
	}
}

func SetQueueEntities(entities []*WorkflowEntity) QueuePropertySetter {
	return func(q *Queue) (QueuePropertySetterOption, error) {
		if q == nil {
			return nil, errors.New("queue is nil")
		}
		q.Entities = make([]*WorkflowEntity, len(entities))
		copy(q.Entities, entities)
		return func(opts *QueueSetterOptions) error {
			workflowIDs := make([]WorkflowEntityID, len(entities))
			for i, entity := range entities {
				workflowIDs[i] = entity.ID
			}
			opts.WorkflowIDs = workflowIDs
			return nil
		}, nil
	}
}

// Hierarchy property getters/setters

func GetHierarchyID(id *HierarchyID) HierarchyPropertyGetter {
	return func(h *Hierarchy) (HierarchyPropertyGetterOption, error) {
		if h == nil {
			return nil, errors.New("hierarchy is nil")
		}
		*id = h.ID
		return nil, nil
	}
}

func GetHierarchyRunID(runID *RunID) HierarchyPropertyGetter {
	return func(h *Hierarchy) (HierarchyPropertyGetterOption, error) {
		if h == nil {
			return nil, errors.New("hierarchy is nil")
		}
		*runID = h.RunID
		return nil, nil
	}
}

func GetHierarchyParentEntityID(parentEntityID *int) HierarchyPropertyGetter {
	return func(h *Hierarchy) (HierarchyPropertyGetterOption, error) {
		if h == nil {
			return nil, errors.New("hierarchy is nil")
		}
		*parentEntityID = h.ParentEntityID
		return nil, nil
	}
}

func GetHierarchyChildEntityID(childEntityID *int) HierarchyPropertyGetter {
	return func(h *Hierarchy) (HierarchyPropertyGetterOption, error) {
		if h == nil {
			return nil, errors.New("hierarchy is nil")
		}
		*childEntityID = h.ChildEntityID
		return nil, nil
	}
}

func GetHierarchyParentExecutionID(parentExecutionID *int) HierarchyPropertyGetter {
	return func(h *Hierarchy) (HierarchyPropertyGetterOption, error) {
		if h == nil {
			return nil, errors.New("hierarchy is nil")
		}
		*parentExecutionID = h.ParentExecutionID
		return nil, nil
	}
}

func GetHierarchyChildExecutionID(childExecutionID *int) HierarchyPropertyGetter {
	return func(h *Hierarchy) (HierarchyPropertyGetterOption, error) {
		if h == nil {
			return nil, errors.New("hierarchy is nil")
		}
		*childExecutionID = h.ChildExecutionID
		return nil, nil
	}
}

func GetHierarchyParentStepID(parentStepID *string) HierarchyPropertyGetter {
	return func(h *Hierarchy) (HierarchyPropertyGetterOption, error) {
		if h == nil {
			return nil, errors.New("hierarchy is nil")
		}
		*parentStepID = h.ParentStepID
		return nil, nil
	}
}

func GetHierarchyChildStepID(childStepID *string) HierarchyPropertyGetter {
	return func(h *Hierarchy) (HierarchyPropertyGetterOption, error) {
		if h == nil {
			return nil, errors.New("hierarchy is nil")
		}
		*childStepID = h.ChildStepID
		return nil, nil
	}
}

func GetHierarchyParentType(parentType *EntityType) HierarchyPropertyGetter {
	return func(h *Hierarchy) (HierarchyPropertyGetterOption, error) {
		if h == nil {
			return nil, errors.New("hierarchy is nil")
		}
		*parentType = h.ParentType
		return nil, nil
	}
}

func GetHierarchyChildType(childType *EntityType) HierarchyPropertyGetter {
	return func(h *Hierarchy) (HierarchyPropertyGetterOption, error) {
		if h == nil {
			return nil, errors.New("hierarchy is nil")
		}
		*childType = h.ChildType
		return nil, nil
	}
}

func SetHierarchyRunID(runID RunID) HierarchyPropertySetter {
	return func(h *Hierarchy) (HierarchyPropertySetterOption, error) {
		if h == nil {
			return nil, errors.New("hierarchy is nil")
		}
		h.RunID = runID
		return nil, nil
	}
}

func SetHierarchyParentEntityID(parentEntityID int) HierarchyPropertySetter {
	return func(h *Hierarchy) (HierarchyPropertySetterOption, error) {
		if h == nil {
			return nil, errors.New("hierarchy is nil")
		}
		h.ParentEntityID = parentEntityID
		return nil, nil
	}
}

func SetHierarchyChildEntityID(childEntityID int) HierarchyPropertySetter {
	return func(h *Hierarchy) (HierarchyPropertySetterOption, error) {
		if h == nil {
			return nil, errors.New("hierarchy is nil")
		}
		h.ChildEntityID = childEntityID
		return nil, nil
	}
}

func SetHierarchyParentExecutionID(parentExecutionID int) HierarchyPropertySetter {
	return func(h *Hierarchy) (HierarchyPropertySetterOption, error) {
		if h == nil {
			return nil, errors.New("hierarchy is nil")
		}
		h.ParentExecutionID = parentExecutionID
		return nil, nil
	}
}

func SetHierarchyChildExecutionID(childExecutionID int) HierarchyPropertySetter {
	return func(h *Hierarchy) (HierarchyPropertySetterOption, error) {
		if h == nil {
			return nil, errors.New("hierarchy is nil")
		}
		h.ChildExecutionID = childExecutionID
		return nil, nil
	}
}

func SetHierarchyParentStepID(parentStepID string) HierarchyPropertySetter {
	return func(h *Hierarchy) (HierarchyPropertySetterOption, error) {
		if h == nil {
			return nil, errors.New("hierarchy is nil")
		}
		h.ParentStepID = parentStepID
		return nil, nil
	}
}

func SetHierarchyChildStepID(childStepID string) HierarchyPropertySetter {
	return func(h *Hierarchy) (HierarchyPropertySetterOption, error) {
		if h == nil {
			return nil, errors.New("hierarchy is nil")
		}
		h.ChildStepID = childStepID
		return nil, nil
	}
}

func SetHierarchyParentType(parentType EntityType) HierarchyPropertySetter {
	return func(h *Hierarchy) (HierarchyPropertySetterOption, error) {
		if h == nil {
			return nil, errors.New("hierarchy is nil")
		}
		h.ParentType = parentType
		return nil, nil
	}
}

func SetHierarchyChildType(childType EntityType) HierarchyPropertySetter {
	return func(h *Hierarchy) (HierarchyPropertySetterOption, error) {
		if h == nil {
			return nil, errors.New("hierarchy is nil")
		}
		h.ChildType = childType
		return nil, nil
	}
}

// Base entity getters/setters

func GetBaseEntityHandlerName(name *string) BaseEntityPropertyGetter {
	return func(e *BaseEntity) (BaseEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		*name = e.HandlerName
		return nil, nil
	}
}

func GetBaseEntityType(entityType *EntityType) BaseEntityPropertyGetter {
	return func(e *BaseEntity) (BaseEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		*entityType = e.Type
		return nil, nil
	}
}

func GetBaseEntityStatus(status *EntityStatus) BaseEntityPropertyGetter {
	return func(e *BaseEntity) (BaseEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		*status = e.Status
		return nil, nil
	}
}

func GetBaseEntityStepID(stepID *string) BaseEntityPropertyGetter {
	return func(e *BaseEntity) (BaseEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		*stepID = e.StepID
		return nil, nil
	}
}

func GetBaseEntityCreatedAt(createdAt *time.Time) BaseEntityPropertyGetter {
	return func(e *BaseEntity) (BaseEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		*createdAt = e.CreatedAt
		return nil, nil
	}
}

func GetBaseEntityUpdatedAt(updatedAt *time.Time) BaseEntityPropertyGetter {
	return func(e *BaseEntity) (BaseEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		*updatedAt = e.UpdatedAt
		return nil, nil
	}
}

// GetBaseEntityID needs to be generic to work with all entity types now
func GetBaseEntityQueueID(queueID *QueueID) BaseEntityPropertyGetter {
	return func(e *BaseEntity) (BaseEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		*queueID = e.QueueID
		return nil, nil
	}
}

func GetBaseEntityRunID(runID *RunID) BaseEntityPropertyGetter {
	return func(e *BaseEntity) (BaseEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		*runID = e.RunID
		return nil, nil
	}
}

func GetBaseEntityRetryPolicy(policy *RetryPolicy) BaseEntityPropertyGetter {
	return func(e *BaseEntity) (BaseEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		*policy = RetryPolicy{
			MaxAttempts: e.RetryPolicy.MaxAttempts,
			MaxInterval: time.Duration(e.RetryPolicy.MaxInterval),
		}
		return nil, nil
	}
}

func GetBaseEntityRetryState(state *RetryState) BaseEntityPropertyGetter {
	return func(e *BaseEntity) (BaseEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		*state = e.RetryState
		return nil, nil
	}
}

func SetBaseEntityHandlerName(name string) BaseEntityPropertySetter {
	return func(e *BaseEntity) (BaseEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		e.HandlerName = name
		return nil, nil
	}
}

func SetBaseEntityQueueID(queueID QueueID) BaseEntityPropertySetter {
	return func(e *BaseEntity) (BaseEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		e.QueueID = queueID
		return nil, nil
	}
}

func SetBaseEntityType(entityType EntityType) BaseEntityPropertySetter {
	return func(e *BaseEntity) (BaseEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		e.Type = entityType
		return nil, nil
	}
}

func SetBaseEntityStatus(status EntityStatus) BaseEntityPropertySetter {
	return func(e *BaseEntity) (BaseEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		e.Status = status
		return nil, nil
	}
}

func SetBaseEntityStepID(stepID string) BaseEntityPropertySetter {
	return func(e *BaseEntity) (BaseEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		e.StepID = stepID
		return nil, nil
	}
}

func SetBaseEntityCreatedAt(createdAt time.Time) BaseEntityPropertySetter {
	return func(e *BaseEntity) (BaseEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		e.CreatedAt = createdAt
		return nil, nil
	}
}

func SetBaseEntityUpdatedAt(updatedAt time.Time) BaseEntityPropertySetter {
	return func(e *BaseEntity) (BaseEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		e.UpdatedAt = updatedAt
		return nil, nil
	}
}

func SetBaseEntityRunID(runID RunID) BaseEntityPropertySetter {
	return func(e *BaseEntity) (BaseEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		e.RunID = runID
		return nil, nil
	}
}

func SetBaseEntityRetryPolicy(policy RetryPolicy) BaseEntityPropertySetter {
	return func(e *BaseEntity) (BaseEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		e.RetryPolicy = *ToInternalRetryPolicy(&policy)
		return nil, nil
	}
}

func SetBaseEntityRetryState(state RetryState) BaseEntityPropertySetter {
	return func(e *BaseEntity) (BaseEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		e.RetryState = state
		return nil, nil
	}
}

// Base execution getters/setters

func GetBaseExecutionStartedAt(startedAt *time.Time) BaseExecutionPropertyGetter {
	return func(e *BaseExecution) (BaseExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base execution is nil")
		}
		*startedAt = e.StartedAt
		return func(opts *BaseExecutionGetterOptions) error {
			opts.IncludeStarted = true
			return nil
		}, nil
	}
}

func GetBaseExecutionCompletedAt(completedAt *time.Time) BaseExecutionPropertyGetter {
	return func(e *BaseExecution) (BaseExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base execution is nil")
		}
		if e.CompletedAt != nil {
			*completedAt = *e.CompletedAt
		}
		return func(opts *BaseExecutionGetterOptions) error {
			opts.IncludeCompleted = true
			return nil
		}, nil
	}
}

func GetBaseExecutionStatus(status *ExecutionStatus) BaseExecutionPropertyGetter {
	return func(e *BaseExecution) (BaseExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base execution is nil")
		}
		*status = e.Status
		return nil, nil
	}
}

func GetBaseExecutionError(errStr *string) BaseExecutionPropertyGetter {
	return func(e *BaseExecution) (BaseExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base execution is nil")
		}
		*errStr = e.Error
		return func(opts *BaseExecutionGetterOptions) error {
			opts.IncludeError = true
			return nil
		}, nil
	}
}

func GetBaseExecutionCreatedAt(createdAt *time.Time) BaseExecutionPropertyGetter {
	return func(e *BaseExecution) (BaseExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base execution is nil")
		}
		*createdAt = e.CreatedAt
		return nil, nil
	}
}

func GetBaseExecutionUpdatedAt(updatedAt *time.Time) BaseExecutionPropertyGetter {
	return func(e *BaseExecution) (BaseExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base execution is nil")
		}
		*updatedAt = e.UpdatedAt
		return nil, nil
	}
}

func SetBaseExecutionStartedAt(startedAt time.Time) BaseExecutionPropertySetter {
	return func(e *BaseExecution) (BaseExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base execution is nil")
		}
		e.StartedAt = startedAt
		return nil, nil
	}
}

func SetBaseExecutionCompletedAt(completedAt time.Time) BaseExecutionPropertySetter {
	return func(e *BaseExecution) (BaseExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base execution is nil")
		}
		e.CompletedAt = &completedAt
		return nil, nil
	}
}

func SetBaseExecutionStatus(status ExecutionStatus) BaseExecutionPropertySetter {
	return func(e *BaseExecution) (BaseExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base execution is nil")
		}
		e.Status = status
		return nil, nil
	}
}

func SetBaseExecutionError(err string) BaseExecutionPropertySetter {
	return func(e *BaseExecution) (BaseExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base execution is nil")
		}
		e.Error = err
		return nil, nil
	}
}

func SetBaseExecutionCreatedAt(createdAt time.Time) BaseExecutionPropertySetter {
	return func(e *BaseExecution) (BaseExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base execution is nil")
		}
		e.CreatedAt = createdAt
		return nil, nil
	}
}

func SetBaseExecutionUpdatedAt(updatedAt time.Time) BaseExecutionPropertySetter {
	return func(e *BaseExecution) (BaseExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base execution is nil")
		}
		e.UpdatedAt = updatedAt
		return nil, nil
	}
}

// WorkflowExecutionData getters/setters
func GetWorkflowExecutionDataID(id *WorkflowExecutionDataID) WorkflowExecutionDataPropertyGetter {
	return func(d *WorkflowExecutionData) (WorkflowExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow execution data is nil")
		}
		*id = d.ID
		return nil, nil
	}
}

func GetWorkflowExecutionDataExecutionID(executionID *WorkflowExecutionID) WorkflowExecutionDataPropertyGetter {
	return func(d *WorkflowExecutionData) (WorkflowExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow execution data is nil")
		}
		*executionID = d.ExecutionID
		return nil, nil
	}
}

func GetWorkflowExecutionDataLastHeartbeat(lastHeartbeat *time.Time) WorkflowExecutionDataPropertyGetter {
	return func(d *WorkflowExecutionData) (WorkflowExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow execution data is nil")
		}
		if d.LastHeartbeat != nil {
			*lastHeartbeat = *d.LastHeartbeat
		}
		return nil, nil
	}
}

func GetWorkflowExecutionDataOutputs(outputs *[][]byte) WorkflowExecutionDataPropertyGetter {
	return func(d *WorkflowExecutionData) (WorkflowExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow execution data is nil")
		}
		*outputs = make([][]byte, len(d.Outputs))
		for i, output := range d.Outputs {
			outputCopy := make([]byte, len(output))
			copy(outputCopy, output)
			(*outputs)[i] = outputCopy
		}
		return func(opts *WorkflowExecutionDataGetterOptions) error {
			opts.IncludeOutputs = true
			return nil
		}, nil
	}
}

func SetWorkflowExecutionDataExecutionID(executionID WorkflowExecutionID) WorkflowExecutionDataPropertySetter {
	return func(d *WorkflowExecutionData) (WorkflowExecutionDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow execution data is nil")
		}
		d.ExecutionID = executionID
		return nil, nil
	}
}

func SetActivityExecutionDataExecutionID(executionID ActivityExecutionID) ActivityExecutionDataPropertySetter {
	return func(d *ActivityExecutionData) (ActivityExecutionDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("activity execution data is nil")
		}
		d.ExecutionID = executionID
		return nil, nil
	}
}

func SetWorkflowExecutionDataLastHeartbeat(lastHeartbeat time.Time) WorkflowExecutionDataPropertySetter {
	return func(d *WorkflowExecutionData) (WorkflowExecutionDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow execution data is nil")
		}
		d.LastHeartbeat = &lastHeartbeat
		return nil, nil
	}
}

func SetWorkflowExecutionDataOutputs(outputs [][]byte) WorkflowExecutionDataPropertySetter {
	return func(d *WorkflowExecutionData) (WorkflowExecutionDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow execution data is nil")
		}
		d.Outputs = make([][]byte, len(outputs))
		for i, output := range outputs {
			outputCopy := make([]byte, len(output))
			copy(outputCopy, output)
			d.Outputs[i] = outputCopy
		}
		return nil, nil
	}
}

// ActivityExecutionData getters/setters
func GetActivityExecutionDataExecutionID(executionID *ActivityExecutionID) ActivityExecutionDataPropertyGetter {
	return func(d *ActivityExecutionData) (ActivityExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("activity execution data is nil")
		}
		*executionID = d.ExecutionID
		return nil, nil
	}
}

func GetActivityExecutionDataLastHeartbeat(lastHeartbeat *time.Time) ActivityExecutionDataPropertyGetter {
	return func(d *ActivityExecutionData) (ActivityExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("activity execution data is nil")
		}
		if d.LastHeartbeat != nil {
			*lastHeartbeat = *d.LastHeartbeat
		}
		return nil, nil
	}
}

// Additional ActivityExecutionData setters (for completeness)
func SetActivityExecutionDataLastHeartbeat(lastHeartbeat time.Time) ActivityExecutionDataPropertySetter {
	return func(d *ActivityExecutionData) (ActivityExecutionDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("activity execution data is nil")
		}
		d.LastHeartbeat = &lastHeartbeat
		return nil, nil
	}
}

// ActivityData getters/setters for Output and ScheduledFor
func GetActivityDataOutput(output *[][]byte) ActivityDataPropertyGetter {
	return func(d *ActivityData) (ActivityDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("activity data is nil")
		}
		*output = make([][]byte, len(d.Output))
		for i, out := range d.Output {
			outputCopy := make([]byte, len(out))
			copy(outputCopy, out)
			(*output)[i] = outputCopy
		}
		return func(opts *ActivityDataGetterOptions) error {
			opts.IncludeOutputs = true
			return nil
		}, nil
	}
}

func SetActivityDataOutput(output [][]byte) ActivityDataPropertySetter {
	return func(d *ActivityData) (ActivityDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("activity data is nil")
		}
		d.Output = make([][]byte, len(output))
		for i, out := range output {
			outputCopy := make([]byte, len(out))
			copy(outputCopy, out)
			d.Output[i] = outputCopy
		}
		return nil, nil
	}
}

func WorkflowExecutionWithData() WorkflowExecutionGetOption {
	return func(opts *WorkflowExecutionGetterOptions) error {
		opts.IncludeData = true
		return nil
	}
}

func ActivityExecutionWithData() ActivityExecutionGetOption {
	return func(opts *ActivityExecutionGetterOptions) error {
		opts.IncludeData = true
		return nil
	}
}

func SideEffectExecutionWithData() SideEffectExecutionGetOption {
	return func(opts *SideEffectExecutionGetterOptions) error {
		opts.IncludeData = true
		return nil
	}
}

func SagaExecutionWithData() SagaExecutionGetOption {
	return func(opts *SagaExecutionGetterOptions) error {
		opts.IncludeData = true
		return nil
	}
}

func WorkflowEntityWithData() WorkflowEntityGetOption {
	return func(opts *WorkflowEntityGetterOptions) error {
		opts.IncludeData = true
		return nil
	}
}

func WorkflowEntityWithVersion() WorkflowEntityGetOption {
	return func(opts *WorkflowEntityGetterOptions) error {
		opts.IncludeVersion = true
		return nil
	}
}

func WorkflowEntityWithQueue() WorkflowEntityGetOption {
	return func(opts *WorkflowEntityGetterOptions) error {
		opts.IncludeQueue = true
		return nil
	}
}

func WorkflowEntityWithChildren() WorkflowEntityGetOption {
	return func(opts *WorkflowEntityGetterOptions) error {
		opts.IncludeChildren = true
		return nil
	}
}

func WorkflowEntityWithRun() WorkflowEntityGetOption {
	return func(opts *WorkflowEntityGetterOptions) error {
		opts.IncludeRun = true
		return nil
	}
}

func ActivityEntityWithData() ActivityEntityGetOption {
	return func(opts *ActivityEntityGetterOptions) error {
		opts.IncludeData = true
		return nil
	}
}

func SideEffectEntityWithData() SideEffectEntityGetOption {
	return func(opts *SideEffectEntityGetterOptions) error {
		opts.IncludeData = true
		return nil
	}
}

func SagaEntityWithData() SagaEntityGetOption {
	return func(opts *SagaEntityGetterOptions) error {
		opts.IncludeData = true
		return nil
	}
}

// WorkflowExecution StackTrace getters/setters
func GetWorkflowExecutionStackTrace(stackTrace **string) WorkflowExecutionPropertyGetter {
	return func(e *WorkflowExecution) (WorkflowExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		*stackTrace = e.StackTrace
		return nil, nil
	}
}

func SetWorkflowExecutionStackTrace(stackTrace string) WorkflowExecutionPropertySetter {
	return func(e *WorkflowExecution) (WorkflowExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		e.StackTrace = &stackTrace
		return nil, nil
	}
}

// ActivityExecution StackTrace getters/setters
func GetActivityExecutionStackTrace(stackTrace **string) ActivityExecutionPropertyGetter {
	return func(e *ActivityExecution) (ActivityExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("activity execution is nil")
		}
		*stackTrace = e.StackTrace
		return nil, nil
	}
}

func SetActivityExecutionStackTrace(stackTrace string) ActivityExecutionPropertySetter {
	return func(e *ActivityExecution) (ActivityExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("activity execution is nil")
		}
		e.StackTrace = &stackTrace
		return nil, nil
	}
}

// SideEffectExecution StackTrace getters/setters
func GetSideEffectExecutionStackTrace(stackTrace **string) SideEffectExecutionPropertyGetter {
	return func(e *SideEffectExecution) (SideEffectExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect execution is nil")
		}
		*stackTrace = e.StackTrace
		return nil, nil
	}
}

func SetSideEffectExecutionStackTrace(stackTrace string) SideEffectExecutionPropertySetter {
	return func(e *SideEffectExecution) (SideEffectExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect execution is nil")
		}
		e.StackTrace = &stackTrace
		return nil, nil
	}
}

// SagaExecution StackTrace getters/setters
func GetSagaExecutionStackTrace(stackTrace **string) SagaExecutionPropertyGetter {
	return func(e *SagaExecution) (SagaExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("saga execution is nil")
		}
		*stackTrace = e.StackTrace
		return nil, nil
	}
}

func SetSagaExecutionStackTrace(stackTrace string) SagaExecutionPropertySetter {
	return func(e *SagaExecution) (SagaExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("saga execution is nil")
		}
		e.StackTrace = &stackTrace
		return nil, nil
	}
}

func GetSagaExecutionType(execType *ExecutionType) SagaExecutionPropertyGetter {
	return func(e *SagaExecution) (SagaExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("saga execution is nil")
		}
		*execType = e.ExecutionType
		return nil, nil
	}
}

// Signal Entity Property getters/setters
func GetSignalEntityStatus(status *EntityStatus) SignalEntityPropertyGetter {
	return func(e *SignalEntity) (SignalEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("signal entity is nil")
		}
		*status = e.Status
		return nil, nil
	}
}

func GetSignalEntityHandlerName(name *string) SignalEntityPropertyGetter {
	return func(e *SignalEntity) (SignalEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("signal entity is nil")
		}
		*name = e.HandlerName
		return nil, nil
	}
}

func GetSignalEntityStepID(stepID *string) SignalEntityPropertyGetter {
	return func(e *SignalEntity) (SignalEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("signal entity is nil")
		}
		*stepID = e.StepID
		return nil, nil
	}
}

func GetSignalEntityRunID(runID *RunID) SignalEntityPropertyGetter {
	return func(e *SignalEntity) (SignalEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("signal entity is nil")
		}
		*runID = e.RunID
		return nil, nil
	}
}

func GetSignalEntityRetryPolicy(policy *RetryPolicy) SignalEntityPropertyGetter {
	return func(e *SignalEntity) (SignalEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("signal entity is nil")
		}
		*policy = RetryPolicy{
			MaxAttempts: e.RetryPolicy.MaxAttempts,
			MaxInterval: time.Duration(e.RetryPolicy.MaxInterval),
		}
		return nil, nil
	}
}

func GetSignalEntityRetryState(state *RetryState) SignalEntityPropertyGetter {
	return func(e *SignalEntity) (SignalEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("signal entity is nil")
		}
		*state = e.RetryState
		return nil, nil
	}
}

func GetSignalEntityData(data **SignalData) SignalEntityPropertyGetter {
	return func(e *SignalEntity) (SignalEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("signal entity is nil")
		}
		*data = e.SignalData
		return func(opts *SignalEntityGetterOptions) error {
			opts.IncludeData = true
			return nil
		}, nil
	}
}

func SetSignalEntityStatus(status EntityStatus) SignalEntityPropertySetter {
	return func(e *SignalEntity) (SignalEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("signal entity is nil")
		}
		e.Status = status
		return nil, nil
	}
}

func SetSignalEntityHandlerName(name string) SignalEntityPropertySetter {
	return func(e *SignalEntity) (SignalEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("signal entity is nil")
		}
		e.HandlerName = name
		return nil, nil
	}
}

func SetSignalEntityStepID(stepID string) SignalEntityPropertySetter {
	return func(e *SignalEntity) (SignalEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("signal entity is nil")
		}
		e.StepID = stepID
		return nil, nil
	}
}

func SetSignalEntityRunID(runID RunID) SignalEntityPropertySetter {
	return func(e *SignalEntity) (SignalEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("signal entity is nil")
		}
		e.RunID = runID
		return nil, nil
	}
}

func SetSignalEntityRetryPolicy(policy RetryPolicy) SignalEntityPropertySetter {
	return func(e *SignalEntity) (SignalEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("signal entity is nil")
		}
		e.RetryPolicy = *ToInternalRetryPolicy(&policy)
		return nil, nil
	}
}

func SetSignalEntityRetryState(state RetryState) SignalEntityPropertySetter {
	return func(e *SignalEntity) (SignalEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("signal entity is nil")
		}
		e.RetryState = state
		return nil, nil
	}
}

// Signal Data Property getters/setters
func GetSignalDataID(id *SignalDataID) SignalDataPropertyGetter {
	return func(d *SignalData) (SignalDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("signal data is nil")
		}
		*id = d.ID
		return nil, nil
	}
}

func GetSignalDataEntityID(entityID *SignalEntityID) SignalDataPropertyGetter {
	return func(d *SignalData) (SignalDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("signal data is nil")
		}
		*entityID = d.EntityID
		return nil, nil
	}
}

func GetSignalDataName(name *string) SignalDataPropertyGetter {
	return func(d *SignalData) (SignalDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("signal data is nil")
		}
		*name = d.Name
		return nil, nil
	}
}

func SetSignalDataEntityID(entityID SignalEntityID) SignalDataPropertySetter {
	return func(d *SignalData) (SignalDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("signal data is nil")
		}
		d.EntityID = entityID
		return nil, nil
	}
}

func SetSignalDataName(name string) SignalDataPropertySetter {
	return func(d *SignalData) (SignalDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("signal data is nil")
		}
		d.Name = name
		return nil, nil
	}
}

// Signal Execution Property getters/setters
func GetSignalExecutionID(id *SignalExecutionID) SignalExecutionPropertyGetter {
	return func(e *SignalExecution) (SignalExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("signal execution is nil")
		}
		*id = e.ID
		return nil, nil
	}
}

func GetSignalExecutionEntityID(entityID *SignalEntityID) SignalExecutionPropertyGetter {
	return func(e *SignalExecution) (SignalExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("signal execution is nil")
		}
		*entityID = e.EntityID
		return nil, nil
	}
}

func GetSignalExecutionStatus(status *ExecutionStatus) SignalExecutionPropertyGetter {
	return func(e *SignalExecution) (SignalExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("signal execution is nil")
		}
		*status = e.Status
		return nil, nil
	}
}

func GetSignalExecutionStartedAt(startedAt *time.Time) SignalExecutionPropertyGetter {
	return func(e *SignalExecution) (SignalExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("signal execution is nil")
		}
		*startedAt = e.StartedAt
		return nil, nil
	}
}

func GetSignalExecutionCompletedAt(completedAt *time.Time) SignalExecutionPropertyGetter {
	return func(e *SignalExecution) (SignalExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("signal execution is nil")
		}
		if e.CompletedAt != nil {
			*completedAt = *e.CompletedAt
		}
		return nil, nil
	}
}

func GetSignalExecutionError(errStr *string) SignalExecutionPropertyGetter {
	return func(e *SignalExecution) (SignalExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("signal execution is nil")
		}
		*errStr = e.Error
		return nil, nil
	}
}

func GetSignalExecutionStackTrace(stackTrace **string) SignalExecutionPropertyGetter {
	return func(e *SignalExecution) (SignalExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("signal execution is nil")
		}
		*stackTrace = e.StackTrace
		return nil, nil
	}
}

func SetSignalExecutionStatus(status ExecutionStatus) SignalExecutionPropertySetter {
	return func(e *SignalExecution) (SignalExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("signal execution is nil")
		}
		e.Status = status
		return nil, nil
	}
}

func SetSignalExecutionStartedAt(startedAt time.Time) SignalExecutionPropertySetter {
	return func(e *SignalExecution) (SignalExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("signal execution is nil")
		}
		e.StartedAt = startedAt
		return nil, nil
	}
}

func SetSignalExecutionCompletedAt(completedAt time.Time) SignalExecutionPropertySetter {
	return func(e *SignalExecution) (SignalExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("signal execution is nil")
		}
		e.CompletedAt = &completedAt
		return nil, nil
	}
}

func SetSignalExecutionError(err string) SignalExecutionPropertySetter {
	return func(e *SignalExecution) (SignalExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("signal execution is nil")
		}
		e.Error = err
		return nil, nil
	}
}

func SetSignalExecutionStackTrace(stackTrace string) SignalExecutionPropertySetter {
	return func(e *SignalExecution) (SignalExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("signal execution is nil")
		}
		e.StackTrace = &stackTrace
		return nil, nil
	}
}

// Signal Execution Data Property getters/setters
func GetSignalExecutionDataID(id *SignalExecutionDataID) SignalExecutionDataPropertyGetter {
	return func(d *SignalExecutionData) (SignalExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("signal execution data is nil")
		}
		*id = d.ID
		return nil, nil
	}
}

func GetSignalExecutionDataExecutionID(executionID *SignalExecutionID) SignalExecutionDataPropertyGetter {
	return func(d *SignalExecutionData) (SignalExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("signal execution data is nil")
		}
		*executionID = d.ExecutionID
		return nil, nil
	}
}

func GetSignalExecutionDataValue(value *[]byte) SignalExecutionDataPropertyGetter {
	return func(d *SignalExecutionData) (SignalExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("signal execution data is nil")
		}
		*value = make([]byte, len(d.Value))
		copy(*value, d.Value)
		return func(opts *SignalExecutionDataGetterOptions) error {
			opts.IncludeValue = true
			return nil
		}, nil
	}
}

func GetSignalExecutionKind(kind *uint) SignalExecutionDataPropertyGetter {
	return func(d *SignalExecutionData) (SignalExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("signal execution data is nil")
		}
		*kind = d.Kind
		return nil, nil
	}
}

func SetSignalExecutionDataExecutionID(executionID SignalExecutionID) SignalExecutionDataPropertySetter {
	return func(d *SignalExecutionData) (SignalExecutionDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("signal execution data is nil")
		}
		d.ExecutionID = executionID
		return nil, nil
	}
}

func SetSignalExecutionDataValue(value []byte) SignalExecutionDataPropertySetter {
	return func(d *SignalExecutionData) (SignalExecutionDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("signal execution data is nil")
		}
		d.Value = make([]byte, len(value))
		copy(d.Value, value)
		return nil, nil
	}
}

func SetSignalExecutionKind(kind uint) SignalExecutionDataPropertySetter {
	return func(d *SignalExecutionData) (SignalExecutionDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("signal execution data is nil")
		}
		d.Kind = kind
		return nil, nil
	}
}

// Convenience options
func SignalEntityWithData() SignalEntityGetOption {
	return func(opts *SignalEntityGetterOptions) error {
		opts.IncludeData = true
		return nil
	}
}

func SignalExecutionWithData() SignalExecutionGetOption {
	return func(opts *SignalExecutionGetterOptions) error {
		opts.IncludeData = true
		return nil
	}
}

func GetWorkflowVersions(versions *[]*Version) WorkflowEntityPropertyGetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		if e.Edges == nil || e.Edges.Versions == nil {
			*versions = make([]*Version, 0)
		} else {
			*versions = make([]*Version, len(e.Edges.Versions))
			copy(*versions, e.Edges.Versions)
		}
		return func(opts *WorkflowEntityGetterOptions) error {
			opts.IncludeVersion = true
			return nil
		}, nil
	}
}

func SetWorkflowVersion(version *Version) WorkflowEntityPropertySetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		if e.Edges == nil {
			e.Edges = &WorkflowEntityEdges{}
		}
		if e.Edges.Versions == nil {
			e.Edges.Versions = make([]*Version, 0)
		}
		e.Edges.Versions = append(e.Edges.Versions, version)
		return func(opts *WorkflowEntitySetterOptions) error {
			opts.Version = version
			return nil
		}, nil
	}
}
