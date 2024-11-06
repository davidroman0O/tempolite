package tempolite

import (
	"context"
	"log"
	"testing"
)

type testOrderSaga struct {
	OrderID string
}

func (o testOrderSaga) Transaction(ctx TransactionContext) (interface{}, error) {
	log.Println("Starting transaction for testOrderSaga")
	// Perform the main transaction logic here, like placing an order.
	orderID := "12345" // Example order ID
	return orderID, nil
}

func (o testOrderSaga) Compensation(ctx CompensationContext) (interface{}, error) {
	log.Println("Compensating for testOrderSaga")
	// Perform compensation logic here, like rolling back the order.
	return "OrderCompensated", nil
}

type testPaymentSaga struct {
	OrderID string
}

func (p testPaymentSaga) Transaction(ctx TransactionContext) (interface{}, error) {
	log.Println("Starting transaction for testPaymentSaga")
	paymentID := "67890"
	return paymentID, nil
}

func (p testPaymentSaga) Compensation(ctx CompensationContext) (interface{}, error) {
	log.Println("Compensating for testPaymentSaga")
	return "PaymentCompensated", nil
}

func TestSaga(t *testing.T) {

	tp, err := New(context.Background(), NewRegistry().Build())
	if err != nil {
		t.Fatalf("Failed to create Tempolite instance: %v", err)
	}

	// Create a new saga builder
	sagaBuilder := NewSaga()

	// Add steps to the saga
	sagaBuilder.AddStep(testOrderSaga{OrderID: "12345"})
	sagaBuilder.AddStep(testPaymentSaga{OrderID: "12345"})

	// Build the saga
	saga, err := sagaBuilder.Build()
	if err != nil {
		t.Fatalf("Failed to build saga: %v", err)
	}

	// Create a mock WorkflowContext for testing
	workflowContext := WorkflowContext{
		tp:           tp,
		workflowID:   "12345",
		executionID:  "67890",
		runID:        "09876",
		workflowType: "TestWorkflow",
		stepID:       "TestStep",
	}

	// Execute the saga within the workflow context
	sagaInfo := workflowContext.Saga("saga-step", saga)

	// Add assertions to check if the saga was registered and executed correctly
	if sagaInfo == nil {
		t.Fatal("Expected non-nil SagaInfo, got nil")
	}

	// Check if the saga was stored in the Tempolite instance
	if _, ok := tp.sagas.Load(sagaInfo.SagaID); !ok {
		t.Error("Saga was not stored in Tempolite")
	}

	/// Yea i think that looks right!
}
