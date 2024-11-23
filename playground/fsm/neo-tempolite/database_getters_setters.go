package tempolite

import (
	"errors"
	"time"
)

// Run property getters/setters
func GetRunStatus(status *RunStatus) RunPropertyGetter {
	return func(r *Run) error {
		if r == nil {
			return errors.New("run is nil")
		}
		*status = r.Status
		return nil
	}
}

func SetRunStatus(status RunStatus) RunPropertySetter {
	return func(r *Run) error {
		if r == nil {
			return errors.New("run is nil")
		}
		r.Status = status
		return nil
	}
}

func GetRunCreatedAt(createdAt *time.Time) RunPropertyGetter {
	return func(r *Run) error {
		if r == nil {
			return errors.New("run is nil")
		}
		*createdAt = r.CreatedAt
		return nil
	}
}

func SetRunCreatedAt(createdAt time.Time) RunPropertySetter {
	return func(r *Run) error {
		if r == nil {
			return errors.New("run is nil")
		}
		r.CreatedAt = createdAt
		return nil
	}
}

func GetRunUpdatedAt(updatedAt *time.Time) RunPropertyGetter {
	return func(r *Run) error {
		if r == nil {
			return errors.New("run is nil")
		}
		*updatedAt = r.UpdatedAt
		return nil
	}
}

func SetRunUpdatedAt(updatedAt time.Time) RunPropertySetter {
	return func(r *Run) error {
		if r == nil {
			return errors.New("run is nil")
		}
		r.UpdatedAt = updatedAt
		return nil
	}
}

func GetRunEntities(entities *[]*WorkflowEntity) RunPropertyGetter {
	return func(r *Run) error {
		if r == nil {
			return errors.New("run is nil")
		}
		*entities = make([]*WorkflowEntity, len(r.Entities))
		copy(*entities, r.Entities)
		return nil
	}
}

func SetRunEntities(entities []*WorkflowEntity) RunPropertySetter {
	return func(r *Run) error {
		if r == nil {
			return errors.New("run is nil")
		}
		r.Entities = entities
		return nil
	}
}

func GetRunHierarchies(hierarchies *[]*Hierarchy) RunPropertyGetter {
	return func(r *Run) error {
		if r == nil {
			return errors.New("run is nil")
		}
		*hierarchies = make([]*Hierarchy, len(r.Hierarchies))
		copy(*hierarchies, r.Hierarchies)
		return nil
	}
}

func SetRunHierarchies(hierarchies []*Hierarchy) RunPropertySetter {
	return func(r *Run) error {
		if r == nil {
			return errors.New("run is nil")
		}
		r.Hierarchies = hierarchies
		return nil
	}
}

// Version property getters/setters
func GetVersionEntityID(entityID *int) VersionPropertyGetter {
	return func(v *Version) error {
		if v == nil {
			return errors.New("version is nil")
		}
		*entityID = v.EntityID
		return nil
	}
}

func SetVersionEntityID(entityID int) VersionPropertySetter {
	return func(v *Version) error {
		if v == nil {
			return errors.New("version is nil")
		}
		v.EntityID = entityID
		return nil
	}
}

func GetVersionChangeID(changeID *string) VersionPropertyGetter {
	return func(v *Version) error {
		if v == nil {
			return errors.New("version is nil")
		}
		*changeID = v.ChangeID
		return nil
	}
}

func SetVersionChangeID(changeID string) VersionPropertySetter {
	return func(v *Version) error {
		if v == nil {
			return errors.New("version is nil")
		}
		v.ChangeID = changeID
		return nil
	}
}

func GetVersionValue(version *int) VersionPropertyGetter {
	return func(v *Version) error {
		if v == nil {
			return errors.New("version is nil")
		}
		*version = v.Version
		return nil
	}
}

func SetVersionValue(version int) VersionPropertySetter {
	return func(v *Version) error {
		if v == nil {
			return errors.New("version is nil")
		}
		v.Version = version
		return nil
	}
}

func GetVersionData(data *map[string]interface{}) VersionPropertyGetter {
	return func(v *Version) error {
		if v == nil {
			return errors.New("version is nil")
		}
		*data = make(map[string]interface{})
		for k, v := range v.Data {
			(*data)[k] = v
		}
		return nil
	}
}

