package tempolite

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
)

type Workflow HandlerInfo

func As[T any, Y Identifier]() HandlerIdentity {
	dataType := reflect.TypeOf((*T)(nil)).Elem()
	return HandlerIdentity(fmt.Sprintf("%s/%s", dataType.PkgPath(), dataType.Name()))
}

type Activity HandlerInfo

type ActivityRegister func() (*Activity, error)

func AsActivity[T any, Y Identifier](instance ...T) ActivityRegister {
	dataType := reflect.TypeOf((*T)(nil)).Elem()

	var defaultInstance T
	if len(instance) > 0 {
		defaultInstance = instance[0]
	} else {
		defaultInstance = reflect.Zero(dataType).Interface().(T)
	}

	return activityRegisterType[T, Y](dataType, defaultInstance)
}

func activityRegisterType[T any, Y Identifier](dataType reflect.Type, instance T) ActivityRegister {
	return func() (*Activity, error) {
		handlerValue := reflect.ValueOf(instance)
		runMethod := handlerValue.MethodByName("Run")
		if !runMethod.IsValid() {
			return nil, fmt.Errorf("%s must have a Run method", dataType)
		}

		runMethodType := runMethod.Type()

		if runMethodType.NumIn() < 1 {
			return nil, fmt.Errorf("Run method of %s must have at least one input parameter (ActivityContext)", dataType)
		}

		if runMethodType.In(0) != reflect.TypeOf(ActivityContext[Y]{}) {
			return nil, fmt.Errorf("first parameter of Run method in %s must be ActivityContext", dataType)
		}

		// Collect all parameter types after the context parameter
		paramTypes := []reflect.Type{}
		paramsKinds := []reflect.Kind{}
		for i := 1; i < runMethodType.NumIn(); i++ {
			paramTypes = append(paramTypes, runMethodType.In(i))
			paramsKinds = append(paramsKinds, runMethodType.In(i).Kind())
		}

		// Collect all return types except the last one if it's an error
		numOut := runMethodType.NumOut()
		if numOut == 0 {
			return nil, fmt.Errorf("Run method in %s must return at least an error", dataType)
		}

		returnTypes := []reflect.Type{}
		returnKinds := []reflect.Kind{}
		for i := 0; i < numOut-1; i++ {
			returnTypes = append(returnTypes, runMethodType.Out(i))
			returnKinds = append(returnKinds, runMethodType.Out(i).Kind())
		}

		// Verify that the last return type is error
		if runMethodType.Out(numOut-1) != reflect.TypeOf((*error)(nil)).Elem() {
			return nil, fmt.Errorf("last return value of Run method in %s must be error", dataType)
		}

		act := &Activity{
			HandlerName:     dataType.Name(),
			HandlerLongName: HandlerIdentity(fmt.Sprintf("%s/%s", dataType.PkgPath(), dataType.Name())),
			Handler:         runMethod.Interface(), // Set Handler to the Run method
			ParamTypes:      paramTypes,
			ParamsKinds:     paramsKinds,
			ReturnTypes:     returnTypes,
			ReturnKinds:     returnKinds,
			NumIn:           runMethodType.NumIn() - 1, // Exclude context
			NumOut:          numOut - 1,                // Exclude error
		}

		log.Printf("Registering activity %s with name %s", dataType.Name(), act.HandlerLongName)

		return act, nil
	}
}

type SideEffect HandlerInfo

// type SideEffectRegister func() (*SideEffect, error)

// func AsSideEffect[T any, Y Identifier](instance ...T) SideEffectRegister {
// 	dataType := reflect.TypeOf((*T)(nil)).Elem()

// 	var defaultInstance T
// 	if len(instance) > 0 {
// 		defaultInstance = instance[0]
// 	} else {
// 		defaultInstance = reflect.Zero(dataType).Interface().(T)
// 	}

// 	return sideEffectRegisterType[T, Y](dataType, defaultInstance)
// }

// func sideEffectRegisterType[T any, Y Identifier](dataType reflect.Type, instance T) SideEffectRegister {
// 	return func() (*SideEffect, error) {
// 		handlerValue := reflect.ValueOf(instance)
// 		runMethod := handlerValue.MethodByName("Run")
// 		if !runMethod.IsValid() {
// 			return nil, fmt.Errorf("%s must have a Run method", dataType)
// 		}

// 		runMethodType := runMethod.Type()

// 		if runMethodType.NumIn() < 1 {
// 			return nil, fmt.Errorf("Run method of %s must have at least one input parameter (SideEffectContext)", dataType)
// 		}

