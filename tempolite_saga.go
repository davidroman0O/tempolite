package tempolite

import (
	"fmt"
	"reflect"
)

type SagaStep interface {
	Transaction(ctx TransactionContext) (interface{}, error)
	Compensation(ctx CompensationContext) (interface{}, error)
}

type SagaDefinition struct {
	Steps       []SagaStep
	HandlerInfo *SagaHandlerInfo
}

type SagaDefinitionBuilder struct {
	steps []SagaStep
}

type SagaHandlerInfo struct {
	TransactionInfo  []HandlerInfo
	CompensationInfo []HandlerInfo
}

// NewSaga creates a new builder instance.
func NewSaga() *SagaDefinitionBuilder {
	return &SagaDefinitionBuilder{
		steps: make([]SagaStep, 0),
	}
}

// AddStep adds a saga step to the builder.
func (b *SagaDefinitionBuilder) AddStep(step SagaStep) *SagaDefinitionBuilder {
	b.steps = append(b.steps, step)
	return b
}

// analyzeMethod helper function to create HandlerInfo for a method
func analyzeMethod(method reflect.Method, name string) (HandlerInfo, error) {
	methodType := method.Type

	if methodType.NumIn() < 2 {
		return HandlerInfo{}, fmt.Errorf("method must have at least two parameters (receiver and context)")
	}

	paramTypes := make([]reflect.Type, methodType.NumIn()-2)
	paramKinds := make([]reflect.Kind, methodType.NumIn()-2)
	for i := 2; i < methodType.NumIn(); i++ {
		paramTypes[i-2] = methodType.In(i)
		paramKinds[i-2] = methodType.In(i).Kind()
	}

	returnTypes := make([]reflect.Type, methodType.NumOut()-1)
	returnKinds := make([]reflect.Kind, methodType.NumOut()-1)
	for i := 0; i < methodType.NumOut()-1; i++ {
		returnTypes[i] = methodType.Out(i)
		returnKinds[i] = methodType.Out(i).Kind()
	}

	handlerName := fmt.Sprintf("%s.%s", name, method.Name)

	return HandlerInfo{
		HandlerName:     handlerName,
		HandlerLongName: HandlerIdentity(name),
		Handler:         method.Func.Interface(),
		ParamTypes:      paramTypes,
		ParamsKinds:     paramKinds,
		ReturnTypes:     returnTypes,
		ReturnKinds:     returnKinds,
		NumIn:           methodType.NumIn() - 2,  // Exclude receiver and context
		NumOut:          methodType.NumOut() - 1, // Exclude error
	}, nil
}

// Build creates a SagaDefinition with the HandlerInfo included.
func (b *SagaDefinitionBuilder) Build() (*SagaDefinition, error) {
	sagaInfo := &SagaHandlerInfo{
		TransactionInfo:  make([]HandlerInfo, len(b.steps)),
		CompensationInfo: make([]HandlerInfo, len(b.steps)),
	}

	for i, step := range b.steps {
		stepType := reflect.TypeOf(step)
		if stepType.Kind() == reflect.Ptr {
			stepType = stepType.Elem()
		}

		transactionMethod, ok := stepType.MethodByName("Transaction")
		if !ok {
			return nil, fmt.Errorf("Transaction method not found for step %d", i)
		}
		compensationMethod, ok := stepType.MethodByName("Compensation")
		if !ok {
			return nil, fmt.Errorf("Compensation method not found for step %d", i)
		}

		transactionInfo, err := analyzeMethod(transactionMethod, stepType.Name())
		if err != nil {
			return nil, fmt.Errorf("error analyzing Transaction method for step %d: %w", i, err)
		}

		compensationInfo, err := analyzeMethod(compensationMethod, stepType.Name())
		if err != nil {
			return nil, fmt.Errorf("error analyzing Compensation method for step %d: %w", i, err)
		}

		sagaInfo.TransactionInfo[i] = transactionInfo
		sagaInfo.CompensationInfo[i] = compensationInfo
	}

	return &SagaDefinition{
		Steps:       b.steps,
		HandlerInfo: sagaInfo,
	}, nil
}

func (w WorkflowContext) Saga(stepID string, saga *SagaDefinition) *SagaInfo {
	if err := w.checkIfPaused(); err != nil {
		w.tp.logger.Debug(w.tp.ctx, "Workflow is paused, skipping Saga execution", "stepID", stepID, "error", err)
		return &SagaInfo{err: err}
	}
	// Enqueue the saga for execution
	return w.tp.saga(w, stepID, saga)
}
