/*
Copyright IBM Corporation All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

/*
Package fan is a type-flexible fan-in pattern implementation
*/
package fan

import (
	"fmt"
	"reflect"
	"sync"
)

// SelectFunc is a function that implements the core logic of a fan-in implementation for a particular
// type. They should contain a single select statement that listens on the `done`
// channel and the `in` channel. They must type-assert the `in` channel to be a
// channel of the proper input type. When they receive an element on the `in` channel
// they must send it on the `out` channel (also type-asserted). They should return
// true *only* if they receive a value from the `done` channel or if their `in` channel
// is closed. All implementations look essentially like this:
//
//		func(done <-chan struct{}, in, out interface{}) bool {
//	 		select {
//	 		case <-done:
//	 			return true
//	 		case element, more := <-in.(chan int):
//	 			if !more {
//	 				return true
//	 			}
//	 			out.(chan int) <- element
//	 		}
//	 		return false
//	 	}
//
// The only variation is the type of channel that `in` and `out` are asserted to be.
type SelectFunc func(done <-chan struct{}, in, out interface{}) (shouldStop bool)

// Config is the configuration for fanning in channels of a particular element type.
type Config struct {
	// SelectFunc is a function that (if set) will be used to listen on a channel and
	// send data on another channel. If it is not provided, a reflect-based default
	// will be used. This has a significant performance penalty, but it will work for
	// all types.
	//
	// To properly implement a SelectFunc, you must specialize it to the type of data
	// that you will be fanning over the channels. See the docs on the SelectFunc type
	// for examples
	SelectFunc
}

// reflectiveSelectFunc is the default implementation of the Fan's SelectFunc. It expects
// slightly different input parameters than the one that a user provides. In particular, the
// concrete type of `in` and `out` should be reflect.Values in instead of being concrete channel
// types. This is to save calling reflect.ValueOf on each of them during every loop iteration.
//
// This function implements exactly the same logic as the example SelectFunc in the docs except
// it works for any channel type. You do pay a pretty stiff performance penalty though.
func reflectiveSelectFunc(done <-chan struct{}, in, out interface{}) (shouldStop bool) {
	const (
		DoneChanClosed = 0
		InputChanRead  = 1
	)
	selectConfig := []reflect.SelectCase{
		DoneChanClosed: reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(done),
		},
		InputChanRead: reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: in.(reflect.Value),
		},
	}
	switch caseChosen, elem, more := reflect.Select(selectConfig); caseChosen {
	case DoneChanClosed:
		return true
	case InputChanRead:
		if !more {
			return true
		}
		out.(reflect.Value).Send(elem)
	}
	return false
}

// FanIn accepts a done channel and a variable number of channels. It returns a
// receive-only channel of the same type as the input channels, which must be type-asserted
// by the caller in order to use it. While the done channel is not closed, values sent over the input
// channels will become available on the returned channel. When all input channels
// close or the done channel closes, the output channel will close.
//
// This will panic if no channels are provided, if values other than channels are provided,
// if send-only channels are provided, or if the provided channels are the not
// the same element type (though a mixture of receive-only and bidirectional channels with the
// same element type is fine).
func (c Config) FanIn(done <-chan struct{}, channels ...interface{}) interface{} {
	if len(channels) < 1 {
		panic(fmt.Errorf("concurrent.FanIn() called with no channels provided"))
	}
	elementType := reflect.TypeOf(nil)
	// make sure all channels are the same type and are actually channels
	for i, channel := range channels {
		t := reflect.TypeOf(channel)
		// panic if it's not a channel
		if t.Kind() != reflect.Chan {
			panic(fmt.Errorf("channels[%d] is not a channel, is %v", i, t.Kind()))
		}
		// panic if we can't receive
		if t.ChanDir() != reflect.BothDir && t.ChanDir() != reflect.RecvDir {
			panic(fmt.Errorf("channels[%d] does not support receive, has dir %v", i, t.ChanDir()))
		}
		// if we are processing the element type of the first channel, set the element type
		// that we will assume for the rest of the channels
		if elementType == reflect.TypeOf(nil) {
			elementType = t.Elem()
		} else if elementType != t.Elem() {
			// if this is not the first channel, this channel's element type needs to match that of the
			// first channel we processed.
			panic(fmt.Errorf("channels[%d] has element type %v, which does not match previous element type %v", i, t.Elem(), elementType))
		}
	}
	output := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, elementType), 0)
	var wg sync.WaitGroup
	wg.Add(len(channels))

	// launch a worker goroutine for each input channel
	for _, channel := range channels {
		go func(loopBody SelectFunc, done <-chan struct{}, inChan, outChan interface{}) {
			// if no select function provided, fall back on a reflection-based implementation
			if loopBody == nil {
				loopBody = reflectiveSelectFunc
				inChan = reflect.ValueOf(inChan)
				outChan = reflect.ValueOf(outChan)
			}
			defer wg.Done()
			for {
				if loopBody(done, inChan, outChan) {
					break
				}
			}
		}(c.SelectFunc, done, channel, output.Interface())
	}
	// make sure we close our output channel when our waitgroup finishes
	go func() {
		defer output.Close()
		wg.Wait()
	}()
	// return output as receive-only
	return output.Convert(reflect.ChanOf(reflect.RecvDir, elementType)).Interface()
}
