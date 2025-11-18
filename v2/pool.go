package v2

import (
	"errors"
	"sync"
)

// Task 任务类型
type Task func()

// workerPool 工作协程池
type workerPool struct {
	taskChan chan Task
	workers  int
	wg       sync.WaitGroup
	once     sync.Once
	closed   bool
	mu       sync.RWMutex
}

// NewWorkerPool 创建新的协程池
func NewWorkerPool(workers, queueSize int) Pool {
	if workers <= 0 {
		workers = 100
	}
	if queueSize <= 0 {
		queueSize = 1000
	}

	pool := &workerPool{
		taskChan: make(chan Task, queueSize),
		workers:  workers,
	}

	// 启动工作协程
	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}

	return pool
}

func (p *workerPool) worker() {
	defer p.wg.Done()

	for task := range p.taskChan {
		task()
	}
}

func (p *workerPool) Submit(task func()) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return errors.New("pool is closed")
	}

	select {
	case p.taskChan <- task:
		return nil
	default:
		return errors.New("task queue is full")
	}
}

func (p *workerPool) Release() {
	p.once.Do(func() {
		p.mu.Lock()
		p.closed = true
		p.mu.Unlock()

		close(p.taskChan)
		p.wg.Wait()
	})
}
