package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type worker struct {
	capacity int
	doneChan chan struct{}
	mutex    sync.Mutex
	taskLen  int
}

func newWorker() *worker {
	return &worker{
		capacity: 5,
		doneChan: make(chan struct{}),
	}
}

func (w *worker) executeTasks() {
	for i := 0; i < w.capacity; i++ {
		w.taskLen++
		fmt.Println("run task ", i+1)
		time.Sleep(time.Second)
		w.remove()
	}
}

func (w *worker) remove() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.taskLen--
	if w.taskLen == 0 {
		select {
		case w.doneChan <- struct{}{}:
		default:
			// If the channel is not ready to receive, handle it accordingly.
			fmt.Println("No receiver for doneChan yet")
		}
	}
}

func (w *worker) gracefulStop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	if w.taskLen == 0 {
		return
	}

	for {
		select {
		case <-w.doneChan:
			fmt.Println("All tasks have completed. proceed to terminating..")
			return
		case <-ticker.C:
			fmt.Printf("Waiting for %d tasks to complete\n", w.taskLen)
		}
	}

}

func main() {
	w := newWorker()
	stopChan := make(chan struct{})

	for i := 0; i < 3; i++ {
		go w.executeTasks()
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		stopChan <- struct{}{}
	}()

	<-stopChan
	w.gracefulStop()
}
