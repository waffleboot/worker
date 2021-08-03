package worker

import (
	"log"
	"runtime"
)

type worker struct {
	id      int
	work    chan Work
	manager *workers
}

func NewWorker(id int, manager *workers) *worker {
	return &worker{
		id:      id,
		manager: manager,
		work:    make(chan Work),
	}
}

func (w *worker) Process(work Work) {
	//Do the work
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("panic running process: %v\n%s\n", r, buf)
		}
	}()
	work.Do()
}

func (w *worker) Start() {
	go func() {
		w.manager.Started()
		defer w.manager.Stopped()
		for {
			w.manager.Submit(w.work) //hey i am ready to work on new job
			select {
			case work := <-w.work: // hey i am waiting for new job
				w.Process(work) // ok i am on it
			case <-w.manager.Done():
				return
			}
		}
	}()
}
