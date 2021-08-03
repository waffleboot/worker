package worker

type workerPoolConfig struct {
	maxWorkers       int
	jobQueueCapacity int
}

type Opts func(*workerPoolConfig)

func WithMaxWorkers(maxWorkers int) Opts {
	return func(cfg *workerPoolConfig) {
		cfg.maxWorkers = maxWorkers
	}
}

func WithJobQueueCapacity(jobQueueCapacity int) Opts {
	return func(cfg *workerPoolConfig) {
		cfg.jobQueueCapacity = jobQueueCapacity
	}
}

func buildWorkerPoolConfig(opts ...Opts) *workerPoolConfig {
	cfg := &workerPoolConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.jobQueueCapacity <= 0 {
		cfg.jobQueueCapacity = 100
	}
	return cfg
}
