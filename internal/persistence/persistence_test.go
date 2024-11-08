package repository

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/comfylite3"
	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/activitydata"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/activityexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/execution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/hierarchy"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/run"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sagadata"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sideeffectdata"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/workflowdata"
)

func TestComplexWorkflowScenario(t *testing.T) {
	optsComfy := []comfylite3.ComfyOption{
		comfylite3.WithPath("tempolite.db"),
	}
	comfy, err := comfylite3.New(optsComfy...)
	if err != nil {
		log.Fatalf("failed creating comfy: %v", err)
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

	// Create queues with different priorities
	queues := make(map[string]*ent.Queue)
	for _, qName := range []string{"critical", "high", "default", "background"} {
		q, err := client.Queue.Create().SetName(qName).Save(ctx)
		if err != nil {
			t.Fatalf("failed creating queue %s: %v", qName, err)
		}
		queues[qName] = q
	}

	// Create a complex order processing run
	run, err := client.Run.Create().
		SetStatus("running").
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating run: %v", err)
	}

	// Create root workflow
	rootWorkflow, err := client.Entity.Create().
		SetType("Workflow").
		SetHandlerName("ComplexOrderWorkflow").
		SetStepID("order-workflow-root").
		SetRun(run).
		SetQueue(queues["high"]).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating root workflow: %v", err)
	}

	// Add workflow data with retry policy
	_, err = client.WorkflowData.Create().
		SetEntity(rootWorkflow).
		SetInput([][]byte{[]byte(`{
			"orderId": "ORD-123",
			"customerId": "CUST-456",
			"items": [
				{"id": "ITEM-1", "quantity": 2, "price": 1000},
				{"id": "ITEM-2", "quantity": 1, "price": 500}
			],
			"shippingAddress": {
				"street": "123 Main St",
				"city": "Example City",
				"country": "US"
			}
		}`)}).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating workflow data: %v", err)
	}

	// Create root workflow execution
	rootExec, err := client.Execution.Create().
		SetEntity(rootWorkflow).
		SetStatus("running").
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating root execution: %v", err)
	}

	// Create workflow execution with its data
	rootWorkflowExec, err := client.WorkflowExecution.Create().
		SetExecution(rootExec).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating workflow execution: %v", err)
	}

	_, err = client.WorkflowExecutionData.Create().
		SetWorkflowExecution(rootWorkflowExec).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating workflow execution data: %v", err)
	}

	// Create validation sub-workflow
	validationWorkflow, validationExec := createSubWorkflow(t, ctx, client, run, rootWorkflow, rootExec, "ValidationWorkflow", "validate-order", queues["high"])

	// Create validation activities
	validationActivities := []struct {
		handler string
		step    string
		queue   *ent.Queue
		input   string
	}{
		{
			"ValidateInventoryActivity",
			"validate-inventory",
			queues["high"],
			`{"items": [{"id": "ITEM-1", "quantity": 2}, {"id": "ITEM-2", "quantity": 1}]}`,
		},
		{
			"ValidateCustomerActivity",
			"validate-customer",
			queues["high"],
			`{"customerId": "CUST-456", "orderTotal": 2500}`,
		},
		{
			"ValidateShippingActivity",
			"validate-shipping",
			queues["default"],
			`{"address": {"street": "123 Main St", "city": "Example City", "country": "US"}}`,
		},
	}

	for _, act := range validationActivities {
		activity, activityExec := createActivity(t, ctx, client, run, validationWorkflow, validationExec, act.handler, act.step, act.queue, act.input)
		createActivityExecutionData(t, ctx, client, activity, activityExec)
	}

	// Create payment saga
	paymentSaga, sagaExec := createSaga(t, ctx, client, run, rootWorkflow, rootExec, "PaymentSaga", "process-payment", queues["critical"])

	// Create saga activities
	sagaActivities := []struct {
		handler string
		step    string
		queue   *ent.Queue
		input   string
	}{
		{
			"AuthorizePaymentActivity",
			"authorize-payment",
			queues["critical"],
			`{"amount": 2500, "customerId": "CUST-456", "paymentMethod": "CARD-789"}`,
		},
		{
			"CapturePaymentActivity",
			"capture-payment",
			queues["critical"],
			`{"authorizationId": "AUTH-123", "amount": 2500}`,
		},
		{
			"NotifyBillingActivity",
			"notify-billing",
			queues["background"],
			`{"orderId": "ORD-123", "amount": 2500, "status": "captured"}`,
		},
	}

	for _, act := range sagaActivities {
		activity, activityExec := createActivity(t, ctx, client, run, paymentSaga, sagaExec, act.handler, act.step, act.queue, act.input)
		createActivityExecutionData(t, ctx, client, activity, activityExec)
	}

	// Create fulfillment workflow
	fulfillmentWorkflow, fulfillmentExec := createSubWorkflow(t, ctx, client, run, rootWorkflow, rootExec, "FulfillmentWorkflow", "fulfill-order", queues["default"])

	// Create fulfillment activities
	fulfillmentActivities := []struct {
		handler string
		step    string
		queue   *ent.Queue
		input   string
	}{
		{
			"ReserveInventoryActivity",
			"reserve-inventory",
			queues["high"],
			`{"items": [{"id": "ITEM-1", "quantity": 2}, {"id": "ITEM-2", "quantity": 1}]}`,
		},
		{
			"CreateShippingLabelActivity",
			"create-shipping-label",
			queues["default"],
			`{"orderId": "ORD-123", "address": {"street": "123 Main St", "city": "Example City", "country": "US"}}`,
		},
		{
			"NotifyWarehouseActivity",
			"notify-warehouse",
			queues["default"],
			`{"orderId": "ORD-123", "items": [{"id": "ITEM-1", "quantity": 2}, {"id": "ITEM-2", "quantity": 1}]}`,
		},
	}

	for _, act := range fulfillmentActivities {
		activity, activityExec := createActivity(t, ctx, client, run, fulfillmentWorkflow, fulfillmentExec, act.handler, act.step, act.queue, act.input)
		createActivityExecutionData(t, ctx, client, activity, activityExec)
	}

	// Create side effects for notifications
	notificationEffects := []struct {
		handler string
		step    string
		queue   *ent.Queue
		input   string
	}{
		{
			"SendCustomerEmailEffect",
			"send-customer-email",
			queues["background"],
			`{"to": "customer@example.com", "template": "order_confirmation", "orderId": "ORD-123"}`,
		},
		{
			"SendMetricsEffect",
			"send-metrics",
			queues["background"],
			`{"orderId": "ORD-123", "processingTime": 120, "orderValue": 2500}`,
		},
		{
			"LogAuditEffect",
			"log-audit",
			queues["background"],
			`{"orderId": "ORD-123", "event": "OrderProcessed", "timestamp": "2024-11-06T10:05:00Z"}`,
		},
	}

	for _, effect := range notificationEffects {
		createSideEffect(t, ctx, client, run, rootWorkflow, rootExec, effect.handler, effect.step, effect.queue, effect.input)
	}

	// Verify the created data structure
	verifyDataStructure(t, ctx, client, run.ID)
}