func SetVersionData(data map[string]interface{}) VersionPropertySetter {
	return func(v *Version) error {
		if v == nil {
			return errors.New("version is nil")
		}
		v.Data = data
		return nil
	}
}

// Hierarchy property getters/setters
func GetHierarchyRunID(runID *int) HierarchyPropertyGetter {
	return func(h *Hierarchy) error {
		if h == nil {
			return errors.New("hierarchy is nil")
		}
		*runID = h.RunID
		return nil
	}
}

func SetHierarchyRunID(runID int) HierarchyPropertySetter {
	return func(h *Hierarchy) error {
		if h == nil {
			return errors.New("hierarchy is nil")
		}
		h.RunID = runID
		return nil
	}
}

func GetHierarchyParentEntityID(parentEntityID *int) HierarchyPropertyGetter {
	return func(h *Hierarchy) error {
		if h == nil {
			return errors.New("hierarchy is nil")
		}
		*parentEntityID = h.ParentEntityID
		return nil
	}
}

func SetHierarchyParentEntityID(parentEntityID int) HierarchyPropertySetter {
	return func(h *Hierarchy) error {
		if h == nil {
			return errors.New("hierarchy is nil")
		}
		h.ParentEntityID = parentEntityID
		return nil
	}
}

func GetHierarchyChildEntityID(childEntityID *int) HierarchyPropertyGetter {
	return func(h *Hierarchy) error {
		if h == nil {
			return errors.New("hierarchy is nil")
		}
		*childEntityID = h.ChildEntityID
		return nil
	}
}

func SetHierarchyChildEntityID(childEntityID int) HierarchyPropertySetter {
	return func(h *Hierarchy) error {
		if h == nil {
			return errors.New("hierarchy is nil")
		}
		h.ChildEntityID = childEntityID
		return nil
	}
}

func GetHierarchyParentExecutionID(parentExecutionID *int) HierarchyPropertyGetter {
	return func(h *Hierarchy) error {
		if h == nil {
			return errors.New("hierarchy is nil")
		}
		*parentExecutionID = h.ParentExecutionID
		return nil
	}
}

func SetHierarchyParentExecutionID(parentExecutionID int) HierarchyPropertySetter {
	return func(h *Hierarchy) error {
		if h == nil {
			return errors.New("hierarchy is nil")
		}
		h.ParentExecutionID = parentExecutionID
		return nil
	}
}

func GetHierarchyChildExecutionID(childExecutionID *int) HierarchyPropertyGetter {
	return func(h *Hierarchy) error {
		if h == nil {
			return errors.New("hierarchy is nil")
		}
		*childExecutionID = h.ChildExecutionID
		return nil
	}
}

func SetHierarchyChildExecutionID(childExecutionID int) HierarchyPropertySetter {
	return func(h *Hierarchy) error {
		if h == nil {
			return errors.New("hierarchy is nil")
		}
		h.ChildExecutionID = childExecutionID
		return nil
	}
}

func GetHierarchyParentStepID(parentStepID *string) HierarchyPropertyGetter {
	return func(h *Hierarchy) error {
		if h == nil {
			return errors.New("hierarchy is nil")
		}
		*parentStepID = h.ParentStepID
		return nil
	}
}

func SetHierarchyParentStepID(parentStepID string) HierarchyPropertySetter {
	return func(h *Hierarchy) error {
		if h == nil {
			return errors.New("hierarchy is nil")
		}
		h.ParentStepID = parentStepID
		return nil
	}
}

func GetHierarchyChildStepID(childStepID *string) HierarchyPropertyGetter {
	return func(h *Hierarchy) error {
		if h == nil {
			return errors.New("hierarchy is nil")
		}
		*childStepID = h.ChildStepID
		return nil
	}
}

func SetHierarchyChildStepID(childStepID string) HierarchyPropertySetter {
	return func(h *Hierarchy) error {
		if h == nil {
			return errors.New("hierarchy is nil")
		}
		h.ChildStepID = childStepID
		return nil
	}
}

