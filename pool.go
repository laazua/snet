package snet

import (
	"sync"
	"sync/atomic"
)

// WorkerPool 协程池
type WorkerPool struct {
	workers   int
	taskQueue chan func()
	wg        sync.WaitGroup
	closed    int32
}

// NewWorkerPool 创建协程池
func newWorkerPool(workers, queueSize int) *WorkerPool {
	pool := &WorkerPool{
		workers:   workers,
		taskQueue: make(chan func(), queueSize),
	}

	for range workers {
		pool.wg.Add(1)
		go pool.worker()
	}

	return pool
}

func (p *WorkerPool) worker() {
	defer p.wg.Done()

	for task := range p.taskQueue {
		task()
	}
}

// Submit 提交任务
func (p *WorkerPool) Submit(task func()) error {
	if atomic.LoadInt32(&p.closed) == 1 {
		return ErrWorkerPoolClosed
	}

	select {
	case p.taskQueue <- task:
		return nil
	default:
		return ErrWorkerPoolQueueFull
	}
}

// Close 关闭协程池
func (p *WorkerPool) Close() {
	if atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		close(p.taskQueue)
		p.wg.Wait()
	}
}
