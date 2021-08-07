package worker

import (
	"time"
)

type Pool struct {
	bufferedQueue chan Work
	singleQueue   chan Work
	workers       *workers
	*stopper
}

func NewWorkerPool(opts ...Option) *Pool {
	options := buildWorkerPoolOptions(opts...)
	workers := NewWorkers(options.maxWorkers)
	for i := 0; i < options.maxWorkers; i++ {
		workers.Add(NewWorker(i+1, workers))
	}
	return &Pool{
		bufferedQueue: make(chan Work, options.jobQueueCapacity),
		singleQueue:   make(chan Work),
		stopper:       NewStopper(),
		workers:       workers,
	}
}

func (p *Pool) Start() {
	p.workers.Start()
	go p.dispatch()
}

func (p *Pool) Stop() {
	p.stopper.Signal()
	p.stopper.Wait()
}

func (p *Pool) dispatch() {
	//open factory gate
	for {
		select {
		case job := <-p.singleQueue:
			workerXChannel := <-p.workers.Work() //free worker x founded
			workerXChannel <- job                // here is your job worker x
		case job := <-p.bufferedQueue:
			workerXChannel := <-p.workers.Work() //free worker x founded
			workerXChannel <- job                // here is your job worker x
		case <-p.stopper.Done():
			// free all workers
			p.workers.Stop()
			p.stopper.Reply()
			return
		}
	}
}

/*This is blocking if all workers are busy*/
func (p *Pool) Submit(job Work) {
	// daily - fill the board with new works
	p.singleQueue <- job
}

/*Tries to enqueue but fails if queue is full*/
func (p *Pool) Enqueue(job Work) bool {
	select {
	case p.bufferedQueue <- job:
		return true
	default:
		return false
	}
}

/*try to enqueue but fails if timeout occurs*/
func (p *Pool) EnqueueWithTimeout(job Work, timeout time.Duration) bool {
	if timeout <= 0 {
		timeout = 1 * time.Second
	}
	select {
	case p.bufferedQueue <- job:
		return true
	case <-time.After(timeout):
		return false
	}
}
