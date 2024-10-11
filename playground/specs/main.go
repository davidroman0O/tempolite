package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"reflect"
	"runtime"
	"sync"

	"github.com/davidroman0O/retrypool"
)

type WorkflowInfo struct{}

func (i *WorkflowInfo) Get(ctx TempoliteContext) ([]interface{}, error) {
	return nil, nil
}

func (i *WorkflowInfo) Cancel() error {
	// todo: implement
	return nil
}

func (i *WorkflowInfo) Pause() error {
	// todo: implement
	return nil
}

func (i *WorkflowInfo) Resume() error {
	// todo: implement
	return nil
}

type ActivityInfo struct{}

func (i *ActivityInfo) Get(ctx TempoliteContext) ([]interface{}, error) {
	// todo: implement
	return nil, nil
}

type SideEffectInfo struct{}

func (i *SideEffectInfo) Get(ctx TempoliteContext) ([]interface{}, error) {
	// todo: implement
	return nil, nil
}

type SagaInfo struct{}

func (i *SagaInfo) Get(ctx TempoliteContext) ([]interface{}, error) {
	// todo: implement
	return nil, nil
}

type TempoliteContext interface {
	context.Context
}

type WorkflowContext struct {
	TempoliteContext
	tp *Tempolite
}

// Since I don't want to hide any implementation, when the WorkflowInfo call Pause/Resume, the moment the Yield() is called, the workflow will be paused or resume if called Resume.
func (w WorkflowContext) Yield() error {
	// todo: implement - could use ConsumerSignalChannel and ProducerSignalChannel behind the scene
	return nil
}

func (w WorkflowContext) ContinueAsNew(ctx WorkflowContext, values ...any) error {
	// todo: implement
	return nil
}

func (w WorkflowContext) GetWorkflow(id string) (*WorkflowInfo, error) {
	// todo: implement
	return nil, nil
}

func (w WorkflowContext) ExecuteWorkflow(name HandlerIdentity, inputs ...any) (*WorkflowInfo, error) {
	// todo: implement
	return nil, nil
}

func (w WorkflowContext) ExecuteActivity(name HandlerIdentity, inputs ...any) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

func (w WorkflowContext) GetActivity(id string) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

func (w WorkflowContext) ExecuteSideEffect(name HandlerIdentity, inputs ...any) (*SideEffectInfo, error) {
	// todo: implement
	return nil, nil
}

type ActivityContext struct {
	TempoliteContext
	tp *Tempolite
}

func (w ActivityContext) ExecuteSideEffect(name HandlerIdentity, inputs ...any) (*SideEffectInfo, error) {
	// todo: implement
	return nil, nil
}

func (w ActivityContext) ExecuteSaga(name HandlerIdentity, inputs interface{}) (*SagaInfo, error) {
	// todo: implement - should build the saga, then step the instaciated steps, then use reflect to get the methods, assign it to a callback on the TransactionTask or CompensationTask, so what's necessary for the Next/Compensate callbacks, generate a bunch of callbacks based on the rules of the Saga Pattern discuess in the documentation, and just dispatch the first TransactionContext
	return nil, nil
}

func (w ActivityContext) ExecuteActivity(name HandlerIdentity, inputs ...any) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

func (w ActivityContext) GetActivity(id string) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

type SideEffectContext struct {
	TempoliteContext
	tp *Tempolite
}

func (w SideEffectContext) ExecuteActivity(name HandlerIdentity, inputs ...any) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

func (w SideEffectContext) ExecuteSideEffect(name HandlerIdentity, inputs ...any) (*SideEffectInfo, error) {
	// todo: implement
	return nil, nil
}

func (w SideEffectContext) GetActivity(id string) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

type TransactionContext struct {
	TempoliteContext
	tp *Tempolite
}

func (w TransactionContext) GetActivity(id string) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

func (w TransactionContext) ExecuteActivity(name HandlerIdentity, inputs ...any) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

func (w TransactionContext) ExecuteSideEffect(name HandlerIdentity, inputs ...any) (*SideEffectInfo, error) {
	// todo: implement
	return nil, nil
}

type CompensationContext struct {
	TempoliteContext
	tp *Tempolite
}

func (w CompensationContext) GetActivity(id string) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

func (w CompensationContext) ExecuteActivity(name HandlerIdentity, inputs ...any) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

// TempoliteContext contains the information from where it was called, so we know the XXXInfo to which it belongs
// Saga only accepts one type of input
func (tp *Tempolite) enqueueSaga(ctx TempoliteContext, input interface{}) (*SagaInfo, error) {
	// todo: implement
	return nil, nil
}

