# worker-graceful-rollover

Planck service for managing graceful rollover during worker deployment.

## Why does it exist

When we deploy a new version of worker, we need to gracefully shut down the process, since there will be jobs that are still being processed. Due to the long-lived SSH and AMQP connections in worker, we cannot shut it down forcefully. However, such a graceful shutdown can take up to two hours (the maximum build timeout).

In order to automate the restarting process we want to allocate a fixed capacity per worker host. All worker processes on that host will share that capacity. `worker-graceful-rollover` is responsible for managing that capacity.

## How

When the `worker-graceful-rollover` process is started, its capacity is defined. It will then allow anyone to request a "slot" by opening a TCP connection. Once the slot is allocated, the server responds with a the byte sequence `.\n` ("\x2e\x0a"). As long as the TCP connection is active, the slot is allocated to that client. Once a client closes the TCP connection, the slot is freed.

Essentially this program is an implementation of a semaphore. A counter for concurrent access that may never go under a certain threshold, meaning it will never be < 0.

## Usage

    $ go run cmd/worker-graceful-rollover.go -capacity 1

    (in separate tab 1)

    $ nc 127.0.0.1 8080

    (in separate tab 2)

    $ nc 127.0.0.1 8080

    (in tab 1)

    ^D

    (in tab 2)

    ^D

## Plans

* Integrate with worker
* Eventually replace with consul or similar

## License

MIT.
