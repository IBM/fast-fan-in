/*
Copyright IBM Corporation All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package fan_test

import (
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	fan "github.com/IBM/fast-fan-in"
)

func TestFanInNoChannels(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatalf("should have panicked with no channels as input")
		}
	}()
	done := make(chan struct{})
	out := fan.Config{}.FanIn(done)
	if out != nil {
		t.Fatalf("should not get output channel if no input channels provided")
	}
}

func TestFanInSingleClose(t *testing.T) {
	in := make(chan int)
	done := make(chan struct{})
	out := fan.Config{}.FanIn(done, in).(<-chan int)
	data := 5
	go func() {
		defer close(in)
		in <- data
	}()
	select {
	case <-time.NewTicker(time.Millisecond * 10).C:
		t.Fatalf("timed out")
	case elem := <-out:
		if elem != data {
			t.Fatalf("expected to receive %v, got %v", data, elem)
		}
	}
	_, more := <-out
	if more {
		t.Fatalf("channel is not closed after input channel closed")
	}
}

func TestFanInMixedTypes(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatalf("should have panicked with mixed channel types as input")
		}
	}()
	in := make(chan int)
	in2 := make(chan string)
	done := make(chan struct{})
	fan.Config{}.FanIn(done, in, in2)
}

func TestFanInSendOnly(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatalf("should have panicked with send-only channel type as input")
		}
	}()
	in := (chan<- int)(make(chan int))
	done := make(chan struct{})
	fan.Config{}.FanIn(done, in)
}

func TestFanInNonChannel(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatalf("should have panicked with non-channel type as input")
		}
	}()
	in := 5
	done := make(chan struct{})
	fan.Config{}.FanIn(done, in)
}

func TestFanInSingleUnclose(t *testing.T) {
	in := make(chan int)
	done := make(chan struct{})
	out := fan.Config{}.FanIn(done, in).(<-chan int)
	data := 5
	go func() {
		in <- data
	}()
	select {
	case <-time.NewTicker(time.Millisecond * 10).C:
		t.Fatalf("timed out")
	case elem, more := <-out:
		if !more {
			t.Fatalf("channel should not be closed since input was not closed")
		}
		if elem != data {
			t.Fatalf("expected to receive %v, got %v", data, elem)
		}
	}
}

func TestFanInMultiplePrematureDone(t *testing.T) {
	in := make([]interface{}, 10)
	for i := range in {
		in[i] = make(chan int)
	}
	done := make(chan struct{})
	out := fan.Config{}.FanIn(done, in...).(<-chan int)
	go func() {
		close(done)
	}()
	select {
	case <-time.NewTicker(time.Millisecond * 10).C:
		t.Fatalf("timed out")
	case _, more := <-out:
		if more {
			t.Fatalf("channel should be closed since done was closed")
		}
	}
}

func TestFanInMultipleClose(t *testing.T) {
	for _, workers := range []int{2, 11, 50} {
		for _, numInputs := range []int{5, 11, 50, 200} {
			t.Run(fmt.Sprintf("workers:%d-inputs:%d", workers, numInputs), func(t *testing.T) {
				FanInMultipleClose(t, workers, numInputs)
			})
		}
	}
}

func FanInMultipleClose(t *testing.T, numChannels, max int) {
	// make some input channels
	ins := make([]interface{}, numChannels)
	for i := range ins {
		ins[i] = make(chan int)
	}
	done := make(chan struct{})
	out := fan.Config{}.FanIn(done, ins...).(<-chan int)

	// make and send some output data (just the numbers 0-(max-1))
	outputs := make([]int, 0, max)
	go func() {
		defer func() {
			for _, in := range ins {
				close(in.(chan int))
			}
			t.Log("all inputs closed")
		}()
		for i := 0; i < max; i++ {
			ins[i%len(ins)].(chan int) <- i
		}
	}()

	// receive all output data and collect into slice
	for i := 0; i < max; i++ {
		select {
		case <-time.NewTicker(time.Millisecond * 10).C:
			t.Fatalf("timed out")
		case elem := <-out:
			outputs = append(outputs, elem)
		}
	}

	// ensure output channel is now closed
	_, more := <-out
	if more {
		t.Fatalf("channel is not closed after input channel closed")
	}

	// make sure we got all of the numbers we expected
	sort.Ints(outputs)
	for i := range outputs {
		if i != outputs[i] {
			t.Fatalf("missing elements in output, expected %d, got %d in %v", i, outputs[i], outputs)
		}
	}
}

// This is an efficient implementation of FanIn for a concrete type. It is used to
// compare the efficiency of the type-agnostic implementation defined in this package
// against a type-specific implementation.
func ConcreteFanIn(done <-chan struct{}, inputs ...<-chan int) <-chan int {
	results := make(chan int)
	var wg sync.WaitGroup
	// define a function to accept input on a single channel and push it onto the
	// shared channel that we return
	fan := func(input <-chan int) {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			case element, more := <-input:
				if !more {
					return
				}
				results <- element
			}
		}
	}
	// launch a goroutine to handle each input channel and push the data onto the
	// returned channel
	for _, input := range inputs {
		wg.Add(1)
		go fan(input)
	}
	// make sure we close our output channel after all workers stop running
	go func() {
		defer close(results)
		wg.Wait()
	}()
	return results
}

func BenchmarkFanIn(b *testing.B) {
	setupConcrete := func(inputs []chan int) (chan<- struct{}, <-chan int) {
		asRcvOnly := make([]<-chan int, len(inputs))
		for i := range inputs {
			asRcvOnly[i] = inputs[i]
		}
		done := make(chan struct{})
		output := ConcreteFanIn(done, asRcvOnly...)
		return done, output
	}
	setupHybridUnspecialized := func(inputs []chan int) (chan<- struct{}, <-chan int) {
		asGeneric := make([]interface{}, len(inputs))
		for i := range inputs {
			asGeneric[i] = inputs[i]
		}
		done := make(chan struct{})
		output := fan.Config{}.FanIn(done, asGeneric...).(<-chan int)
		return done, output
	}
	setupHybridSpecialized := func(inputs []chan int) (chan<- struct{}, <-chan int) {
		asGeneric := make([]interface{}, len(inputs))
		for i := range inputs {
			asGeneric[i] = inputs[i]
		}
		done := make(chan struct{})
		fan := fan.Config{
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
		output := fan.FanIn(done, asGeneric...).(<-chan int)
		return done, output
	}
	type setupFunc func(inputs []chan int) (chan<- struct{}, <-chan int)
	type implDetails struct {
		Name  string
		Setup setupFunc
	}
	for _, numChannels := range []int{1, 10, 100} {
		for _, numElements := range []int{10, 100, 1000, 10000, 100000} {
			for _, setup := range []implDetails{
				{Name: "concrete", Setup: setupConcrete},
				{Name: "hybrid-reflect", Setup: setupHybridUnspecialized},
				{Name: "hybrid-closure", Setup: setupHybridSpecialized},
			} {
				b.Run(fmt.Sprintf("chans:%d,elems:%d,impl:%s", numChannels, numElements, setup.Name), func(b *testing.B) {
					inputs := make([]chan int, numChannels)
					for i := range inputs {
						inputs[i] = make(chan int)
					}
					done, output := setup.Setup(inputs)
					defer close(done)
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						go func() {
							for i := 0; i < numElements; i++ {
								inputs[i%len(inputs)] <- i
							}
						}()
						for i := 0; i < numElements; i++ {
							<-output
						}
					}
				})
			}
		}
	}
}

// Here's a simple example of doubling integers using the fan-out, fan-in
// pattern:
func ExampleConfig() {
	// define a simple worker function that spawns a new goroutine to
	// read numbers from an input channel, double them, and send them on an
	// output channel. Importantly the output channel will close as soon
	// as the input channel does (see the deferred close).
	double := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out) // close output channel when this anonymous func returns
			for i := range in {
				out <- i * 2
			}
		}()
		return out
	}

	// make an input channel of integers and send the numbers 1-10
	ints := make(chan int)
	go func() {
		defer close(ints)
		for i := 0; i < 10; i++ {
			ints <- i
		}
	}()

	// launch a fixed quantity of double() worker goroutines all reading from
	// the same input channel. This is a Fan-Out, as work is being distributed from
	// one goroutine to many.
	numWorkers := 3
	// we allocate this as a slice of interface because otherwise we'd need to cast
	// a []chan int into a []interface{}. Go doesn't allow this as a direct type-cast,
	// so we'd need to allocate a second slice of type []interface{} and copy each
	// element. This is more concise
	workerOuts := make([]interface{}, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workerOuts[i] = double(ints) // this returns the output channel of the worker
	}

	// make a done channel that we could use to terminate the fan-in operation early
	// (we won't use it in this example, but the API requires it).
	done := make(chan struct{})

	// configure our fan-in. This time we'll use the reflect-based approach to keep
	// the code shorter. We'd specify a SelectFunc in this struct to accelerate it.
	out := fan.Config{}.FanIn(done, workerOuts...).(<-chan int)

	// collect the data from the output channel and print it
	outputNums := []int{}
	for i := range out {
		outputNums = append(outputNums, i)
	}
	sort.Ints(outputNums)
	fmt.Println(outputNums)
	/*
		Output:
		[0 2 4 6 8 10 12 14 16 18]
	*/
}