func GetHierarchyParentType(parentType *EntityType) HierarchyPropertyGetter {
	return func(h *Hierarchy) error {
		if h == nil {
			return errors.New("hierarchy is nil")
		}
		*parentType = h.ParentType
		return nil
	}
}

func SetHierarchyParentType(parentType EntityType) HierarchyPropertySetter {
	return func(h *Hierarchy) error {
		if h == nil {
			return errors.New("hierarchy is nil")
		}
		h.ParentType = parentType
		return nil
	}
}

func GetHierarchyChildType(childType *EntityType) HierarchyPropertyGetter {
	return func(h *Hierarchy) error {
		if h == nil {
			return errors.New("hierarchy is nil")
		}
		*childType = h.ChildType
		return nil
	}
}

func SetHierarchyChildType(childType EntityType) HierarchyPropertySetter {
	return func(h *Hierarchy) error {
		if h == nil {
			return errors.New("hierarchy is nil")
		}
		h.ChildType = childType
		return nil
	}
}

// Queue property getters/setters
func GetQueueName(name *string) QueuePropertyGetter {
	return func(q *Queue) error {
		if q == nil {
			return errors.New("queue is nil")
		}
		*name = q.Name
		return nil
	}
}

func SetQueueName(name string) QueuePropertySetter {
	return func(q *Queue) error {
		if q == nil {
			return errors.New("queue is nil")
		}
		q.Name = name
		return nil
	}
}

func GetQueueCreatedAt(createdAt *time.Time) QueuePropertyGetter {
	return func(q *Queue) error {
		if q == nil {
			return errors.New("queue is nil")
		}
		*createdAt = q.CreatedAt
		return nil
	}
}

func SetQueueCreatedAt(createdAt time.Time) QueuePropertySetter {
	return func(q *Queue) error {
		if q == nil {
			return errors.New("queue is nil")
		}
		q.CreatedAt = createdAt
		return nil
	}
}

func GetQueueUpdatedAt(updatedAt *time.Time) QueuePropertyGetter {
	return func(q *Queue) error {
		if q == nil {
			return errors.New("queue is nil")
		}
		*updatedAt = q.UpdatedAt
		return nil
	}
}

func SetQueueUpdatedAt(updatedAt time.Time) QueuePropertySetter {
	return func(q *Queue) error {
		if q == nil {
			return errors.New("queue is nil")
		}
		q.UpdatedAt = updatedAt
		return nil
	}
}

func GetQueueEntities(entities *[]*WorkflowEntity) QueuePropertyGetter {
	return func(q *Queue) error {
		if q == nil {
			return errors.New("queue is nil")
		}
		*entities = make([]*WorkflowEntity, len(q.Entities))
		copy(*entities, q.Entities)
		return nil
	}
}

func SetQueueEntities(entities []*WorkflowEntity) QueuePropertySetter {
	return func(q *Queue) error {
		if q == nil {
			return errors.New("queue is nil")
		}
		q.Entities = entities
		return nil
	}
}

// Entity Data property getters/setters
// WorkflowData
func GetWorkflowDataDuration(duration *string) WorkflowDataPropertyGetter {
	return func(wd *WorkflowData) error {
		if wd == nil {
			return errors.New("workflow data is nil")
		}
		*duration = wd.Duration
		return nil
	}
}

func SetWorkflowEntityStatus(status EntityStatus) WorkflowEntityPropertySetter {
	return func(we *WorkflowEntity) error {
		if we == nil {
			return errors.New("workflow entity is nil")
		}
		we.Status = status
		return nil
	}
}

func SetWorkflowDataDuration(duration string) WorkflowDataPropertySetter {
	return func(wd *WorkflowData) error {
		if wd == nil {
			return errors.New("workflow data is nil")
		}
		wd.Duration = duration
		return nil
	}
}

func GetWorkflowDataPaused(paused *bool) WorkflowDataPropertyGetter {
	return func(wd *WorkflowData) error {
		if wd == nil {
			return errors.New("workflow data is nil")
		}
		*paused = wd.Paused
		return nil
	}
}

func SetWorkflowDataPaused(paused bool) WorkflowDataPropertySetter {
	return func(wd *WorkflowData) error {
		if wd == nil {
			return errors.New("workflow data is nil")
		}
		wd.Paused = paused
		return nil
	}
}

