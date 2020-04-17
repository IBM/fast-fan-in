# fast-fan-in

[![GoDoc](https://godoc.org/github.com/IBM/fast-fan-in?status.svg)](https://godoc.org/github.com/IBM/fast-fan-in)

Golang fan-in pattern efficiently adaptable to any channel type without code generation.

For context on the fan-in pattern (and fan-out, the inverse), see [this blog post](https://blog.golang.org/pipelines).

For an end-to-end example of using this library with fan out and fan in, see [this example](https://github.com/IBM/fast-fan-in/blob/master/fan-in_test.go#L303).

- [Usage](#usage)
- [Rationale](#rationale)
- [Benchmarks](#benchmarks)
- [Notes](#notes)

## Usage

This package provides convenient and efficient mechanisms for combining many channels
with the same element type into one channel with that element type.

### Primitive Types

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

### Custom Types

For non-primitive types, you can achieve good performance by providing an anonymous function
that type-asserts the channels to the appropriate element type (avoiding reflection on the
hot path):

```go
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
```

This small bit of boilerplate captures the necessary type information to avoid performing
any reflection while passing data read from the channels, resulting in the same throughput
as a custom implementation for your type.

All SelectFunc implementations look essentially the same, with the only difference being
the element type of the channels in the two type assertions.

### Custom Types with Reflection

If your use-case is not performance-critical, we also provide a reflection-based fallback
implementation which is used when no SelectFunc is provided. See [benchmarks](#benchmarks)
to understand the performance effect of this implementation.

To use the inefficient reflection-based approach on a custom type, you can do:

```go
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
```

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
for any type of channel element. We also wanted a solution that is approximately as
fast as re-implementing the code for each type.

We've built a system that uses a minimal quantity of reflection (none on the
hot path) combined with anonymous functions that capture type information to achieve
good performance on channels of any element type.

## Benchmarks

The graph below visualizes the relative performance of fanning-in channels three
ways:

- **Concrete**: Code specialized completely to the channel type (like the snippet above in [rationale](#rationale))
- **Hybrid-Closure**: Our [approach with an anonymous function](#custom-types) that provides the type information
- **Hybrid-Reflect**: Our [fallback 100% reflection-based approach](#custom-types-with-reflection) (which is the fastest reflection-based option that we could figure out).

![benchmark visualization](https://raw.githubusercontent.com/IBM/fast-fan-in/master/img/benchmarks.png)

As the chart demonstrates, the **Hybrid-Closure** approach can realize nearly the same performance as using a type-specialized rewrite, even when managing many channels.

You can replicate these benchmarks by running the following (this process takes a long time):

```shell
go test -bench=. -benchtime=30s -timeout 1h -run="" 2>&1 | tee benchmark-data
cd benchutils; go run ./cmd/vizsimple/ -elements 100000 -output 100000-elements.png < benchmark-data
```

The file `100000-elements.png` will be a visualization like the above, but populated with your benchmark results.

The tool supports slicing the benchmark data with different numbers of elements and channels, see the source for details.

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
