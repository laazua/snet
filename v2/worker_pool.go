package v2

import (
	"sync"
)

// WorkerPoolImpl 工作池实现
type WorkerPoolImpl struct {
	taskQueue chan func()
	workerNum int
	wg        sync.WaitGroup
	stopChan  chan struct{}
}

// NewWorkerPool 创建工作池
func NewWorkerPool(workerNum, queueSize int) WorkerPool {
	return &WorkerPoolImpl{
		taskQueue: make(chan func(), queueSize),
		workerNum: workerNum,
		stopChan:  make(chan struct{}),
	}
}

func (wp *WorkerPoolImpl) Start() {
	for i := 0; i < wp.workerNum; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

func (wp *WorkerPoolImpl) Stop() {
	close(wp.stopChan)
	wp.wg.Wait()
	close(wp.taskQueue)
}

func (wp *WorkerPoolImpl) Submit(task func()) {
	select {
	case wp.taskQueue <- task:
	case <-wp.stopChan:
	}
}

func (wp *WorkerPoolImpl) worker() {
	defer wp.wg.Done()

	for {
		select {
		case task := <-wp.taskQueue:
			if task != nil {
				task()
			}
		case <-wp.stopChan:
			return
		}
	}
}
