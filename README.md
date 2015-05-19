## Ringo Mundo

A very simple and optimized package to help manage a ring buffer use between publishers and consumers.

### What is a ring buffer?

TODO

### How this package works

One of more publishers use a Leader object to coordinate and place information into a ringbuffer, while one or more consumers use a Follower object to coordinate and process information from a cyclic buffer of work. The Leader and Follower objects validate they do not overwrite each other in the ring by reading each other's commmit status.

### Instructions

To create a ring-buffer, first pre-allocate a specialized array of work you want to track.

For example:
```
ringSize := PT32Meg
ring []*MyWorkStruct := make([]*MyWorkStruct, ringSize)
for i = 0; i < ringSize; i++ {
	ring[i] = &MyWorkStruct{foo: 0}
}
```
Next, we create the network of components to manage this ring.  This is simply done using a factory function to wire the network together.  For example, a simple publisher/subscriber network:
```
network := ringo.SimplePair(ringSize)
```
Inside the network is a Leader and Follower object:

A Leader is used to synchronize the work to be written to the buffer.
A Follower is used to synchonize the work to be consumed from the buffer.
Each is dependent on the other.  The Follower cannot pass the leader.  The Leader cannot
pass the Follower if the Follower has not completed the previous iteration of work.
In essense, each head is chasing the others tail.

Each CPU should be assigned to access either a Leader or Follower. We want one native thread per component.

The size should be a power of two.  See the constant file for example sizes. You should pick a ring buffer size based on available machine memory and fine tune the size as needed.  The larger the buffer the less number of rotations and possible contentions.

A publisher go routine would be coded like this to get work into the buffer:
```
mask := network.Leader.Mask()
index := network.Leader.Reserve()  // Reserve a new index
ring[index&mask].foo = 99          // Store some data into a work slot of the ring.
network.Leader.Commit(index)       // Mark as committed = done.
```
A consumer thread would be coded something like this to remove and process work:
```
mask := network.Follower.Mask()
index := network.Follower.Reserve()  // Reserve a new index.
data = ring[index&mask].foo          // Read some data from the slot.
Process(data)						   // Process it.
network.Follower.Commit(index)       // Mark as committed = done.

```
