/*
Copyright IBM Corporation All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fan_test

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

	fan "github.com/IBM/fast-fan-in"
)

// Here's a simple example of doubling integers using the fan-out, fan-in
// pattern:
func ExampleInts() {
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

	// configure our fan-in.
	out := fan.Ints().FanIn(done, workerOuts...).(<-chan int)

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

// Here's a simple example of concatenating strings using the fan-out, fan-in
// pattern:
func ExampleStrings() {
	// define a simple worker function that spawns a new goroutine to
	// read numbers from an input channel, double them, and send them on an
	// output channel. Importantly the output channel will close as soon
	// as the input channel does (see the deferred close).
	concat := func(in <-chan string) <-chan string {
		out := make(chan string)
		go func() {
			defer close(out) // close output channel when this anonymous func returns
			for i := range in {
				out <- i + i
			}
		}()
		return out
	}

	// make an input channel of integers and send the numbers 1-10
	strs := make(chan string)
	go func() {
		defer close(strs)
		for i := 0; i < 10; i++ {
			strs <- strconv.Itoa(i)
		}
	}()

	// launch a fixed quantity of double() worker goroutines all reading from
	// the same input channel. This is a Fan-Out, as work is being distributed from
	// one goroutine to many.
	numWorkers := 3
	// we allocate this as a slice of interface because otherwise we'd need to cast
	// a []chan string into a []interface{}. Go doesn't allow this as a direct type-cast,
	// so we'd need to allocate a second slice of type []interface{} and copy each
	// element. This is more concise
	workerOuts := make([]interface{}, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workerOuts[i] = concat(strs) // this returns the output channel of the worker
	}

	// make a done channel that we could use to terminate the fan-in operation early
	// (we won't use it in this example, but the API requires it).
	done := make(chan struct{})

	// configure our fan-in.
	out := fan.Strings().FanIn(done, workerOuts...).(<-chan string)

	// collect the data from the output channel and print it
	outputNums := []string{}
	for i := range out {
		outputNums = append(outputNums, i)
	}
	sort.Strings(outputNums)
	fmt.Println(outputNums)
	/*
		Output:
		[00 11 22 33 44 55 66 77 88 99]
	*/
}

func ExampleUints() {
	a, b := make(chan uint), make(chan uint)
	go func() {
		defer close(a)
		defer close(b)
		a <- 1
		b <- 2
		a <- 3
		b <- 4
	}()

	done := make(chan struct{})
	out := fan.Uints().FanIn(done, a, b).(<-chan uint)

	var results []uint
	for result := range out {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})
	fmt.Println(results)
	/*
		Output:
		[1 2 3 4]
	*/
}

func ExampleUint8s() {
	a, b := make(chan uint8), make(chan uint8)
	go func() {
		defer close(a)
		defer close(b)
		a <- 1
		b <- 2
		a <- 3
		b <- 4
	}()

	done := make(chan struct{})
	out := fan.Uint8s().FanIn(done, a, b).(<-chan uint8)

	var results []uint8
	for result := range out {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})
	fmt.Println(results)
	/*
		Output:
		[1 2 3 4]
	*/
}
func ExampleUint16s() {
	a, b := make(chan uint16), make(chan uint16)
	go func() {
		defer close(a)
		defer close(b)
		a <- 1
		b <- 2
		a <- 3
		b <- 4
	}()

	done := make(chan struct{})
	out := fan.Uint16s().FanIn(done, a, b).(<-chan uint16)

	var results []uint16
	for result := range out {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})
	fmt.Println(results)
	/*
		Output:
		[1 2 3 4]
	*/
}
func ExampleUint32s() {
	a, b := make(chan uint32), make(chan uint32)
	go func() {
		defer close(a)
		defer close(b)
		a <- 1
		b <- 2
		a <- 3
		b <- 4
	}()

	done := make(chan struct{})
	out := fan.Uint32s().FanIn(done, a, b).(<-chan uint32)

	var results []uint32
	for result := range out {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})
	fmt.Println(results)
	/*
		Output:
		[1 2 3 4]
	*/
}
func ExampleUint64s() {
	a, b := make(chan uint64), make(chan uint64)
	go func() {
		defer close(a)
		defer close(b)
		a <- 1
		b <- 2
		a <- 3
		b <- 4
	}()

	done := make(chan struct{})
	out := fan.Uint64s().FanIn(done, a, b).(<-chan uint64)

	var results []uint64
	for result := range out {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})
	fmt.Println(results)
	/*
		Output:
		[1 2 3 4]
	*/
}

