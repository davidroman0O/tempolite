package tempolite

// func TestTempoliteBasic(t *testing.T) {

// 	database := NewDefaultDatabase()
// 	ctx := context.Background()

// 	workflow := func(ctx WorkflowContext) error {
// 		t.Log("workflow started")
// 		return nil
// 	}

// 	engine, err := New(ctx, database)
// 	if err != nil {
// 		t.Fatalf("failed to create engine: %v", err)
// 	}

// 	engine.RegisterWorkflow(workflow)

// 	future := engine.Workflow(workflow, nil)
// 	if err := future.Get(); err != nil {
// 		t.Fatalf("failed to run workflow: %v", err)
// 	}

// 	t.Log("workflow completed successfully")
// 	engine.Wait()
// }
