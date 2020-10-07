package goob_test

import (
	"sync"
	"testing"

	"github.com/ysmood/goob"
)

func TestPipe(t *testing.T) {
	checkLeak(t)

	const pipeCount = 10
	const msgCount = 10

	round := func() {
		p := goob.NewPipe()
		defer p.Stop()

		wg := sync.WaitGroup{}
		wg.Add(msgCount)
		for i := 0; i < msgCount*2; i++ {
			if i%2 == 0 {
				go p.Write(i)
			} else {
				go func() {
					<-p.Events
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
	checkLeak(t)

	const count = 1000

	for i := 0; i < count; i++ {
		p := goob.NewPipe()
		go p.Write(1)
		go p.Stop()
	}
}