func GetWorkflowDataResumable(resumable *bool) WorkflowDataPropertyGetter {
	return func(wd *WorkflowData) error {
		if wd == nil {
			return errors.New("workflow data is nil")
		}
		*resumable = wd.Resumable
		return nil
	}
}

func SetWorkflowDataResumable(resumable bool) WorkflowDataPropertySetter {
	return func(wd *WorkflowData) error {
		if wd == nil {
			return errors.New("workflow data is nil")
		}
		wd.Resumable = resumable
		return nil
	}
}

// Continue WorkflowData getters/setters
func GetWorkflowDataInputs(inputs *[][]byte) WorkflowDataPropertyGetter {
	return func(wd *WorkflowData) error {
		if wd == nil {
			return errors.New("workflow data is nil")
		}
		*inputs = make([][]byte, len(wd.Inputs))
		for i, input := range wd.Inputs {
			inputCopy := make([]byte, len(input))
			copy(inputCopy, input)
			(*inputs)[i] = inputCopy
		}
		return nil
	}
}

func SetWorkflowDataInputs(inputs [][]byte) WorkflowDataPropertySetter {
	return func(wd *WorkflowData) error {
		if wd == nil {
			return errors.New("workflow data is nil")
		}
		wd.Inputs = inputs
		return nil
	}
}

func GetWorkflowDataAttempt(attempt *int) WorkflowDataPropertyGetter {
	return func(wd *WorkflowData) error {
		if wd == nil {
			return errors.New("workflow data is nil")
		}
		*attempt = wd.Attempt
		return nil
	}
}

func SetWorkflowDataAttempt(attempt int) WorkflowDataPropertySetter {
	return func(wd *WorkflowData) error {
		if wd == nil {
			return errors.New("workflow data is nil")
		}
		wd.Attempt = attempt
		return nil
	}
}

// ActivityData getters/setters
func GetActivityDataTimeout(timeout *int64) ActivityDataPropertyGetter {
	return func(ad *ActivityData) error {
		if ad == nil {
			return errors.New("activity data is nil")
		}
		*timeout = ad.Timeout
		return nil
	}
}

func SetActivityEntityStatus(status EntityStatus) ActivityEntityPropertySetter {
	return func(ae *ActivityEntity) error {
		if ae == nil {
			return errors.New("activity entity is nil")
		}
		ae.Status = status
		return nil
	}
}

func SetActivityDataTimeout(timeout int64) ActivityDataPropertySetter {
	return func(ad *ActivityData) error {
		if ad == nil {
			return errors.New("activity data is nil")
		}
		ad.Timeout = timeout
		return nil
	}
}

func GetActivityDataMaxAttempts(maxAttempts *int) ActivityDataPropertyGetter {
	return func(ad *ActivityData) error {
		if ad == nil {
			return errors.New("activity data is nil")
		}
		*maxAttempts = ad.MaxAttempts
		return nil
	}
}

func SetActivityDataMaxAttempts(maxAttempts int) ActivityDataPropertySetter {
	return func(ad *ActivityData) error {
		if ad == nil {
			return errors.New("activity data is nil")
		}
		ad.MaxAttempts = maxAttempts
		return nil
	}
}

func GetActivityDataScheduledFor(scheduledFor *time.Time) ActivityDataPropertyGetter {
	return func(ad *ActivityData) error {
		if ad == nil {
			return errors.New("activity data is nil")
		}
		if ad.ScheduledFor != nil {
			*scheduledFor = *ad.ScheduledFor
		}
		return nil
	}
}

func SetActivityDataScheduledFor(scheduledFor time.Time) ActivityDataPropertySetter {
	return func(ad *ActivityData) error {
		if ad == nil {
			return errors.New("activity data is nil")
		}
		ad.ScheduledFor = &scheduledFor
		return nil
	}
}

func GetActivityDataInputs(inputs *[][]byte) ActivityDataPropertyGetter {
	return func(ad *ActivityData) error {
		if ad == nil {
			return errors.New("activity data is nil")
		}
		*inputs = make([][]byte, len(ad.Inputs))
		for i, input := range ad.Inputs {
			inputCopy := make([]byte, len(input))
			copy(inputCopy, input)
			(*inputs)[i] = inputCopy
		}
		return nil
	}
}

