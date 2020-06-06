package goob

import (
	"context"
	"runtime"
	"time"
)

// Event interface
type Event interface{}

type null struct{}

// Observable hub
type Observable struct {
	subscribers map[*subscriber]null

	produce     chan Event
	subscribe   chan *subscriber
	consume     chan *subscriber
	unsubscribe chan *subscriber
}

// Subscriber object
type subscriber struct {
	buffer []Event
	list   chan []Event
}

// New observable instance
func New(ctx context.Context) *Observable {
	ob := &Observable{
		subscribers: map[*subscriber]null{},
		produce:     make(chan Event),
		subscribe:   make(chan *subscriber),
		consume:     make(chan *subscriber),
		unsubscribe: make(chan *subscriber),
	}

	go ob.start(ctx)

	return ob
}

// Publish message to the queue
func (ob *Observable) Publish(e Event) {
	ob.produce <- e
}

// Subscribe message
func (ob *Observable) Subscribe(ctx context.Context) chan Event {
	s := &subscriber{
		buffer: []Event{},
		list:   make(chan []Event),
	}
	ob.subscribe <- s

	ch := make(chan Event)

	go func() {
		ticked := true
		wait := time.Nanosecond
		for {
			if ctx.Err() != nil {
				ob.unsubscribe <- s
				close(ch)
				break
			}

			ob.consume <- s
			list := <-s.list

			if len(list) == 0 {
				// to reduce the overhead use the go scheduler first, then use the syscall sleep
				if ticked {
					ticked = false
					runtime.Gosched()
					continue
				}

				// backoff
				if wait < time.Millisecond {
					wait *= 2
				}
				time.Sleep(wait)
				continue
			}
			ticked = true
			wait = time.Nanosecond

			for _, e := range list {
				ch <- e
			}
		}
	}()

	return ch
}

func (ob *Observable) start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			break

		case e := <-ob.produce:
			for s := range ob.subscribers {
				s.buffer = append(s.buffer, e)
			}

		case s := <-ob.subscribe:
			ob.subscribers[s] = null{}

		case s := <-ob.consume:
			s.list <- s.buffer
			s.buffer = []Event{}

		case s := <-ob.unsubscribe:
			delete(ob.subscribers, s)
		}
	}
}
