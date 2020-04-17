/*
Copyright IBM Corporation All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

/*
Package fan is a type-flexible fan-in pattern implementation

This package provides convenient and efficient mechanisms for combining many channels
with the same element type into one channel with that element type.

For any primitive type in Go, this package provides helper methods to "fan-in" many
channels with that primitive as the element type. For instance, if you have several
channels of integers:

    var a, b, c chan int // assume these are created elsewhere and are in use

    // We can close this done channel to stop the fan-in operation before all of the
    // input channels close
    done := make(chan struct{})

    // all values sent on a, b, and c will be readable from combined, which will only
    // close when either all of a, b, and c close OR done closes
    combined := fan.Ints().FanIn(done, a, b, c).(<-chan int)

For non-primitive types, this package provides both an easy-to-use (but inefficient)
reflection-based approach that requires no boilerplate and an efficient implementation
that can be customized via a helper function to work with any Go type.

To use the inefficient reflection-based approach on a custom type, you can do:

    type MyCustomType struct {
        Foo, Bar int
        Baz string
    }
    var a, b, c chan MyCustomType // assume these are created elsewhere and are in use

    // We can close this done channel to stop the fan-in operation before all of the
    // input channels close
    done := make(chan struct{})

    // all values sent on a, b, and c will be readable from combined, which will only
    // close when either all of a, b, and c close OR done closes
    combined := fan.Config{}.FanIn(done, a, b, c).(<-chan MyCustomType)

To accelerate the fan-in operation to nearly the same speed as an implementation specialized
to your custom type, simply provide a SelectFunc implementation within the fan.Config:

    type MyCustomType struct {
        Foo, Bar int
        Baz string
    }
    var a, b, c chan MyCustomType // assume these are created elsewhere and are in use

    // We can close this done channel to stop the fan-in operation before all of the
    // input channels close
    done := make(chan struct{})

    // all values sent on a, b, and c will be readable from combined, which will only
    // close when either all of a, b, and c close OR done closes
    combined := fan.Config{
        SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
	 		select {
	 		case <-done:
	 			return true
	 		case element, more := <-in.(<-chan MyCustomType):
	 			if !more {
	 				return true
	 			}
	 			out.(chan MyCustomType) <- element
	 		}
	 		return false
	 	}
    }.FanIn(done, a, b, c).(<-chan MyCustomType)

This small bit of boilerplate captures the necessary type information to avoid performing
any reflection while passing data read from the channels, resulting in the same throughput
as a custom implementation for your type.

All SelectFunc implementations look essentially the same, with the only difference being
the element type of the channels in the two type assertions.

*/
package fan

import (
	"fmt"
	"reflect"
	"sync"
)

// Interfaces returns a config intended to fan-in channels with the empty interface
// as their element type.
func Interfaces() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan interface{}):
				if !more {
					return true
				}
				out.(chan interface{}) <- element
			}
			return false
		},
	}
}

// Strings returns a config intended to fan-in channels with string
// as their element type.
func Strings() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan string):
				if !more {
					return true
				}
				out.(chan string) <- element
			}
			return false
		},
	}
}

// ByteSlices returns a config intended to fan-in channels with byte slice
// as their element type.
func ByteSlices() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan []byte):
				if !more {
					return true
				}
				out.(chan []byte) <- element
			}
			return false
		},
	}
}

// Uintptrs returns a config intended to fan-in channels with uintptr
// as their element type.
func Uintptrs() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan uintptr):
				if !more {
					return true
				}
				out.(chan uintptr) <- element
			}
			return false
		},
	}
}

// Bools returns a config intended to fan-in channels with bool
// as their element type.
func Bools() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan bool):
				if !more {
					return true
				}
				out.(chan bool) <- element
			}
			return false
		},
	}
}

// Bytes returns a config intended to fan-in channels with byte
// as their element type.
func Bytes() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan byte):
				if !more {
					return true
				}
				out.(chan byte) <- element
			}
			return false
		},
	}
}

// Runes returns a config intended to fan-in channels with rune
// as their element type.
func Runes() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan rune):
				if !more {
					return true
				}
				out.(chan rune) <- element
			}
			return false
		},
	}
}

// Complex64s returns a config intended to fan-in channels with complex64
// as their element type.
func Complex64s() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan complex64):
				if !more {
					return true
				}
				out.(chan complex64) <- element
			}
			return false
		},
	}
}

// Complex128s returns a config intended to fan-in channels with complex128
// as their element type.
func Complex128s() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan complex128):
				if !more {
					return true
				}
				out.(chan complex128) <- element
			}
			return false
		},
	}
}

// Float32s returns a config intended to fan-in channels with float32
// as their element type.
func Float32s() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan float32):
				if !more {
					return true
				}
				out.(chan float32) <- element
			}
			return false
		},
	}
}

// Float64s returns a config intended to fan-in channels with float64
// as their element type.
func Float64s() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan float64):
				if !more {
					return true
				}
				out.(chan float64) <- element
			}
			return false
		},
	}
}

// Ints returns a config intended to fan-in channels with int
// as their element type.
func Ints() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan int):
				if !more {
					return true
				}
				out.(chan int) <- element
			}
			return false
		},
	}
}

// Uints returns a config intended to fan-in channels with uint
// as their element type.
func Uints() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan uint):
				if !more {
					return true
				}
				out.(chan uint) <- element
			}
			return false
		},
	}
}

// Int8s returns a config intended to fan-in channels with int8
// as their element type.
func Int8s() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan int8):
				if !more {
					return true
				}
				out.(chan int8) <- element
			}
			return false
		},
	}
}

// Uint8s returns a config intended to fan-in channels with uint8
// as their element type.
func Uint8s() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan uint8):
				if !more {
					return true
				}
				out.(chan uint8) <- element
			}
			return false
		},
	}
}

// Int16s returns a config intended to fan-in channels with int16
// as their element type.
func Int16s() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan int16):
				if !more {
					return true
				}
				out.(chan int16) <- element
			}
			return false
		},
	}
}

// Uint16s returns a config intended to fan-in channels with uint16
// as their element type.
func Uint16s() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan uint16):
				if !more {
					return true
				}
				out.(chan uint16) <- element
			}
			return false
		},
	}
}

// Int32s returns a config intended to fan-in channels with int32
// as their element type.
func Int32s() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan int32):
				if !more {
					return true
				}
				out.(chan int32) <- element
			}
			return false
		},
	}
}

// Uint32s returns a config intended to fan-in channels with uint32
// as their element type.
func Uint32s() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan uint32):
				if !more {
					return true
				}
				out.(chan uint32) <- element
			}
			return false
		},
	}
}

// Int64s returns a config intended to fan-in channels with int64
// as their element type.
func Int64s() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan int64):
				if !more {
					return true
				}
				out.(chan int64) <- element
			}
			return false
		},
	}
}

// Uint64s returns a config intended to fan-in channels with uint64
// as their element type.
func Uint64s() Config {
	return Config{
		SelectFunc: func(done <-chan struct{}, in, out interface{}) bool {
			select {
			case <-done:
				return true
			case element, more := <-in.(<-chan uint64):
				if !more {
					return true
				}
				out.(chan uint64) <- element
			}
			return false
		},
	}
}

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
//	 		case element, more := <-in.(<-chan int):
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
// by the caller in order to be usable. While the done channel is not closed, values sent over the input
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
			// ensure that the inChan to each fan-in worker is receive-only
			inChan = reflect.ValueOf(inChan).Convert(reflect.ChanOf(reflect.RecvDir, elementType)).Interface()
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