func ExampleInt8s() {
	a, b := make(chan int8), make(chan int8)
	go func() {
		defer close(a)
		defer close(b)
		a <- 1
		b <- 2
		a <- 3
		b <- 4
	}()

	done := make(chan struct{})
	out := fan.Int8s().FanIn(done, a, b).(<-chan int8)

	var results []int8
	for result := range out {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})
	fmt.Println(results)
	/*
		Output:
		[1 2 3 4]
	*/
}
func ExampleInt16s() {
	a, b := make(chan int16), make(chan int16)
	go func() {
		defer close(a)
		defer close(b)
		a <- 1
		b <- 2
		a <- 3
		b <- 4
	}()

	done := make(chan struct{})
	out := fan.Int16s().FanIn(done, a, b).(<-chan int16)

	var results []int16
	for result := range out {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})
	fmt.Println(results)
	/*
		Output:
		[1 2 3 4]
	*/
}
func ExampleInt32s() {
	a, b := make(chan int32), make(chan int32)
	go func() {
		defer close(a)
		defer close(b)
		a <- 1
		b <- 2
		a <- 3
		b <- 4
	}()

	done := make(chan struct{})
	out := fan.Int32s().FanIn(done, a, b).(<-chan int32)

	var results []int32
	for result := range out {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})
	fmt.Println(results)
	/*
		Output:
		[1 2 3 4]
	*/
}
func ExampleInt64s() {
	a, b := make(chan int64), make(chan int64)
	go func() {
		defer close(a)
		defer close(b)
		a <- 1
		b <- 2
		a <- 3
		b <- 4
	}()

	done := make(chan struct{})
	out := fan.Int64s().FanIn(done, a, b).(<-chan int64)

	var results []int64
	for result := range out {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})
	fmt.Println(results)
	/*
		Output:
		[1 2 3 4]
	*/
}

func ExampleFloat32s() {
	a, b := make(chan float32), make(chan float32)
	go func() {
		defer close(a)
		defer close(b)
		a <- 1
		b <- 2
		a <- 3
		b <- 4
	}()

	done := make(chan struct{})
	out := fan.Float32s().FanIn(done, a, b).(<-chan float32)

	var results []float32
	for result := range out {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})
	fmt.Println(results)
	/*
		Output:
		[1 2 3 4]
	*/
}
func ExampleFloat64s() {
	a, b := make(chan float64), make(chan float64)
	go func() {
		defer close(a)
		defer close(b)
		a <- 1
		b <- 2
		a <- 3
		b <- 4
	}()

	done := make(chan struct{})
	out := fan.Float64s().FanIn(done, a, b).(<-chan float64)

	var results []float64
	for result := range out {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})
	fmt.Println(results)
	/*
		Output:
		[1 2 3 4]
	*/
}
func ExampleComplex64s() {
	a, b := make(chan complex64), make(chan complex64)
	go func() {
		defer close(a)
		defer close(b)
		a <- 1 + 3i
		b <- 2 + 4i
		a <- 3 + 9i
		b <- 4 + 2i
	}()

	done := make(chan struct{})
	out := fan.Complex64s().FanIn(done, a, b).(<-chan complex64)

	var results []complex64
	for result := range out {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return real(results[i]) < real(results[j])
	})
	fmt.Println(results)
	/*
		Output:
		[(1+3i) (2+4i) (3+9i) (4+2i)]
	*/
}
func ExampleComplex128s() {
	a, b := make(chan complex128), make(chan complex128)
	go func() {
		defer close(a)
		defer close(b)
		a <- 1 + 3i
		b <- 2 + 4i
		a <- 3 + 9i
		b <- 4 + 2i
	}()

	done := make(chan struct{})
	out := fan.Complex128s().FanIn(done, a, b).(<-chan complex128)

	var results []complex128
	for result := range out {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return real(results[i]) < real(results[j])
	})
	fmt.Println(results)
	/*
		Output:
		[(1+3i) (2+4i) (3+9i) (4+2i)]
	*/
}