// Helper functions

func createSubWorkflow(t *testing.T, ctx context.Context, client *ent.Client, run *ent.Run, parentWorkflow *ent.Entity, parentExec *ent.Execution, handler, stepID string, queue *ent.Queue) (*ent.Entity, *ent.Execution) {
	workflow, err := client.Entity.Create().
		SetType("Workflow").
		SetHandlerName(handler).
		SetStepID(stepID).
		SetRun(run).
		SetQueue(queue).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating sub-workflow: %v", err)
	}

	// Add workflow data
	_, err = client.WorkflowData.Create().
		SetEntity(workflow).
		SetInput([][]byte{[]byte(fmt.Sprintf(`{"parentWorkflow": "%s"}`, parentWorkflow.StepID))}).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating workflow data: %v", err)
	}

	// Create execution
	exec, err := client.Execution.Create().
		SetEntity(workflow).
		SetStatus("running").
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating workflow execution: %v", err)
	}

	// Create workflow execution and its data together
	workflowExec, err := client.WorkflowExecution.Create().
		SetExecution(exec).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating workflow execution: %v", err)
	}

	_, err = client.WorkflowExecutionData.Create().
		SetWorkflowExecution(workflowExec). // Set the required edge
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating workflow execution data: %v", err)
	}

	// Create hierarchy
	_, err = client.Hierarchy.Create().
		SetRun(run).
		SetParentEntity(parentWorkflow).
		SetChildEntity(workflow).
		SetParentStepID(parentWorkflow.StepID).
		SetChildStepID(workflow.StepID).
		SetParentExecutionID(parentExec.ID).
		SetChildExecutionID(exec.ID).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating hierarchy: %v", err)
	}

	return workflow, exec
}

