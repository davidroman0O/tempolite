package types

type WorkflowID int

func (w WorkflowID) ID() int {
	return int(w)
}

func (w WorkflowID) IsNoID() bool {
	return w == NoWorkflowID
}

var NoWorkflowID = WorkflowID(-1)

type WorkflowExecutionID int

func (w WorkflowExecutionID) ID() int {
	return int(w)
}

func (w WorkflowExecutionID) IsNoID() bool {
	return w == NoWorkflowExecutionID
}

var NoWorkflowExecutionID = WorkflowExecutionID(-1)

type ActivityID int

func (a ActivityID) ID() int {
	return int(a)
}

func (a ActivityID) IsNoID() bool {
	return a == NoActivityID
}

var NoActivityID = ActivityID(-1)

type ActivityExecutionID int

func (a ActivityExecutionID) ID() int {
	return int(a)
}

func (a ActivityExecutionID) IsNoID() bool {
	return a == NoActivityExecutionID
}

var NoActivityExecutionID = ActivityExecutionID(-1)

type SideEffectID int

func (s SideEffectID) ID() int {
	return int(s)
}

func (s SideEffectID) IsNoID() bool {
	return s == NoSideEffectID
}

var NoSideEffectID = SideEffectID(-1)

type SideEffectExecutionID int

func (s SideEffectExecutionID) ID() int {
	return int(s)
}

func (s SideEffectExecutionID) IsNoID() bool {
	return s == NoSideEffectExecutionID
}

var NoSideEffectExecutionID = SideEffectExecutionID(-1)

type SagaID int

func (s SagaID) ID() int {
	return int(s)
}

func (s SagaID) IsNoID() bool {
	return s == NoSagaID
}

var NoSagaID = SagaID(-1)

type SignalID int

func (s SignalID) ID() int {
	return int(s)
}

func (s SignalID) IsNoID() bool {
	return s == NoSignalID
}

var NoSignalID = SignalID(-1)