func (tp *Tempolite) enqueueActivty(ctx TempoliteContext, input ...interface{}) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

func (tp *Tempolite) enqueueSideEffect(ctx TempoliteContext, input ...interface{}) (*SideEffectInfo, error) {
	// todo: implement
	return nil, nil
}

func (tp *Tempolite) getSaga(id string) (*SagaInfo, error) {
	// todo: implement
	return nil, nil
}

func (tp *Tempolite) getActivity(id string) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

func (tp *Tempolite) getWorkflow(id string) (*WorkflowInfo, error) {
	// todo: implement
	return nil, nil
}

func (tp *Tempolite) getSideEffect(id string) (*SideEffectInfo, error) {
	// todo: implement
	return nil, nil
}

type SagaStep interface {
	Transaction(ctx TransactionContext) (interface{}, error)
	Compensation(ctx CompensationContext) (interface{}, error)
}

type SagaDefinition struct {
	Steps []SagaStep
}

// Signals are notified on the database but store no data which is done in-memory
type SignalConsumerChannel[T any] struct {
	C chan T // when a signal is created or getted, the Producer and Consumer structs will share the same signal instance
}

type SignalProducerChannel[T any] struct {
	C chan T // when a signal is created or getted, the Producer and Consumer structs will share the same signal instance
}

func (s SignalConsumerChannel[T]) Close() {}

func ConsumeSignal[T any](ctx TempoliteContext, name string) (SignalConsumerChannel[T], error) {
	// use tempolite context to call functions from Tempolite to create/save (database) and return the signal channel
	// should use ctx to call a private function on *Tempolite to manage the signal
	return SignalConsumerChannel[T]{
		C: make(chan T, 1),
	}, nil
}

func ProduceSignal[T any](ctx TempoliteContext, name string) (SignalProducerChannel[T], error) {
	// use tempolite context to call functions from Tempolite to get (database) and return the signal channel
	// should use ctx to call a private function on *Tempolite to manage the signal
	return SignalProducerChannel[T]{
		C: make(chan T, 1),
	}, nil
}

type HandlerInfo struct {
	HandlerName     string
	HandlerLongName HandlerIdentity
	Handler         interface{}
	ParamTypes      []reflect.Type
	ReturnTypes     []reflect.Type // Changed from ReturnType to ReturnTypes slice
	NumIn           int
	NumOut          int
}

func (hi HandlerInfo) ToInterface(data []byte) ([]interface{}, error) {
	var paramData []json.RawMessage
	if err := json.Unmarshal(data, &paramData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %v", err)
	}

	if len(paramData) != len(hi.ParamTypes) {
		return nil, fmt.Errorf("parameter count mismatch: expected %d, got %d", len(hi.ParamTypes), len(paramData))
	}

	params := make([]interface{}, len(hi.ParamTypes))
	for i, paramType := range hi.ParamTypes {
		paramPtr := reflect.New(paramType)
		if err := json.Unmarshal(paramData[i], paramPtr.Interface()); err != nil {
			return nil, fmt.Errorf("failed to unmarshal parameter %d: %v", i, err)
		}
		params[i] = paramPtr.Elem().Interface()
	}

	return params, nil
}

type TaskUnit struct{}

type Tempolite struct {
	workflows   sync.Map
	activities  sync.Map
	sideEffects sync.Map
	sagas       sync.Map

	sagaBuilders sync.Map

	pool retrypool.Pool[*TaskUnit]
}

func (tp *Tempolite) RegisterWorkflow(workflowFunc interface{}) error {
	handlerType := reflect.TypeOf(workflowFunc)

	if handlerType.Kind() != reflect.Func {
		return fmt.Errorf("workflow must be a function")
	}

	if handlerType.NumIn() < 1 {
		return fmt.Errorf("workflow function must have at least one input parameter (WorkflowContext)")
	}

	if handlerType.In(0) != reflect.TypeOf(WorkflowContext{}) {
		return fmt.Errorf("first parameter of workflow function must be WorkflowContext")
	}

	// Collect parameter types after the context parameter
	paramTypes := []reflect.Type{}
	for i := 1; i < handlerType.NumIn(); i++ {
		paramTypes = append(paramTypes, handlerType.In(i))
	}

	// Collect return types excluding error
	numOut := handlerType.NumOut()
	if numOut == 0 {
		return fmt.Errorf("workflow function must return at least an error")
	}

	returnTypes := []reflect.Type{}
	for i := 0; i < numOut-1; i++ {
		returnTypes = append(returnTypes, handlerType.Out(i))
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
		ReturnTypes:     returnTypes,
		NumIn:           handlerType.NumIn() - 1, // Exclude context
		NumOut:          numOut - 1,              // Exclude error
	}

	log.Printf("Registering workflow %s with name %s", funcName, handlerIdentity)

	tp.workflows.Store(workflow.HandlerLongName, *workflow)
	return nil
}