func createActivity(t *testing.T, ctx context.Context, client *ent.Client, run *ent.Run, parentEntity *ent.Entity, parentExec *ent.Execution, handler, stepID string, queue *ent.Queue, input string) (*ent.Entity, *ent.Execution) {
	activity, err := client.Entity.Create().
		SetType("Activity").
		SetHandlerName(handler).
		SetStepID(stepID).
		SetRun(run).
		SetQueue(queue).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating activity: %v", err)
	}

	// Add activity data
	_, err = client.ActivityData.Create().
		SetEntity(activity).
		SetMaxAttempts(3).
		SetTimeout(int64(5 * time.Minute)).
		SetInput([][]byte{[]byte(input)}).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating activity data: %v", err)
	}

	// Create execution
	exec, err := client.Execution.Create().
		SetEntity(activity).
		SetStatus("running").
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating activity execution: %v", err)
	}

	// Create activity execution
	_, err = client.ActivityExecution.Create().
		SetExecution(exec).
		SetAttempt(1).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating activity execution: %v", err)
	}

	// Create hierarchy
	_, err = client.Hierarchy.Create().
		SetRun(run).
		SetParentEntity(parentEntity).
		SetChildEntity(activity).
		SetParentStepID(parentEntity.StepID).
		SetChildStepID(activity.StepID).
		SetParentExecutionID(parentExec.ID).
		SetChildExecutionID(exec.ID).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating hierarchy: %v", err)
	}

	return activity, exec
}

func createActivityExecutionData(t *testing.T, ctx context.Context, client *ent.Client, activity *ent.Entity, exec *ent.Execution) {
	// Add activity execution data
	activityExec, err := client.ActivityExecution.Query().
		Where(activityexecution.HasExecutionWith(execution.IDEQ(exec.ID))).
		Only(ctx)
	if err != nil {
		t.Fatalf("failed querying activity execution: %v", err)
	}

	_, err = client.ActivityExecutionData.Create().
		SetActivityExecution(activityExec).
		SetHeartbeats([][]byte{
			[]byte(`{"timestamp": "2024-11-06T10:00:00Z", "status": "processing"}`),
			[]byte(`{"timestamp": "2024-11-06T10:00:30Z", "status": "processing"}`),
		}).
		SetLastHeartbeat(time.Now()).
		SetProgress([]byte(`{"percent_complete": 50}`)).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating activity execution data: %v", err)
	}
}

