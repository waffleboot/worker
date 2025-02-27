package main

import (
	"context"
	"fmt"
	"github.com/ahmetask/worker"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

var (
	Pool *worker.Pool
)

type Job1 struct {
	Id int
	Wg *sync.WaitGroup
}

/*implement work interface*/
func (j *Job1) Do() {
	time.Sleep(1 * time.Second)
	log.Println(fmt.Sprintf("Job1 Finished:%d", j.Id))
	j.Wg.Done()
}

type Job2 struct {
	Id string
	Wg *sync.WaitGroup
}

/*implement work interface*/
func (j *Job2) Do() {
	time.Sleep(2 * time.Second)
	log.Println(fmt.Sprintf("Job2 Finished:%s", j.Id))
	j.Wg.Done()
}

func ScheduledJob1(ctx context.Context) {
	log.Println("ScheduledJob1 Started")

	wg := &sync.WaitGroup{}
	jobs := []int{1, 2, 3, 4, 5, 6, ctx.Value("foo1").(int)}

	for _, j := range jobs {

		//Try to Enqueue the job with given timeout else return false
		queued := Pool.EnqueueWithTimeout(&Job1{
			Id: j,
			Wg: wg,
		}, 1*time.Second)
		if queued {
			wg.Add(1)
		}
		log.Println(fmt.Sprintf("Enqueue Job Finished: %d Result:%v", j, queued))
	}
	wg.Wait()
	log.Println("ScheduledJob1 Finished")
}

func ScheduledJob2(ctx context.Context) {
	log.Println("ScheduledJob2 Started")
	wg := &sync.WaitGroup{}
	jobs := []string{"s2-a", "s2-b", "s2-c", "s2-d", "s2-e", "s2-f"}

	for _, j := range jobs {

		// Try to enqueue the job but this call returns false if queue is full
		queued := Pool.Enqueue(&Job2{
			Id: j,
			Wg: wg,
		})

		if queued {
			wg.Add(1)
		}

		log.Println(fmt.Sprintf("Enqueue Job Finished:%s Result:%v", j, queued))
	}

	wg.Wait()
	log.Println("ScheduledJob2 Finished")
}

func ScheduledJob3(ctx context.Context) {
	log.Println("ScheduledJob3 Started")
	wg := &sync.WaitGroup{}
	jobs := []string{"s3-a", "s3-b", "s3-c", "s3-d", "s3-e", "s3-f"}
	for _, j := range jobs {

		wg.Add(1)

		// Blocking call wait for available workers if workers are fast enough use this
		Pool.Submit(&Job2{
			Id: j,
			Wg: wg,
		})
	}

	wg.Wait()

	log.Println("ScheduledJob3 Finished")
}

func main() {
	Pool = worker.NewWorkerPool(
		worker.WithMaxWorkers(4),
		worker.WithJobQueueCapacity(4))
	Pool.Start()

	scheduler := worker.NewScheduler()

	context1 := context.Background()
	c1 := context.WithValue(context1, "foo1", 7)
	trigger1, active := scheduler.Add(c1, ScheduledJob1, time.Second*10, false)

	// Soft Start/Stop
	time.AfterFunc(10*time.Second, func() {
		fmt.Println("ScheduledJob1 Starting")
		active <- true // or false for stopping the scheduler
	})

	//Specific Trigger
	time.AfterFunc(15*time.Second, func() {
		fmt.Println("ScheduledJob1 Triggered")
		trigger1 <- struct{}{}
	})

	context2 := context.Background()
	scheduler.Add(context2, ScheduledJob2, time.Minute*1, true)

	context3 := context.Background()
	scheduler.Add(context3, ScheduledJob3, time.Minute*1, true)

	//Manual Trigger
	time.AfterFunc(30*time.Second, func() {
		fmt.Println("Trigger All")
		scheduler.TriggerAll()
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	scheduler.Stop()
	Pool.Stop()
}
