# worker-graceful-rollover

Planck service for managing graceful rollover during worker deployment.

## Usage

    $ go run cmd/worker-graceful-rollover/main.go -capacity 1

    (in separate tab 1)

    $ nc 127.0.0.1 8080

    (in separate tab 2)

    $ nc 127.0.0.1 8080

    (in tab 1)

    ^D

    (in tab 2)

    ^D

## License

MIT.