func ExampleBools() {
	a, b := make(chan bool), make(chan bool)
	go func() {
		defer close(a)
		defer close(b)
		a <- true
		b <- true
		a <- false
		b <- true
	}()

	done := make(chan struct{})
	out := fan.Bools().FanIn(done, a, b).(<-chan bool)

	var results []bool
	for result := range out {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return !results[i]
	})
	fmt.Println(results)
	/*
		Output:
		[false true true true]
	*/
}

func ExampleBytes() {
	a, b := make(chan byte), make(chan byte)
	go func() {
		defer close(a)
		defer close(b)
		a <- 1
		b <- 2
		a <- 3
		b <- 4
	}()

	done := make(chan struct{})
	out := fan.Bytes().FanIn(done, a, b).(<-chan byte)

	var results []byte
	for result := range out {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})
	fmt.Println(results)
	/*
		Output:
		[1 2 3 4]
	*/
}

func ExampleRunes() {
	a, b := make(chan rune), make(chan rune)
	go func() {
		defer close(a)
		defer close(b)
		a <- '1'
		b <- '2'
		a <- '3'
		b <- '4'
	}()

	done := make(chan struct{})
	out := fan.Runes().FanIn(done, a, b).(<-chan rune)

	var results []rune
	for result := range out {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})
	fmt.Println(results)
	/*
		Output:
		[49 50 51 52]
	*/
}

func ExampleInterfaces() {
	a, b := make(chan interface{}), make(chan interface{})
	go func() {
		defer close(a)
		defer close(b)
		a <- "hello"
		b <- "world"
		a <- "here's an"
		b <- "example"
	}()

	done := make(chan struct{})
	out := fan.Interfaces().FanIn(done, a, b).(<-chan interface{})

	var results []interface{}
	for result := range out {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return strings.Compare(results[i].(string), results[j].(string)) < 1
	})
	fmt.Println(results)
	/*
		Output:
		[example hello here's an world]
	*/
}

func ExampleByteSlices() {
	a, b := make(chan []byte), make(chan []byte)
	go func() {
		defer close(a)
		defer close(b)
		a <- []byte("hello")
		b <- []byte("world")
		a <- []byte("here's an")
		b <- []byte("example")
	}()

	done := make(chan struct{})
	out := fan.ByteSlices().FanIn(done, a, b).(<-chan []byte)

	var results [][]byte
	for result := range out {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return bytes.Compare(results[i], results[j]) < 1
	})
	fmt.Println(results)
	/*
		Output:
		[[101 120 97 109 112 108 101] [104 101 108 108 111] [104 101 114 101 39 115 32 97 110] [119 111 114 108 100]]
	*/
}

func ExampleUintptrs() {
	a, b := make(chan uintptr), make(chan uintptr)
	go func() {
		defer close(a)
		defer close(b)
		a <- 2
		b <- 3
		a <- 5
		b <- 7
	}()

	done := make(chan struct{})
	out := fan.Uintptrs().FanIn(done, a, b).(<-chan uintptr)

	var results []uintptr
	for result := range out {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})
	fmt.Println(results)
	/*
		Output:
		[2 3 5 7]
	*/
}