// 		if runMethodType.In(0) != reflect.TypeOf(SideEffectContext[Y]{}) {
// 			return nil, fmt.Errorf("first parameter of Run method in %s must be SideEffectContext", dataType)
// 		}

// 		// Collect all parameter types after the context parameter
// 		paramTypes := []reflect.Type{}
// 		paramsKinds := []reflect.Kind{}
// 		for i := 1; i < runMethodType.NumIn(); i++ {
// 			paramTypes = append(paramTypes, runMethodType.In(i))
// 			paramsKinds = append(paramsKinds, runMethodType.In(i).Kind())
// 		}

// 		// Collect all return types except the last one if it's an error
// 		numOut := runMethodType.NumOut()
// 		if numOut == 0 {
// 			return nil, fmt.Errorf("Run method in %s must return at least an error", dataType)
// 		}

// 		returnTypes := []reflect.Type{}
// 		returnKinds := []reflect.Kind{}
// 		for i := 0; i < numOut-1; i++ {
// 			returnTypes = append(returnTypes, runMethodType.Out(i))
// 			returnKinds = append(returnKinds, runMethodType.Out(i).Kind())
// 		}

// 		// Verify that the last return type is error
// 		if runMethodType.Out(numOut-1) != reflect.TypeOf((*error)(nil)).Elem() {
// 			return nil, fmt.Errorf("last return value of Run method in %s must be error", dataType)
// 		}

// 		se := &SideEffect{
// 			HandlerName:     dataType.Name(),
// 			HandlerLongName: HandlerIdentity(fmt.Sprintf("%s/%s", dataType.PkgPath(), dataType.Name())),
// 			Handler:         runMethod.Interface(), // Set Handler to the Run method
// 			ParamTypes:      paramTypes,
// 			ParamsKinds:     paramsKinds,
// 			ReturnTypes:     returnTypes,
// 			ReturnKinds:     returnKinds,
// 			NumIn:           runMethodType.NumIn() - 1, // Exclude context
// 			NumOut:          numOut - 1,                // Exclude error
// 		}

// 		log.Printf("Registering side effect %s with name %s", dataType.Name(), se.HandlerLongName)

// 		return se, nil
// 	}
// }

// type SagaTransaction HandlerInfo
// type SagaCompensation HandlerInfo

// type SagaRegister func() (*SagaTransaction, *SagaCompensation, error)

// func AsSaga[T any, Y Identifier]() SagaRegister {
// 	dataType := reflect.TypeOf((*T)(nil)).Elem() // Get the type of T as a non-pointer
// 	return sagaRegisterType[Y](dataType)
// }

// func sagaRegisterType[Y Identifier](dataType reflect.Type) SagaRegister {
// 	return func() (*SagaTransaction, *SagaCompensation, error) {
// 		handlerType := dataType

// 		if handlerType.Kind() != reflect.Struct {
// 			return nil, nil, fmt.Errorf("%s must be a struct", handlerType)
// 		}

// 		// Check for Transaction method
// 		transactionMethod, txExists := handlerType.MethodByName("Transaction")
// 		if !txExists {
// 			return nil, nil, fmt.Errorf("%s must have a Transaction method", handlerType)
// 		}

// 		// Validate the Transaction method signature
// 		if transactionMethod.Type.NumIn() < 2 || transactionMethod.Type.In(1) != reflect.TypeOf(TransactionContext[Y]{}) {
// 			return nil, nil, fmt.Errorf("Transaction method in %s must accept (TransactionContext) as its second parameter", handlerType)
// 		}

// 		// Collect parameter types after context
// 		txParamTypes := []reflect.Type{}
// 		txParamsKinds := []reflect.Kind{}
// 		for i := 2; i < transactionMethod.Type.NumIn(); i++ {
// 			txParamTypes = append(txParamTypes, transactionMethod.Type.In(i))
// 			txParamsKinds = append(txParamsKinds, transactionMethod.Type.In(i).Kind())
// 		}

// 		// Collect return types excluding error
// 		txNumOut := transactionMethod.Type.NumOut()
// 		if txNumOut == 0 {
// 			return nil, nil, fmt.Errorf("Transaction method in %s must return at least an error", handlerType)
// 		}

// 		txReturnTypes := []reflect.Type{}
// 		txReturnKinds := []reflect.Kind{}
// 		for i := 0; i < txNumOut-1; i++ {
// 			txReturnTypes = append(txReturnTypes, transactionMethod.Type.Out(i))
// 			txReturnKinds = append(txReturnKinds, transactionMethod.Type.Out(i).Kind())
// 		}

