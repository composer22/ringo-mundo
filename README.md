# Ringo Mundo
[![License MIT](https://img.shields.io/npm/l/express.svg)](http://opensource.org/licenses/MIT)
[![Build Status](https://travis-ci.org/composer22/ringo-blingo.svg?branch=master)](http://travis-ci.org/composer22/ringo-blingo)
[![Current Release](https://img.shields.io/badge/release-v0.1.0-brightgreen.svg)](https://github.com/composer22/chattypantz/releases/tag/v0.1.0)
[![Coverage Status](https://coveralls.io/repos/composer22/ringo-blingo/badge.svg?branch=master)](https://coveralls.io/r/composer22/ringo-blingo?branch=master)

A very simple and optimized package to help manage a ring buffer use between publishers and consumers  written in [Go.](http://golang.org).

## About

TODO

## What is a ring buffer?

TODO

## How this package works

One of more publishers use a Master object to coordinate and place information into a ringbuffer, while one or more consumers use Slave objects to coordinate and process information from a cyclic buffer of work. The Master and Slave objects validate they do not overwrite each other in the ring by reading each other's counters or status arrays.

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
Next, we create the network of components to manage this ring.  This is simply done using a factory function to wire the network together.  For example, a simple publisher/subscriber queue:
```
queue := ringo.SimpleQueue(ringSize)
```
Inside the queue is a Master and Slave object:

A Master is used to synchronize the work to be written to the buffer.
A Slave is used to synchonize the work to be consumed from the buffer.
Each is dependent on the other from rotation to rotation.  The Slave cannot pass the Master.  The Master cannot pass the Slave if the Slave has not completed the previous iteration of work. In essense, each head is chasing the others tail.

Each CPU should be assigned to access either a Master or Slave. We want one native thread per component.

The size should be a power of two.  See the constant file for example sizes. You should pick a ring buffer size based on available machine memory and fine tune the size as needed.  The larger the buffer the less number of rotations and possible contentions.

A publisher go routine would be coded like this to get work into the buffer:
```
index := queue.Master.Reserve()  // Reserve a new index
ring[index&mask].foo = 99        // Store some data into a work slot of the ring.
queue.Master.Commit()            // Mark as done.
```
A consumer thread would be coded something like this to remove and process work:
```
index := queue.Slave.Reserve()  // Reserve a new index.
data = ring[index&mask].foo     // Read some data from the slot.
Process(data)                   // Process it.
queue.Slave.Commit()            // Mark as done.

```
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
