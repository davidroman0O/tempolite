package tempolite

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type TempolitePoolOptions struct {
	MaxFileSize  int64  // in bytes
	MaxPageCount int64  // in pages
	BaseFolder   string // base folder path
	BaseName     string // base name for database files
}

type TempolitePool[T Identifier] struct {
	mu sync.RWMutex

	// Current active instance
	current     *Tempolite[T]
	currentPath string

	// Configuration
	maxFileSize   int64
	maxPageCount  int64
	baseFolder    string
	baseName      string
	registry      *Registry[T]
	tempoliteOpts []tempoliteOption

	// File cache
	latestDBCache atomic.Pointer[string]
	cacheMu       sync.RWMutex

	// Coordination
	ctx         context.Context
	cancel      context.CancelFunc
	needRotate  atomic.Bool
	monitorDone chan struct{}

	rotateMu sync.RWMutex
}

func (p *TempolitePool[T]) findLatestDatabase() (string, error) {
	// Check cache first
	if latest := p.latestDBCache.Load(); latest != nil {
		// Verify the file still exists
		if _, err := os.Stat(*latest); err == nil {
			return *latest, nil
		}
	}

	// Cache miss or file doesn't exist, do the full lookup
	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()

	pattern := filepath.Join(p.baseFolder, fmt.Sprintf("%s_*.db", p.baseName))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to glob database files: %w", err)
	}

	if len(matches) == 0 {
		p.latestDBCache.Store(nil)
		return "", nil
	}

	sort.Slice(matches, func(i, j int) bool {
		iInfo, _ := os.Stat(matches[i])
		jInfo, _ := os.Stat(matches[j])
		return iInfo.ModTime().After(jInfo.ModTime())
	})

	p.latestDBCache.Store(&matches[0])
	return matches[0], nil
}

func (p *TempolitePool[T]) createNewDatabase() string {
	timestamp := time.Now().Unix()
	path := filepath.Join(p.baseFolder, fmt.Sprintf("%s_%d.db", p.baseName, timestamp))
	p.latestDBCache.Store(&path)
	return path
}

// Rest of the implementation remains exactly the same as the original
func NewTempolitePool[T Identifier](ctx context.Context, registry *Registry[T], poolOpts TempolitePoolOptions, tempoliteOpts ...tempoliteOption) (*TempolitePool[T], error) {
	if poolOpts.BaseFolder == "" {
		return nil, fmt.Errorf("base folder is required")
	}
	if poolOpts.BaseName == "" {
		return nil, fmt.Errorf("base name is required")
	}

	ctx, cancel := context.WithCancel(ctx)

	pool := &TempolitePool[T]{
		maxFileSize:   poolOpts.MaxFileSize,
		maxPageCount:  poolOpts.MaxPageCount,
		baseFolder:    poolOpts.BaseFolder,
		baseName:      poolOpts.BaseName,
		registry:      registry,
		tempoliteOpts: tempoliteOpts,
		ctx:           ctx,
		cancel:        cancel,
		monitorDone:   make(chan struct{}),
	}

	if err := os.MkdirAll(poolOpts.BaseFolder, 0755); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create base folder: %w", err)
	}

	if err := pool.initializeInstance(); err != nil {
		cancel()
		return nil, err
	}

	// Start size monitor
	go pool.monitorSize()

	return pool, nil
}

func (p *TempolitePool[T]) monitorSize() {
	defer close(p.monitorDone)
	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			if needsRotation, _ := p.checkLimits(); needsRotation {
				log.Println("DEBUG: Initiating rotation due to size limit.")
				if err := p.rotate(); err != nil {
					log.Printf("ERROR: Failed to rotate: %v", err)
				}
			}
		}
	}
}

func (p *TempolitePool[T]) checkLimits() (bool, error) {
	if p.current == nil {
		return false, nil
	}

	if p.maxFileSize > 0 {
		fileInfo, err := os.Stat(p.currentPath)
		if err != nil {
			return false, fmt.Errorf("failed to stat database file: %w", err)
		}
		if fileInfo.Size() >= p.maxFileSize {
			log.Printf("DEBUG: Size limit reached: %d >= %d", fileInfo.Size(), p.maxFileSize)
			return true, nil
		}
	}

	if p.maxPageCount > 0 {
		var pageCount int64
		row := p.current.db.QueryRow("PRAGMA page_count")
		if err := row.Scan(&pageCount); err != nil {
			return false, fmt.Errorf("failed to get page count: %w", err)
		}
		if pageCount >= p.maxPageCount {
			return true, nil
		}
	}

	return false, nil
}

func (p *TempolitePool[T]) initializeInstance() error {
	latestDB, err := p.findLatestDatabase()
	if err != nil {
		return err
	}

	var dbPath string
	if latestDB == "" {
		dbPath = p.createNewDatabase()
	} else {
		dbPath = latestDB
	}

	opts := append(p.tempoliteOpts, WithPath(dbPath))
	instance, err := New[T](p.ctx, p.registry, opts...)
	if err != nil {
		return fmt.Errorf("failed to create Tempolite instance: %w", err)
	}

	p.current = instance
	p.currentPath = dbPath
	return nil
}

