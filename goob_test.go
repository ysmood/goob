package goob_test

import (
	"math/rand"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ysmood/goob"
)

func checkLeak(t *testing.T) {
	// testleak.Check(t, 0)
}

type null struct{}

func eq(t *testing.T, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Error(expected, "not equal", actual)
	}
}

func TestNew(t *testing.T) {
	checkLeak(t)

	ob := goob.New()
	defer ob.Close()
	s := ob.Subscribe()
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

func TestUnsubscribe(t *testing.T) {
	checkLeak(t)

	ob := goob.New()

	s := ob.Subscribe()
	ob.Unsubscribe(s)

	time.Sleep(10 * time.Millisecond)

	eq(t, ob.Len(), 0)
}

func TestClosed(t *testing.T) {
	checkLeak(t)

	ob := goob.New()
	ob.Close()

	s := ob.Subscribe()
	_, ok := <-s

	eq(t, ok, false)
	eq(t, ob.Len(), 0)
}

func TestMultipleConsumers(t *testing.T) {
	checkLeak(t)

	ob := goob.New()
	defer ob.Close()
	s1 := ob.Subscribe()
	s2 := ob.Subscribe()
	s3 := ob.Subscribe()
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

	ob := goob.New()
	defer ob.Close()
	s := ob.Subscribe()

	ob.Publish(1)

	time.Sleep(20 * time.Millisecond)

	<-s
}

func TestMonkey(t *testing.T) {
	checkLeak(t)

	count := int32(0)
	roundSize := 2
	size := 100

	wg := sync.WaitGroup{}
	wg.Add(roundSize)

	run := func() {
		defer wg.Done()

		ob := goob.New()
		defer ob.Close()
		s := ob.Subscribe()

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
	ob := goob.New()
	defer ob.Close()
	s := ob.Subscribe()

	go func() {
		for range s {
		}
	}()

	for i := 0; i < b.N; i++ {
		ob.Publish(nil)
	}
}

func BenchmarkConsume(b *testing.B) {
	ob := goob.New()
	defer ob.Close()
	s := ob.Subscribe()

	for i := 0; i < b.N; i++ {
		ob.Publish(nil)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		<-s
	}
}
