package worker

import "strconv"

type options struct {
	maxWorkers       int
	jobQueueCapacity int
}

type Option interface {
	apply(*options)
}

type maxWorkersOption int

func WithMaxWorkers(maxWorkers int) Option {
	return maxWorkersOption(maxWorkers)
}

func (o maxWorkersOption) apply(opts *options) {
	opts.maxWorkers = int(o)
}

func (o maxWorkersOption) String() string {
	return strconv.Itoa(int(o))
}

type jobQueueCapacityOption int

func WithJobQueueCapacity(jobQueueCapacity int) Option {
	return jobQueueCapacityOption(jobQueueCapacity)
}

func (o jobQueueCapacityOption) apply(opts *options) {
	opts.jobQueueCapacity = int(o)
	if opts.jobQueueCapacity <= 0 {
		opts.jobQueueCapacity = 100
	}
}

func buildWorkerPoolOptions(opts ...Option) options {
	options := options{}
	for _, o := range opts {
		o.apply(&options)
	}
	return options
}
