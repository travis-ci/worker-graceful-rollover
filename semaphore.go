package rollover

import (
	"bufio"
	"fmt"
	"io"
)

type Semaphore chan interface{}

func HandleConnection(sem Semaphore, conn io.ReadCloser, notify io.Writer, sink *chan error) {
	fmt.Println("accepted connection")

	defer conn.Close()
	done := make(chan interface{})

	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
		}
		if err := scanner.Err(); err != nil {
			if sink != nil {
				*sink <- err
			}
		}
		done <- struct{}{}
	}()

	select {
	case sem <- struct{}{}:
		// acquire semaphore
		fmt.Println("acquired")
		notify.Write([]byte(".\n"))
		defer func() {
			// release semaphore
			<-sem
			fmt.Println("released")
		}()
		// wait for connection to complete
		<-done
		fmt.Println("done")
	case <-done:
		// cancel acquisition
		fmt.Println("canceled")
	}
}
