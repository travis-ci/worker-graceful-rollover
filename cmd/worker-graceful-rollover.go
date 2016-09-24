package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var capacity = flag.Int("capacity", 0, "available capacity")
var addr = flag.String("addr", "127.0.0.1:8080", "address to bind to")

type state struct {
	Capacity int
	Locks    map[string]time.Time
}

func server(control, queue, heartbeats chan string) {
	st := new(state)
	st.Locks = make(map[string]time.Time)

	err := readStateFile("state.json", st)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading state: %v\n", err)
	}

	expiryTimer := time.NewTimer(time.Second * 2)
	persistTimer := time.NewTimer(time.Second * 10)
	pollTimer := time.NewTimer(time.Second * 1)

	for {
		log.Print("server: entering loop")

		available := st.Capacity - len(st.Locks)
		if available > 0 {
			log.Print("server: slots are available, polling queue")
			select {
			case id := <-queue:
				log.Printf("server: acquire %v", id)
				st.Locks[id] = time.Now().UTC()
			default:
				log.Printf("server: queue empty")
			}
		}

		select {
		case cmd := <-control:
			switch {
			case cmd == "inc-capacity":
				log.Print("server: inc-capacity")
				st.Capacity++

			case cmd == "dec-capacity":
				log.Print("server: dec-capacity")
				st.Capacity--
				if st.Capacity < 0 {
					st.Capacity = 0
				}

			case strings.HasPrefix(cmd, "set-capacity:"):
				log.Print("server: set-capacity")
				args := strings.SplitN(cmd, ":", 2)
				n, err := strconv.Atoi(args[1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "error parsing set-capacity args: %v\n", err)
					continue
				}
				st.Capacity = n
				if st.Capacity < 0 {
					st.Capacity = 0
				}

			case cmd == "status":
				log.Print("server: status")
				available := st.Capacity - len(st.Locks)
				fmt.Printf("capacity: %v\n", st.Capacity)
				fmt.Printf("slots in use: %v\n", len(st.Locks))
				fmt.Printf("available slots: %v\n", available)
			}

		case id := <-heartbeats:
			log.Printf("server: heartbeat %v", id)
			if _, ok := st.Locks[id]; ok {
				st.Locks[id] = time.Now().UTC()
			}

		case <-expiryTimer.C:
			log.Print("server: expiry")
			var expiredIDs []string
			expiry := time.Now().Add(-1 * time.Minute)

			for id, timestamp := range st.Locks {
				if timestamp.Before(expiry) {
					expiredIDs = append(expiredIDs, id)
				}
			}

			for _, id := range expiredIDs {
				log.Printf("server: expired id %v", id)
				delete(st.Locks, id)
			}

		case <-persistTimer.C:
			log.Print("server: persist")
			err := writeStateFile("state.json", st)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error persisting state: %v\n", err)
			}

		case <-pollTimer.C:
			log.Print("server: poll")
		}
	}
}

func readStateFile(filename string, st *state) error {
	if _, err := os.Stat(filename); err == nil {
		log.Printf("reading state file")

		f, err := os.Open(filename)
		if err != nil {
			return errors.Wrap(err, "error opening state file")
		}

		err = json.NewDecoder(f).Decode(st)
		if err != nil {
			return errors.Wrap(err, "error decoding state from json")
		}

		err = f.Close()
		if err != nil {
			return errors.Wrap(err, "error closing state file")
		}

		log.Printf("resetting lock timestamps")

		for id := range st.Locks {
			st.Locks[id] = time.Now().UTC()
		}
	}

	return nil
}

func writeStateFile(filename string, st *state) error {
	f, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err, "error opening state file")
	}

	err = json.NewEncoder(f).Encode(st)
	if err != nil {
		return errors.Wrap(err, "error encoding state as json")
	}

	err = f.Sync()
	if err != nil {
		return errors.Wrap(err, "error syncing state file to disk")
	}

	err = f.Close()
	if err != nil {
		return errors.Wrap(err, "error closing state file")
	}

	return nil
}

func handleConnection(conn net.Conn, queue, heartbeats chan string) error {
	defer conn.Close()

	log.Print("conn: accept")

	id, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return err
	}
	id = strings.TrimSpace(id)

	log.Printf("conn[%s]: received id", id)

	queue <- id

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		switch line {
		case "ping":
			heartbeats <- id
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func main() {
	flag.Parse()

	control := make(chan string)
	queue := make(chan string)
	heartbeats := make(chan string)

	log.Print("starting server")

	go server(control, queue, heartbeats)

	if *capacity > 0 {
		log.Print("setting capacity")
		control <- "set-capacity:" + strconv.Itoa(*capacity)
	}

	log.Print("setting up status ticker")

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				control <- "status"
			}
		}
	}()

	log.Printf("listening on port %s", *addr)

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("error: could not bind to %v: %v", *addr, err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf("error: could not accept connection: %v", err)
			continue
		}
		go func() {
			err := handleConnection(conn, queue, heartbeats)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error handling connection: %v\n", err)
				fmt.Fprintf(os.Stderr, "error handling connection: %v\n", err)
			}
		}()
	}
}
