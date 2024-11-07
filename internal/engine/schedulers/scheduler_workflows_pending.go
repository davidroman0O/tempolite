package schedulers

import "github.com/davidroman0O/tempolite/internal/persistence/repository"

type SchedulerWorkflowsPending struct {
	db repository.Repository
}

func (s SchedulerWorkflowsPending) Tick() {

}
