package tempolite

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMemoryDatabaseRaces(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T, *MemoryDatabase)
	}{
		{"ConcurrentWorkflowOperations", testConcurrentWorkflowOperations},
		{"ConcurrentActivityOperations", testConcurrentActivityOperations},
		{"ConcurrentSagaOperations", testConcurrentSagaOperations},
		{"ConcurrentSideEffectOperations", testConcurrentSideEffectOperations},
		{"ConcurrentSignalOperations", testConcurrentSignalOperations},
		{"ConcurrentQueueOperations", testConcurrentQueueOperations},
		{"ConcurrentVersionOperations", testConcurrentVersionOperations},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewMemoryDatabase()
			tt.test(t, db)
		})
	}
}

func testConcurrentWorkflowOperations(t *testing.T, db *MemoryDatabase) {
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()

			// Create workflow entity
			workflowEntity := &WorkflowEntity{
				BaseEntity: BaseEntity{
					HandlerName: fmt.Sprintf("handler_%d", i),
					Status:      StatusPending,
					Type:        EntityWorkflow,
					StepID:      fmt.Sprintf("step_%d", i),
				},
				WorkflowData: &WorkflowData{
					Inputs: [][]byte{[]byte(fmt.Sprintf("input_%d", i))},
				},
			}

			// Add workflow
			id, err := db.AddWorkflowEntity(workflowEntity)
			if err != nil {
				t.Errorf("Failed to add workflow: %v", err)
				return
			}

			// Get workflow
			_, err = db.GetWorkflowEntity(id)
			if err != nil {
				t.Errorf("Failed to get workflow: %v", err)
				return
			}

			// Update status
			err = db.SetWorkflowEntityProperties(id, SetWorkflowEntityStatus(StatusRunning))
			if err != nil {
				t.Errorf("Failed to update workflow status: %v", err)
				return
			}
		}(i)
	}

	wg.Wait()
}

func testConcurrentActivityOperations(t *testing.T, db *MemoryDatabase) {
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Create a parent workflow first
	parentWorkflowID, err := db.AddWorkflowEntity(&WorkflowEntity{
		BaseEntity: BaseEntity{
			HandlerName: "parent_workflow",
			Status:      StatusPending,
			Type:        EntityWorkflow,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create parent workflow: %v", err)
	}

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()

			// Create activity entity
			activityEntity := &ActivityEntity{
				BaseEntity: BaseEntity{
					HandlerName: fmt.Sprintf("handler_%d", i),
					Status:      StatusPending,
					Type:        EntityActivity,
					StepID:      fmt.Sprintf("step_%d", i),
				},
				ActivityData: &ActivityData{
					Inputs: [][]byte{[]byte(fmt.Sprintf("input_%d", i))},
				},
			}

			// Add activity
			id, err := db.AddActivityEntity(activityEntity, parentWorkflowID)
			if err != nil {
				t.Errorf("Failed to add activity: %v", err)
				return
			}

			// Get activity
			_, err = db.GetActivityEntity(id)
			if err != nil {
				t.Errorf("Failed to get activity: %v", err)
				return
			}

			// Update status
			err = db.SetActivityEntityProperties(id, SetActivityEntityStatus(StatusRunning))
			if err != nil {
				t.Errorf("Failed to update activity status: %v", err)
				return
			}
		}(i)
	}

	wg.Wait()
}

func testConcurrentSagaOperations(t *testing.T, db *MemoryDatabase) {
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Create a parent workflow first
	parentWorkflowID, err := db.AddWorkflowEntity(&WorkflowEntity{
		BaseEntity: BaseEntity{
			HandlerName: "parent_workflow",
			Status:      StatusPending,
			Type:        EntityWorkflow,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create parent workflow: %v", err)
	}

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()

			// Create saga entity
			sagaEntity := &SagaEntity{
				BaseEntity: BaseEntity{
					HandlerName: fmt.Sprintf("handler_%d", i),
					Status:      StatusPending,
					Type:        EntitySaga,
					StepID:      fmt.Sprintf("step_%d", i),
				},
				SagaData: &SagaData{},
			}

			// Add saga
			id, err := db.AddSagaEntity(sagaEntity, parentWorkflowID)
			if err != nil {
				t.Errorf("Failed to add saga: %v", err)
				return
			}

			// Get saga
			_, err = db.GetSagaEntity(id)
			if err != nil {
				t.Errorf("Failed to get saga: %v", err)
				return
			}

			// Update status
			err = db.SetSagaEntityProperties(id, SetSagaEntityStatus(StatusRunning))
			if err != nil {
				t.Errorf("Failed to update saga status: %v", err)
				return
			}
		}(i)
	}

	wg.Wait()
}

