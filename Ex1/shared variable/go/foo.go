// Use `go run foo.go` to run your program

package main

import (
    . "fmt"
    "runtime"
    "time"
)

var i = 0

func server(increment chan int, quitInc chan int, decrement chan int, quitDec chan int){
    for {
        select{
        case <- increment:
            i++
        case <- decrement:
            i--
        case <- quitDec:
            return
        }
    }

}

func incrementing(increment chan int, quitInc chan int) {
    for a := 0; a < 1000; a++{
        increment <- 1
    }
    quitInc <- 1
}

func decrementing(decrement chan int, quitDec chan int) {
    for a := 0; a < 10; a++{
        decrement <- 1
    }
    quitDec <- 1
}


func main() {
    // What does GOMAXPROCS do? What happens if you set it to 1?
    runtime.GOMAXPROCS(2)    
	
    increment := make(chan int)
    decrement := make(chan int)
    quitInc := make(chan int)
    quitDec := make(chan int)



    // TODO: Spawn both functions as goroutines
    go incrementing(increment, quitInc)
    go decrementing(decrement, quitDec)
    go server(increment, quitInc, decrement, quitDec)



    
	
    // We have no direct way to wait for the completion of a goroutine (without additional synchronization of some sort)
    // We will do it properly with channels soon. For now: Sleep.
    time.Sleep(500*time.Millisecond)
    Println("The magic number is:", i)
}
