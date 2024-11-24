package tempolite

import (
	"errors"
	"time"
)

// Run property getters/setters

// Basic property getters
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

// Relationship property getters
func GetRunEntities(entities *[]*WorkflowEntity) RunPropertyGetter {
	return func(r *Run) (RunPropertyGetterOption, error) {
		if r == nil {
			return nil, errors.New("run is nil")
		}
		*entities = make([]*WorkflowEntity, len(r.Entities))
		copy(*entities, r.Entities)
		return func(opts *RunPropertyGetterOptions) error {
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
		return func(opts *RunPropertyGetterOptions) error {
			opts.IncludeHierarchies = true
			return nil
		}, nil
	}
}

// Basic property setters
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

// Relationship property setters
func SetRunWorkflowEntity(workflowID int) RunPropertySetter {
	return func(r *Run) (RunPropertySetterOption, error) {
		if r == nil {
			return nil, errors.New("run is nil")
		}
		return func(opts *RunPropertySetterOptions) error {
			opts.WorkflowID = &workflowID
			return nil
		}, nil
	}
}

// Version property getters/setters

// Basic property getters
func GetVersionEntityID(entityID *int) VersionPropertyGetter {
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

// Basic property setters
func SetVersionEntityID(entityID int) VersionPropertySetter {
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
		v.Data = data
		return nil, nil
	}
}

// Workflow Entity property getters
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

func GetWorkflowEntityCreatedAt(createdAt *time.Time) WorkflowEntityPropertyGetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		*createdAt = e.CreatedAt
		return nil, nil
	}
}

func GetWorkflowEntityUpdatedAt(updatedAt *time.Time) WorkflowEntityPropertyGetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		*updatedAt = e.UpdatedAt
		return nil, nil
	}
}

func GetWorkflowEntityRunID(runID *int) WorkflowEntityPropertyGetter {
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
			MaxAttempts:        e.RetryPolicy.MaxAttempts,
			InitialInterval:    time.Duration(e.RetryPolicy.InitialInterval),
			BackoffCoefficient: e.RetryPolicy.BackoffCoefficient,
			MaxInterval:        time.Duration(e.RetryPolicy.MaxInterval),
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

func GetWorkflowEntityHandlerInfo(info *HandlerInfo) WorkflowEntityPropertyGetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		*info = e.HandlerInfo
		return nil, nil
	}
}

// Workflow Entity property setters
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

func SetWorkflowEntityRunID(runID int) WorkflowEntityPropertySetter {
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

func SetWorkflowEntityHandlerInfo(info HandlerInfo) WorkflowEntityPropertySetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		e.HandlerInfo = info
		return nil, nil
	}
}

// Workflow relationship setters
func SetWorkflowEntityQueue(queueID int) WorkflowEntityPropertySetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		return func(opts *WorkflowEntitySetterOptions) error {
			opts.QueueID = &queueID
			return nil
		}, nil
	}
}

func SetWorkflowEntityQueueByName(queueName string) WorkflowEntityPropertySetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		return func(opts *WorkflowEntitySetterOptions) error {
			opts.QueueName = &queueName
			return nil
		}, nil
	}
}

func SetWorkflowEntityVersion(version *Version) WorkflowEntityPropertySetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		return func(opts *WorkflowEntitySetterOptions) error {
			opts.Version = version
			return nil
		}, nil
	}
}

func AddWorkflowEntityChild(childID int, childType EntityType) WorkflowEntityPropertySetter {
	return func(e *WorkflowEntity) (WorkflowEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow entity is nil")
		}
		return func(opts *WorkflowEntitySetterOptions) error {
			opts.ChildID = &childID
			opts.ChildType = &childType
			return nil
		}, nil
	}
}

// WorkflowData property getters
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

func GetWorkflowDataAttempt(attempt *int) WorkflowDataPropertyGetter {
	return func(d *WorkflowData) (WorkflowDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow data is nil")
		}
		*attempt = d.Attempt
		return nil, nil
	}
}

// WorkflowData property setters
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

func SetWorkflowDataAttempt(attempt int) WorkflowDataPropertySetter {
	return func(d *WorkflowData) (WorkflowDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("workflow data is nil")
		}
		d.Attempt = attempt
		return nil, nil
	}
}

// ActivityEntity property getters
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

func GetActivityEntityCreatedAt(createdAt *time.Time) ActivityEntityPropertyGetter {
	return func(e *ActivityEntity) (ActivityEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		*createdAt = e.CreatedAt
		return nil, nil
	}
}

