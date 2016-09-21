package main

import (
	"flag"
	"log"
	"net"
	"os"

	"github.com/travis-ci/worker-graceful-rollover"
)

var capacity = flag.Int("capacity", 1, "available capacity")
var addr = flag.String("addr", "127.0.0.1:8080", "address to bind to")

func main() {
	flag.Parse()

	sem := make(rollover.Semaphore, *capacity)

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("error: could not bind to %v: %v", *addr, err)
	}

	sink := make(chan error)
	go func() {
		for err := range sink {
			os.Stderr.Write([]byte(err.Error()))
		}
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf("error: could not accept connection: %v", err)
			panic(err)
		}
		go rollover.HandleConnection(sem, conn, conn, &sink)
	}
}
