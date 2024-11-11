package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/comfylite3"
	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/hierarchy"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/schema"
)

func TestComplexWorkflowScenario(t *testing.T) {
	optsComfy := []comfylite3.ComfyOption{
		comfylite3.WithPath("tempolite.db"),
	}
	comfy, err := comfylite3.New(optsComfy...)
	if err != nil {
		t.Fatalf("failed creating comfy: %v", err)
	}

	db := comfylite3.OpenDB(
		comfy,
		comfylite3.WithOption("_fk=1"),
		comfylite3.WithOption("cache=shared"),
		comfylite3.WithOption("mode=rwc"),
		comfylite3.WithForeignKeys(),
	)
	defer db.Close()

	client := ent.NewClient(ent.Driver(sql.OpenDB(dialect.SQLite, db)))
	defer client.Close()

	if err := client.Schema.Create(context.Background()); err != nil {
		t.Fatalf("failed creating schema resources: %v", err)
	}

	ctx := context.Background()
	repo := NewRepository(ctx, client)
	tx, err := client.Tx(ctx)
	if err != nil {
		t.Fatalf("failed starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Create multiple runs for different domains
	runs := make(map[string]*RunInfo)
	for _, runName := range []string{
		"order-processing-123", "shipping-456", "inventory-789",
		"payment-batch-001", "user-registration-002", "data-sync-003",
		"billing-cycle-004", "marketing-campaign-005", "audit-process-006",
		"notification-batch-007"} {
		run, err := repo.Runs().Create(tx)
		if err != nil {
			t.Fatalf("failed creating run %s: %v", runName, err)
		}
		runs[runName] = run
	}

	// Create queues with different priorities
	queues := make(map[string]*QueueInfo)
	for _, qName := range []string{
		"critical", "high", "default", "background", "batch",
		"analytics", "cleanup", "maintenance", "scheduler", "audit"} {
		q, err := repo.Queues().Create(tx, qName)
		if err != nil {
			t.Fatalf("failed creating queue %s: %v", qName, err)
		}
		queues[qName] = q
	}

	// Create workflows for each run
	for runName, run := range runs {
		// Create root workflow
		rootWorkflow, err := repo.Workflows().Create(tx, CreateWorkflowInput{
			RunID:       run.ID,
			HandlerName: fmt.Sprintf("%sRootWorkflow", runName),
			StepID:      fmt.Sprintf("%s-root", runName),
			QueueID:     queues["high"].ID,
			RetryPolicy: &schema.RetryPolicy{
				MaxAttempts:        3,
				InitialInterval:    1000000000,
				BackoffCoefficient: 2.0,
				MaxInterval:        300000000000,
			},
			Input: [][]byte{[]byte(fmt.Sprintf(`{"runId":"%s","timestamp":"%s"}`,
				runName, time.Now().Format(time.RFC3339)))},
		})
		if err != nil {
			t.Fatalf("failed creating root workflow for %s: %v", runName, err)
		}

		// Create multiple sub-workflows
		subWorkflowNames := []string{"validation", "processing", "notification", "cleanup", "monitoring"}
		subWorkflows := make([]*WorkflowInfo, len(subWorkflowNames))

		for i, name := range subWorkflowNames {
			subWorkflows[i], err = repo.Workflows().Create(tx, CreateWorkflowInput{
				RunID:       run.ID,
				HandlerName: fmt.Sprintf("%s%sWorkflow", runName, name),
				StepID:      fmt.Sprintf("%s-%s", runName, name),
				QueueID:     queues["default"].ID,
				RetryPolicy: &schema.RetryPolicy{MaxAttempts: 2},
				Input:       [][]byte{[]byte(fmt.Sprintf(`{"parentWorkflow":"%s-root"}`, runName))},
			})
			if err != nil {
				t.Fatalf("failed creating sub-workflow %s for %s: %v", name, runName, err)
			}
		}

		// Create saga for each workflow
		saga, err := repo.Sagas().Create(tx, CreateSagaInput{
			RunID:       run.ID,
			HandlerName: fmt.Sprintf("%sTransactionSaga", runName),
			StepID:      fmt.Sprintf("%s-saga", runName),
			QueueID:     queues["critical"].ID,
			CompensationData: [][]byte{[]byte(fmt.Sprintf(`{
				"steps": [
					{"step": "Step1", "compensation": "CompensateStep1"},
					{"step": "Step2", "compensation": "CompensateStep2"},
					{"step": "Step3", "compensation": "CompensateStep3"}
				],
				"workflowId": "%s"
			}`, runName))},
		})
		if err != nil {
			t.Fatalf("failed creating saga for %s: %v", runName, err)
		}

		// Create multiple activities for each sub-workflow
		activityTypes := []string{"validate", "process", "update", "notify", "log"}
		for _, subWorkflow := range subWorkflows {
			for _, actType := range activityTypes {
				activity, err := repo.Activities().Create(tx, CreateActivityInput{
					RunID:       run.ID,
					HandlerName: fmt.Sprintf("%s%sActivity", subWorkflow.EntityInfo.HandlerName, actType),
					StepID:      fmt.Sprintf("%s-%s", subWorkflow.EntityInfo.StepID, actType),
					QueueID:     queues["default"].ID,
					RetryPolicy: &schema.RetryPolicy{MaxAttempts: 3},
					Input:       [][]byte{[]byte(fmt.Sprintf(`{"workflowId":"%s","type":"%s"}`, runName, actType))},
				})
				if err != nil {
					t.Fatalf("failed creating activity %s for %s: %v", actType, runName, err)
				}

				// Create hierarchy between sub-workflow and activity
				_, err = repo.Hierarchies().Create(tx, run.ID,
					subWorkflow.EntityInfo.ID, activity.EntityInfo.ID,
					subWorkflow.EntityInfo.StepID, activity.EntityInfo.StepID,
					subWorkflow.Execution.ID, activity.Execution.ID, hierarchy.ChildTypeActivity, hierarchy.ParentTypeWorkflow)
				if err != nil {
					t.Fatalf("failed creating hierarchy for activity %s: %v", actType, err)
				}
			}
		}

		// Create side effects for metrics and notifications
		sideEffectTypes := []string{"metrics", "notification", "audit", "cleanup"}
		for _, effectType := range sideEffectTypes {
			effect, err := repo.SideEffects().Create(tx, CreateSideEffectInput{
				RunID:       run.ID,
				HandlerName: fmt.Sprintf("%s%sEffect", runName, effectType),
				StepID:      fmt.Sprintf("%s-%s-effect", runName, effectType),
				QueueID:     queues["background"].ID,
				Input:       [][]byte{[]byte(fmt.Sprintf(`{"workflowId":"%s","type":"%s"}`, runName, effectType))},
				Metadata:    []byte(fmt.Sprintf(`{"category":"%s","priority":"low"}`, effectType)),
			})
			if err != nil {
				t.Fatalf("failed creating side effect %s for %s: %v", effectType, runName, err)
			}

			// Create hierarchy between root workflow and side effect
			_, err = repo.Hierarchies().Create(tx, run.ID,
				rootWorkflow.EntityInfo.ID, effect.EntityInfo.ID,
				rootWorkflow.EntityInfo.StepID, effect.EntityInfo.StepID,
				rootWorkflow.Execution.ID, effect.Execution.ID, hierarchy.ChildTypeSideEffect, hierarchy.ParentTypeWorkflow)
			if err != nil {
				t.Fatalf("failed creating hierarchy for side effect %s: %v", effectType, err)
			}
		}

		// Create hierarchies between root workflow and sub-workflows
		for _, subWorkflow := range subWorkflows {
			_, err = repo.Hierarchies().Create(tx, run.ID,
				rootWorkflow.EntityInfo.ID, subWorkflow.EntityInfo.ID,
				rootWorkflow.EntityInfo.StepID, subWorkflow.EntityInfo.StepID,
				rootWorkflow.Execution.ID, subWorkflow.Execution.ID, hierarchy.ChildTypeWorkflow, hierarchy.ParentTypeWorkflow)
			if err != nil {
				t.Fatalf("failed creating hierarchy for sub-workflow in %s: %v", runName, err)
			}
		}

		// Create hierarchy between root workflow and saga
		_, err = repo.Hierarchies().Create(tx, run.ID,
			rootWorkflow.EntityInfo.ID, saga.EntityInfo.ID,
			rootWorkflow.EntityInfo.StepID, saga.EntityInfo.StepID,
			rootWorkflow.Execution.ID, saga.Execution.ID, hierarchy.ChildTypeSaga, hierarchy.ParentTypeWorkflow)
		if err != nil {
			t.Fatalf("failed creating hierarchy for saga in %s: %v", runName, err)
		}
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("failed committing transaction: %v", err)
	}
}