func GetActivityEntityUpdatedAt(updatedAt *time.Time) ActivityEntityPropertyGetter {
	return func(e *ActivityEntity) (ActivityEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		*updatedAt = e.UpdatedAt
		return nil, nil
	}
}

func GetActivityEntityRunID(runID *int) ActivityEntityPropertyGetter {
	return func(e *ActivityEntity) (ActivityEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		*runID = e.RunID
		return nil, nil
	}
}

func GetActivityEntityRetryPolicy(policy *RetryPolicy) ActivityEntityPropertyGetter {
	return func(e *ActivityEntity) (ActivityEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		*policy = RetryPolicy{
			MaxAttempts:        e.RetryPolicy.MaxAttempts,
			InitialInterval:    time.Duration(e.RetryPolicy.InitialInterval),
			BackoffCoefficient: e.RetryPolicy.BackoffCoefficient,
			MaxInterval:        time.Duration(e.RetryPolicy.MaxInterval),
		}
		return nil, nil
	}
}

// ActivityEntity property setters
func SetActivityEntityStatus(status EntityStatus) ActivityEntityPropertySetter {
	return func(e *ActivityEntity) (ActivityEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		e.Status = status
		return nil, nil
	}
}

func SetActivityEntityWorkflowID(workflowID int) ActivityEntityPropertySetter {
	return func(e *ActivityEntity) (ActivityEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("activity entity is nil")
		}
		return func(opts *ActivityEntitySetterOptions) error {
			opts.ParentWorkflowID = &workflowID
			return nil
		}, nil
	}
}

// ActivityData property getters
func GetActivityDataTimeout(timeout *int64) ActivityDataPropertyGetter {
	return func(d *ActivityData) (ActivityDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("activity data is nil")
		}
		*timeout = d.Timeout
		return nil, nil
	}
}

func GetActivityDataMaxAttempts(maxAttempts *int) ActivityDataPropertyGetter {
	return func(d *ActivityData) (ActivityDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("activity data is nil")
		}
		*maxAttempts = d.MaxAttempts
		return nil, nil
	}
}

func GetActivityDataScheduledFor(scheduledFor *time.Time) ActivityDataPropertyGetter {
	return func(d *ActivityData) (ActivityDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("activity data is nil")
		}
		if d.ScheduledFor != nil {
			*scheduledFor = *d.ScheduledFor
		}
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

func GetActivityDataOutput(output *[][]byte) ActivityDataPropertyGetter {
	return func(d *ActivityData) (ActivityDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("activity data is nil")
		}
		*output = make([][]byte, len(d.Output))
		for i, out := range d.Output {
			outCopy := make([]byte, len(out))
			copy(outCopy, out)
			(*output)[i] = outCopy
		}
		return func(opts *ActivityDataGetterOptions) error {
			opts.IncludeOutputs = true
			return nil
		}, nil
	}
}

// ActivityData property setters
func SetActivityDataTimeout(timeout int64) ActivityDataPropertySetter {
	return func(d *ActivityData) (ActivityDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("activity data is nil")
		}
		d.Timeout = timeout
		return nil, nil
	}
}

func SetActivityDataMaxAttempts(maxAttempts int) ActivityDataPropertySetter {
	return func(d *ActivityData) (ActivityDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("activity data is nil")
		}
		d.MaxAttempts = maxAttempts
		return nil, nil
	}
}

func SetActivityDataScheduledFor(scheduledFor time.Time) ActivityDataPropertySetter {
	return func(d *ActivityData) (ActivityDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("activity data is nil")
		}
		d.ScheduledFor = &scheduledFor
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

func SetActivityDataOutput(output [][]byte) ActivityDataPropertySetter {
	return func(d *ActivityData) (ActivityDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("activity data is nil")
		}
		d.Output = make([][]byte, len(output))
		for i, out := range output {
			outCopy := make([]byte, len(out))
			copy(outCopy, out)
			d.Output[i] = outCopy
		}
		return nil, nil
	}
}

// SagaEntity property getters
func GetSagaEntityStatus(status *EntityStatus) SagaEntityPropertyGetter {
	return func(e *SagaEntity) (SagaEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("saga entity is nil")
		}
		*status = e.Status
		return nil, nil
	}
}

