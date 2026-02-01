package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Task represents a unit of work to be executed.
// It takes a context.Context to support cancellation and timeouts.
type Task interface {
	Execute(ctx context.Context) error
}

// WorkerPool manages a pool of goroutines to execute tasks.
type WorkerPool struct {
	tasks    chan Task
	wg       sync.WaitGroup
	quit     chan struct{}
	workers  int
	taskName string
}

// NewWorkerPool creates and starts a new WorkerPool.
// workers: number of concurrent workers.
// buffer: size of the task buffer.
// taskName: a descriptive name for the type of tasks this pool handles (for logging).
func NewWorkerPool(workers, buffer int, taskName string) *WorkerPool {
	pool := &WorkerPool{
		tasks:    make(chan Task, buffer),
		quit:     make(chan struct{}),
		workers:  workers,
		taskName: taskName,
	}

	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker(i + 1)
	}
	return pool
}

// Submit adds a task to the pool.
// Returns an error if the pool is shutting down or the task channel is closed.
func (p *WorkerPool) Submit(task Task) error {
	select {
	case <-p.quit:
		return fmt.Errorf("worker pool '%s' is shutting down, cannot submit new tasks", p.taskName)
	case p.tasks <- task:
		return nil
	default:
		return fmt.Errorf("worker pool '%s' is busy, task channel buffer is full", p.taskName)
	}
}

// worker is a goroutine that continuously fetches and executes tasks.
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()
	log.Printf("Worker #%d for '%s' started.", id, p.taskName)

	for {
		select {
		case task, ok := <-p.tasks:
			if !ok {
				log.Printf("Worker #%d for '%s' stopping: task channel closed.", id, p.taskName)
				return
			}
			log.Printf("Worker #%d for '%s' executing task.", id, p.taskName)
			// Apply a timeout context for the task
			taskCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // 30s timeout for IO tasks
			err := task.Execute(taskCtx)
			cancel() // Release resources associated with the context

			if err != nil {
				log.Printf("Worker #%d for '%s' task execution failed: %v", id, p.taskName, err)
			} else {
				log.Printf("Worker #%d for '%s' task executed successfully.", id, p.taskName)
			}
		case <-p.quit:
			log.Printf("Worker #%d for '%s' stopping: graceful shutdown initiated.", id, p.taskName)
			return
		}
	}
}

// Shutdown initiates a graceful shutdown of the WorkerPool.
// It stops accepting new tasks and waits for existing tasks to complete
// within a specified timeout.
func (p *WorkerPool) Shutdown(timeout time.Duration) {
	log.Printf("Initiating graceful shutdown for worker pool '%s'.", p.taskName)

	// Signal to stop accepting new tasks
	close(p.quit)

	// Close the tasks channel after a brief delay to allow pending submits to complete
	// and signal workers to drain remaining tasks.
	// This approach avoids deadlocks if tasks are still being submitted during shutdown.
	time.AfterFunc(100*time.Millisecond, func() {
		close(p.tasks)
	})

	done := make(chan struct{})
	go func() {
		p.wg.Wait() // Wait for all workers to finish
		close(done)
	}()

	select {
	case <-done:
		log.Printf("Worker pool '%s' shutdown complete: all tasks processed.", p.taskName)
	case <-time.After(timeout):
		log.Printf("Worker pool '%s' shutdown timed out after %v: some tasks might not have completed.", p.taskName, timeout)
	}
}

// ListenForGracefulShutdown sets up a listener for OS signals (SIGINT, SIGTERM)
// to trigger a graceful shutdown of the worker pool.
func ListenForGracefulShutdown(pool *WorkerPool, shutdownTimeout time.Duration) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %s. Initiating graceful shutdown...", sig)
		pool.Shutdown(shutdownTimeout)
		// Optionally, exit the program after all pools have shut down
		// For now, let the main goroutine handle the overall application exit
	}()
}
