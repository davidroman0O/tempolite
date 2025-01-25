package data

import (
	"errors"
	"fmt"
	"reflect"
)

type SagaStep interface {
	Transaction(ctx TransactionContext) error
	Compensation(ctx CompensationContext) error
}

type FunctionStep struct {
	transactionFn  func(ctx TransactionContext) error
	compensationFn func(ctx CompensationContext) error
}

func (s *FunctionStep) Transaction(ctx TransactionContext) error {
	return s.transactionFn(ctx)
}

func (s *FunctionStep) Compensation(ctx CompensationContext) error {
	return s.compensationFn(ctx)
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

func (b *SagaDefinitionBuilder) Add(transaction func(TransactionContext) error, compensation func(CompensationContext) error) *SagaDefinitionBuilder {
	b.steps = append(b.steps, &FunctionStep{
		transactionFn:  transaction,
		compensationFn: compensation,
	})
	return b
}

// AddStep adds a saga step to the builder.
func (b *SagaDefinitionBuilder) AddStep(step SagaStep) *SagaDefinitionBuilder {
	b.steps = append(b.steps, step)
	return b
}

// Build creates a SagaDefinition with the HandlerInfo included.
func (b *SagaDefinitionBuilder) Build() (*SagaDefinition, error) {

	// logger.Debug(context.Background(), "building saga definition", "saga_builder.steps", len(b.steps))

	sagaInfo := &SagaHandlerInfo{
		TransactionInfo:  make([]HandlerInfo, len(b.steps)),
		CompensationInfo: make([]HandlerInfo, len(b.steps)),
	}

	for i, step := range b.steps {
		stepType := reflect.TypeOf(step)
		originalType := stepType // Keep original type for handler name
		isPtr := stepType.Kind() == reflect.Ptr

		// Get the base type for method lookup
		if isPtr {
			stepType = stepType.Elem()
		}

		// Try to find methods on both pointer and value receivers
		var transactionMethod, compensationMethod reflect.Method
		var transactionOk, compensationOk bool

		// First try the original type (whether pointer or value)
		if transactionMethod, transactionOk = originalType.MethodByName("Transaction"); !transactionOk {
			// If not found and original wasn't a pointer, try pointer
			if !isPtr {
				if ptrMethod, ok := reflect.PtrTo(stepType).MethodByName("Transaction"); ok {
					transactionMethod = ptrMethod
					transactionOk = true
				}
			}
		}

		if compensationMethod, compensationOk = originalType.MethodByName("Compensation"); !compensationOk {
			// If not found and original wasn't a pointer, try pointer
			if !isPtr {
				if ptrMethod, ok := reflect.PtrTo(stepType).MethodByName("Compensation"); ok {
					compensationMethod = ptrMethod
					compensationOk = true
				}
			}
		}

		if !transactionOk {
			err := errors.Join(ErrTransactionMethodNotFound, fmt.Errorf("not found for step %d", i))
			// logger.Error(context.Background(), err.Error(), "saga_builder.step", i)
			return nil, err
		}
		if !compensationOk {
			err := errors.Join(ErrCompensationMethodNotFound, fmt.Errorf("not found for step %d", i))
			// logger.Error(context.Background(), err.Error(), "saga_builder.step", i)
			return nil, err
		}

		// Use the actual type name for the handler
		typeName := stepType.Name()
		if isPtr {
			typeName = "*" + typeName
		}

		// Note: Since Transaction/Compensation now only return error, we modify the analysis
		transactionInfo := HandlerInfo{
			HandlerName:     typeName + ".Transaction",
			HandlerLongName: HandlerIdentity(typeName),
			Handler:         transactionMethod.Func.Interface(),
			ParamTypes:      []reflect.Type{reflect.TypeOf((*TransactionContext)(nil)).Elem()},
			ReturnTypes:     []reflect.Type{}, // No return types except error
			NumIn:           1,
			NumOut:          0,
		}

		compensationInfo := HandlerInfo{
			HandlerName:     typeName + ".Compensation",
			HandlerLongName: HandlerIdentity(typeName),
			Handler:         compensationMethod.Func.Interface(),
			ParamTypes:      []reflect.Type{reflect.TypeOf((*CompensationContext)(nil)).Elem()},
			ReturnTypes:     []reflect.Type{}, // No return types except error
			NumIn:           1,
			NumOut:          0,
		}

		sagaInfo.TransactionInfo[i] = transactionInfo
		sagaInfo.CompensationInfo[i] = compensationInfo
	}

	// logger.Debug(context.Background(), "built saga definition", "saga_builder.steps", len(b.steps), "saga_builder.transaction_info", len(sagaInfo.TransactionInfo), "saga_builder.compensation_info", len(sagaInfo.CompensationInfo))

	return &SagaDefinition{
		Steps:       b.steps,
		HandlerInfo: sagaInfo,
	}, nil
}