func (tp *Tempolite) RegisterActivityFunc(activityFunc interface{}) error {
	handlerType := reflect.TypeOf(activityFunc)

	if handlerType.Kind() != reflect.Func {
		return fmt.Errorf("activity must be a function")
	}

	if handlerType.NumIn() < 1 {
		return fmt.Errorf("activity function must have at least one input parameter (ActivityContext)")
	}

	if handlerType.In(0) != reflect.TypeOf(ActivityContext{}) {
		return fmt.Errorf("first parameter of activity function must be ActivityContext")
	}

	// Collect parameter types after the context parameter
	paramTypes := []reflect.Type{}
	for i := 1; i < handlerType.NumIn(); i++ {
		paramTypes = append(paramTypes, handlerType.In(i))
	}

	// Collect return types excluding error
	numOut := handlerType.NumOut()
	if numOut == 0 {
		return fmt.Errorf("activity function must return at least an error")
	}

	returnTypes := []reflect.Type{}
	for i := 0; i < numOut-1; i++ {
		returnTypes = append(returnTypes, handlerType.Out(i))
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
		ReturnTypes:     returnTypes,
		NumIn:           handlerType.NumIn() - 1, // Exclude context
		NumOut:          numOut - 1,              // Exclude error
	}

	log.Printf("Registering activity function %s with name %s", funcName, handlerIdentity)

	tp.activities.Store(activity.HandlerLongName, *activity)
	return nil
}

