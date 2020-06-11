package goob

import (
	"context"
	"runtime"
	"sync"
	"time"
)

// Event interface
type Event interface{}

type null struct{}

// Observable hub
type Observable struct {
	lock sync.Mutex

	subscribers map[*subscriber]null
}

// Subscriber object
type subscriber struct {
	lock   sync.Mutex
	buffer []Event
}

// New observable instance
func New() *Observable {
	ob := &Observable{
		subscribers: map[*subscriber]null{},
	}

	return ob
}

// Publish message to the queue
func (ob *Observable) Publish(e Event) {
	ob.lock.Lock()
	defer ob.lock.Unlock()

	for s := range ob.subscribers {
		s.lock.Lock()
		s.buffer = append(s.buffer, e)
		s.lock.Unlock()
	}
}

// Count of the subscribers
func (ob *Observable) Count() int {
	return len(ob.subscribers)
}

// Subscribe message, the ctx is used to cancel the subscription
func (ob *Observable) Subscribe(ctx context.Context) chan Event {
	s := &subscriber{
		buffer: []Event{},
	}

	ob.lock.Lock()
	ob.subscribers[s] = null{}
	ob.lock.Unlock()

	ch := make(chan Event)

	go func() {
		defer func() {
			close(ch)

			ob.lock.Lock()
			delete(ob.subscribers, s)
			ob.lock.Unlock()
		}()

		ticked := true
		wait := time.Nanosecond
		for {
			if ctx.Err() != nil {
				break
			}

			s.lock.Lock()
			list := s.buffer
			s.buffer = []Event{}
			s.lock.Unlock()

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
