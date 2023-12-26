package main

import (
	"fmt"
	"time"
)

type MyJob struct {
	ID int
}

func (job MyJob) RunTask(req interface{}) {
	fmt.Println("job-", job.ID)
	time.Sleep(500 * time.Millisecond)
}
func main() {
	jobSize := 100
	workerPool := NewWorkerPool(100, jobSize)
	workerPool.Start()
	for i := 0; i < 100; i++ {
		task := MyJob{ID: i}
		workerPool.Submit(task)
	}
	//// 阻塞主线程
	//time.Sleep(2 * time.Second)
	//fmt.Println("runtime.NumGoroutine() :", runtime.NumGoroutine())

	workerPool.Shutdown()
	fmt.Println("all done")
}
