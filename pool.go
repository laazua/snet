package snet

import (
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrWorkerPoolFull   = errors.New("worker pool is full")
	ErrWorkerPoolClosed = errors.New("worker pool is closed")
)

// Task 任务接口
type Task interface {
	Execute()
}

// taskWrapper 任务包装器
type taskWrapper struct {
	task Task
}

// worker 工作协程
type worker struct {
	pool *WorkerPool
	task chan Task
}

// WorkerPool 协程池
type WorkerPool struct {
	workers     []*worker
	taskQueue   chan Task
	workerQueue chan *worker
	closed      int32
	wg          sync.WaitGroup
}

// NewWorkerPool 创建协程池
func NewWorkerPool(poolSize, queueSize int) *WorkerPool {
	pool := &WorkerPool{
		workers:     make([]*worker, poolSize),
		taskQueue:   make(chan Task, queueSize),
		workerQueue: make(chan *worker, poolSize),
	}

	// 创建工作协程
	for i := 0; i < poolSize; i++ {
		worker := &worker{
			pool: pool,
			task: make(chan Task),
		}
		pool.workers[i] = worker
		pool.workerQueue <- worker
		pool.wg.Add(1)

		go worker.run()
	}

	// 启动任务分发器
	go pool.dispatcher()

	return pool
}

// dispatcher 任务分发器
func (p *WorkerPool) dispatcher() {
	for task := range p.taskQueue {
		if atomic.LoadInt32(&p.closed) == 1 {
			return
		}

		// 获取可用工作协程
		worker := <-p.workerQueue
		worker.task <- task
	}
}

// worker.run 工作协程执行循环
func (w *worker) run() {
	defer w.pool.wg.Done()

	for task := range w.task {
		if task == nil {
			return
		}

		task.Execute()
		w.pool.workerQueue <- w
	}
}

// Submit 提交任务
func (p *WorkerPool) Submit(task Task) error {
	if atomic.LoadInt32(&p.closed) == 1 {
		return ErrWorkerPoolClosed
	}

	select {
	case p.taskQueue <- task:
		return nil
	default:
		return ErrWorkerPoolFull
	}
}

// Close 关闭协程池
func (p *WorkerPool) Close() {
	if atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		close(p.taskQueue)

		// 通知所有工作协程退出
		for _, worker := range p.workers {
			close(worker.task)
		}

		p.wg.Wait()
	}
}