func createSaga(t *testing.T, ctx context.Context, client *ent.Client, run *ent.Run, parentWorkflow *ent.Entity, parentExec *ent.Execution, handler, stepID string, queue *ent.Queue) (*ent.Entity, *ent.Execution) {
	saga, err := client.Entity.Create().
		SetType("Saga").
		SetHandlerName(handler).
		SetStepID(stepID).
		SetRun(run).
		SetQueue(queue).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating saga: %v", err)
	}

	// Add saga data
	_, err = client.SagaData.Create().
		SetEntity(saga).
		SetCompensationData([][]byte{
			[]byte(`{
				"steps": [
					{"activity": "AuthorizePayment", "compensation": "VoidAuthorization"},
					{"activity": "CapturePayment", "compensation": "RefundPayment"},
					{"activity": "NotifyBilling", "compensation": "CancelBillingNotification"}
				]
			}`),
		}).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating saga data: %v", err)
	}

	// Create execution
	exec, err := client.Execution.Create().
		SetEntity(saga).
		SetStatus("running").
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating saga execution: %v", err)
	}

	// Create saga execution
	sagaExec, err := client.SagaExecution.Create().
		SetExecution(exec).
		SetStepType("transaction").
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating saga execution: %v", err)
	}

	// Add saga execution data
	_, err = client.SagaExecutionData.Create().
		SetSagaExecution(sagaExec).
		SetTransactionHistory([][]byte{
			[]byte(`{"step": "AuthorizePayment", "status": "completed", "timestamp": "2024-11-06T10:02:00Z"}`),
			[]byte(`{"step": "CapturePayment", "status": "in_progress", "timestamp": "2024-11-06T10:02:30Z"}`),
		}).
		SetLastTransaction(time.Now()).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating saga execution data: %v", err)
	}

	// Create hierarchy
	_, err = client.Hierarchy.Create().
		SetRun(run).
		SetParentEntity(parentWorkflow).
		SetChildEntity(saga).
		SetParentStepID(parentWorkflow.StepID).
		SetChildStepID(saga.StepID).
		SetParentExecutionID(parentExec.ID).
		SetChildExecutionID(exec.ID).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating hierarchy: %v", err)
	}

	return saga, exec
}

func createSideEffect(t *testing.T, ctx context.Context, client *ent.Client, run *ent.Run, parentWorkflow *ent.Entity, parentExec *ent.Execution, handler, stepID string, queue *ent.Queue, input string) {
	sideEffect, err := client.Entity.Create().
		SetType("SideEffect").
		SetHandlerName(handler).
		SetStepID(stepID).
		SetRun(run).
		SetQueue(queue).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating side effect: %v", err)
	}

	// Add side effect data
	_, err = client.SideEffectData.Create().
		SetEntity(sideEffect).
		SetInput([][]byte{[]byte(input)}).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating side effect data: %v", err)
	}

	// Create execution
	exec, err := client.Execution.Create().
		SetEntity(sideEffect).
		SetStatus("running").
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating side effect execution: %v", err)
	}

	// Create side effect execution
	sideEffectExec, err := client.SideEffectExecution.Create().
		SetExecution(exec).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating side effect execution: %v", err)
	}

	// Add side effect execution data
	_, err = client.SideEffectExecutionData.Create().
		SetSideEffectExecution(sideEffectExec).
		SetEffectTime(time.Now()).
		SetEffectMetadata([]byte(`{"correlation_id": "CORR-123", "attempt": 1}`)).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating side effect execution data: %v", err)
	}

	// Create hierarchy
	_, err = client.Hierarchy.Create().
		SetRun(run).
		SetParentEntity(parentWorkflow).
		SetChildEntity(sideEffect).
		SetParentStepID(parentWorkflow.StepID).
		SetChildStepID(sideEffect.StepID).
		SetParentExecutionID(parentExec.ID).
		SetChildExecutionID(exec.ID).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating hierarchy: %v", err)
	}
}

