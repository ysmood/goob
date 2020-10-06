package goob

import (
	"context"
	"sync"
)

// Observable hub
type Observable struct {
	lock        *sync.Mutex
	subscribers map[*func(Event)]struct{}
}

// New observable instance
func New() *Observable {
	ob := &Observable{
		lock:        &sync.Mutex{},
		subscribers: map[*func(Event)]struct{}{},
	}
	return ob
}

// Publish message to the queue
func (ob *Observable) Publish(e Event) {
	ob.lock.Lock()
	defer ob.lock.Unlock()

	for write := range ob.subscribers {
		(*write)(e)
	}
}

// Subscribe message, the ctx is used to cancel the subscription
func (ob *Observable) Subscribe(ctx context.Context) <-chan Event {
	ob.lock.Lock()
	defer ob.lock.Unlock()

	s, r := Pipe(ctx)

	ob.subscribers[&s] = struct{}{}
	go func() {
		<-ctx.Done()
		ob.lock.Lock()
		defer ob.lock.Unlock()
		delete(ob.subscribers, &s)
	}()

	return r
}

// Count of the subscribers
func (ob *Observable) Count() int {
	ob.lock.Lock()
	defer ob.lock.Unlock()
	return len(ob.subscribers)
}
