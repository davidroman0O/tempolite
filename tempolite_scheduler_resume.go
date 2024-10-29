package tempolite

// // When workflows were paused, but will be flagged as IsReady
// func (tp *Tempolite) resumeWorkflowsWorker() {
// 	ticker := time.NewTicker(time.Second / 16)
// 	defer ticker.Stop()
// 	defer close(tp.resumeWorkflowsWorkerDone)
// 	for {
// 		select {
// 		case <-tp.ctx.Done():
// 			tp.logger.Debug(tp.ctx, "Resume workflows worker due to context done")
// 			return
// 		case <-tp.resumeWorkflowsWorkerDone:
// 			tp.logger.Debug(tp.ctx, "Resume workflows worker done")
// 			return
// 		case <-ticker.C:
// 			// Get distinct queue names that have workflows to resume
// 			queues, err := tp.client.Workflow.Query().
// 				Where(
// 					workflow.StatusEQ(workflow.StatusRunning),
// 					workflow.IsPausedEQ(false),
// 					workflow.IsReadyEQ(true),
// 				).
// 				Select(workflow.FieldQueueName).
// 				Strings(tp.ctx)
// 			if err != nil {
// 				tp.logger.Error(tp.ctx, "Error querying queues", "error", err)
// 				continue
// 			}

// 			// Process each queue
// 			for _, queueName := range queues {

// 				if queueName == "" {
// 					queueName = "default"
// 				}

// 				var targetPool *retrypool.Pool[*workflowTask]
// 				if poolRaw, exists := tp.queues.Load(queueName); exists {
// 					targetPool = poolRaw.(*retrypool.Pool[*workflowTask])
// 				} else {
// 					continue
// 				}

// 				availableSlots := len(targetPool.GetWorkerIDs()) - targetPool.ProcessingCount()
// 				if availableSlots <= 0 {
// 					continue
// 				}

// 				workflows, err := tp.client.Workflow.Query().
// 					Where(
// 						workflow.StatusEQ(workflow.StatusRunning),
// 						workflow.IsPausedEQ(false),
// 						workflow.IsReadyEQ(true),
// 						workflow.QueueNameEQ(queueName),
// 					).
// 					Limit(availableSlots).
// 					All(tp.ctx)
// 				if err != nil {
// 					tp.logger.Error(tp.ctx, "Error querying workflows to resume", "error", err)
// 					continue
// 				}

// 				for _, wf := range workflows {
// 					tp.logger.Debug(tp.ctx, "Resuming workflow", "workflowID", wf.ID, "queue", queueName)
// 					if err := tp.redispatchWorkflow(WorkflowID(wf.ID)); err != nil {
// 						tp.logger.Error(tp.ctx, "Error redispatching workflow", "workflowID", wf.ID, "error", err)
// 					}

// 					tx, err := tp.client.Tx(tp.ctx)
// 					if err != nil {
// 						tp.logger.Error(tp.ctx, "Error starting transaction for updating workflow", "workflowID", wf.ID, "error", err)
// 						continue
// 					}
// 					tp.logger.Debug(tp.ctx, "Started transaction for updating workflow", "workflowID", wf.ID)

// 					_, err = tx.Workflow.UpdateOne(wf).SetIsReady(false).Save(tp.ctx)
// 					if err != nil {
// 						tp.logger.Error(tp.ctx, "Error updating workflow", "workflowID", wf.ID, "error", err)
// 						if rollbackErr := tx.Rollback(); rollbackErr != nil {
// 							tp.logger.Error(tp.ctx, "Error rolling back transaction", "workflowID", wf.ID, "error", rollbackErr)
// 						} else {
// 							tp.logger.Debug(tp.ctx, "Successfully rolled back transaction", "workflowID", wf.ID)
// 						}
// 						continue
// 					}

// 					if err := tx.Commit(); err != nil {
// 						tp.logger.Error(tp.ctx, "Error committing transaction for updating workflow", "workflowID", wf.ID, "error", err)
// 					} else {
// 						tp.logger.Debug(tp.ctx, "Successfully committed transaction for updating workflow", "workflowID", wf.ID)
// 					}
// 				}
// 			}
// 		}
// 	}
// }

