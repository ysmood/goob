package goob_test

import (
	"context"
	"fmt"

	"github.com/ysmood/goob"
)

func Example_basic() {
	// create an observable instance
	ob := goob.New()

	// use context primitive to unsubscribe observable
	ctx, unsubscribe := context.WithCancel(context.Background())
	defer unsubscribe()

	// create 2 subscribers
	s1 := ob.Subscribe(ctx)
	s2 := ob.Subscribe(ctx)

	// publish events
	ob.Publish(1)
	ob.Publish(2)
	ob.Publish(3)

	// s1 consume events
	for e := range s1 {
		fmt.Print(e)

		if e.(int) == 3 {
			break
		}
	}

	// s2 consume events with goob.Each helper, it will auto cast the type
	goob.Each(s2, func(i int) bool {
		fmt.Print(i)
		return i == 3
	})

	// Output: 123123
}