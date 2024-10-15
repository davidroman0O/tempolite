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

// func (hi HandlerInfo) ToInterface(data []byte) ([]interface{}, error) {
// 	var paramData []json.RawMessage
// 	if err := json.Unmarshal(data, &paramData); err != nil {
// 		return nil, fmt.Errorf("failed to unmarshal parameters: %v", err)
// 	}

// 	if len(paramData) != len(hi.ParamTypes) {
// 		return nil, fmt.Errorf("parameter count mismatch: expected %d, got %d", len(hi.ParamTypes), len(paramData))
// 	}

// 	params := make([]interface{}, len(hi.ParamTypes))
// 	for i, paramType := range hi.ParamTypes {
// 		paramPtr := reflect.New(paramType)
// 		if err := json.Unmarshal(paramData[i], paramPtr.Interface()); err != nil {
// 			return nil, fmt.Errorf("failed to unmarshal parameter %d: %v", i, err)
// 		}
// 		params[i] = paramPtr.Elem().Interface()
// 	}

// 	return params, nil
// }
