package goob_test

import (
	"fmt"

	"github.com/ysmood/goob"
)

func Example_basic() {
	// create an observable instance
	ob := goob.New()
	defer ob.Close()

	events := ob.Subscribe()

	// publish events without blocking
	ob.Publish(1)
	ob.Publish(2)
	ob.Publish(3)

	// consume events
	for e := range events {
		fmt.Print(e)

		if e.(int) == 3 {
			break
		}
	}

	// Output: 123
}