func SetActivityDataInputs(inputs [][]byte) ActivityDataPropertySetter {
	return func(ad *ActivityData) error {
		if ad == nil {
			return errors.New("activity data is nil")
		}
		ad.Inputs = inputs
		return nil
	}
}

func GetActivityDataOutput(output *[][]byte) ActivityDataPropertyGetter {
	return func(ad *ActivityData) error {
		if ad == nil {
			return errors.New("activity data is nil")
		}
		*output = make([][]byte, len(ad.Output))
		for i, out := range ad.Output {
			outCopy := make([]byte, len(out))
			copy(outCopy, out)
			(*output)[i] = outCopy
		}
		return nil
	}
}

func SetActivityDataOutput(output [][]byte) ActivityDataPropertySetter {
	return func(ad *ActivityData) error {
		if ad == nil {
			return errors.New("activity data is nil")
		}
		ad.Output = output
		return nil
	}
}

func GetActivityDataAttempt(attempt *int) ActivityDataPropertyGetter {
	return func(ad *ActivityData) error {
		if ad == nil {
			return errors.New("activity data is nil")
		}
		*attempt = ad.Attempt
		return nil
	}
}

func SetActivityDataAttempt(attempt int) ActivityDataPropertySetter {
	return func(ad *ActivityData) error {
		if ad == nil {
			return errors.New("activity data is nil")
		}
		ad.Attempt = attempt
		return nil
	}
}

// SagaData getters/setters
func GetSagaDataCompensating(compensating *bool) SagaDataPropertyGetter {
	return func(sd *SagaData) error {
		if sd == nil {
			return errors.New("saga data is nil")
		}
		*compensating = sd.Compensating
		return nil
	}
}

func SetSagaDataCompensating(compensating bool) SagaDataPropertySetter {
	return func(sd *SagaData) error {
		if sd == nil {
			return errors.New("saga data is nil")
		}
		sd.Compensating = compensating
		return nil
	}
}

func GetSagaDataCompensationData(compensationData *[][]byte) SagaDataPropertyGetter {
	return func(sd *SagaData) error {
		if sd == nil {
			return errors.New("saga data is nil")
		}
		*compensationData = make([][]byte, len(sd.CompensationData))
		for i, data := range sd.CompensationData {
			dataCopy := make([]byte, len(data))
			copy(dataCopy, data)
			(*compensationData)[i] = dataCopy
		}
		return nil
	}
}

func SetSagaDataCompensationData(compensationData [][]byte) SagaDataPropertySetter {
	return func(sd *SagaData) error {
		if sd == nil {
			return errors.New("saga data is nil")
		}
		sd.CompensationData = compensationData
		return nil
	}
}

// Execution Data getters/setters
// WorkflowExecutionData
func GetWorkflowExecutionDataLastHeartbeat(lastHeartbeat *time.Time) WorkflowExecutionDataPropertyGetter {
	return func(wed *WorkflowExecutionData) error {
		if wed == nil {
			return errors.New("workflow execution data is nil")
		}
		if wed.LastHeartbeat != nil {
			*lastHeartbeat = *wed.LastHeartbeat
		}
		return nil
	}
}

func SetWorkflowExecutionDataLastHeartbeat(lastHeartbeat time.Time) WorkflowExecutionDataPropertySetter {
	return func(wed *WorkflowExecutionData) error {
		if wed == nil {
			return errors.New("workflow execution data is nil")
		}
		wed.LastHeartbeat = &lastHeartbeat
		return nil
	}
}

// Property setters for WorkflowExecution
func SetWorkflowExecutionStatus(status ExecutionStatus) WorkflowExecutionPropertySetter {
	return func(we *WorkflowExecution) error {
		if we == nil {
			return errors.New("workflow execution is nil")
		}
		we.Status = status
		return nil
	}
}

func SetWorkflowExecutionCompletedAt(completedAt time.Time) WorkflowExecutionPropertySetter {
	return func(we *WorkflowExecution) error {
		if we == nil {
			return errors.New("workflow execution is nil")
		}
		we.CompletedAt = &completedAt
		return nil
	}
}

