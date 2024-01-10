// Use `go run foo.go` to run your program

package main

import (
    . "fmt"
    "runtime"
)

func server(increment chan int, read chan int, decrement chan int){
    var i = 0
    for {
        select{
        case <- increment:
            i++
        case <- decrement:
            i--
        case read <- i:

    }
}
}

func incrementing(increment chan int, quit chan int) {
    for a := 0; a <= 1000; a++{
        increment <- 1
    }
    quit <- 1
}

func decrementing(decrement chan int, quit chan int) {
    for a := 0; a <= 10; a++{
        decrement <- 1
    }
    quit <- 1
}


func main() {
    // What does GOMAXPROCS do? What happens if you set it to 1?
    runtime.GOMAXPROCS(2)    
	
    increment := make(chan int)
    decrement := make(chan int)
    quit := make(chan int)
    read := make(chan int)




    // TODO: Spawn both functions as goroutines
    go incrementing(increment, quit)
    go decrementing(decrement, quit)
    go server(increment, read, decrement)

    <- quit
    <- quit
    
    Println(<-read)

    return

    
	
    // We have no direct way to wait for the completion of a goroutine (without additional synchronization of some sort)
    // We will do it properly with channels soon. For now: Sleep.
    //time.Sleep(500*time.Millisecond)
    
}
