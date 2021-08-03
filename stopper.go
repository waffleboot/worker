package worker

type stopper struct {
	signal   chan struct{}
	response chan struct{}
}

func NewStopper() *stopper {
	return &stopper{
		signal:   make(chan struct{}),
		response: make(chan struct{}),
	}
}

func (s *stopper) Signal() {
	close(s.signal)
}

func (s *stopper) Wait() {
	<-s.response
}

func (s *stopper) Reply() {
	close(s.response)
}

func (s *stopper) Done() chan struct{} {
	return s.signal
}
