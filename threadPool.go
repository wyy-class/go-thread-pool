package main

import (
	"fmt"
	"sync"
)

type Job interface {
	RunTask(req interface{})
}

type JobChan chan Job

type Worker struct {
	JobQueue JobChan
	Quit     chan struct{}
}

func NewWorker() Worker {
	return Worker{
		JobQueue: make(JobChan),
		Quit:     make(chan struct{}),
	}
}
func (w *Worker) Start(workerPool *WorkerPool) {
	workerPool.WorkerQueue <- w

	go func() {
		for {
			select {
			case job := <-w.JobQueue:
				job.RunTask(nil)
				workerPool.WorkerQueue <- w
				workerPool.wg.Done() // 在任务完成后减少 WaitGroup 计数
			case <-w.Quit:
				return
			}
		}
	}()
}

type WorkerPool struct {
	PoolSize    int
	JobQueue    JobChan
	WorkerQueue chan *Worker
	Quit        chan struct{}
	wg          sync.WaitGroup // 新增 WaitGroup 用于等待任务完成
}

func NewWorkerPool(poolSize, jobSize int) WorkerPool {
	return WorkerPool{
		PoolSize:    poolSize,
		JobQueue:    make(JobChan, jobSize),
		WorkerQueue: make(chan *Worker, poolSize),
		Quit:        make(chan struct{}),
	}
}
func (pool *WorkerPool) Start() {
	//启动所有worker
	for i := 0; i < pool.PoolSize; i++ {
		w := NewWorker()
		w.Start(pool)
	}
	go func() {
		for {
			select {
			case job, ok := <-pool.JobQueue:
				if !ok { //提前退出循环，会导致pool.Quit的死锁
					fmt.Println("task queue is empty.")
					<-pool.Quit //防止pool.Quit通道的发送操作因为没有接收者而被阻塞，从而导致死锁。
					return
				}
				worker := <-pool.WorkerQueue
				worker.JobQueue <- job
			case <-pool.Quit:
				return
			}
		}
	}()
}
func (pool *WorkerPool) Submit(job Job) {
	pool.wg.Add(1) // 每次提交任务增加 WaitGroup 计数
	pool.JobQueue <- job
}
func (pool *WorkerPool) Shutdown() {
	close(pool.JobQueue) // 关闭 JobQueue，不再接受新任务
	pool.wg.Wait()       // 等待所有任务完成

	close(pool.WorkerQueue)
	for worker := range pool.WorkerQueue {
		worker.Quit <- struct{}{}
	}
	pool.Quit <- struct{}{}
}
