package goob

import (
	"context"
	"sync"
)

// Event interface
type Event interface{}

// Pipe w to r, r won't block w because it uses an internal buffer.
func Pipe(ctx context.Context) (w func(Event), r <-chan Event) {
	ch := make(chan Event)
	lock := sync.Mutex{}
	buf := []Event{}
	wait := make(chan struct{}, 1)

	w = func(e Event) {
		lock.Lock()
		buf = append(buf, e)
		lock.Unlock()

		if len(wait) == 0 {
			select {
			case <-ctx.Done():
				return
			case wait <- struct{}{}:
			}
		}
	}

	go func() {
		defer close(ch)

		for ctx.Err() == nil {
			lock.Lock()
			events := buf
			buf = []Event{}
			lock.Unlock()

			for _, e := range events {
				select {
				case <-ctx.Done():
					return
				case ch <- e:
				}
			}

			select {
			case <-ctx.Done():
				return
			case <-wait:
			}
		}
	}()

	return w, ch
}