func SetWorkflowExecutionError(err string) WorkflowExecutionPropertySetter {
	return func(we *WorkflowExecution) error {
		if we == nil {
			return errors.New("workflow execution is nil")
		}
		we.Error = err
		return nil
	}
}

func GetWorkflowExecutionDataError(err string) WorkflowExecutionPropertySetter {
	return func(we *WorkflowExecution) error {
		if we == nil {
			return errors.New("workflow execution is nil")
		}
		we.Error = err
		return nil
	}
}

func GetWorkflowExecutionDataOutputs(outputs *[][]byte) WorkflowExecutionDataPropertyGetter {
	return func(wed *WorkflowExecutionData) error {
		if wed == nil {
			return errors.New("workflow execution data is nil")
		}
		*outputs = make([][]byte, len(wed.Outputs))
		for i, output := range wed.Outputs {
			outputCopy := make([]byte, len(output))
			copy(outputCopy, output)
			(*outputs)[i] = outputCopy
		}
		return nil
	}
}

func SetWorkflowExecutionDataOutputs(outputs [][]byte) WorkflowExecutionDataPropertySetter {
	return func(wed *WorkflowExecutionData) error {
		if wed == nil {
			return errors.New("workflow execution data is nil")
		}
		wed.Outputs = outputs
		return nil
	}
}

// ActivityExecutionData
func GetActivityExecutionDataLastHeartbeat(lastHeartbeat *time.Time) ActivityExecutionDataPropertyGetter {
	return func(aed *ActivityExecutionData) error {
		if aed == nil {
			return errors.New("activity execution data is nil")
		}
		if aed.LastHeartbeat != nil {
			*lastHeartbeat = *aed.LastHeartbeat
		}
		return nil
	}
}

func SetActivityExecutionStatus(status ExecutionStatus) ActivityExecutionPropertySetter {
	return func(ae *ActivityExecution) error {
		if ae == nil {
			return errors.New("activity execution is nil")
		}
		ae.Status = status
		return nil
	}
}

func SetActivityExecutionCompletedAt(completedAt time.Time) ActivityExecutionPropertySetter {
	return func(ae *ActivityExecution) error {
		if ae == nil {
			return errors.New("activity execution is nil")
		}
		ae.CompletedAt = &completedAt
		return nil
	}
}

func SetActivityEntityError(err string) ActivityExecutionPropertySetter {
	return func(ae *ActivityExecution) error {
		if ae == nil {
			return errors.New("activity execution is nil")
		}
		ae.Error = err
		return nil
	}
}

func SetActivityExecutionDataLastHeartbeat(lastHeartbeat time.Time) ActivityExecutionDataPropertySetter {
	return func(aed *ActivityExecutionData) error {
		if aed == nil {
			return errors.New("activity execution data is nil")
		}
		aed.LastHeartbeat = &lastHeartbeat
		return nil
	}
}

func GetActivityExecutionDataOutputs(outputs *[][]byte) ActivityExecutionDataPropertyGetter {
	return func(aed *ActivityExecutionData) error {
		if aed == nil {
			return errors.New("activity execution data is nil")
		}
		*outputs = make([][]byte, len(aed.Outputs))
		for i, output := range aed.Outputs {
			outputCopy := make([]byte, len(output))
			copy(outputCopy, output)
			(*outputs)[i] = outputCopy
		}
		return nil
	}
}

func SetActivityExecutionDataOutputs(outputs [][]byte) ActivityExecutionDataPropertySetter {
	return func(aed *ActivityExecutionData) error {
		if aed == nil {
			return errors.New("activity execution data is nil")
		}
		aed.Outputs = outputs
		return nil
	}
}

// SagaExecutionData
func GetSagaExecutionDataLastHeartbeat(lastHeartbeat *time.Time) SagaExecutionDataPropertyGetter {
	return func(sed *SagaExecutionData) error {
		if sed == nil {
			return errors.New("saga execution data is nil")
		}
		if sed.LastHeartbeat != nil {
			*lastHeartbeat = *sed.LastHeartbeat
		}
		return nil
	}
}

func SetSagaExecutionStatus(status ExecutionStatus) SagaExecutionPropertySetter {
	return func(se *SagaExecution) error {
		if se == nil {
			return errors.New("saga execution is nil")
		}
		se.Status = status
		return nil
	}
}

