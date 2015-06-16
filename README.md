# Ringo Mundo
[![License MIT](https://img.shields.io/npm/l/express.svg)](http://opensource.org/licenses/MIT)
[![Build Status](https://travis-ci.org/composer22/ringo-mundo.svg?branch=master)](http://travis-ci.org/composer22/ringo-mundo)
[![Current Release](https://img.shields.io/badge/release-v0.1.0-brightgreen.svg)](https://github.com/composer22/chattypantz/releases/tag/v0.1.0)
[![Coverage Status](https://coveralls.io/repos/composer22/ringo-mundo/badge.svg?branch=master)](https://coveralls.io/r/composer22/ringo-mundo?branch=master)

A very simple and optimized package to help manage ring buffers, written in [Golang.](http://golang.org).

## What This Package Does

In creating queues for an application, Go channels, object allocation, and locks take up great amounts of processing time. This package provides a toolkit to eliminate contentions and unecessary memory allocation and processing.

## What is a Ring Buffer?

A ring buffer is simply an array of values who's head wraps around to the first slot.  For example, an 8 position ring buffer:
```
Actual Cell        [0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 0 | 1 | 2 | 3 | ...]
Increment Index     0   1   2   3   4   5   6   7   8   9   10  11  ...

Using modulus arithmetic any incremented index can be converted to the actual index of an array.

IncrIndex % size = actualIndex e.g. cell
9 % 8 = 1
10 % 8 = 2

A faster computated way is to use masking:
arr := make([]int, 8)
mask = len(arr) - 1
one := 9&mask // is 1
two := 10&mask // is 2
```
Ring buffers are a good structure to use in queuing work for processing. A publisher can move through the buffer writing work to be done.  A consumer can follow this trail processing the jobs. When the publisher reaches the end, it can return to the beginning without checking for the end of the array or resetting it's index and continue to write new work to the beginning slots that have been processed by the consumer. Locks can be eliminated by reading each other's position and avoiding passing.

For more information on ring buffers, see: [Circular Buffers](http://en.wikipedia.org/wiki/Circular_buffer).

## How This Package Works

One of more publishers use a Publisher object to coordinate and place information into a ringbuffer, while one or more processor routines use Consumer objects to coordinate and process information from the cyclic buffer of work. The Publisher and Consumer objects validate they do not overwrite each other in the ring by reading each other's counters or status arrays.

To help facilitate multiplexing of Consumers, and to gate any dependencies, a Barrier object is also provided.

## Performance

The following benchmarks were gathered on a development machine.
```
MBP 15-inch, Mid 2014
2.5 GHz Intel Core i7 Quad
16 GB 1600 MHz DDR3
OSX 10.10.3

Simple Queue - One CPU worked best in testing.
=============================================
SinglePublisher:         219.7 million transactions per second (4.55 ns/op)
MultiPublisher:          72.4 million transactions per second (13.8 ns/op)
Using Go Channel:        30.2 million transactions per second (33.1 ns/op)

Disruptor Pattern - One CPU worked best in testing.
===================================================
SinglePublisher:         98.0 million transactions per second (10.2 ns/op)
MultiPublisher:          50.0 million transactions per second (20.0 ns/op)
```
## Getting Started

To create a ring-buffer, first pre-allocate a specialized array of work you want to track.

For example:
```
ringSize := PT32Meg
mask := ringSize - 1
ring []*MyWorkStruct := make([]*MyWorkStruct, ringSize)
for i = 0; i < ringSize; i++ {
	ring[i] = &MyWorkStruct{foo: 0}
}
```
Note that the size is expressed as a power of two and is quite large. See the constant file for example sizes. You should pick a ring buffer size based on available machine memory and fine tune the size as needed.  The larger the buffer the less number of rotations and possible contentions.

Next, we create the network of components to manage this ring. For example, a simple publish/subscribe queue:
```
publisher := ringo.SimplePublishNodeNew(ringSize)
consumer := ringo.SimpleConsumeNodeNew()
// Set each component to check the others committed counters.
publisher.SetDependency(consumer.Committed())
consumer.SetDependency(publisher.Committed())
```

The publisher is used to synchronize the work to be written to the buffer.
A consumer is used to synchonize the work to be consumed from the buffer for processing.
Each is dependent on the other to finish its work, from rotation to rotation.  The consumer cannot pass the publisher.  The publisher cannot pass the consumer. In essense, each head is chasing the others tail in a circle.

A publisher go routine would be coded to get work into the buffer:
```
for {
  index := publisher.Reserve()  // Reserve a new index
  ring[index&mask].foo = 99  // Store some data into a work slot of the ring.
  publisher.Commit()            // Mark as done.
}
```
A consumer go routine would be coded to remove and process work:
```
for {
  index := consumer.Reserve()  // Reserve a new index.
  data = ring[*index&mask].foo // Read some data from the slot.
  Process(data)                // Process it.
  consumer.Commit()            // Mark as done.
}

```
The *index&mask is the same as index % size.

For a more complex example, see the Disruptor pattern, which demonstrates mutiplexing and chained consumers (disruptor_test.go).

## Building

This code currently requires version 1.42 or higher of Go.

Information on Golang installation, including pre-built binaries, is available at
<http://golang.org/doc/install>.

Run `go version` to see the version of Go which you have installed.

Run `go build` inside the directory to build.

Run `go test ./...` to run the unit regression tests.

Run `go install` installs the package into your local repo.

A successful build run produces no messages and publishes the package to your path.

Run `go help` for more guidance, and visit <http://golang.org/> for tutorials, presentations, references and more.

## License

(The MIT License)

Copyright (c) 2015 Pyxxel Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to
deal in the Software without restriction, including without limitation the
rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
IN THE SOFTWARE.
