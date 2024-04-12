// Use `go run foo.go` to run your program

package main

import (
	. "fmt"
	"runtime"
)

func counter_server(inc, dec, finished, get_counter chan int) {
	i := 0
	finished_counter := 0
	for {
		select {
		case <-inc:
			i++
		case <-dec:
			i--
		case <-finished:
			finished_counter++
			if finished_counter == 2 {
				get_counter <- i
			}
		}
	}
}

func incrementing(inc, finished chan int) {

	for j := 0; j < 1000000; j++ {
		inc <- 0
	}
	finished <- 0
}

func decrementing(dec, finished chan int) {
	
	for j := 0; j < 1000000; j++ {
		dec <- 0
	}
	finished <- 0
}

func main() {
	// What does GOMAXPROCS do? What happens if you set it to 1?
	runtime.GOMAXPROCS(2)

	inc := make(chan int)
	dec := make(chan int)
	get_counter := make(chan int)
	finished := make(chan int)

	go incrementing(inc, finished)
	go decrementing(dec, finished)
	go counter_server(inc, dec, finished, get_counter)

	counter_value := <-get_counter
	Println("The magic number is:", counter_value)
}
