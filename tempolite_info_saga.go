package tempolite

import (
	"errors"
	"runtime"
	"time"

	"github.com/davidroman0O/go-tempolite/ent/saga"
)

type SagaInfo[T Identifier] struct {
	tp     *Tempolite[T]
	SagaID SagaID
	err    error
}

func (i *SagaInfo[T]) Get(output ...interface{}) error {
	if i.err != nil {
		return i.err
	}

	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()

	for {
		select {
		case <-i.tp.ctx.Done():
			return i.tp.ctx.Err()
		case <-ticker.C:
			sagaEntity, err := i.tp.client.Saga.Query().
				Where(saga.IDEQ(i.SagaID.String())).
				WithSteps().
				Only(i.tp.ctx)
			if err != nil {
				return err
			}

			switch sagaEntity.Status {
			case saga.StatusCompleted:
				return nil
			case saga.StatusFailed:
				return errors.New(sagaEntity.Error)
			case saga.StatusCompensated:
				return errors.New("saga was compensated")
			case saga.StatusPending, saga.StatusRunning:
				runtime.Gosched()
				continue
			}
			runtime.Gosched()
		}
	}
}
