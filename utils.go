package goob

import (
	"context"
	"reflect"
)

// Each event, if iteratee returns true, the iteration will stop
func Each(s chan Event, iteratee interface{}) {
	for e := range s {
		fnVal := reflect.ValueOf(iteratee)
		ret := fnVal.Call([]reflect.Value{reflect.ValueOf(e)})
		if len(ret) == 1 && ret[0].Bool() {
			break
		}
	}
}

// Map events
func (ob *Observable) Map(ctx context.Context, iteratee interface{}) *Observable {
	mapped := New()
	s := ob.Subscribe(ctx)

	go func() {
		for e := range s {
			fnVal := reflect.ValueOf(iteratee)
			ret := fnVal.Call([]reflect.Value{reflect.ValueOf(e)})
			mapped.Publish(ret[0].Interface())
		}
	}()

	return mapped
}

// Filter events
func (ob *Observable) Filter(ctx context.Context, iteratee interface{}) *Observable {
	filterred := New()
	s := ob.Subscribe(ctx)

	go func() {
		for e := range s {
			fnVal := reflect.ValueOf(iteratee)
			ret := fnVal.Call([]reflect.Value{reflect.ValueOf(e)})
			if ret[0].Bool() {
				filterred.Publish(e)
			}
		}
	}()

	return filterred
}

func noPanic(fn func()) {
	defer func() {
		_ = recover()
	}()

	fn()
}