// 		// Verify that the last return type is error
// 		if transactionMethod.Type.Out(txNumOut-1) != reflect.TypeOf((*error)(nil)).Elem() {
// 			return nil, nil, fmt.Errorf("last return value of Transaction method in %s must be error", handlerType)
// 		}

// 		// Create the SagaTransaction object
// 		sagaTransaction := &SagaTransaction{
// 			HandlerName:     handlerType.Name() + "Transaction",
// 			HandlerLongName: HandlerIdentity(fmt.Sprintf("%s/%s/Transaction", handlerType.PkgPath(), handlerType.Name())),
// 			Handler:         handlerType, // Set to handlerType
// 			ParamTypes:      txParamTypes,
// 			ParamsKinds:     txParamsKinds,
// 			ReturnTypes:     txReturnTypes,
// 			ReturnKinds:     txReturnKinds,
// 			NumIn:           transactionMethod.Type.NumIn() - 2, // Exclude receiver and context
// 			NumOut:          txNumOut - 1,                       // Exclude error
// 		}

// 		// Check for Compensation method
// 		compensationMethod, compExists := handlerType.MethodByName("Compensation")
// 		if !compExists {
// 			return nil, nil, fmt.Errorf("%s must have a Compensation method", handlerType)
// 		}

// 		// Validate the Compensation method signature
// 		if compensationMethod.Type.NumIn() < 2 || compensationMethod.Type.In(1) != reflect.TypeOf(CompensationContext[Y]{}) {
// 			return nil, nil, fmt.Errorf("Compensation method in %s must accept (CompensationContext) as its second parameter", handlerType)
// 		}

// 		// Collect parameter types after context
// 		compParamTypes := []reflect.Type{}
// 		compParamsKinds := []reflect.Kind{}
// 		for i := 2; i < compensationMethod.Type.NumIn(); i++ {
// 			compParamTypes = append(compParamTypes, compensationMethod.Type.In(i))
// 			compParamsKinds = append(compParamsKinds, compensationMethod.Type.In(i).Kind())
// 		}

// 		// Collect return types excluding error
// 		compNumOut := compensationMethod.Type.NumOut()
// 		if compNumOut == 0 {
// 			return nil, nil, fmt.Errorf("Compensation method in %s must return at least an error", handlerType)
// 		}

// 		compReturnTypes := []reflect.Type{}
// 		compReturnKinds := []reflect.Kind{}
// 		for i := 0; i < compNumOut-1; i++ {
// 			compReturnTypes = append(compReturnTypes, compensationMethod.Type.Out(i))
// 			compReturnKinds = append(compReturnKinds, compensationMethod.Type.Out(i).Kind())
// 		}

// 		// Verify that the last return type is error
// 		if compensationMethod.Type.Out(compNumOut-1) != reflect.TypeOf((*error)(nil)).Elem() {
// 			return nil, nil, fmt.Errorf("last return value of Compensation method in %s must be error", handlerType)
// 		}

// 		// Create the SagaCompensation object
// 		sagaCompensation := &SagaCompensation{
// 			HandlerName:     handlerType.Name() + "Compensation",
// 			HandlerLongName: HandlerIdentity(fmt.Sprintf("%s/%s/Compensation", handlerType.PkgPath(), handlerType.Name())),
// 			Handler:         handlerType, // Set to handlerType
// 			ParamTypes:      compParamTypes,
// 			ParamsKinds:     compParamsKinds,
// 			ReturnTypes:     compReturnTypes,
// 			ReturnKinds:     compReturnKinds,
// 			NumIn:           compensationMethod.Type.NumIn() - 2, // Exclude receiver and context
// 			NumOut:          compNumOut - 1,                      // Exclude error
// 		}

// 		log.Printf("Registering saga %s with transaction and compensation methods", handlerType.Name())

// 		return sagaTransaction, sagaCompensation, nil
// 	}
// }

// type SagaActivity struct {
// 	dataType reflect.Type
// 	builder  interface{}
// }

// func AsSagaActivity[T any, Y Identifier](builder SagaActivityBuilder[T, Y]) *SagaActivity {
// 	var dataType reflect.Type

// 	// Get the type of T and ensure that we get the underlying type if T is a pointer.
// 	t := reflect.TypeOf((*T)(nil)).Elem()

// 	if t.Kind() == reflect.Ptr {
// 		dataType = t.Elem()
// 	} else {
// 		dataType = t
// 	}

// 	return &SagaActivity{
// 		dataType: dataType,
// 		builder:  builder,
// 	}
// }