func verifyDataStructure(t *testing.T, ctx context.Context, client *ent.Client, runID int) {
	t.Run("VerifyDataStructure", func(t *testing.T) {
		// Count various entities
		counts := struct {
			Entities       int
			Executions     int
			Hierarchies    int
			WorkflowData   int
			ActivityData   int
			SagaData       int
			SideEffectData int
			ExecutionData  int
		}{}

		var err error
		counts.Entities, err = client.Entity.Query().Where(entity.HasRunWith(run.IDEQ(runID))).Count(ctx)
		if err != nil {
			t.Errorf("failed counting entities: %v", err)
		}

		counts.Executions, err = client.Execution.Query().
			Where(execution.HasEntityWith(entity.HasRunWith(run.IDEQ(runID)))).
			Count(ctx)
		if err != nil {
			t.Errorf("failed counting executions: %v", err)
		}

		counts.Hierarchies, err = client.Hierarchy.Query().Where(hierarchy.IDEQ(runID)).Count(ctx)
		if err != nil {
			t.Errorf("failed counting hierarchies: %v", err)
		}

		// Log the structure
		t.Logf("Data structure counts: %+v", counts)

		// Find root workflow (one without a parent in hierarchies)
		rootWorkflow, err := client.Entity.Query().
			Where(
				entity.TypeEQ("Workflow"),
				entity.HasRunWith(run.IDEQ(runID)),
			).
			First(ctx)
		if err != nil {
			t.Errorf("failed finding root workflow: %v", err)
		}

		// Verify it's really the root by checking hierarchies
		parentCount, err := client.Hierarchy.Query().
			Where(
				hierarchy.IDEQ(runID),
				hierarchy.ChildEntityIDEQ(rootWorkflow.ID),
			).
			Count(ctx)
		if err != nil {
			t.Errorf("failed checking root workflow's parents: %v", err)
		}
		if parentCount > 0 {
			t.Errorf("supposed root workflow has parents")
		}

		printEntityTree(t, ctx, client, rootWorkflow, "", runID)
	})
}

func printEntityTree(t *testing.T, ctx context.Context, client *ent.Client, parent *ent.Entity, indent string, runID int) {
	t.Logf("%s%s (%s) [%s]", indent, parent.StepID, parent.Type, parent.HandlerName)

	// Find child entities through hierarchy
	children, err := client.Hierarchy.Query().
		Where(
			hierarchy.IDEQ(runID),
			hierarchy.ParentEntityIDEQ(parent.ID),
		).
		All(ctx)
	if err != nil {
		t.Errorf("failed querying children: %v", err)
		return
	}

	for _, child := range children {
		childEntity, err := client.Entity.Get(ctx, child.ChildEntityID)
		if err != nil {
			t.Errorf("failed getting child entity: %v", err)
			continue
		}
		printEntityTree(t, ctx, client, childEntity, indent+"  ", runID)
	}

	// Print additional details for each entity type
	switch parent.Type {
	case "Workflow":
		workflowData, err := client.WorkflowData.Query().
			Where(workflowdata.HasEntityWith(entity.IDEQ(parent.ID))).
			Only(ctx)
		if err != nil {
			t.Logf("%s  [No workflow data found]", indent)
		} else {
			t.Logf("%s  Paused: %v, Resumable: %v", indent, workflowData.Paused, workflowData.Resumable)
		}

	case "Activity":
		activityData, err := client.ActivityData.Query().
			Where(activitydata.HasEntityWith(entity.IDEQ(parent.ID))).
			Only(ctx)
		if err != nil {
			t.Logf("%s  [No activity data found]", indent)
		} else {
			t.Logf("%s  Max Attempts: %d, Timeout: %d", indent, activityData.MaxAttempts, activityData.Timeout)
		}

	case "Saga":
		sagaData, err := client.SagaData.Query().
			Where(sagadata.HasEntityWith(entity.IDEQ(parent.ID))).
			Only(ctx)
		if err != nil {
			t.Logf("%s  [No saga data found]", indent)
		} else {
			t.Logf("%s  Compensating: %v", indent, sagaData.Compensating)
		}

	case "SideEffect":
		sideEffectData, err := client.SideEffectData.Query().
			Where(sideeffectdata.HasEntityWith(entity.IDEQ(parent.ID))).
			Only(ctx)
		if err != nil {
			t.Logf("%s  [No side effect data found]", indent)
		} else {
			t.Logf("%s  SideEffect: %v", indent, sideEffectData.ID)
		}
	}

	// Print executions
	executions, err := client.Execution.Query().
		Where(execution.HasEntityWith(entity.IDEQ(parent.ID))).
		All(ctx)
	if err != nil {
		t.Errorf("failed querying executions: %v", err)
	} else {
		for _, exec := range executions {
			t.Logf("%s  Execution: Status=%s, Started=%s", indent, exec.Status, exec.StartedAt.Format(time.RFC3339))
		}
	}
}