// func (tp *Tempolite) redispatchWorkflow(id WorkflowID) error {
// 	// Start a transaction
// 	tx, err := tp.client.Tx(tp.ctx)
// 	if err != nil {
// 		tp.logger.Error(tp.ctx, "Error starting transaction", "error", err)
// 		return fmt.Errorf("error starting transaction: %w", err)
// 	}

// 	// Ensure the transaction is rolled back in case of an error
// 	defer func() {
// 		if r := recover(); r != nil {
// 			if rerr := tx.Rollback(); rerr != nil {
// 				tp.logger.Error(tp.ctx, "Failed to rollback transaction after panic", "error", rerr)
// 			}
// 			panic(r)
// 		}
// 	}()

// 	wf, err := tx.Workflow.Get(tp.ctx, id.String())
// 	if err != nil {
// 		if rerr := tx.Rollback(); rerr != nil {
// 			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
// 		}
// 		tp.logger.Error(tp.ctx, "Error fetching workflow", "workflowID", id, "error", err)
// 		return fmt.Errorf("error fetching workflow: %w", err)
// 	}

// 	tp.logger.Debug(tp.ctx, "Redispatching workflow", "workflowID", id)
// 	tp.logger.Debug(tp.ctx, "Querying workflow execution", "workflowID", id)

// 	wfEx, err := tx.WorkflowExecution.Query().
// 		Where(workflowexecution.HasWorkflowWith(workflow.IDEQ(wf.ID))).
// 		WithWorkflow().
// 		Order(ent.Desc(workflowexecution.FieldStartedAt)).
// 		First(tp.ctx)

// 	if err != nil {
// 		if rerr := tx.Rollback(); rerr != nil {
// 			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
// 		}
// 		tp.logger.Error(tp.ctx, "Error querying workflow execution", "workflowID", id, "error", err)
// 		return fmt.Errorf("error querying workflow execution: %w", err)
// 	}

// 	// Create WorkflowContext
// 	ctx := WorkflowContext{
// 		tp:           tp,
// 		workflowID:   wf.ID,
// 		executionID:  wfEx.ID,
// 		runID:        wfEx.RunID,
// 		workflowType: wf.Identity,
// 		stepID:       wf.StepID,
// 		queueName:    wf.QueueName,
// 	}

// 	tp.logger.Debug(tp.ctx, "Getting handler info for workflow", "workflowID", id, "handlerIdentity", wf.Identity)

// 	handlerInfo, ok := tp.workflows.Load(HandlerIdentity(wf.Identity))
// 	if !ok {
// 		if rerr := tx.Rollback(); rerr != nil {
// 			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
// 		}
// 		tp.logger.Error(tp.ctx, "Workflow handler not found", "workflowID", id)
// 		return fmt.Errorf("workflow handler not found for %s", wf.Identity)
// 	}

// 	workflowHandler, ok := handlerInfo.(Workflow)
// 	if !ok {
// 		if rerr := tx.Rollback(); rerr != nil {
// 			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
// 		}
// 		tp.logger.Error(tp.ctx, "Invalid workflow handler", "workflowID", id)
// 		return fmt.Errorf("invalid workflow handler for %s", wf.Identity)
// 	}

// 	inputs, err := tp.convertInputsFromSerialization(HandlerInfo(workflowHandler), wf.Input)
// 	if err != nil {
// 		if rerr := tx.Rollback(); rerr != nil {
// 			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
// 		}
// 		tp.logger.Error(tp.ctx, "Error converting inputs from serialization", "workflowID", id, "error", err)
// 		return fmt.Errorf("error converting inputs from serialization: %w", err)
// 	}

// 	tp.logger.Debug(tp.ctx, "Creating workflow task", "workflowID", id, "handlerName", workflowHandler.HandlerLongName)

// 	// Create and dispatch the workflow task
// 	task := &workflowTask{
// 		ctx:         ctx,
// 		handlerName: workflowHandler.HandlerLongName,
// 		handler:     workflowHandler.Handler,
// 		params:      inputs,
// 		maxRetry:    wf.RetryPolicy.MaximumAttempts,
// 	}

