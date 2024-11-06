package tempolite

import (
	"reflect"
)

const (
	DefaultVersion = -1 // Equivalent to workflow.DefaultVersion in Temporal
)

type WorkflowID string

func (s WorkflowID) String() string {
	return string(s)
}

type WorkflowExecutionID string

func (s WorkflowExecutionID) String() string {
	return string(s)
}

type ActivityID string

func (s ActivityID) String() string {
	return string(s)
}

type ActivityExecutionID string

func (s ActivityExecutionID) String() string {
	return string(s)
}

type SideEffectID string

func (s SideEffectID) String() string {
	return string(s)
}

type SideEffectExecutionID string

func (s SideEffectExecutionID) String() string {
	return string(s)
}

type SagaID string

func (s SagaID) String() string {
	return string(s)
}

type SignalID string

func (s SignalID) String() string {
	return string(s)
}

type HandlerIdentity string

type HandlerInfo struct {
	HandlerName     string
	HandlerLongName HandlerIdentity
	Handler         interface{}
	ParamsKinds     []reflect.Kind
	ParamTypes      []reflect.Type
	ReturnTypes     []reflect.Type
	ReturnKinds     []reflect.Kind
	NumIn           int
	NumOut          int
}
