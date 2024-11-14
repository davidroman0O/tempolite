package tempolite

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/comfylite3"
	"github.com/davidroman0O/tempolite/internal/clock"
	"github.com/davidroman0O/tempolite/internal/engine/orchestrator"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/internal/types"
	"github.com/davidroman0O/tempolite/pkg/logs"
)

type tempoliteConfig struct {
	path            *string
	destructive     bool
	logger          logs.Logger
	defaultLogLevel logs.Level
	queues          []queueConfig
}

type tempoliteOption func(*tempoliteConfig)

func WithDefaultLogLevel(level logs.Level) tempoliteOption {
	return func(c *tempoliteConfig) {
		c.defaultLogLevel = level
	}
}

func WithQueueConfig(queue queueConfig) tempoliteOption {
	return func(c *tempoliteConfig) {
		c.queues = append(c.queues, queue)
	}
}

func WithLogger(logger logs.Logger) tempoliteOption {
	return func(c *tempoliteConfig) {
		c.logger = logger
	}
}

func WithPath(path string) tempoliteOption {
	return func(c *tempoliteConfig) {
		c.path = &path
	}
}

func WithMemory() tempoliteOption {
	return func(c *tempoliteConfig) {
		c.path = nil
	}
}

func WithDestructive() tempoliteOption {
	return func(c *tempoliteConfig) {
		c.destructive = true
	}
}

type queueConfig struct {
	Name                string
	WorkflowWorkers     int
	ActivityWorkers     int
	SideEffectWorkers   int
	TransactionWorkers  int
	CompensationWorkers int
}

type queueOption func(*queueConfig)

func WithWorkflowWorkers(n int) queueOption {
	return func(c *queueConfig) {
		c.WorkflowWorkers = n
	}
}

func WithActivityWorkers(n int) queueOption {
	return func(c *queueConfig) {
		c.ActivityWorkers = n
	}
}

func WithSideEffectWorkers(n int) queueOption {
	return func(c *queueConfig) {
		c.SideEffectWorkers = n
	}
}

func WithTransactionWorkers(n int) queueOption {
	return func(c *queueConfig) {
		c.TransactionWorkers = n
	}
}

func WithCompensationWorkers(n int) queueOption {
	return func(c *queueConfig) {
		c.CompensationWorkers = n
	}
}

