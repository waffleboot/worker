package worker

import (
	"context"
	"log"
	"runtime"
	"sync"
	"time"
)

type Job func(ctx context.Context)

type Scheduler struct {
	wg            *sync.WaitGroup
	cancellations []context.CancelFunc
	triggers      []chan struct{}
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		wg:            new(sync.WaitGroup),
		cancellations: make([]context.CancelFunc, 0),
		triggers:      make([]chan struct{}, 0),
	}
}

func (s *Scheduler) Add(ctx context.Context, j Job, interval time.Duration, active bool) (chan struct{}, chan bool) {
	ctx, cancel := context.WithCancel(ctx)
	s.cancellations = append(s.cancellations, cancel)

	triggerChannel := make(chan struct{})

	activeChannel := make(chan bool)

	s.triggers = append(s.triggers, triggerChannel)
	s.wg.Add(1)
	go s.process(ctx, j, interval, triggerChannel, activeChannel, active)
	return triggerChannel, activeChannel
}

func (s *Scheduler) TriggerAll() {
	for _, ch := range s.triggers {
		ch <- struct{}{}
	}
}

func (s *Scheduler) Stop() {
	for _, cancel := range s.cancellations {
		cancel()
	}
	s.wg.Wait()
}

func (s *Scheduler) Run(j Job, ctx context.Context, active bool) {
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("panic scheduled job: %v\n%s\n", r, buf)
		}
	}()
	if active {
		j(ctx)
	}
}

func (s *Scheduler) process(ctx context.Context, j Job, interval time.Duration, trigger chan struct{}, activeCh chan bool, active bool) {
	ticker := time.NewTicker(interval)
	first := make(chan bool, 1)
	first <- true

	for {
		select {
		case active = <-activeCh:
		case <-first:
			s.Run(j, ctx, active)
		case <-ticker.C:
			s.Run(j, ctx, active)
		case <-trigger:
			s.Run(j, ctx, active)
			<-ticker.C
		case <-ctx.Done():
			s.wg.Done()
			ticker.Stop()
			close(first)
			return
		}
	}
}
