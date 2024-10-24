package tempolite

import (
	"errors"
	"runtime"
	"time"

	"github.com/davidroman0O/tempolite/ent/saga"
)

type SagaInfo struct {
	tp     *Tempolite
	SagaID SagaID
	err    error
}

func (i *SagaInfo) Get(output ...interface{}) error {
	if i.err != nil {
		i.tp.logger.Error(i.tp.ctx, "SagaInfo.Get", "sagaID", i.SagaID, "error", i.err)
		return i.err
	}

	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()

	for {
		select {
		case <-i.tp.ctx.Done():
			i.tp.logger.Error(i.tp.ctx, "SagaInfo.Get: context done", "sagaID", i.SagaID)
			return i.tp.ctx.Err()
		case <-ticker.C:
			sagaEntity, err := i.tp.client.Saga.Query().
				Where(saga.IDEQ(i.SagaID.String())).
				WithSteps().
				Only(i.tp.ctx)
			if err != nil {
				i.tp.logger.Error(i.tp.ctx, "SagaInfo.Get: failed to query saga", "sagaID", i.SagaID, "error", err)
				return err
			}

			switch sagaEntity.Status {
			case saga.StatusCompleted:
				i.tp.logger.Debug(i.tp.ctx, "SagaInfo.Get: saga completed", "sagaID", i.SagaID)
				return nil
			case saga.StatusFailed:
				i.tp.logger.Debug(i.tp.ctx, "SagaInfo.Get: saga failed", "sagaID", i.SagaID, "error", errors.New(sagaEntity.Error))
				return errors.New(sagaEntity.Error)
			case saga.StatusCompensated:
				i.tp.logger.Debug(i.tp.ctx, "SagaInfo.Get: saga compensated", "sagaID", i.SagaID)
				return errors.New("saga was compensated")
			case saga.StatusPending, saga.StatusRunning:
				i.tp.logger.Debug(i.tp.ctx, "SagaInfo.Get: saga still running", "sagaID", i.SagaID)
				runtime.Gosched()
				continue
			}
			runtime.Gosched()
		}
	}
}
