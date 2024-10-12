package tempolite

import (
	"fmt"
	"log"
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

type HandlerIdentity string

type SagaStep interface {
	Transaction(ctx TransactionContext) (interface{}, error)
	Compensation(ctx CompensationContext) (interface{}, error)
}

type SagaDefinition struct {
	Steps []SagaStep
}

type SagaDefinitionBuilder struct {
	tp    *Tempolite
	steps []SagaStep
}

// NewSaga creates a new builder instance with a reference to Tempolite.
func NewSaga(tp *Tempolite) *SagaDefinitionBuilder {
	return &SagaDefinitionBuilder{
		tp:    tp,
		steps: make([]SagaStep, 0),
	}
}

// AddStep adds a registered saga step to the builder using a direct instance.
func (b *SagaDefinitionBuilder) AddStep(step SagaStep) (*SagaDefinitionBuilder, error) {
	stepType := reflect.TypeOf(step)
	if stepType.Kind() == reflect.Ptr {
		stepType = stepType.Elem() // Get the underlying element type if it's a pointer
	}
	stepName := stepType.Name()

	// Check if the step is already registered in Tempolite
	if _, exists := b.tp.sagas.Load(stepName); !exists {
		// If not registered, attempt to register it using Tempolite's method
		if err := b.tp.RegisterSaga(sagaRegisterType(stepType)); err != nil {
			return b, fmt.Errorf("failed to register saga step %s: %w", stepName, err)
		}
		log.Printf("Registered saga step %s", stepName)
	}

	// After registration, add the step to the list
	b.steps = append(b.steps, step)
	return b, nil
}

// Build creates a SagaDefinition with the registered steps.
func (b *SagaDefinitionBuilder) Build() *SagaDefinition {
	return &SagaDefinition{
		Steps: b.steps,
	}
}

type SagaActivityBuilder[T any] func(input T, builder *SagaDefinitionBuilder) *SagaDefinition

func NewSagaActvityBuilder[T any](builder SagaActivityBuilder[T]) interface{} {
	return builder
}

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
