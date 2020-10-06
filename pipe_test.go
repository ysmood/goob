package goob_test

import (
	"context"
	"sync"
	"testing"

	"github.com/ysmood/goob"
	"github.com/ysmood/gotrace/pkg/testleak"
)

func TestPipe(t *testing.T) {
	testleak.Check(t, 0)

	const pipeCount = 10
	const msgCount = 10

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	round := func() {
		w, r := goob.Pipe(ctx)

		wg := sync.WaitGroup{}
		wg.Add(msgCount)
		for i := 0; i < msgCount*2; i++ {
			if i%2 == 0 {
				go w(i)
			} else {
				go func() {
					<-r
					wg.Done()
				}()
			}
		}
		wg.Wait()
	}

	wg := sync.WaitGroup{}
	wg.Add(pipeCount)
	for i := 0; i < pipeCount; i++ {
		go func() {
			round()
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestPipeCancel(t *testing.T) {
	testleak.Check(t, 0)

	const count = 100

	for i := 0; i < count; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		w, _ := goob.Pipe(ctx)
		go w(1)
		go cancel()
	}
}
