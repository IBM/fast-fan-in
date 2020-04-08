# fast-fan-in

[![GoDoc](https://godoc.org/github.com/IBM/fast-fan-in?status.svg)](https://godoc.org/github.com/IBM/fast-fan-in)

Golang fan-in pattern efficiently adaptable to any channel type without code generation

- [Usage](#usage)
- [Rationale](#rationale)
- [Benchmarks](#benchmarks)
- [Notes](#notes)

## Usage

This package provides convenient and efficient mechanisms for combining many channels
with the same element type into one channel with that element type.

For any primitive type in Go, this package provides helper methods to "fan-in" many
channels with that primitive as the element type. For instance, if you have several
channels of integers:

```go
    var a, b, c chan int // assume these are created elsewhere and are in use

    // We can close this done channel to stop the fan-in operation before all of the
    // input channels close
    done := make(chan struct{})

    // all values sent on a, b, and c will be readable from combined, which will only
    // close when either all of a, b, and c close OR done closes
    combined := fan.Ints().FanIn(done, a, b, c).(<-chan int)
```

For non-primitive types, this package provides both an easy-to-use (but inefficient)
reflection-based approach that requires no boilerplate and an efficient implementation
that can be customized via a helper function to work with any Go type.

To use the inefficient reflection-based approach on a custom type, you can do:

```go
    type MyCustomType struct {
        A, B int
        C string
    }
    var a, b, c chan MyCustomType // assume these are created elsewhere and are in use

    // We can close this done channel to stop the fan-in operation before all of the
    // input channels close
    done := make(chan struct{})

    // all values sent on a, b, and c will be readable from combined, which will only
    // close when either all of a, b, and c close OR done closes
    combined := fan.Config{}.FanIn(done, a, b, c).(<-chan MyCustomType)
```

To accelerate the fan-in operation to nearly the same speed as an implementation specialized
to your custom type, simply provide a SelectFunc implementation within the fan.Config:

```go
    type MyCustomType struct {
        A, B int
        C string
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
```

This small bit of boilerplate captures the necessary type information to avoid performing
any reflection while passing data read from the channels, resulting in the same throughput
as a custom implementation for your type.

All SelectFunc implementations look essentially the same, with the only difference being
the element type of the channels in the two type assertions.


## Rationale

Channels provide an elegant mechanism for distributing work among many goroutines, and
there are [many excellent concurrent design patterns](https://blog.golang.org/pipelines) for doing scalable processing
of data using channels.

In particular, fanning data out to a group of worker goroutines and collecting the results
of their work is a useful pattern, but implementing fan-in is nontrivial. A complete
implementation for integers looks like this:

```go
func FanIn(done <-chan struct{}, inputs ...<-chan int) <-chan int {
	results := make(chan int)
	var wg sync.WaitGroup
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
	for _, input := range inputs {
		wg.Add(1)
		go fan(input)
	}
	go func() {
		defer close(results)
		wg.Wait()
	}()
	return results
}
```

This code must be re-implemented for every type of data that you want to send
over the channels, and therefore is extremely non-DRY in pratice.

We wrote this package to implement fan-in once, but in a way that can be re-used
any type of channel element. We also wanted a solution that is approximately as
fast as re-implementing the code for each type.

We've built a system that uses a minimal quantity of reflection (none on the
hot path) combined with anonymous functions that capture type information to achieve
good performance on channels of any element type.

## Benchmarks

The graph below visualizes the relative performance of fanning-in channels three
ways:

- **Concrete**: Code specialized completely to the channel type (like the snippet above)
- **Hybrid**: Our approach with an anonymous function that provides the type information
- **Reflect**: Our fallback 100% reflection-based approach (which is the fastest reflection-
  based option that we could figure out).

![benchmark visualization](https://raw.githubusercontent.com/IBM/fast-fan-in/master/img/benchmarks.png)

Each cluster of bars represents a different number of elements sent through the
fan.

The nine bars in each cluster are the three different approaches using three different
quantities of channels. Here is a key for the bars (there's also one in the bottom right,
but it's hard to read):

1. **Concrete** with 1 channel
2. **Concrete** with 10 channels
3. **Concrete** with 100 channels
4. **Hybrid** with 1 channel
5. **Hybrid** with 10 channels
6. **Hybrid** with 100 channels
7. **Reflect** with 1 channel
8. **Reflect** with 10 channels
9. **Reflect** with 100 channels

The rightmost set of bars reflects the results from the benchmark which sent 100,000
elements through the channels. You can see that the y-axis indicates that the
**Concrete** and **Hybrid** approaches perform nearly the same for all quantities of
channels tested, while the **Reflect** approach suffers a steep penalty.

## Notes

If you have any questions or issues you can create a new [issue here](https://github.com/ibm/fast-fan-in/issues).

Pull requests are very welcome! Make sure your patches are well tested.
Ideally create a topic branch for every separate change you make. For
example:

1. Fork the repo
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Added some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request

## License & Authors

If you would like to see the detailed LICENSE click [here](LICENSE).

- Author: Christopher Waldon  <chris.waldon@ibm.com>

```text
Copyright:: 2019-2020 IBM, Inc

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
