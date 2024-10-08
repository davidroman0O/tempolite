package types

// Enums
type TaskStatus int

const (
	TaskStatusPending TaskStatus = iota
	TaskStatusInProgress
	TaskStatusCompleted
	TaskStatusFailed
	TaskStatusCancelled
	TaskStatusTerminated
)

func TaskStatusValues() []string {
	return []string{
		"Pending",
		"InProgress",
		"Completed",
		"Failed",
		"Cancelled",
		"Terminated",
	}
}

// Convert any TaskStatus to a string
func TaskStatusToString(status TaskStatus) string {
	switch status {
	case TaskStatusPending:
		return "Pending"
	case TaskStatusInProgress:
		return "InProgress"
	case TaskStatusCompleted:
		return "Completed"
	case TaskStatusFailed:
		return "Failed"
	case TaskStatusCancelled:
		return "Cancelled"
	case TaskStatusTerminated:
		return "Terminated"
	default:
		return "Unknown"
	}
}

type SagaStatus int

const (
	SagaStatusPending SagaStatus = iota
	SagaStatusInProgress
	SagaStatusPaused
	SagaStatusCompleted
	SagaStatusFailed
	SagaStatusCancelled
	SagaStatusTerminating
	SagaStatusTerminated
	SagaStatusCriticallyFailed
)

type ExecutionStatus int

const (
	ExecutionStatusPending ExecutionStatus = iota
	ExecutionStatusInProgress
	ExecutionStatusCompleted
	ExecutionStatusFailed
	ExecutionStatusCancelled
	ExecutionStatusCriticallyFailed
)

type ExecutionNodeType int

const (
	ExecutionNodeTypeHandler ExecutionNodeType = iota
	ExecutionNodeTypeSagaHandler
	ExecutionNodeTypeSagaStep
	ExecutionNodeTypeSideEffect
	ExecutionNodeTypeCompensation
)
