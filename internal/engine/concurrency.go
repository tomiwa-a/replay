package engine

import (
	"sync"

	"github.com/replay/replay/internal/workflow"
)

// WorkerPool manages the parallel execution of multiple workflows.
type WorkerPool struct {
	concurrency int
	tasks       chan *workflow.Workflow
	wg          sync.WaitGroup
	engine      *Engine
}

// NewWorkerPool creates a new pool with the specified N workers.
func NewWorkerPool(n int, engine *Engine) *WorkerPool {
	return &WorkerPool{
		concurrency: n,
		tasks:       make(chan *workflow.Workflow, 512),
		engine:      engine,
	}
}

// Start spawns the worker goroutines.
func (p *WorkerPool) Start() {
	for i := 0; i < p.concurrency; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

func (p *WorkerPool) worker() {
	defer p.wg.Done()
	// Each worker logic will go here
	for wf := range p.tasks {
		_ = p.engine.Run(wf)
	}
}

// Submit adds a workflow to the execution queue.
func (p *WorkerPool) Submit(wf *workflow.Workflow) {
	p.tasks <- wf
}

// Wait closes the queue and waits for all workers to finish.
func (p *WorkerPool) Wait() {
	close(p.tasks)
	p.wg.Wait()
}
