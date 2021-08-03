package worker

import (
	"sync"
	"time"
)

type Pool struct {
	singleJob         chan Work
	internalQueue     chan Work
	readyPool         chan chan Work //boss says hey i have a new job at my desk workers who available can get it in this way he does not have to ask current status of workers
	workers           []*worker
	dispatcherStopped sync.WaitGroup
	workersStopped    *sync.WaitGroup
	quit              chan struct{}
}

func NewWorkerPool(opts ...Opts) *Pool {

	cfg := buildWorkerPoolConfig(opts...)

	workersStopped := sync.WaitGroup{}

	readyPool := make(chan chan Work, cfg.maxWorkers)
	workers := make([]*worker, cfg.maxWorkers, cfg.maxWorkers)

	// create workers
	for i := 0; i < cfg.maxWorkers; i++ {
		workers[i] = NewWorker(i+1, readyPool, &workersStopped)
	}

	return &Pool{
		internalQueue:     make(chan Work, cfg.jobQueueCapacity),
		singleJob:         make(chan Work),
		readyPool:         readyPool,
		workers:           workers,
		dispatcherStopped: sync.WaitGroup{},
		workersStopped:    &workersStopped,
		quit:              make(chan struct{}),
	}
}

func (p *Pool) Start() {
	//tell workers to get ready
	for _, w := range p.workers {
		w.Start()
	}
	// open factory
	go p.dispatch()
}

func (p *Pool) Stop() {
	p.quit <- struct{}{}
	p.dispatcherStopped.Wait()
}

func (p *Pool) dispatch() {
	//open factory gate
	p.dispatcherStopped.Add(1)
	for {
		select {
		case job := <-p.singleJob:
			workerXChannel := <-p.readyPool //free worker x founded
			workerXChannel <- job           // here is your job worker x
		case job := <-p.internalQueue:
			workerXChannel := <-p.readyPool //free worker x founded
			workerXChannel <- job           // here is your job worker x
		case <-p.quit:
			// free all workers
			for i := 0; i < len(p.workers); i++ {
				p.workers[i].Stop()
			}
			// wait for all workers to finish their job
			p.workersStopped.Wait()
			//close factory gate
			p.dispatcherStopped.Done()
			return
		}
	}
}

/*This is blocking if all workers are busy*/
func (q *Pool) Submit(job Work) {
	// daily - fill the board with new works
	q.singleJob <- job
}

/*Tries to enqueue but fails if queue is full*/
func (q *Pool) Enqueue(job Work) bool {
	select {
	case q.internalQueue <- job:
		return true
	default:
		return false
	}
}

/*try to enqueue but fails if timeout occurs*/
func (q *Pool) EnqueueWithTimeout(job Work, timeout time.Duration) bool {
	if timeout <= 0 {
		timeout = 1 * time.Second
	}

	ch := make(chan bool)
	t := time.AfterFunc(timeout, func() { ch <- false })
	defer func() {
		t.Stop()
		close(ch)
	}()

	for {
		select {
		case q.internalQueue <- job:
			return true
		case <-ch:
			return false
		}
	}
}
