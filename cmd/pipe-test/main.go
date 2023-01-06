package main

import (
	"fmt"
	"io"
	"sync"
	"time"
)

func main() {
	// Create a pipe.
	pr, pw := io.Pipe()

	wg := sync.WaitGroup{}

	wg.Add(1)
	// Use a goroutine to write to the pipe.
	go func() {
		defer wg.Done()
		for {
			fmt.Printf("Writing to pipe\n")
			_, err := fmt.Fprintln(pw, "hello, world")
			// check for closed pipe
			if err == io.ErrClosedPipe {
				fmt.Printf("Pipe closed\n")
				return
			}
			if err != nil {
				fmt.Println(err)
				return
			}
			time.Sleep(time.Second)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Read from the pipe.
		bytes, err := io.ReadAll(pr)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Read %d\n", len(bytes))

	}()

	time.Sleep(4 * time.Second)

	fmt.Println("closing writer")
	pw.Close()

	wg.Wait()
}