func GetSagaEntityWorkflowID(workflowID *int) SagaEntityPropertyGetter {
	return func(e *SagaEntity) (SagaEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("saga entity is nil")
		}
		*workflowID = e.RunID // This will be replaced by the actual workflow ID from relationship
		return func(opts *SagaEntityGetterOptions) error {
			opts.IncludeWorkflow = true
			return nil
		}, nil
	}
}

// SagaEntity property setters
func SetSagaEntityStatus(status EntityStatus) SagaEntityPropertySetter {
	return func(e *SagaEntity) (SagaEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("saga entity is nil")
		}
		e.Status = status
		return nil, nil
	}
}

func SetSagaEntityWorkflowID(workflowID int) SagaEntityPropertySetter {
	return func(e *SagaEntity) (SagaEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("saga entity is nil")
		}
		return func(opts *SagaEntitySetterOptions) error {
			opts.ParentWorkflowID = &workflowID
			return nil
		}, nil
	}
}

// SagaData property getters
func GetSagaDataCompensating(compensating *bool) SagaDataPropertyGetter {
	return func(d *SagaData) (SagaDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("saga data is nil")
		}
		*compensating = d.Compensating
		return nil, nil
	}
}

func GetSagaDataCompensationData(compensationData *[][]byte) SagaDataPropertyGetter {
	return func(d *SagaData) (SagaDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("saga data is nil")
		}
		*compensationData = make([][]byte, len(d.CompensationData))
		for i, cd := range d.CompensationData {
			cdCopy := make([]byte, len(cd))
			copy(cdCopy, cd)
			(*compensationData)[i] = cdCopy
		}
		return func(opts *SagaDataGetterOptions) error {
			opts.IncludeCompensationData = true
			return nil
		}, nil
	}
}

// SagaData property setters
func SetSagaDataCompensating(compensating bool) SagaDataPropertySetter {
	return func(d *SagaData) (SagaDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("saga data is nil")
		}
		d.Compensating = compensating
		return nil, nil
	}
}

func SetSagaDataCompensationData(compensationData [][]byte) SagaDataPropertySetter {
	return func(d *SagaData) (SagaDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("saga data is nil")
		}
		d.CompensationData = make([][]byte, len(compensationData))
		for i, cd := range compensationData {
			cdCopy := make([]byte, len(cd))
			copy(cdCopy, cd)
			d.CompensationData[i] = cdCopy
		}
		return nil, nil
	}
}

// SideEffectEntity property getters
func GetSideEffectEntityStatus(status *EntityStatus) SideEffectEntityPropertyGetter {
	return func(e *SideEffectEntity) (SideEffectEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect entity is nil")
		}
		*status = e.Status
		return nil, nil
	}
}

func GetSideEffectEntityWorkflowID(workflowID *int) SideEffectEntityPropertyGetter {
	return func(e *SideEffectEntity) (SideEffectEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect entity is nil")
		}
		*workflowID = e.RunID // Will be replaced by actual workflow ID from relationship
		return func(opts *SideEffectEntityGetterOptions) error {
			opts.IncludeWorkflow = true
			return nil
		}, nil
	}
}

// SideEffectEntity property setters
func SetSideEffectEntityStatus(status EntityStatus) SideEffectEntityPropertySetter {
	return func(e *SideEffectEntity) (SideEffectEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect entity is nil")
		}
		e.Status = status
		return nil, nil
	}
}

func SetSideEffectEntityWorkflowID(workflowID int) SideEffectEntityPropertySetter {
	return func(e *SideEffectEntity) (SideEffectEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("side effect entity is nil")
		}
		return func(opts *SideEffectEntitySetterOptions) error {
			opts.ParentWorkflowID = &workflowID
			return nil
		}, nil
	}
}

// Workflow Execution property getters
func GetWorkflowExecutionStatus(status *ExecutionStatus) WorkflowExecutionPropertyGetter {
	return func(e *WorkflowExecution) (WorkflowExecutionDataPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		*status = e.Status
		return nil, nil
	}
}

func GetWorkflowExecutionStartedAt(startedAt *time.Time) WorkflowExecutionPropertyGetter {
	return func(e *WorkflowExecution) (WorkflowExecutionDataPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		*startedAt = e.StartedAt
		return nil, nil
	}
}

func GetWorkflowExecutionCompletedAt(completedAt *time.Time) WorkflowExecutionPropertyGetter {
	return func(e *WorkflowExecution) (WorkflowExecutionDataPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		if e.CompletedAt != nil {
			*completedAt = *e.CompletedAt
		}
		return nil, nil
	}
}