func (p *TempolitePool[T]) rotate() error {
	p.rotateMu.Lock() // Acquire write lock to block API functions
	defer p.rotateMu.Unlock()

	// Wait for current tasks to complete
	p.mu.RLock()
	current := p.current
	p.mu.RUnlock()

	if err := current.Wait(); err != nil {
		return err
	}

	// Proceed with rotation
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.current != nil {
		p.current.Close()
	}

	newPath := p.createNewDatabase()
	opts := append(p.tempoliteOpts, WithPath(newPath))

	newInstance, err := New[T](p.ctx, p.registry, opts...)
	if err != nil {
		return fmt.Errorf("failed to create new Tempolite instance: %w", err)
	}

	p.current = newInstance
	p.currentPath = newPath

	return nil
}

func (p *TempolitePool[T]) Close() {
	p.cancel()
	<-p.monitorDone

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.current != nil {
		p.current.Close()
	}
}

func (p *TempolitePool[T]) Wait() error {
	p.mu.RLock()
	current := p.current
	p.mu.RUnlock()

	if err := current.Wait(); err != nil {
		return err
	}

	if p.needRotate.Load() {
		p.needRotate.Store(false)
		return p.rotate()
	}

	return nil
}

/// API functions

func (p *TempolitePool[T]) Workflow(stepID T, workflowFunc interface{}, options tempoliteWorkflowOptions, params ...interface{}) *WorkflowInfo[T] {
	p.rotateMu.RLock() // Acquire read lock
	defer p.rotateMu.RUnlock()

	p.mu.RLock()
	current := p.current
	p.mu.RUnlock()

	return current.Workflow(stepID, workflowFunc, options, params...)
}

func (p *TempolitePool[T]) GetWorkflow(id WorkflowID) *WorkflowInfo[T] {
	p.rotateMu.RLock()
	defer p.rotateMu.RUnlock()

	p.mu.RLock()
	current := p.current
	p.mu.RUnlock()

	return current.GetWorkflow(id)
}

func (p *TempolitePool[T]) GetActivity(id ActivityID) (*ActivityInfo[T], error) {
	p.rotateMu.RLock()
	defer p.rotateMu.RUnlock()

	p.mu.RLock()
	current := p.current
	p.mu.RUnlock()

	return current.GetActivity(id)
}

func (p *TempolitePool[T]) GetVersion(workflowType, workflowID, changeID string, minSupported, maxSupported int) (int, error) {
	p.rotateMu.RLock()
	defer p.rotateMu.RUnlock()

	p.mu.RLock()
	current := p.current
	p.mu.RUnlock()

	return current.getOrCreateVersion(workflowType, workflowID, changeID, minSupported, maxSupported)
}

func (p *TempolitePool[T]) RetryWorkflow(workflowID WorkflowID) *WorkflowInfo[T] {
	p.rotateMu.RLock()
	defer p.rotateMu.RUnlock()

	p.mu.RLock()
	current := p.current
	p.mu.RUnlock()

	return current.RetryWorkflow(workflowID)
}

func (p *TempolitePool[T]) ReplayWorkflow(workflowID WorkflowID) *WorkflowInfo[T] {
	p.rotateMu.RLock()
	defer p.rotateMu.RUnlock()

	p.mu.RLock()
	current := p.current
	p.mu.RUnlock()

	return current.ReplayWorkflow(workflowID)
}

func (p *TempolitePool[T]) PauseWorkflow(id WorkflowID) error {
	p.rotateMu.RLock()
	defer p.rotateMu.RUnlock()

	p.mu.RLock()
	current := p.current
	p.mu.RUnlock()

	return current.PauseWorkflow(id)
}

func (p *TempolitePool[T]) ResumeWorkflow(id WorkflowID) error {
	p.rotateMu.RLock()
	defer p.rotateMu.RUnlock()

	p.mu.RLock()
	current := p.current
	p.mu.RUnlock()

	return current.ResumeWorkflow(id)
}

func (p *TempolitePool[T]) CancelWorkflow(id WorkflowID) error {
	p.rotateMu.RLock()
	defer p.rotateMu.RUnlock()

	p.mu.RLock()
	current := p.current
	p.mu.RUnlock()

	return current.CancelWorkflow(id)
}

func (p *TempolitePool[T]) ListPausedWorkflows() ([]WorkflowID, error) {
	p.rotateMu.RLock()
	defer p.rotateMu.RUnlock()

	p.mu.RLock()
	current := p.current
	p.mu.RUnlock()

	return current.ListPausedWorkflows()
}

func (p *TempolitePool[T]) PublishSignal(workflowID WorkflowID, stepID T, value interface{}) error {
	p.rotateMu.RLock()
	defer p.rotateMu.RUnlock()

	p.mu.RLock()
	current := p.current
	p.mu.RUnlock()

	return current.PublishSignal(workflowID, stepID, value)
}

func (p *TempolitePool[T]) GetLatestWorkflowExecution(originalWorkflowID WorkflowID) (WorkflowID, error) {
	p.rotateMu.RLock()
	defer p.rotateMu.RUnlock()

	p.mu.RLock()
	current := p.current
	p.mu.RUnlock()

	return current.GetLatestWorkflowExecution(originalWorkflowID)
}

func (p *TempolitePool[T]) IsActivityRegistered(longName HandlerIdentity) bool {
	p.rotateMu.RLock()
	defer p.rotateMu.RUnlock()

	p.mu.RLock()
	current := p.current
	p.mu.RUnlock()

	return current.IsActivityRegistered(longName)
}
