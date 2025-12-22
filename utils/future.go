package utils

type Signal struct {
	ch chan string
}

func NewSignal() *Signal {
	return &Signal{ch: make(chan string, 1)}
}

func (s *Signal) Wait() string {
	return <-s.ch
}

func (s *Signal) Signal(response string) {
	s.ch <- response
}