func GetWorkflowExecutionAttempt(attempt *int) WorkflowExecutionPropertyGetter {
	return func(e *WorkflowExecution) (WorkflowExecutionDataPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		*attempt = e.Attempt
		return nil, nil
	}
}

func GetWorkflowExecutionError(errStr *string) WorkflowExecutionPropertyGetter {
	return func(e *WorkflowExecution) (WorkflowExecutionDataPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		*errStr = e.Error
		return nil, nil
	}
}

// Workflow Execution property setters
func SetWorkflowExecutionStatus(status ExecutionStatus) WorkflowExecutionPropertySetter {
	return func(e *WorkflowExecution) (WorkflowExecutionDataPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		e.Status = status
		return nil, nil
	}
}

func SetWorkflowExecutionStartedAt(startedAt time.Time) WorkflowExecutionPropertySetter {
	return func(e *WorkflowExecution) (WorkflowExecutionDataPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		e.StartedAt = startedAt
		return nil, nil
	}
}

func SetWorkflowExecutionCompletedAt(completedAt time.Time) WorkflowExecutionPropertySetter {
	return func(e *WorkflowExecution) (WorkflowExecutionDataPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		e.CompletedAt = &completedAt
		return nil, nil
	}
}

func SetWorkflowExecutionAttempt(attempt int) WorkflowExecutionPropertySetter {
	return func(e *WorkflowExecution) (WorkflowExecutionDataPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		e.Attempt = attempt
		return nil, nil
	}
}

func SetWorkflowExecutionError(errStr string) WorkflowExecutionPropertySetter {
	return func(e *WorkflowExecution) (WorkflowExecutionDataPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("workflow execution is nil")
		}
		e.Error = errStr
		return nil, nil
	}
}

// WorkflowExecutionData property getters/setters
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

// ActivityExecutionData property getters/setters
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

func SetActivityExecutionDataLastHeartbeat(lastHeartbeat time.Time) ActivityExecutionDataPropertySetter {
	return func(d *ActivityExecutionData) (ActivityExecutionDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("activity execution data is nil")
		}
		d.LastHeartbeat = &lastHeartbeat
		return nil, nil
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

// SagaExecutionData property getters/setters
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

func GetSagaExecutionDataOutput(output *[][]byte) SagaExecutionDataPropertyGetter {
	return func(d *SagaExecutionData) (SagaExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("saga execution data is nil")
		}
		*output = make([][]byte, len(d.Output))
		for i, o := range d.Output {
			outputCopy := make([]byte, len(o))
			copy(outputCopy, o)
			(*output)[i] = outputCopy
		}
		return func(opts *SagaExecutionDataGetterOptions) error {
			opts.IncludeOutput = true
			return nil
		}, nil
	}
}

