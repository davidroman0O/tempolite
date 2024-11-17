package tempolite

// State and Trigger definitions
type state string

const (
	StateIdle          state = "Idle"
	StateExecuting     state = "Executing"
	StateCompleted     state = "Completed"
	StateFailed        state = "Failed"
	StateRetried       state = "Retried"
	StatePaused        state = "Paused"
	StateTransactions  state = "Transactions"
	StateCompensations state = "Compensations"
)

type trigger string

const (
	TriggerStart      trigger = "Start"
	TriggerComplete   trigger = "Complete"
	TriggerFail       trigger = "Fail"
	TriggerPause      trigger = "Pause"
	TriggerResume     trigger = "Resume"
	TriggerCompensate trigger = "Compensate"
)
