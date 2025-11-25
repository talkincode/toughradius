package golimit

type GoLimit struct {
	ch chan int
}

func NewGoLimit(max int) *GoLimit {
	return &GoLimit{ch: make(chan int, max)}
}

func (g *GoLimit) Add() {
	g.ch <- 1
}

func (g *GoLimit) Done() {
	<-g.ch
}
