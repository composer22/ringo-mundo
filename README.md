## Ringo Mundo

A very simple and optimized package to help manage a ring buffer use between publishers and consumers.

### What is a ring buffer?

TODO

### How this package works

One of more publishers use a Master object to coordinate and place information into a ringbuffer, while one or more consumers use Slave objects to coordinate and process information from a cyclic buffer of work. The Master and Slave objects validate they do not overwrite each other in the ring by reading each other's counters or status arrays.

### Instructions

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