// func (tp *Tempolite[T]) RegisterSagaActivity(sagaActivity *SagaActivity) error {
// 	// Store the builder function in the sagaBuilders map, using the dataType as the key
// 	tp.sagaBuilders.Store(sagaActivity.dataType, sagaActivity.builder)
// 	log.Printf("Registered saga activity for type %v", sagaActivity.dataType)
// 	return nil
// }

func (tp *Tempolite[T]) RegisterWorkflow(workflowFunc interface{}) error {
	handlerType := reflect.TypeOf(workflowFunc)

	if handlerType.Kind() != reflect.Func {
		return fmt.Errorf("workflow must be a function")
	}

	if handlerType.NumIn() < 1 {
		return fmt.Errorf("workflow function must have at least one input parameter (WorkflowContext)")
	}

	if handlerType.In(0) != reflect.TypeOf(WorkflowContext[T]{}) {
		return fmt.Errorf("first parameter of workflow function must be WorkflowContext")
	}

	// Collect parameter types after the context parameter
	paramsKinds := []reflect.Kind{}
	paramTypes := []reflect.Type{}
	for i := 1; i < handlerType.NumIn(); i++ {
		paramTypes = append(paramTypes, handlerType.In(i))
		paramsKinds = append(paramsKinds, handlerType.In(i).Kind())
	}

	// Collect return types excluding error
	numOut := handlerType.NumOut()
	if numOut == 0 {
		return fmt.Errorf("workflow function must return at least an error")
	}

	returnKinds := []reflect.Kind{}
	returnTypes := []reflect.Type{}
	for i := 0; i < numOut-1; i++ {
		returnTypes = append(returnTypes, handlerType.Out(i))
		returnKinds = append(returnKinds, handlerType.Out(i).Kind())
	}

	// Verify that the last return type is error
	if handlerType.Out(numOut-1) != reflect.TypeOf((*error)(nil)).Elem() {
		return fmt.Errorf("last return value of workflow function must be error")
	}

	// Generate a unique identifier for the workflow function
	funcName := runtime.FuncForPC(reflect.ValueOf(workflowFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)

	workflow := &Workflow{
		HandlerName:     funcName,
		HandlerLongName: handlerIdentity,
		Handler:         workflowFunc,
		ParamTypes:      paramTypes,
		ParamsKinds:     paramsKinds,
		ReturnTypes:     returnTypes,
		ReturnKinds:     returnKinds,
		NumIn:           handlerType.NumIn() - 1, // Exclude context
		NumOut:          numOut - 1,              // Exclude error
	}

	log.Printf("Registering workflow %s with name %s", funcName, handlerIdentity)

	tp.workflows.Store(workflow.HandlerLongName, *workflow)
	return nil
}

func (tp *Tempolite[T]) RegisterActivityFunc(activityFunc interface{}) error {
	handlerType := reflect.TypeOf(activityFunc)

	if handlerType.Kind() != reflect.Func {
		return fmt.Errorf("activity must be a function")
	}

	if handlerType.NumIn() < 1 {
		return fmt.Errorf("activity function must have at least one input parameter (ActivityContext)")
	}

	if handlerType.In(0) != reflect.TypeOf(ActivityContext[T]{}) {
		return fmt.Errorf("first parameter of activity function must be ActivityContext")
	}

	// Collect parameter types after the context parameter
	paramTypes := []reflect.Type{}
	paramsKinds := []reflect.Kind{}
	for i := 1; i < handlerType.NumIn(); i++ {
		paramTypes = append(paramTypes, handlerType.In(i))
		paramsKinds = append(paramsKinds, handlerType.In(i).Kind())
	}

	// Collect return types excluding error
	numOut := handlerType.NumOut()
	if numOut == 0 {
		return fmt.Errorf("activity function must return at least an error")
	}

	returnTypes := []reflect.Type{}
	returnKinds := []reflect.Kind{}
	for i := 0; i < numOut-1; i++ {
		returnTypes = append(returnTypes, handlerType.Out(i))
		returnKinds = append(returnKinds, handlerType.Out(i).Kind())
	}

	// Verify that the last return type is error
	if handlerType.Out(numOut-1) != reflect.TypeOf((*error)(nil)).Elem() {
		return fmt.Errorf("last return value of activity function must be error")
	}

	// Generate a unique identifier for the activity function
	funcName := runtime.FuncForPC(reflect.ValueOf(activityFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)

	activity := &Activity{
		HandlerName:     funcName,
		HandlerLongName: handlerIdentity,
		Handler:         activityFunc,
		ParamTypes:      paramTypes,
		ParamsKinds:     paramsKinds,
		ReturnTypes:     returnTypes,
		ReturnKinds:     returnKinds,
		NumIn:           handlerType.NumIn() - 1, // Exclude context
		NumOut:          numOut - 1,              // Exclude error
	}

	log.Printf("Registering activity function %s with name %s", funcName, handlerIdentity)

	tp.activities.Store(activity.HandlerLongName, *activity)
	return nil
}

func (tp *Tempolite[T]) IsActivityRegistered(longName HandlerIdentity) bool {
	_, ok := tp.activities.Load(longName)
	return ok
}

func (tp *Tempolite[T]) RegisterActivity(register ActivityRegister) error {
	activity, err := register()
	if err != nil {
		return err
	}
	tp.activities.Store(activity.HandlerLongName, *activity)
	return nil
}

// func (tp *Tempolite[T]) IsSagaRegistered(longName HandlerIdentity) bool {
// 	_, ok := tp.sagas.Load(longName)
// 	return ok
// }

// func (tp *Tempolite[T]) RegisterSaga(register SagaRegister) error {
// 	transaction, compensation, err := register()
// 	if err != nil {
// 		return err
// 	}

// 	// Store the transaction and compensation separately, or together as needed
// 	tp.sagas.Store(transaction.HandlerLongName, *transaction)
// 	tp.sagas.Store(compensation.HandlerLongName, *compensation)

// 	log.Printf("Successfully registered saga with transaction %s and compensation %s", transaction.HandlerName, compensation.HandlerName)
// 	return nil
// }

// func (tp *Tempolite[T]) RegisterSideEffectFunc(sideEffectFunc interface{}) error {
// 	handlerType := reflect.TypeOf(sideEffectFunc)

// 	if handlerType.Kind() != reflect.Func {
// 		return fmt.Errorf("side effect must be a function")
// 	}

// 	if handlerType.NumIn() < 1 {
// 		return fmt.Errorf("side effect function must have at least one input parameter (SideEffectContext)")
// 	}

// 	if handlerType.In(0) != reflect.TypeOf(SideEffectContext[T]{}) {
// 		return fmt.Errorf("first parameter of side effect function must be SideEffectContext")
// 	}

// 	// Collect parameter types after the context parameter
// 	paramTypes := []reflect.Type{}
// 	paramKinds := []reflect.Kind{}
// 	for i := 1; i < handlerType.NumIn(); i++ {
// 		paramTypes = append(paramTypes, handlerType.In(i))
// 		paramKinds = append(paramKinds, handlerType.In(i).Kind())
// 	}

// 	// Collect return types excluding error
// 	numOut := handlerType.NumOut()
// 	if numOut == 0 {
// 		return fmt.Errorf("side effect function must return at least an error")
// 	}

// 	returnTypes := []reflect.Type{}
// 	returnKinds := []reflect.Kind{}
// 	for i := 0; i < numOut-1; i++ {
// 		returnTypes = append(returnTypes, handlerType.Out(i))
// 		returnKinds = append(returnKinds, handlerType.Out(i).Kind())
// 	}

// 	// Verify that the last return type is error
// 	if handlerType.Out(numOut-1) != reflect.TypeOf((*error)(nil)).Elem() {
// 		return fmt.Errorf("last return value of side effect function must be error")
// 	}

// 	// Generate a unique identifier for the side effect function
// 	funcName := runtime.FuncForPC(reflect.ValueOf(sideEffectFunc).Pointer()).Name()
// 	handlerIdentity := HandlerIdentity(funcName)

// 	sideEffect := &SideEffect{
// 		HandlerName:     funcName,
// 		HandlerLongName: handlerIdentity,
// 		Handler:         sideEffectFunc,
// 		ParamTypes:      paramTypes,
// 		ParamsKinds:     paramKinds,
// 		ReturnTypes:     returnTypes,
// 		ReturnKinds:     returnKinds,
// 		NumIn:           handlerType.NumIn() - 1, // Exclude context
// 		NumOut:          numOut - 1,              // Exclude error
// 	}

// 	log.Printf("Registering side effect function %s with name %s", funcName, handlerIdentity)

// 	tp.sideEffects.Store(sideEffect.HandlerLongName, *sideEffect)
// 	return nil
// }

// func (tp *Tempolite[T]) IsSideEffectRegistered(longName HandlerIdentity) bool {
// 	_, ok := tp.sideEffects.Load(longName)
// 	return ok
// }

// func (tp *Tempolite[T]) RegisterSideEffect(register SideEffectRegister) error {
// 	sideEffect, err := register()
// 	if err != nil {
// 		return err
// 	}
// 	tp.sideEffects.Store(sideEffect.HandlerLongName, *sideEffect)
// 	return nil
// }