func NewQueue(name string, opts ...queueOption) queueConfig {
	c := queueConfig{
		Name:                name,
		WorkflowWorkers:     1,
		ActivityWorkers:     1,
		SideEffectWorkers:   1,
		TransactionWorkers:  1,
		CompensationWorkers: 1,
	}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

type Tempolite struct {
	mu sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	// technically, for now, just one orchestrator
	orchestrator *orchestrator.Orchestrator

	comfy  *comfylite3.ComfyDB
	client *ent.Client

	clock *clock.FuncClock
}

func New(ctx context.Context, builder registry.RegistryBuildFn, opts ...tempoliteOption) (*Tempolite, error) {
	cfg := tempoliteConfig{
		defaultLogLevel: logs.LevelError,
	}

	for _, opt := range opts {
		opt(&cfg)
	}
	var err error

	if cfg.logger != nil {
		logs.NewLogger(cfg.logger)
	} else {
		logs.Initialize(cfg.defaultLogLevel)
	}

	ctx, cancel := context.WithCancel(ctx)

	logs.Debug(ctx, "Creating Tempolite")
	tp := &Tempolite{
		ctx:    ctx,
		cancel: cancel,
	}

	logs.Debug(ctx, "Creating db clients")
	// create comfy and ent client
	if err := tp.createClient(cfg); err != nil {
		cancel()
		logs.Error(ctx, "Error creating db clients", "error", err)
		return nil, err
	}

	// logs.Debug(ctx, "Creating engine")
	// if tp.engine, err = engine.New(ctx, builder, tp.client); err != nil {
	// 	logs.Error(ctx, "Error creating engine", "error", err)
	// 	return nil, err
	// }

	db := repository.NewRepository(ctx, tp.client)

	tx, err := db.Tx()
	if err != nil {
		logs.Error(ctx, "Error creating transaction", "error", err)
		return nil, err
	}

	if _, err = db.Queues().Create(tx, "default"); err != nil {
		tx.Rollback()
		logs.Error(ctx, "Error creating queue", "error", err)
	} else {
		if err = tx.Commit(); err != nil {
			logs.Error(ctx, "Error committing transaction", "error", err)
			return nil, err
		}
	}

	reg, err := builder()
	if err != nil {
		logs.Error(ctx, "Error creating registry", "error", err)
		return nil, err
	}

	if tp.orchestrator, err = orchestrator.New(ctx, db, reg); err != nil {
		logs.Error(ctx, "Error creating orchestrator", "error", err)
		return nil, err
	}

	tp.clock = clock.NewFuncClock(time.Microsecond*10, func(err error) {
		logs.Error(ctx, "Error in clock", "error", err)
	})

	// tp.clock.AddFunc("orchestrator", tp.orchestrator.Tick, clock.WithFuncCleanup(tp.orchestrator.Stop))
	// tp.clock.Start()

	defer logs.Debug(ctx, "Tempolite created")

	return tp, nil
}

func (tp *Tempolite) createClient(cfg tempoliteConfig) error {

	optsComfy := []comfylite3.ComfyOption{}

	var firstTime bool
	if cfg.path != nil {
		logs.Debug(tp.ctx, "Database got a path", "path", *cfg.path)
		logs.Debug(tp.ctx, "Checking if first time or not", "path", *cfg.path)
		info, err := os.Stat(*cfg.path)
		if err == nil && !info.IsDir() {
			firstTime = false
		} else {
			firstTime = true
		}
		logs.Debug(tp.ctx, "Fist timer check", "firstTime", firstTime)
		if cfg.destructive {
			logs.Debug(tp.ctx, "Destructive option triggered", "firstTime", firstTime)
			if err := os.Remove(*cfg.path); err != nil {
				if !os.IsNotExist(err) {
					logs.Error(tp.ctx, "Error removing file", "error", err)
					return err
				}
			}
		}
		logs.Debug(tp.ctx, "Creating directory recursively if necessary", "path", *cfg.path)
		if err := os.MkdirAll(filepath.Dir(*cfg.path), os.ModePerm); err != nil {
			logs.Error(tp.ctx, "Error creating directory", "error", err)
			return err
		}
		optsComfy = append(optsComfy, comfylite3.WithPath(*cfg.path))
	} else {
		logs.Debug(tp.ctx, "Memory database option")
		optsComfy = append(optsComfy, comfylite3.WithMemory())
		firstTime = true
	}

	logs.Debug(tp.ctx, "Opening/Creating database")
	comfy, err := comfylite3.New(optsComfy...)
	if err != nil {
		logs.Error(tp.ctx, "Error opening/creating database", "error", err)
		return err
	}

	//	TODO: think how should I allow sql.DB option
	db := comfylite3.OpenDB(
		comfy,
		comfylite3.WithOption("_fk=1"),
		comfylite3.WithOption("cache=shared"),
		comfylite3.WithOption("mode=rwc"),
		comfylite3.WithForeignKeys(),
	)

	tp.comfy = comfy
	tp.client = ent.NewClient(ent.Driver(sql.OpenDB(dialect.SQLite, db)))

	if firstTime || (cfg.destructive && cfg.path != nil) {
		logs.Debug(tp.ctx, "Creating schema")
		if err = tp.client.Schema.Create(tp.ctx); err != nil {
			logs.Error(tp.ctx, "Error creating schema", "error", err)
			return err
		}
	}

	return nil
}

func (tp *Tempolite) Shutdown() error {
	<-time.After(time.Second)
	tp.orchestrator.Wait()
	logs.Debug(tp.ctx, "Shutting down Tempolite")
	tp.mu.Lock()
	defer tp.mu.Unlock()

	defer logs.Debug(tp.ctx, "Tempolite shutdown complete")

	// tp.clock.Stop()

	// if err := tp.engine.Shutdown(); err != nil {
	// 	if err != context.Canceled {
	// 		logs.Error(tp.ctx, "Error shutting down engine", "error", err)
	// 		return err
	// 	}
	// }

	tp.cancel()

	if err := tp.client.Close(); err != nil {
		logs.Error(tp.ctx, "Error closing client", "error", err)
		return err
	}

	if err := tp.comfy.Close(); err != nil {
		logs.Error(tp.ctx, "Error closing comfy", "error", err)
		return err
	}

	return nil
}

func (tp *Tempolite) Workflow(workflowFunc interface{}, opts types.WorkflowOptions, params ...any) (types.WorkflowID, error) {
	logs.Debug(tp.ctx, "Creating workflow")
	return tp.orchestrator.Workflow(workflowFunc, opts, params...)
}
