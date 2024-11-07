package tempolite

import "sync/atomic"

func (tp *Tempolite) getWorkerWorkflowID(queue string) (int, error) {
	counter, _ := tp.queuePoolWorkflowCounter.LoadOrStore(queue, &atomic.Int64{})
	return int(counter.(*atomic.Int64).Add(1)), nil
}

func (tp *Tempolite) getWorkerActivityID(queue string) (int, error) {
	counter, _ := tp.queuePoolActivityCounter.LoadOrStore(queue, &atomic.Int64{})
	return int(counter.(*atomic.Int64).Add(1)), nil
}

func (tp *Tempolite) getWorkerSideEffectID(queue string) (int, error) {
	counter, _ := tp.queuePoolSideEffectCounter.LoadOrStore(queue, &atomic.Int64{})
	return int(counter.(*atomic.Int64).Add(1)), nil
}

func (tp *Tempolite) getWorkerTransactionID(queue string) (int, error) {
	counter, _ := tp.queuePoolTransactionCounter.LoadOrStore(queue, &atomic.Int64{})
	return int(counter.(*atomic.Int64).Add(1)), nil
}

func (tp *Tempolite) getWorkerCompensationID(queue string) (int, error) {
	counter, _ := tp.queuePoolCompensationCounter.LoadOrStore(queue, &atomic.Int64{})
	return int(counter.(*atomic.Int64).Add(1)), nil
}

func (tp *Tempolite) initQueueCounters(queue string) {
	tp.queuePoolWorkflowCounter.Store(queue, &atomic.Int64{})
	tp.queuePoolActivityCounter.Store(queue, &atomic.Int64{})
	tp.queuePoolSideEffectCounter.Store(queue, &atomic.Int64{})
	tp.queuePoolTransactionCounter.Store(queue, &atomic.Int64{})
	tp.queuePoolCompensationCounter.Store(queue, &atomic.Int64{})
}

func (tp *Tempolite) AddWorkerWorkflow(queue string) error {
	pool, err := tp.getWorkflowPoolQueue(queue)
	if err != nil {
		return err
	}
	id, err := tp.getWorkerWorkflowID(queue)
	if err != nil {
		return err
	}
	pool.AddWorker(workflowWorker{id: id, tp: tp})
	return nil
}

func (tp *Tempolite) AddWorkerActivity(queue string) error {
	pool, err := tp.getActivityPoolQueue(queue)
	if err != nil {
		return err
	}

	id, err := tp.getWorkerActivityID(queue)
	if err != nil {
		return err
	}
	pool.AddWorker(activityWorker{id: id, tp: tp})
	return nil
}

func (tp *Tempolite) AddWorkerSideEffect(queue string) error {
	pool, err := tp.getSideEffectPoolQueue(queue)
	if err != nil {
		return err
	}
	id, err := tp.getWorkerSideEffectID(queue)
	if err != nil {
		return err
	}
	pool.AddWorker(sideEffectWorker{id: id, tp: tp})
	return nil
}

func (tp *Tempolite) AddWorkerTransaction(queue string) error {
	pool, err := tp.getTransactionPoolQueue(queue)
	if err != nil {
		return err
	}
	id, err := tp.getWorkerTransactionID(queue)
	if err != nil {
		return err
	}
	pool.AddWorker(transactionWorker{id: id, tp: tp})
	return nil
}

func (tp *Tempolite) AddWorkerCompensation(queue string) error {
	pool, err := tp.getCompensationPoolQueue(queue)
	if err != nil {
		return err
	}
	id, err := tp.getWorkerCompensationID(queue)
	if err != nil {
		return err
	}
	pool.AddWorker(compensationWorker{id: id, tp: tp})
	return nil
}

func (tp *Tempolite) RemoveWorkerWorkflowByID(queue string, id int) error {
	pool, err := tp.getWorkflowPoolQueue(queue)
	if err != nil {
		return err
	}
	pool.ForceClose()
	return pool.RemoveWorker(id)
}

func (tp *Tempolite) RemoveWorkerActivityByID(queue string, id int) error {
	pool, err := tp.getActivityPoolQueue(queue)
	if err != nil {
		return err
	}
	return pool.RemoveWorker(id)
}

func (tp *Tempolite) RemoveWorkerSideEffectByID(queue string, id int) error {
	pool, err := tp.getSideEffectPoolQueue(queue)
	if err != nil {
		return err
	}
	return pool.RemoveWorker(id)
}

func (tp *Tempolite) RemoveWorkerTransactionByID(queue string, id int) error {
	pool, err := tp.getTransactionPoolQueue(queue)
	if err != nil {
		return err
	}
	return pool.RemoveWorker(id)
}

func (tp *Tempolite) RemoveWorkerCompensationByID(queue string, id int) error {
	pool, err := tp.getCompensationPoolQueue(queue)
	if err != nil {
		return err
	}
	return pool.RemoveWorker(id)
}

func (tp *Tempolite) ListWorkersWorkflow(queue string) ([]int, error) {
	pool, err := tp.getWorkflowPoolQueue(queue)
	if err != nil {
		return nil, err
	}
	return pool.GetWorkerIDs(), nil
}

func (tp *Tempolite) ListWorkersActivity(queue string) ([]int, error) {
	pool, err := tp.getActivityPoolQueue(queue)
	if err != nil {
		return nil, err
	}
	return pool.GetWorkerIDs(), nil
}

func (tp *Tempolite) ListWorkersSideEffect(queue string) ([]int, error) {
	pool, err := tp.getSideEffectPoolQueue(queue)
	if err != nil {
		return nil, err
	}
	return pool.GetWorkerIDs(), nil
}

func (tp *Tempolite) ListWorkersTransaction(queue string) ([]int, error) {
	pool, err := tp.getTransactionPoolQueue(queue)
	if err != nil {
		return nil, err
	}
	return pool.GetWorkerIDs(), nil
}

func (tp *Tempolite) ListWorkersCompensation(queue string) ([]int, error) {
	pool, err := tp.getCompensationPoolQueue(queue)
	if err != nil {
		return nil, err
	}
	return pool.GetWorkerIDs(), nil
}

func (tp *Tempolite) ListQueues() []string {
	queues := []string{}
	tp.queues.Range(func(key, _ interface{}) bool {
		queues = append(queues, key.(string))
		return true
	})
	return queues
}
