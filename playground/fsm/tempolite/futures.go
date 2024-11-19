package tempolite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"
)

type DatabaseFuture struct {
	ctx      context.Context
	entityID int
	database Database
	results  []interface{}
	err      error
}

func NewDatabaseFuture(ctx context.Context, database Database) *DatabaseFuture {
	return &DatabaseFuture{
		ctx:      ctx,
		database: database,
	}
}

func (f *DatabaseFuture) setEntityID(entityID int) {
	f.entityID = entityID
}

func (f *DatabaseFuture) setError(err error) {
	f.err = err
}

func (f *DatabaseFuture) Get(out ...interface{}) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-f.ctx.Done():
			return errors.New("context cancelled")
		case <-ticker.C:
			if f.entityID == 0 {
				continue
			}
			if f.err != nil {
				return f.err
			}
			if completed := f.checkCompletion(); completed {
				return f.handleResults(out...)
			}
			continue
		}
	}
}

func (f *DatabaseFuture) handleResults(out ...interface{}) error {
	if f.err != nil {
		return f.err
	}

	if len(out) == 0 {
		return nil
	}

	// // Handle single result case
	// if len(f.results) == 1 && len(out) == 1 {
	// 	val := reflect.ValueOf(out[0])
	// 	if val.Kind() != reflect.Ptr {
	// 		return fmt.Errorf("output parameter must be a pointer")
	// 	}
	// 	val = val.Elem()

	// 	result := reflect.ValueOf(f.results[0])
	// 	if !result.Type().AssignableTo(val.Type()) {
	// 		return fmt.Errorf("cannot assign type %v to %v", result.Type(), val.Type())
	// 	}

	// 	val.Set(result)
	// 	return nil
	// }

	// Handle multiple results
	if len(out) > len(f.results) {
		return fmt.Errorf("number of outputs (%d) exceeds number of results (%d)", len(out), len(f.results))
	}

	for i := 0; i < len(out); i++ {
		val := reflect.ValueOf(out[i])
		if val.Kind() != reflect.Ptr {
			return fmt.Errorf("output parameter %d must be a pointer", i)
		}
		val = val.Elem()

		result := reflect.ValueOf(f.results[i])
		if !result.Type().AssignableTo(val.Type()) {
			return fmt.Errorf("cannot assign type %v to %v for parameter %d", result.Type(), val.Type(), i)
		}

		val.Set(result)
	}

	return nil
}

func (f *DatabaseFuture) checkCompletion() bool {
	entity := f.database.GetEntity(f.entityID)
	if entity == nil {
		f.err = fmt.Errorf("entity not found: %d", f.entityID)
		return true
	}

	fmt.Println("\t entity", entity.ID, " status: ", entity.Status)
	// TODO: should we check for pause?
	switch entity.Status {
	case StatusCompleted:
		// Get results from latest execution
		if latestExec := f.database.GetLatestExecution(f.entityID); latestExec != nil &&
			latestExec.WorkflowExecutionData != nil {
			outputs, err := convertOutputsFromSerialization(*entity.HandlerInfo,
				latestExec.WorkflowExecutionData.Outputs)
			if err != nil {
				f.err = err
				return true
			}
			f.results = outputs
			return true
		}
	case StatusFailed:
		if latestExec := f.database.GetLatestExecution(f.entityID); latestExec != nil {
			f.err = errors.New(latestExec.Error)
		} else {
			f.err = errors.New("workflow failed with no error details")
		}
		return true
	case StatusCancelled:
		f.err = errors.New("workflow was cancelled")
		return true
	}

	return false
}

// RuntimeFuture represents an asynchronous result.
type RuntimeFuture struct {
	results    []interface{}
	err        error
	done       chan struct{}
	workflowID int
}

func (f *RuntimeFuture) WorkflowID() int {
	return f.workflowID
}

func NewRuntimeFuture() *RuntimeFuture {
	return &RuntimeFuture{
		done: make(chan struct{}),
	}
}

func (f *RuntimeFuture) setEntityID(entityID int) {
	f.workflowID = entityID
}

func (f *RuntimeFuture) setResult(results []interface{}) {
	log.Printf("Future.setResult called with results: %v", results)
	f.results = results
	close(f.done)
}

func (f *RuntimeFuture) setError(err error) {
	log.Printf("Future.setError called with error: %v", err)
	f.err = err
	close(f.done)
}

func (f *RuntimeFuture) Get(out ...interface{}) error {
	<-f.done
	if f.err != nil {
		return f.err
	}

	if len(out) == 0 {
		return nil
	}

	// // Handle the case where we have a single result
	// if len(f.results) == 1 && len(out) == 1 {
	// 	val := reflect.ValueOf(out[0])
	// 	if val.Kind() != reflect.Ptr {
	// 		return fmt.Errorf("output parameter must be a pointer")
	// 	}
	// 	val = val.Elem()

	// 	result := reflect.ValueOf(f.results[0])
	// 	if !result.Type().AssignableTo(val.Type()) {
	// 		return fmt.Errorf("cannot assign type %v to %v", result.Type(), val.Type())
	// 	}

	// 	val.Set(result)
	// 	return nil
	// } else if len(f.results) == 0 {
	// 	return nil
	// }

	if len(out) > len(f.results) {
		return fmt.Errorf("number of outputs (%d) exceeds number of results (%d)", len(out), len(f.results))
	}

	for i := 0; i < len(out); i++ {
		val := reflect.ValueOf(out[i])
		if val.Kind() != reflect.Ptr {
			return fmt.Errorf("output parameter %d must be a pointer", i)
		}
		val = val.Elem()

		result := reflect.ValueOf(f.results[i])
		if !result.Type().AssignableTo(val.Type()) {
			return fmt.Errorf("cannot assign type %v to %v for parameter %d", result.Type(), val.Type(), i)
		}

		val.Set(result)
	}

	return nil
}
