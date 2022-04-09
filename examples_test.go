package goob_test

import (
	"context"
	"fmt"
	"time"

	"github.com/ysmood/goob"
)

func Example_basic() {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// create an observable instance
	ob := goob.New(ctx)

	events := ob.Subscribe(context.TODO())

	// publish events without blocking
	ob.Publish(1)
	ob.Publish(2)
	ob.Publish(3)

	// consume events
	for e := range events {
		fmt.Print(e)
	}

	// Output: 123
}
