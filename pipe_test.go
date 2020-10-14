package goob_test

import (
	"sync"
	"testing"
	"time"

	"github.com/ysmood/goob"
)

func TestPipeOrder(t *testing.T) {
	checkLeak(t)

	p := goob.NewPipe()

	p.Write(1)
	p.Write(2)
	p.Write(3)

	if 1 != <-p.Events {
		t.Fatal()
	}
	if 2 != <-p.Events {
		t.Fatal()
	}
	if 3 != <-p.Events {
		t.Fatal()
	}
}

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

func TestPipeMonkey(t *testing.T) {
	p := goob.NewPipe()
	t.Cleanup(p.Stop)

	round := 30
	count := 10000

	for i := 0; i < round; i++ {
		go func() {
			for i := 0; i < count; i++ {
				p.Write(i)
			}
		}()
	}

	for i := 0; i < count*round; i++ {
		time.Sleep(100 * time.Nanosecond)
		<-p.Events
	}
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
