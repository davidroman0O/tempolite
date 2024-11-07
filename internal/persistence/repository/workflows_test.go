package repository

import (
	"context"
	"testing"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/comfylite3"
	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*ent.Client, func()) {
	optsComfy := []comfylite3.ComfyOption{
		comfylite3.WithPath("tempolite_test.db"),
	}
	comfy, err := comfylite3.New(optsComfy...)
	require.NoError(t, err)

	db := comfylite3.OpenDB(
		comfy,
		comfylite3.WithOption("_fk=1"),
		comfylite3.WithOption("cache=shared"),
		comfylite3.WithForeignKeys(),
	)

	client := ent.NewClient(ent.Driver(sql.OpenDB(dialect.SQLite, db)))
	err = client.Schema.Create(context.Background())
	require.NoError(t, err)

	return client, func() {
		client.Close()
		db.Close()
	}
}

func TestWorkflowLifecycle(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewRepository(ctx, client)

	t.Run("Basic Workflow Creation and Retrieval", func(t *testing.T) {
		tx, err := client.Tx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		// Create a run first
		run, err := repo.Runs().Create(tx)
		require.NoError(t, err)

		// Create a queue
		queue, err := repo.Queues().Create(tx, "test-queue")
		require.NoError(t, err)

		// Create workflow
		workflow, err := repo.Workflows().Create(tx, CreateWorkflowInput{
			RunID:       run.ID,
			HandlerName: "TestWorkflow",
			StepID:      "test-workflow-1",
			QueueID:     queue.ID,
			RetryPolicy: &schema.RetryPolicy{
				MaxAttempts:        3,
				InitialInterval:    1000000000,
				BackoffCoefficient: 2.0,
				MaxInterval:        300000000000,
			},
			Input: [][]byte{[]byte(`{"test":"data"}`)},
		})
		require.NoError(t, err)
		assert.NotNil(t, workflow)
		assert.Equal(t, "TestWorkflow", workflow.EntityInfo.HandlerName)
		assert.Equal(t, ComponentWorkflow, workflow.EntityInfo.Type)
		assert.Equal(t, StatusRunning, workflow.Execution.Status)

		// Test retrieval by ID
		retrieved, err := repo.Workflows().Get(tx, workflow.EntityInfo.ID)
		require.NoError(t, err)
		assert.Equal(t, workflow.EntityInfo.ID, retrieved.EntityInfo.ID)

		// Test retrieval by StepID
		byStepID, err := repo.Workflows().GetByStepID(tx, "test-workflow-1")
		require.NoError(t, err)
		assert.Equal(t, workflow.EntityInfo.ID, byStepID.EntityInfo.ID)

		require.NoError(t, tx.Commit())
	})

	t.Run("Workflow Hierarchy with Sub-workflows", func(t *testing.T) {
		tx, err := client.Tx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		run, err := repo.Runs().Create(tx)
		require.NoError(t, err)

		queue, err := repo.Queues().Create(tx, "hierarchy-queue")
		require.NoError(t, err)

		// Create parent workflow
		parentWorkflow, err := repo.Workflows().Create(tx, CreateWorkflowInput{
			RunID:       run.ID,
			HandlerName: "ParentWorkflow",
			StepID:      "parent-workflow",
			QueueID:     queue.ID,
			RetryPolicy: &schema.RetryPolicy{MaxAttempts: 3},
			Input:       [][]byte{[]byte(`{"parent":"data"}`)},
		})
		require.NoError(t, err)

		// Create multiple sub-workflows
		subWorkflows := make([]*WorkflowInfo, 3)
		for i := 0; i < 3; i++ {
			subWorkflow, err := repo.Workflows().Create(tx, CreateWorkflowInput{
				RunID:       run.ID,
				HandlerName: "SubWorkflow",
				StepID:      "sub-workflow-" + string(rune(i+1)),
				QueueID:     queue.ID,
				RetryPolicy: &schema.RetryPolicy{MaxAttempts: 2},
				Input:       [][]byte{[]byte(`{"sub":"data"}`)},
			})
			require.NoError(t, err)

			// Create hierarchy relationship
			_, err = repo.Hierarchies().Create(tx, run.ID,
				parentWorkflow.EntityInfo.ID,
				subWorkflow.EntityInfo.ID,
				parentWorkflow.EntityInfo.StepID,
				subWorkflow.EntityInfo.StepID,
				parentWorkflow.Execution.ID,
				subWorkflow.Execution.ID)
			require.NoError(t, err)

			subWorkflows[i] = subWorkflow
		}

		// Verify parent-child relationships
		for _, subWorkflow := range subWorkflows {
			hierarchy, err := repo.Hierarchies().GetParent(tx, subWorkflow.EntityInfo.ID)
			require.NoError(t, err)
			assert.Equal(t, parentWorkflow.EntityInfo.ID, hierarchy.ParentEntityID)
		}

		children, err := repo.Hierarchies().GetChildren(tx, parentWorkflow.EntityInfo.ID)
		require.NoError(t, err)
		assert.Len(t, children, 3)

		require.NoError(t, tx.Commit())
	})

	t.Run("Workflow State Management", func(t *testing.T) {
		tx, err := client.Tx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		run, err := repo.Runs().Create(tx)
		require.NoError(t, err)

		workflow, err := repo.Workflows().Create(tx, CreateWorkflowInput{
			RunID:       run.ID,
			HandlerName: "StateWorkflow",
			StepID:      "state-workflow",
			RetryPolicy: &schema.RetryPolicy{MaxAttempts: 1},
		})
		require.NoError(t, err)

		// Test pausing
		err = repo.Workflows().Pause(tx, workflow.EntityInfo.ID)
		require.NoError(t, err)

		paused, err := repo.Workflows().Get(tx, workflow.EntityInfo.ID)
		require.NoError(t, err)
		assert.True(t, paused.Data.Paused)

		// Make workflow resumable and test resuming
		_, err = repo.Workflows().UpdateData(tx, workflow.EntityInfo.ID, UpdateWorkflowDataInput{
			Resumable: boolPtr(true),
		})
		require.NoError(t, err)

		err = repo.Workflows().Resume(tx, workflow.EntityInfo.ID)
		require.NoError(t, err)

		resumed, err := repo.Workflows().Get(tx, workflow.EntityInfo.ID)
		require.NoError(t, err)
		assert.False(t, resumed.Data.Paused)

		// Test execution status updates
		_, err = repo.Executions().Complete(tx, workflow.Execution.ID)
		require.NoError(t, err)

		completed, err := repo.Workflows().Get(tx, workflow.EntityInfo.ID)
		require.NoError(t, err)
		assert.Equal(t, StatusCompleted, completed.Execution.Status)
		assert.NotNil(t, completed.Execution.CompletedAt)

		require.NoError(t, tx.Commit())
	})

	t.Run("Workflow Data Updates", func(t *testing.T) {
		tx, err := client.Tx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		run, err := repo.Runs().Create(tx)
		require.NoError(t, err)

		workflow, err := repo.Workflows().Create(tx, CreateWorkflowInput{
			RunID:       run.ID,
			HandlerName: "DataWorkflow",
			StepID:      "data-workflow",
			Input:       [][]byte{[]byte(`{"initial":"data"}`)},
		})
		require.NoError(t, err)

		// Update workflow data
		updated, err := repo.Workflows().UpdateData(tx, workflow.EntityInfo.ID, UpdateWorkflowDataInput{
			Input: [][]byte{[]byte(`{"updated":"data"}`)},
			RetryPolicy: &schema.RetryPolicy{
				MaxAttempts:        5,
				InitialInterval:    2000000000,
				BackoffCoefficient: 1.5,
				MaxInterval:        600000000000,
			},
		})
		require.NoError(t, err)
		assert.Equal(t, 5, updated.Data.RetryPolicy.MaxAttempts)
		assert.Equal(t, [][]byte{[]byte(`{"updated":"data"}`)}, updated.Data.Input)

		require.NoError(t, tx.Commit())
	})

	t.Run("List Workflows", func(t *testing.T) {
		tx, err := client.Tx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		run, err := repo.Runs().Create(tx)
		require.NoError(t, err)

		// Create multiple workflows
		for i := 0; i < 5; i++ {
			_, err := repo.Workflows().Create(tx, CreateWorkflowInput{
				RunID:       run.ID,
				HandlerName: "ListWorkflow",
				StepID:      "list-workflow-" + string(rune(i+1)),
			})
			require.NoError(t, err)
		}

		// List all workflows for the run
		workflows, err := repo.Workflows().List(tx, run.ID)
		require.NoError(t, err)
		assert.Len(t, workflows, 5)

		require.NoError(t, tx.Commit())
	})
}

func boolPtr(b bool) *bool {
	return &b
}