func testConcurrentSideEffectOperations(t *testing.T, db *MemoryDatabase) {
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Create a parent workflow first
	parentWorkflowID, err := db.AddWorkflowEntity(&WorkflowEntity{
		BaseEntity: BaseEntity{
			HandlerName: "parent_workflow",
			Status:      StatusPending,
			Type:        EntityWorkflow,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create parent workflow: %v", err)
	}

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()

			// Create side effect entity
			sideEffectEntity := &SideEffectEntity{
				BaseEntity: BaseEntity{
					HandlerName: fmt.Sprintf("handler_%d", i),
					Status:      StatusPending,
					Type:        EntitySideEffect,
					StepID:      fmt.Sprintf("step_%d", i),
				},
				SideEffectData: &SideEffectData{},
			}

			// Add side effect
			id, err := db.AddSideEffectEntity(sideEffectEntity, parentWorkflowID)
			if err != nil {
				t.Errorf("Failed to add side effect: %v", err)
				return
			}

			// Get side effect
			_, err = db.GetSideEffectEntity(id)
			if err != nil {
				t.Errorf("Failed to get side effect: %v", err)
				return
			}

			// Update status
			err = db.SetSideEffectEntityProperties(id, SetSideEffectEntityStatus(StatusRunning))
			if err != nil {
				t.Errorf("Failed to update side effect status: %v", err)
				return
			}
		}(i)
	}

	wg.Wait()
}

func testConcurrentSignalOperations(t *testing.T, db *MemoryDatabase) {
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Create a parent workflow first
	parentWorkflowID, err := db.AddWorkflowEntity(&WorkflowEntity{
		BaseEntity: BaseEntity{
			HandlerName: "parent_workflow",
			Status:      StatusPending,
			Type:        EntityWorkflow,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create parent workflow: %v", err)
	}

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()

			// Create signal entity
			signalEntity := &SignalEntity{
				BaseEntity: BaseEntity{
					HandlerName: fmt.Sprintf("handler_%d", i),
					Status:      StatusPending,
					Type:        EntitySignal,
					StepID:      fmt.Sprintf("step_%d", i),
				},
				SignalData: &SignalData{
					Name: fmt.Sprintf("signal_%d", i),
				},
			}

			// Add signal
			id, err := db.AddSignalEntity(signalEntity, parentWorkflowID)
			if err != nil {
				t.Errorf("Failed to add signal: %v", err)
				return
			}

			// Get signal
			_, err = db.GetSignalEntity(id)
			if err != nil {
				t.Errorf("Failed to get signal: %v", err)
				return
			}

			// Update status
			err = db.SetSignalEntityProperties(id, SetSignalEntityStatus(StatusRunning))
			if err != nil {
				t.Errorf("Failed to update signal status: %v", err)
				return
			}
		}(i)
	}

	wg.Wait()
}

func testConcurrentQueueOperations(t *testing.T, db *MemoryDatabase) {
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()

			// Create queue
			queue := &Queue{
				Name: fmt.Sprintf("queue_%d", i),
			}

			// Add queue
			id, err := db.AddQueue(queue)
			if err != nil && err != ErrQueueExists {
				t.Errorf("Failed to add queue: %v", err)
				return
			}

			// Get queue
			_, err = db.GetQueue(id)
			if err != nil {
				t.Errorf("Failed to get queue: %v", err)
				return
			}

			// Get queue by name
			_, err = db.GetQueueByName(fmt.Sprintf("queue_%d", i))
			if err != nil {
				t.Errorf("Failed to get queue by name: %v", err)
				return
			}
		}(i)
	}

	wg.Wait()
}

func testConcurrentVersionOperations(t *testing.T, db *MemoryDatabase) {
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()

			// Create version
			version := &Version{
				EntityID: WorkflowEntityID(i),
				ChangeID: fmt.Sprintf("change_%d", i),
				Version:  i,
				Data: map[string]interface{}{
					"key": fmt.Sprintf("value_%d", i),
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Add version
			id, err := db.AddVersion(version)
			if err != nil {
				t.Errorf("Failed to add version: %v", err)
				return
			}

			// Get version
			_, err = db.GetVersion(id)
			if err != nil {
				t.Errorf("Failed to get version: %v", err)
				return
			}

			// Update version
			version.Version = i + 1
			err = db.UpdateVersion(version)
			if err != nil {
				t.Errorf("Failed to update version: %v", err)
				return
			}
		}(i)
	}

	wg.Wait()
}