func (tp *Tempolite) RegisterSideEffectFunc(sideEffectFunc interface{}) error {
	handlerType := reflect.TypeOf(sideEffectFunc)

	if handlerType.Kind() != reflect.Func {
		return fmt.Errorf("side effect must be a function")
	}

	if handlerType.NumIn() < 1 {
		return fmt.Errorf("side effect function must have at least one input parameter (SideEffectContext)")
	}

	if handlerType.In(0) != reflect.TypeOf(SideEffectContext{}) {
		return fmt.Errorf("first parameter of side effect function must be SideEffectContext")
	}

	// Collect parameter types after the context parameter
	paramTypes := []reflect.Type{}
	for i := 1; i < handlerType.NumIn(); i++ {
		paramTypes = append(paramTypes, handlerType.In(i))
	}

	// Collect return types excluding error
	numOut := handlerType.NumOut()
	if numOut == 0 {
		return fmt.Errorf("side effect function must return at least an error")
	}

	returnTypes := []reflect.Type{}
	for i := 0; i < numOut-1; i++ {
		returnTypes = append(returnTypes, handlerType.Out(i))
	}

	// Verify that the last return type is error
	if handlerType.Out(numOut-1) != reflect.TypeOf((*error)(nil)).Elem() {
		return fmt.Errorf("last return value of side effect function must be error")
	}

	// Generate a unique identifier for the side effect function
	funcName := runtime.FuncForPC(reflect.ValueOf(sideEffectFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)

	sideEffect := &SideEffect{
		HandlerName:     funcName,
		HandlerLongName: handlerIdentity,
		Handler:         sideEffectFunc,
		ParamTypes:      paramTypes,
		ReturnTypes:     returnTypes,
		NumIn:           handlerType.NumIn() - 1, // Exclude context
		NumOut:          numOut - 1,              // Exclude error
	}

	log.Printf("Registering side effect function %s with name %s", funcName, handlerIdentity)

	tp.sideEffects.Store(sideEffect.HandlerLongName, *sideEffect)
	return nil
}

func (tp *Tempolite) IsActivityRegistered(longName HandlerIdentity) bool {
	_, ok := tp.activities.Load(longName)
	return ok
}

func (tp *Tempolite) RegisterActivity(register ActivityRegister) error {
	activity, err := register()
	if err != nil {
		return err
	}
	tp.activities.Store(activity.HandlerLongName, *activity)
	return nil
}

func (tp *Tempolite) IsSideEffectRegistered(longName HandlerIdentity) bool {
	_, ok := tp.sideEffects.Load(longName)
	return ok
}

func (tp *Tempolite) RegisterSideEffect(register SideEffectRegister) error {
	sideEffect, err := register()
	if err != nil {
		return err
	}
	tp.sideEffects.Store(sideEffect.HandlerLongName, *sideEffect)
	return nil
}

func (tp *Tempolite) IsSagaRegistered(longName HandlerIdentity) bool {
	_, ok := tp.sagas.Load(longName)
	return ok
}

func (tp *Tempolite) RegisterSaga(register SagaRegister) error {
	transaction, compensation, err := register()
	if err != nil {
		return err
	}

	// Store the transaction and compensation separately, or together as needed
	tp.sagas.Store(transaction.HandlerLongName, *transaction)
	tp.sagas.Store(compensation.HandlerLongName, *compensation)

	log.Printf("Successfully registered saga with transaction %s and compensation %s", transaction.HandlerName, compensation.HandlerName)
	return nil
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

type HandlerIdentity string

type Workflow HandlerInfo

type Activity HandlerInfo

type ActivityRegister func() (*Activity, error)

func AsActivity[T any](instance ...T) ActivityRegister {
	dataType := reflect.TypeOf((*T)(nil)).Elem()

	var defaultInstance T
	if len(instance) > 0 {
		defaultInstance = instance[0]
	} else {
		defaultInstance = reflect.Zero(dataType).Interface().(T)
	}

	return activityRegisterType(dataType, defaultInstance)
}

func activityRegisterType[T any](dataType reflect.Type, instance T) ActivityRegister {
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

		if runMethodType.In(0) != reflect.TypeOf(ActivityContext{}) {
			return nil, fmt.Errorf("first parameter of Run method in %s must be ActivityContext", dataType)
		}

		// Collect all parameter types after the context parameter
		paramTypes := []reflect.Type{}
		for i := 1; i < runMethodType.NumIn(); i++ {
			paramTypes = append(paramTypes, runMethodType.In(i))
		}

		// Collect all return types except the last one if it's an error
		numOut := runMethodType.NumOut()
		if numOut == 0 {
			return nil, fmt.Errorf("Run method in %s must return at least an error", dataType)
		}

		returnTypes := []reflect.Type{}
		for i := 0; i < numOut-1; i++ {
			returnTypes = append(returnTypes, runMethodType.Out(i))
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
			ReturnTypes:     returnTypes,
			NumIn:           runMethodType.NumIn() - 1, // Exclude context
			NumOut:          numOut - 1,                // Exclude error
		}

		log.Printf("Registering activity %s with name %s", dataType.Name(), act.HandlerLongName)

		return act, nil
	}
}

type SideEffect HandlerInfo

type SideEffectRegister func() (*SideEffect, error)

func AsSideEffect[T any](instance ...T) SideEffectRegister {
	dataType := reflect.TypeOf((*T)(nil)).Elem()

	var defaultInstance T
	if len(instance) > 0 {
		defaultInstance = instance[0]
	} else {
		defaultInstance = reflect.Zero(dataType).Interface().(T)
	}

	return sideEffectRegisterType(dataType, defaultInstance)
}

func sideEffectRegisterType[T any](dataType reflect.Type, instance T) SideEffectRegister {
	return func() (*SideEffect, error) {
		handlerValue := reflect.ValueOf(instance)
		runMethod := handlerValue.MethodByName("Run")
		if !runMethod.IsValid() {
			return nil, fmt.Errorf("%s must have a Run method", dataType)
		}

		runMethodType := runMethod.Type()

		if runMethodType.NumIn() < 1 {
			return nil, fmt.Errorf("Run method of %s must have at least one input parameter (SideEffectContext)", dataType)
		}

		if runMethodType.In(0) != reflect.TypeOf(SideEffectContext{}) {
			return nil, fmt.Errorf("first parameter of Run method in %s must be SideEffectContext", dataType)
		}

		// Collect all parameter types after the context parameter
		paramTypes := []reflect.Type{}
		for i := 1; i < runMethodType.NumIn(); i++ {
			paramTypes = append(paramTypes, runMethodType.In(i))
		}

		// Collect all return types except the last one if it's an error
		numOut := runMethodType.NumOut()
		if numOut == 0 {
			return nil, fmt.Errorf("Run method in %s must return at least an error", dataType)
		}

		returnTypes := []reflect.Type{}
		for i := 0; i < numOut-1; i++ {
			returnTypes = append(returnTypes, runMethodType.Out(i))
		}

		// Verify that the last return type is error
		if runMethodType.Out(numOut-1) != reflect.TypeOf((*error)(nil)).Elem() {
			return nil, fmt.Errorf("last return value of Run method in %s must be error", dataType)
		}

		se := &SideEffect{
			HandlerName:     dataType.Name(),
			HandlerLongName: HandlerIdentity(fmt.Sprintf("%s/%s", dataType.PkgPath(), dataType.Name())),
			Handler:         runMethod.Interface(), // Set Handler to the Run method
			ParamTypes:      paramTypes,
			ReturnTypes:     returnTypes,
			NumIn:           runMethodType.NumIn() - 1, // Exclude context
			NumOut:          numOut - 1,                // Exclude error
		}

		log.Printf("Registering side effect %s with name %s", dataType.Name(), se.HandlerLongName)

		return se, nil
	}
}

type SagaTransaction HandlerInfo
type SagaCompensation HandlerInfo

type SagaRegister func() (*SagaTransaction, *SagaCompensation, error)

func AsSaga[T any]() SagaRegister {
	dataType := reflect.TypeOf((*T)(nil)).Elem() // Get the type of T as a non-pointer
	return sagaRegisterType(dataType)
}

func sagaRegisterType(dataType reflect.Type) SagaRegister {
	return func() (*SagaTransaction, *SagaCompensation, error) {
		handlerType := dataType

		if handlerType.Kind() != reflect.Struct {
			return nil, nil, fmt.Errorf("%s must be a struct", handlerType)
		}

		// Check for Transaction method
		transactionMethod, txExists := handlerType.MethodByName("Transaction")
		if !txExists {
			return nil, nil, fmt.Errorf("%s must have a Transaction method", handlerType)
		}

		// Validate the Transaction method signature
		if transactionMethod.Type.NumIn() < 2 || transactionMethod.Type.In(1) != reflect.TypeOf(TransactionContext{}) {
			return nil, nil, fmt.Errorf("Transaction method in %s must accept (TransactionContext) as its second parameter", handlerType)
		}

		// Collect parameter types after context
		txParamTypes := []reflect.Type{}
		for i := 2; i < transactionMethod.Type.NumIn(); i++ {
			txParamTypes = append(txParamTypes, transactionMethod.Type.In(i))
		}

		// Collect return types excluding error
		txNumOut := transactionMethod.Type.NumOut()
		if txNumOut == 0 {
			return nil, nil, fmt.Errorf("Transaction method in %s must return at least an error", handlerType)
		}

		txReturnTypes := []reflect.Type{}
		for i := 0; i < txNumOut-1; i++ {
			txReturnTypes = append(txReturnTypes, transactionMethod.Type.Out(i))
		}

		// Verify that the last return type is error
		if transactionMethod.Type.Out(txNumOut-1) != reflect.TypeOf((*error)(nil)).Elem() {
			return nil, nil, fmt.Errorf("last return value of Transaction method in %s must be error", handlerType)
		}

		// Create the SagaTransaction object
		sagaTransaction := &SagaTransaction{
			HandlerName:     handlerType.Name() + "Transaction",
			HandlerLongName: HandlerIdentity(fmt.Sprintf("%s/%s/Transaction", handlerType.PkgPath(), handlerType.Name())),
			Handler:         handlerType, // Set to handlerType
			ParamTypes:      txParamTypes,
			ReturnTypes:     txReturnTypes,
			NumIn:           transactionMethod.Type.NumIn() - 2, // Exclude receiver and context
			NumOut:          txNumOut - 1,                       // Exclude error
		}

		// Check for Compensation method
		compensationMethod, compExists := handlerType.MethodByName("Compensation")
		if !compExists {
			return nil, nil, fmt.Errorf("%s must have a Compensation method", handlerType)
		}

		// Validate the Compensation method signature
		if compensationMethod.Type.NumIn() < 2 || compensationMethod.Type.In(1) != reflect.TypeOf(CompensationContext{}) {
			return nil, nil, fmt.Errorf("Compensation method in %s must accept (CompensationContext) as its second parameter", handlerType)
		}

		// Collect parameter types after context
		compParamTypes := []reflect.Type{}
		for i := 2; i < compensationMethod.Type.NumIn(); i++ {
			compParamTypes = append(compParamTypes, compensationMethod.Type.In(i))
		}

		// Collect return types excluding error
		compNumOut := compensationMethod.Type.NumOut()
		if compNumOut == 0 {
			return nil, nil, fmt.Errorf("Compensation method in %s must return at least an error", handlerType)
		}

		compReturnTypes := []reflect.Type{}
		for i := 0; i < compNumOut-1; i++ {
			compReturnTypes = append(compReturnTypes, compensationMethod.Type.Out(i))
		}

		// Verify that the last return type is error
		if compensationMethod.Type.Out(compNumOut-1) != reflect.TypeOf((*error)(nil)).Elem() {
			return nil, nil, fmt.Errorf("last return value of Compensation method in %s must be error", handlerType)
		}

		// Create the SagaCompensation object
		sagaCompensation := &SagaCompensation{
			HandlerName:     handlerType.Name() + "Compensation",
			HandlerLongName: HandlerIdentity(fmt.Sprintf("%s/%s/Compensation", handlerType.PkgPath(), handlerType.Name())),
			Handler:         handlerType, // Set to handlerType
			ParamTypes:      compParamTypes,
			ReturnTypes:     compReturnTypes,
			NumIn:           compensationMethod.Type.NumIn() - 2, // Exclude receiver and context
			NumOut:          compNumOut - 1,                      // Exclude error
		}

		log.Printf("Registering saga %s with transaction and compensation methods", handlerType.Name())

		return sagaTransaction, sagaCompensation, nil
	}
}

type SagaActivity struct {
	dataType reflect.Type
	builder  interface{}
}

func AsSagaActivity[T any](builder SagaActivityBuilder[T]) *SagaActivity {
	var dataType reflect.Type

	// Get the type of T and ensure that we get the underlying type if T is a pointer.
	t := reflect.TypeOf((*T)(nil)).Elem()

	if t.Kind() == reflect.Ptr {
		dataType = t.Elem()
	} else {
		dataType = t
	}

	return &SagaActivity{
		dataType: dataType,
		builder:  builder,
	}
}

func (tp *Tempolite) RegisterSagaActivity(sagaActivity *SagaActivity) error {
	// Store the builder function in the sagaBuilders map, using the dataType as the key
	tp.sagaBuilders.Store(sagaActivity.dataType, sagaActivity.builder)
	log.Printf("Registered saga activity for type %v", sagaActivity.dataType)
	return nil
}

func main() {
	tp := Tempolite{}

	tp.RegisterWorkflow(Workflow1)

	// Registering the activity
	if err := tp.RegisterActivity(AsActivity[SimpleActivity](SimpleActivity{SpecialValue: "Helloworld"})); err != nil {
		log.Fatalf("failed to register activity: %v", err)
	}

	acty := &ActivityShared{counter: 0}

	if err := tp.RegisterActivityFunc(acty.Increment); err != nil {
		log.Fatalf("failed to register activity: %v", err)
	}

	if err := tp.RegisterActivityFunc(acty.Decrement); err != nil {
		log.Fatalf("failed to register activity: %v", err)
	}

	if err := tp.RegisterActivityFunc(acty.Add); err != nil {
		log.Fatalf("failed to register activity: %v", err)
	}

	// Registering the side effect
	if err := tp.RegisterSideEffect(AsSideEffect[LoggingSideEffect]()); err != nil {
		log.Fatalf("failed to register side effect: %v", err)
	}

	if err := tp.RegisterSideEffectFunc(getRandomNumber); err != nil {
		log.Fatalf("failed to register side effect: %v", err)
	}

	// Registering the saga
	if err := tp.RegisterSaga(AsSaga[OrderSaga]()); err != nil {
		log.Fatalf("failed to register saga: %v", err)
	}

	log.Println("All components registered successfully")

	// Register a SagaActivity for OrderPaymentTask
	// if relationship input type -> builder exists, then it will return an error
	if err := tp.RegisterSagaActivity(
		AsSagaActivity[OrderPaymentTask](
			func(input OrderPaymentTask, builder *SagaDefinitionBuilder) *SagaDefinition {
				// Add steps directly as instances, with error handling
				if _, err := builder.AddStep(OrderSaga{OrderID: input.OrderID}); err != nil {
					log.Fatalf("failed to add OrderSaga: %v", err)
				}

				if _, err := builder.AddStep(PaymentSaga{OrderID: input.OrderID}); err != nil {
					log.Fatalf("failed to add PaymentSaga: %v", err)
				}

				return builder.Build()
			},
		),
	); err != nil {
		// might already exists too
		log.Fatalf("failed to register saga activity: %v", err)
	}

	log.Println("All saga steps registered successfully")

	{
		/// this is just a example, it won't be in tempolite real code

		// Example input data
		task := OrderPaymentTask{OrderID: "12345"}

		// Build the saga using the registered builder
		saga, err := tp.BuildSagaFromInput(task)
		if err != nil {
			log.Fatalf("failed to build saga: %v", err)
		}

		log.Println("Saga built successfully", saga)
	}

	id, _ := tp.EnqueueActivityFunc(
		acty.Increment,
	)

	id, _ = tp.EnqueueActivityFunc(
		acty.Increment,
	)

	id, _ = tp.EnqueueActivityFunc(
		acty.Decrement,
	)

	id, _ = tp.EnqueueActivityFunc(
		acty.Add,
		12,
	)

	id, _ = tp.EnqueueActivity(
		As[SimpleActivity](),
		MessageTask{Message: "Hello, World!"},
	)

	go func() {
		chn := tp.ProduceSignal(id)
		chn <- "data"
		close(chn) // if closed, then the SignalConsumerChannel/SignalProducerChannel channel store on a sync.Map will be closed too
	}()

	_, _ = tp.EnqueueWorkflow(Workflow1, "hello", 42)
}

func (tp *Tempolite) ProduceSignal(id string) chan interface{} {
	// whatever happen here, we have to create a channel that will then send the data to the other channel used by the consumer ON the correct type!!!
	return make(chan interface{}, 1)
}

func As[T any]() HandlerIdentity {
	dataType := reflect.TypeOf((*T)(nil)).Elem()
	return HandlerIdentity(fmt.Sprintf("%s/%s", dataType.PkgPath(), dataType.Name()))
}

func (tp *Tempolite) EnqueueActivityFunc(activityFunc interface{}, params ...interface{}) (string, error) {
	funcName := runtime.FuncForPC(reflect.ValueOf(activityFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)

	return tp.EnqueueActivity(handlerIdentity, params...)
}

func (tp *Tempolite) EnqueueActivity(longName HandlerIdentity, params ...interface{}) (string, error) {
	// The whole implementation will be totally different, it's just for fun here
	// The real implementation will be with ActivityTask save on a db, then fetched from a scheduler to be sent to a pool.
	handlerInterface, ok := tp.activities.Load(longName)
	if !ok {
		return "", fmt.Errorf("activity %s not found", longName)
	}

	activity := handlerInterface.(Activity)

	if len(params) != activity.NumIn {
		return "", fmt.Errorf("parameter count mismatch: expected %d, got %d", activity.NumIn, len(params))
	}

	// Serialize parameters
	payloadBytes, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("failed to marshal task parameters: %v", err)
	}

	// Enqueue the activity task (implementation depends on your task queue system)
	taskID := "task-id" // Generate or obtain a unique task ID
	log.Printf("Enqueued activity %s with payload: %s", longName, string(payloadBytes))

	// values := []reflect.Value{
	// 	reflect.ValueOf(ActivityContext{}),
	// }
	// for _, param := range params {
	// 	values = append(values, reflect.ValueOf(param))
	// }

	// reflect.ValueOf(activity.Handler).Call(values)

	return taskID, nil
}

func (tp *Tempolite) EnqueueWorkflow(workfow interface{}, params ...interface{}) (string, error) {
	return "", nil
}

func (tp *Tempolite) RemoveWorkflow(id string) (string, error) {
	return "", nil
}

func (tp *Tempolite) PauseWorkflow(id string) (string, error) {
	return "", nil
}

func (tp *Tempolite) ResumeWorkflow(id string) (string, error) {
	return "", nil
}

func (tp *Tempolite) GetWorkflow(id string) (*WorkflowInfo, error) {
	return tp.getWorkflow(id)
}

func (tp *Tempolite) GetActivity(id string) (*ActivityInfo, error) {
	return tp.getActivity(id)
}

func (tp *Tempolite) GetSideEffect(id string) (*SideEffectInfo, error) {
	return tp.getSideEffect(id)
}

func (tp *Tempolite) GetSaga(id string) (*SagaInfo, error) {
	return tp.getSaga(id)
}

func Workflow1(ctx WorkflowContext, input1 string, input2 int) (int, int, error) {
	fmt.Println("hello ", input1)

	// we're going to switch it
	seinfo, err := ctx.SideEffect(func(ctx SideEffect) interface{} {
		return getRandomNumber(5, 10)
	}, WithName("randomize"))
	if err != nil {
		return 0, 0, err
	}

	var value int
	payload, err := seinfo.Get(&value)

	activityInfo, err := ctx.ExecuteActivity(As[SimpleActivity](), MessageTask{Message: "Hello, World!"})
	if err != nil {
		return 0, 0, err
	}

	// If you wanted to call Pause, it will stop here, and when Resume is called, it will continue from here
	if err := ctx.Yield(); err != nil {
		return 0, 0, err
	}

	payload, err := activityInfo.Get(ctx)
	if err != nil {
		return 0, 0, err
	}

	if value, ok := payload[0].(int); ok {
		fmt.Println("Activity1 returned", value)
	}

	return 1, 2, nil
}

type OrderPaymentTask struct {
	OrderID string
}

type MessageTask struct {
	Message string
}

type SimpleActivity struct {
	SpecialValue string
}

func (h SimpleActivity) Run(ctx ActivityContext, task MessageTask) (int, string, error) {

	signal, err := ConsumeSignal[int](ctx, "waiting-number")
	if err != nil {
		return 0, "", err
	}

	<-signal.C

	return 420, "cool", nil
}

type ActivityShared struct {
	counter int
}

func (h *ActivityShared) Increment(ctx ActivityContext) (int, error) {
	defer fmt.Println("counter is ", h.counter)
	h.counter++
	return h.counter, nil
}

func (h *ActivityShared) Decrement(ctx ActivityContext) (int, error) {
	defer fmt.Println("counter is ", h.counter)
	h.counter--
	return h.counter, nil
}

func (h *ActivityShared) Add(ctx ActivityContext, value int) (int, error) {
	defer fmt.Println("counter is ", h.counter)
	h.counter += value
	return h.counter, nil
}

func (h *ActivityShared) Random(ctx ActivityContext, a int, b int) (int, error) {
	defer fmt.Println("counter is ", h.counter)

	info, err := ctx.ExecuteSideEffect("randomize", a, b)
	if err != nil {
		return 0, err
	}

	payload, err := info.Get(ctx)
	if err != nil {
		return 0, err
	}

	h.counter = payload[0].(int)

	return h.counter, nil
}

func getRandomNumber(ctx SideEffectContext, a int, b int) (int, error) {
	if a >= b {
		return 0, fmt.Errorf("invalid range: a must be less than b")
	}

	return rand.IntN(b-a+1) + a, nil
}

type LoggingSideEffect struct{}

func (l LoggingSideEffect) Run(ctx SideEffectContext, newInput string) (interface{}, error) {
	log.Println("Executing side effect", newInput)
	// Perform some side effect logic here, like logging or emitting an event.
	return "SideEffectResult", nil
}

type OrderSaga struct {
	OrderID string
}

func (o OrderSaga) Transaction(ctx TransactionContext) (interface{}, error) {
	log.Println("Starting transaction for OrderSaga")
	// Perform the main transaction logic here, like placing an order.
	orderID := "12345" // Example order ID
	return orderID, nil
}

func (o OrderSaga) Compensation(ctx CompensationContext) (interface{}, error) {
	log.Println("Compensating for OrderSaga")
	// Perform compensation logic here, like rolling back the order.
	return "OrderCompensated", nil
}

type PaymentSaga struct {
	OrderID string
}

func (p PaymentSaga) Transaction(ctx TransactionContext) (interface{}, error) {
	log.Println("Starting transaction for PaymentSaga")
	paymentID := "67890"
	return paymentID, nil
}

func (p PaymentSaga) Compensation(ctx CompensationContext) (interface{}, error) {
	log.Println("Compensating for PaymentSaga")
	return "PaymentCompensated", nil
}

// This function is simple an example, IT SHOULD NOW BE LIKE THAT IN TEMPORAL
func (tp *Tempolite) BuildSagaFromInput(input interface{}) (*SagaDefinition, error) {
	// Retrieve the type of the input
	dataType := reflect.TypeOf(input)

	// Check if a builder function is registered for this input type
	builderInterface, ok := tp.sagaBuilders.Load(dataType)
	if !ok {
		return nil, fmt.Errorf("no saga activity builder registered for input type %v", dataType)
	}

	// Use reflection to call the builder function
	builderValue := reflect.ValueOf(builderInterface)

	// Ensure the builder is a function
	if builderValue.Kind() != reflect.Func {
		return nil, fmt.Errorf("registered builder is not a function for type %v", dataType)
	}

	// Create a new saga definition builder
	sagaDefBuilder := NewSaga(tp)

	// Call the builder function using reflection
	results := builderValue.Call([]reflect.Value{
		reflect.ValueOf(input),
		reflect.ValueOf(sagaDefBuilder),
	})

	// The first result should be the constructed Saga
	if len(results) != 1 {
		return nil, fmt.Errorf("builder function returned unexpected results")
	}

	saga, ok := results[0].Interface().(*SagaDefinition)
	if !ok {
		return nil, fmt.Errorf("builder function did not return a *Saga")
	}

	return saga, nil
}
