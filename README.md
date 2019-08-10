# pipes

## Pipelines

A **pipeline** is a series of _stages_ connected by channels, where each stage is a group of goroutines running the same function.
In each stage, the goroutines

* receive values from _upstream_ via **inbound** channels
* perform some function on that data, usually producing new values
* send values _downstream_ via **outbound** channels

Each stage has any number of inbound and outbound channels, except the first and the last stages, which have only outbound and inbound channels, respectively.
The first stage is sometimes called the _source_ or _producer_; the last stage, the _sink_ or _consumer_. [1]

There are some guidelines for pipeline construction:

* stages close their outbound channels when all the send operations are done
* stages keep receiving values from inbound channels until those channels are closed or the senders are unblocked

Pipelines unblock senders either by ensuring there's enough buffer for all the values that are sent or by explicitly signalling sender when the receiver may abandon the channel.

## Fan-out, fan-in

Multiple functions can read from the same channel until that channel is closed; this is called _fan-out_.
This provides a way to distribute work amongst a group of workers to parallelize CPU use and I/O.

A function can read from multiple inputs and proceed until all are closed by multiplexing the input channels onto a single channel that's closed when all the inputs are closed.
This is called _fan-in_.

## Explicit cancellation

When we need a way to tell an unknown and unbounded number of goroutines to stop sending their values downstream, we can do it in Go by closing a channel, because a receive operation on a closed channel can always proceed immediately, yielding the element type's zero value.

This means that `main` can unblock all the senders simply by closing a `done` channel.
This close is effectively a broadcast signal to the senders.

## References

[1](https://blog.golang.org/pipelines) _Go Concurrency Patterns: Pipelines and cancellation_

## License

TBD
