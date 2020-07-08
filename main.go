package goob

import (
	"context"
	"sync"
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
	ctx    context.Context
	wait   chan null
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
		var wait chan null

		s.lock.Lock()
		s.buffer = append(s.buffer, e)
		wait = s.wait
		if wait != nil {
			s.wait = nil
		}
		s.lock.Unlock()

		if wait != nil {
			select {
			case <-s.ctx.Done():
			case wait <- null{}:
			}
		}
	}
}

// Count of the subscribers
func (ob *Observable) Count() int {
	return len(ob.subscribers)
}

// Subscribe message, the ctx is used to cancel the subscription
func (ob *Observable) Subscribe(ctx context.Context) <-chan Event {
	s := &subscriber{
		ctx:    ctx,
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

		for {
			if ctx.Err() != nil {
				return
			}

			var wait chan null

			s.lock.Lock()
			list := s.buffer
			s.buffer = []Event{}
			if len(list) == 0 {
				s.wait = make(chan null)
				wait = s.wait
			}
			s.lock.Unlock()

			if len(list) == 0 {
				select {
				case <-ctx.Done():
					return
				case <-wait:
				}
			}

			for _, e := range list {
				select {
				case <-ctx.Done():
					return
				case ch <- e:
				}
			}
		}
	}()

	return ch
}
