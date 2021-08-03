package worker

import "sync"

type workers struct {
	readyPool chan chan Work
	workers   []*worker
	done      chan struct{}
	wg        *sync.WaitGroup
}

func NewWorkers(maxWorkers int) *workers {
	return &workers{
		readyPool: make(chan chan Work, maxWorkers),
		workers:   make([]*worker, 0, maxWorkers),
		done:      make(chan struct{}),
		wg:        &sync.WaitGroup{},
	}
}

func (p *workers) Add(w *worker) {
	p.workers = append(p.workers, w)
}

func (w *workers) Start() {
	for _, w := range w.workers {
		w.Start()
	}
}

func (w *workers) Stop() {
	close(w.done)
	w.wg.Wait()
}

func (w *workers) Started() {
	w.wg.Add(1)
}

func (w *workers) Stopped() {
	w.wg.Done()
}

func (w *workers) Done() <-chan struct{} {
	return w.done
}

func (w *workers) Work() <-chan chan Work {
	return w.readyPool
}

func (w *workers) Submit(work chan Work) {
	w.readyPool <- work
}