func SetSagaExecutionCompletedAt(completedAt time.Time) SagaExecutionPropertySetter {
	return func(se *SagaExecution) error {
		if se == nil {
			return errors.New("saga execution is nil")
		}
		se.CompletedAt = &completedAt
		return nil
	}
}

func SetSagaEntityStatus(status EntityStatus) SagaEntityPropertySetter {
	return func(se *SagaEntity) error {
		if se == nil {
			return errors.New("saga entity is nil")
		}
		se.Status = status
		return nil
	}
}

func SetSagaExecutionError(err string) SagaExecutionPropertySetter {
	return func(se *SagaExecution) error {
		if se == nil {
			return errors.New("saga execution is nil")
		}
		se.Error = err
		return nil
	}
}

func SetSagaExecutionDataLastHeartbeat(lastHeartbeat time.Time) SagaExecutionDataPropertySetter {
	return func(sed *SagaExecutionData) error {
		if sed == nil {
			return errors.New("saga execution data is nil")
		}
		sed.LastHeartbeat = &lastHeartbeat
		return nil
	}
}

func GetSagaExecutionDataOutput(output *[][]byte) SagaExecutionDataPropertyGetter {
	return func(sed *SagaExecutionData) error {
		if sed == nil {
			return errors.New("saga execution data is nil")
		}
		*output = make([][]byte, len(sed.Output))
		for i, out := range sed.Output {
			outCopy := make([]byte, len(out))
			copy(outCopy, out)
			(*output)[i] = outCopy
		}
		return nil
	}
}

func SetSagaExecutionDataOutput(output [][]byte) SagaExecutionDataPropertySetter {
	return func(sed *SagaExecutionData) error {
		if sed == nil {
			return errors.New("saga execution data is nil")
		}
		sed.Output = output
		return nil
	}
}

func GetSagaExecutionDataHasOutput(hasOutput *bool) SagaExecutionDataPropertyGetter {
	return func(sed *SagaExecutionData) error {
		if sed == nil {
			return errors.New("saga execution data is nil")
		}
		*hasOutput = sed.HasOutput
		return nil
	}
}

func SetSagaExecutionDataHasOutput(hasOutput bool) SagaExecutionDataPropertySetter {
	return func(sed *SagaExecutionData) error {
		if sed == nil {
			return errors.New("saga execution data is nil")
		}
		sed.HasOutput = hasOutput
		return nil
	}
}

// SideEffectExecutionData
func GetSideEffectExecutionDataOutputs(outputs *[][]byte) SideEffectExecutionDataPropertyGetter {
	return func(seed *SideEffectExecutionData) error {
		if seed == nil {
			return errors.New("side effect execution data is nil")
		}
		*outputs = make([][]byte, len(seed.Outputs))
		for i, output := range seed.Outputs {
			outputCopy := make([]byte, len(output))
			copy(outputCopy, output)
			(*outputs)[i] = outputCopy
		}
		return nil
	}
}

func SetSideEffectExecutionDataOutputs(outputs [][]byte) SideEffectExecutionDataPropertySetter {
	return func(seed *SideEffectExecutionData) error {
		if seed == nil {
			return errors.New("side effect execution data is nil")
		}
		seed.Outputs = outputs
		return nil
	}
}

func SetSideEffectExecutionStatus(status ExecutionStatus) SideEffectExecutionPropertySetter {
	return func(see *SideEffectExecution) error {
		if see == nil {
			return errors.New("side effect execution is nil")
		}
		see.Status = status
		return nil
	}
}

func SetSideEffectExecutionCompletedAt(completedAt time.Time) SideEffectExecutionPropertySetter {
	return func(see *SideEffectExecution) error {
		if see == nil {
			return errors.New("side effect execution is nil")
		}
		see.CompletedAt = &completedAt
		return nil
	}
}

func SetSideEffectEntityStatus(status EntityStatus) SideEffectEntityPropertySetter {
	return func(see *SideEffectEntity) error {
		if see == nil {
			return errors.New("side effect entity is nil")
		}
		see.Status = status
		return nil
	}
}
