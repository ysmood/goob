package goob_test

import (
	"context"
	"math/rand"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ysmood/goob"
	"github.com/ysmood/gotrace"
)

func checkLeak(t *testing.T) {
	gotrace.CheckLeak(t, 0)
}

type null struct{}

func eq(t *testing.T, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Error(expected, "not equal", actual)
	}
}

func TestNew(t *testing.T) {
	checkLeak(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer t.Cleanup(cancel)

	ob := goob.New(ctx)
	s := ob.Subscribe(ctx)
	size := 1000

	expected := []int{}
	go func() {
		for i := range make([]null, size) {
			expected = append(expected, i)
			ob.Publish(i)
		}
	}()

	result := []int{}
	for msg := range s {
		result = append(result, msg.(int))
		if len(result) == size {
			break
		}
	}

	eq(t, expected, result)
}

func TestCancel(t *testing.T) {
	checkLeak(t)

	ob := goob.New(context.Background())

	ctx, cancel := context.WithCancel(context.Background())
	ob.Subscribe(ctx)
	cancel()
	time.Sleep(10 * time.Millisecond)
	eq(t, ob.Len(), 0)
}

func TestClosed(t *testing.T) {
	checkLeak(t)

	ctx, cancel := context.WithCancel(context.Background())

	ob := goob.New(ctx)
	ob.Subscribe(ctx)
	cancel()

	s := ob.Subscribe(context.Background())
	_, ok := <-s

	ob.Publish(1)

	eq(t, ok, false)
	eq(t, ob.Len(), 0)
}

func TestMultipleConsumers(t *testing.T) {
	checkLeak(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer t.Cleanup(cancel)

	ob := goob.New(ctx)
	s1 := ob.Subscribe(ctx)
	s2 := ob.Subscribe(ctx)
	s3 := ob.Subscribe(ctx)
	size := 1000

	expected := []int{}
	go func() {
		for i := range make([]null, size) {
			expected = append(expected, i)
			time.Sleep(time.Duration(rand.Intn(100)) * time.Nanosecond)
			ob.Publish(i)
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(2)

	r1 := []int{}
	go func() {
		for e := range s1 {
			r1 = append(r1, e.(int))
			if len(r1) == size {
				wg.Done()
			}
		}
	}()

	r2 := []int{}
	go func() {
		for e := range s2 {
			r2 = append(r2, e.(int))
			if len(r2) == size {
				wg.Done()
			}
		}
	}()

	go func() {
		<-s3 // simulate slow consumer
	}()

	wg.Wait()

	eq(t, expected, r1)
	eq(t, expected, r2)
}

func TestSlowConsumer(t *testing.T) {
	checkLeak(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer t.Cleanup(cancel)

	ob := goob.New(ctx)
	s := ob.Subscribe(ctx)

	ob.Publish(1)

	time.Sleep(20 * time.Millisecond)

	<-s
}

func TestMonkey(t *testing.T) {
	checkLeak(t)

	count := int32(0)
	roundSize := 20
	size := 1000

	wg := sync.WaitGroup{}
	wg.Add(roundSize)

	run := func() {
		defer wg.Done()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ob := goob.New(ctx)
		s := ob.Subscribe(ctx)

		go func() {
			for range make([]null, size) {
				time.Sleep(time.Duration(rand.Intn(100)) * time.Nanosecond)
				ob.Publish(nil)
			}
		}()

		wait := make(chan null)
		go func() {
			for i := range make([]null, size) {
				time.Sleep(time.Duration(rand.Intn(100)) * time.Nanosecond)

				<-s

				atomic.AddInt32(&count, 1)

				if i == size-1 {
					wait <- null{}
				}
			}
		}()
		<-wait
	}

	for range make([]null, roundSize) {
		go run()
	}

	wg.Wait()

	eq(t, roundSize*size, int(count))
}

func BenchmarkPublish(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer b.Cleanup(cancel)

	ob := goob.New(ctx)
	s := ob.Subscribe(ctx)

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for range s {
			}
		}()
	}

	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			ob.Publish(nil)
		}
	})
}

func BenchmarkConsume(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer b.Cleanup(cancel)

	ob := goob.New(ctx)
	s := ob.Subscribe(ctx)

	for i := 0; i < b.N; i++ {
		ob.Publish(nil)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		<-s
	}
}
