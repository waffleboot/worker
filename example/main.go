package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/ahmetask/worker"
)

var Pool *worker.Pool

type Job1 struct {
	id int
	wg *sync.WaitGroup
}

/*implement work interface*/
func (j *Job1) Do() {
	defer j.wg.Done()
	time.Sleep(1 * time.Second)
	log.Println(fmt.Sprintf("Job1 Finished:%d", j.id))
}

type Job2 struct {
	id string
	wg *sync.WaitGroup
}

/*implement work interface*/
func (j *Job2) Do() {
	defer j.wg.Done()
	time.Sleep(2 * time.Second)
	log.Println(fmt.Sprintf("Job2 Finished:%s", j.id))
}

func EnqueueWithTimeout(ctx context.Context) {
	log.Println("EnqueueWithTimeout Started")
	wg := &sync.WaitGroup{}
	jobs := []int{1, 2, 3, 4, 5, 6, ctx.Value("foo1").(int)}
	for _, j := range jobs {
		//Try to Enqueue the job with given timeout else return false
		queued := Pool.EnqueueWithTimeout(&Job1{
			id: j,
			wg: wg,
		}, 1*time.Second)
		if queued {
			wg.Add(1)
		}
		log.Println(fmt.Sprintf("Enqueue Job Finished: %d Result:%v", j, queued))
	}
	wg.Wait()
	log.Println("EnqueueWithTimeout Finished")
}

func Enqueue(ctx context.Context) {
	log.Println("Enqueue Started")
	wg := &sync.WaitGroup{}
	jobs := []string{"s2-a", "s2-b", "s2-c", "s2-d", "s2-e", "s2-f"}
	for _, j := range jobs {
		// Try to enqueue the job but this call returns false if queue is full
		queued := Pool.Enqueue(&Job2{
			id: j,
			wg: wg,
		})
		if queued {
			wg.Add(1)
		}
		log.Println(fmt.Sprintf("Enqueue Job Finished:%s Result:%v", j, queued))
	}
	wg.Wait()
	log.Println("Enqueue Finished")
}

func Submit(ctx context.Context) {
	log.Println("Submit Started")
	wg := &sync.WaitGroup{}
	jobs := []string{"s3-a", "s3-b", "s3-c", "s3-d", "s3-e", "s3-f"}
	for _, j := range jobs {
		wg.Add(1)
		// Blocking call wait for available workers if workers are fast enough use this
		Pool.Submit(&Job2{
			id: j,
			wg: wg,
		})
	}
	wg.Wait()
	log.Println("Submit Finished")
}

func main() {

	Pool = worker.NewWorkerPool(
		worker.WithMaxWorkers(4),
		worker.WithJobQueueCapacity(4))

	Pool.Start()

	scheduler := worker.NewScheduler()

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		<-ch
		scheduler.Stop()
		Pool.Stop()
		os.Exit(1)
	}()

	ctx := context.WithValue(context.Background(), "foo1", 7)

	trigger, active := scheduler.Add(ctx, EnqueueWithTimeout, time.Second*10, false)

	// Soft Start/Stop
	time.AfterFunc(10*time.Second, func() {
		fmt.Println("ScheduledJob1 Starting")
		active <- true // or false for stopping the scheduler
	})

	//Specific Trigger
	time.AfterFunc(15*time.Second, func() {
		fmt.Println("ScheduledJob1 Triggered")
		trigger <- struct{}{}
	})

	scheduler.Add(context.Background(), Enqueue, time.Minute*1, true)

	scheduler.Add(context.Background(), Submit, time.Minute*1, true)

	//Manual Trigger
	time.AfterFunc(30*time.Second, func() {
		fmt.Println("Trigger All")
		scheduler.TriggerAll()
	})

	scheduler.Stop()
	Pool.Stop()
}