// 	retryIt := func() error {
// 		tp.logger.Debug(tp.ctx, "Retrying workflow", "workflowID", id, "handlerName", workflowHandler.HandlerLongName)

// 		retryTx, err := tp.client.Tx(tp.ctx)
// 		if err != nil {
// 			tp.logger.Error(tp.ctx, "Error starting transaction for retry", "workflowID", id, "error", err)
// 			return fmt.Errorf("error starting transaction for retry: %w", err)
// 		}

// 		var workflowExecution *ent.WorkflowExecution
// 		if workflowExecution, err = retryTx.WorkflowExecution.
// 			Create().
// 			SetID(uuid.NewString()).
// 			SetRunID(wfEx.RunID).
// 			SetWorkflow(wf).
// 			Save(tp.ctx); err != nil {
// 			if rerr := retryTx.Rollback(); rerr != nil {
// 				tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
// 			}
// 			tp.logger.Error(tp.ctx, "Error creating workflow execution during retry", "workflowID", id, "error", err)
// 			return err
// 		}

// 		task.ctx.executionID = workflowExecution.ID
// 		task.retryCount++
// 		tp.logger.Debug(tp.ctx, "Retrying workflow", "workflowID", id, "handlerName", workflowHandler.HandlerLongName, "retryCount", task.retryCount)

// 		if _, err = retryTx.Workflow.UpdateOneID(ctx.workflowID).SetStatus(workflow.StatusRunning).Save(tp.ctx); err != nil {
// 			if rerr := retryTx.Rollback(); rerr != nil {
// 				tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
// 			}
// 			tp.logger.Error(tp.ctx, "Error updating workflow status during retry", "workflowID", id, "error", err)
// 			return fmt.Errorf("error updating workflow status during retry: %w", err)
// 		}

// 		if err = retryTx.Commit(); err != nil {
// 			tp.logger.Error(tp.ctx, "Error committing retry transaction", "workflowID", id, "error", err)
// 			return fmt.Errorf("error committing retry transaction: %w", err)
// 		}

// 		return nil
// 	}

// 	task.retry = retryIt

// 	tp.logger.Debug(tp.ctx, "Fetching total execution workflow related to workflow entity", "workflowID", id, "handlerName", workflowHandler.HandlerLongName)
// 	total, err := tx.WorkflowExecution.
// 		Query().
// 		Where(workflowexecution.HasWorkflowWith(workflow.IDEQ(wf.ID))).Count(tp.ctx)
// 	if err != nil {
// 		if rerr := tx.Rollback(); rerr != nil {
// 			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
// 		}
// 		tp.logger.Error(tp.ctx, "Error fetching total execution workflow related to workflow entity", "workflowID", id, "handlerName", workflowHandler.HandlerLongName, "error", err)
// 		return err
// 	}

// 	if total > 1 {
// 		task.retryCount = total
// 	}

// 	if wf.QueueName == "" {
// 		wf.QueueName = "default"
// 	}

// 	// Get the appropriate pool based on queue name
// 	var targetPool *QueueWorkers
// 	if poolRaw, exists := tp.queues.Load(wf.QueueName); exists {
// 		targetPool = poolRaw.(*QueueWorkers)
// 	} else {
// 		if rerr := tx.Rollback(); rerr != nil {
// 			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
// 		}
// 		return fmt.Errorf("queue %s not found", wf.QueueName)
// 	}

// 	tp.logger.Debug(tp.ctx, "Dispatching workflow", "workflowID", id, "handlerName", workflowHandler.HandlerLongName, "queue", wf.QueueName)

// 	if err := targetPool.Workflows.Dispatch(task); err != nil {
// 		if rerr := tx.Rollback(); rerr != nil {
// 			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
// 		}
// 		tp.logger.Error(tp.ctx, "Error dispatching workflow task", "workflowID", id, "error", err)
// 		return fmt.Errorf("error dispatching workflow task: %w", err)
// 	}

// 	if err := tx.Commit(); err != nil {
// 		tp.logger.Error(tp.ctx, "Error committing transaction", "workflowID", id, "error", err)
// 		return fmt.Errorf("error committing transaction: %w", err)
// 	}

// 	return nil
// }