func GetSagaExecutionDataHasOutput(hasOutput *bool) SagaExecutionDataPropertyGetter {
	return func(d *SagaExecutionData) (SagaExecutionDataPropertyGetterOption, error) {
		if d == nil {
			return nil, errors.New("saga execution data is nil")
		}
		*hasOutput = d.HasOutput
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

func SetSagaExecutionDataOutput(output [][]byte) SagaExecutionDataPropertySetter {
	return func(d *SagaExecutionData) (SagaExecutionDataPropertySetterOption, error) {
		if d == nil {
			return nil, errors.New("saga execution data is nil")
		}
		d.Output = make([][]byte, len(output))
		for i, o := range output {
			outputCopy := make([]byte, len(o))
			copy(outputCopy, o)
			d.Output[i] = outputCopy
		}
		d.HasOutput = len(output) > 0
		return nil, nil
	}
}

// SideEffectExecutionData property getters/setters
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

func GetQueueWorkflows(workflows *[]*WorkflowEntity) QueuePropertyGetter {
	return func(q *Queue) (QueuePropertyGetterOption, error) {
		if q == nil {
			return nil, errors.New("queue is nil")
		}
		*workflows = make([]*WorkflowEntity, len(q.Entities))
		copy(*workflows, q.Entities)
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

func SetQueueWorkflows(workflowIDs []int) QueuePropertySetter {
	return func(q *Queue) (QueuePropertySetterOption, error) {
		if q == nil {
			return nil, errors.New("queue is nil")
		}
		return func(opts *QueueSetterOptions) error {
			opts.WorkflowIDs = workflowIDs
			return nil
		}, nil
	}
}

// Hierarchy property getters
func GetHierarchyRunID(runID *int) HierarchyPropertyGetter {
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

// Hierarchy property setters
func SetHierarchyRunID(runID int) HierarchyPropertySetter {
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

// Base Entity property getters
func GetBaseEntityID(id *int) BaseEntityPropertyGetter {
	return func(e *BaseEntity) (BaseEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		*id = e.ID
		return nil, nil
	}
}

func GetBaseEntityHandlerName(handlerName *string) BaseEntityPropertyGetter {
	return func(e *BaseEntity) (BaseEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		*handlerName = e.HandlerName
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

func GetBaseEntityRunID(runID *int) BaseEntityPropertyGetter {
	return func(e *BaseEntity) (BaseEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		*runID = e.RunID
		return nil, nil
	}
}

func GetBaseEntityRetryPolicy(retryPolicy *RetryPolicy) BaseEntityPropertyGetter {
	return func(e *BaseEntity) (BaseEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		internal := e.RetryPolicy
		*retryPolicy = RetryPolicy{
			MaxAttempts:        internal.MaxAttempts,
			InitialInterval:    time.Duration(internal.InitialInterval),
			BackoffCoefficient: internal.BackoffCoefficient,
			MaxInterval:        time.Duration(internal.MaxInterval),
		}
		return nil, nil
	}
}

func GetBaseEntityRetryState(retryState *RetryState) BaseEntityPropertyGetter {
	return func(e *BaseEntity) (BaseEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		*retryState = e.RetryState
		return nil, nil
	}
}

func GetBaseEntityHandlerInfo(handlerInfo *HandlerInfo) BaseEntityPropertyGetter {
	return func(e *BaseEntity) (BaseEntityPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		*handlerInfo = e.HandlerInfo
		return nil, nil
	}
}

// Base Entity property setters
func SetBaseEntityHandlerName(handlerName string) BaseEntityPropertySetter {
	return func(e *BaseEntity) (BaseEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		e.HandlerName = handlerName
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

func SetBaseEntityRetryPolicy(retryPolicy RetryPolicy) BaseEntityPropertySetter {
	return func(e *BaseEntity) (BaseEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		e.RetryPolicy = *ToInternalRetryPolicy(&retryPolicy)
		return nil, nil
	}
}

func SetBaseEntityRetryState(retryState RetryState) BaseEntityPropertySetter {
	return func(e *BaseEntity) (BaseEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		e.RetryState = retryState
		return nil, nil
	}
}

func SetBaseEntityHandlerInfo(handlerInfo HandlerInfo) BaseEntityPropertySetter {
	return func(e *BaseEntity) (BaseEntityPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base entity is nil")
		}
		e.HandlerInfo = handlerInfo
		return nil, nil
	}
}

// Base Execution property getters
func GetBaseExecutionID(id *int) BaseExecutionPropertyGetter {
	return func(e *BaseExecution) (BaseExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base execution is nil")
		}
		*id = e.ID
		return nil, nil
	}
}

func GetBaseExecutionEntityID(entityID *int) BaseExecutionPropertyGetter {
	return func(e *BaseExecution) (BaseExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base execution is nil")
		}
		*entityID = e.EntityID
		return nil, nil
	}
}

func GetBaseExecutionStartedAt(startedAt *time.Time) BaseExecutionPropertyGetter {
	return func(e *BaseExecution) (BaseExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base execution is nil")
		}
		*startedAt = e.StartedAt
		return nil, nil
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
		return nil, nil
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

func GetBaseExecutionError(err *string) BaseExecutionPropertyGetter {
	return func(e *BaseExecution) (BaseExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base execution is nil")
		}
		*err = e.Error
		return nil, nil
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

func GetBaseExecutionAttempt(attempt *int) BaseExecutionPropertyGetter {
	return func(e *BaseExecution) (BaseExecutionPropertyGetterOption, error) {
		if e == nil {
			return nil, errors.New("base execution is nil")
		}
		*attempt = e.Attempt
		return nil, nil
	}
}

// Base Execution property setters
func SetBaseExecutionEntityID(entityID int) BaseExecutionPropertySetter {
	return func(e *BaseExecution) (BaseExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base execution is nil")
		}
		e.EntityID = entityID
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

func SetBaseExecutionAttempt(attempt int) BaseExecutionPropertySetter {
	return func(e *BaseExecution) (BaseExecutionPropertySetterOption, error) {
		if e == nil {
			return nil, errors.New("base execution is nil")
		}
		e.Attempt = attempt
		return nil, nil
	}
}
