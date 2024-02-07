package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"time"
    "strconv"
)




func listen(port int,received chan int) *net.UDPConn{
    // Resolve UDP address to listen on
    addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", port))

    // Create UDP listener
    conn, _ := net.ListenUDP("udp", addr)

    fmt.Println("Listening for broadcasts...")

    // Start the timer
    

    // Continuously listen for incoming UDP packets
    
    //for  time.Now().Before(deadline) {

        buffer := make([]byte, 1024)
        n, _, _ := conn.ReadFromUDP(buffer)
        fmt.Printf("Received broadcast: %s\n", string(buffer[:n]))
        num, _ := strconv.Atoi(string(buffer[:n]))

        received <- num


        return conn
        //deadline = time.Now().Add(5 * time.Second)
    //}
}

func start_node() *exec.Cmd {
    //Starts a new node
    fmt.Println("Starting new node")

    cmd := exec.Command("gnome-terminal", "--", "go", "run", "counter.go")
    err := cmd.Run()
    if err != nil {
        fmt.Println("Error:", err)
        return cmd
    }
    fmt.Println("Created new node")
    return cmd
}

func kill_self(){
	//Kills itself after a random amount of time
	randTime := time.Duration(rand.Intn(3)+10)*time.Second
    time.Sleep(randTime)
	os.Exit(0)
}

func master(a int){
    //start_node()
    go kill_self()

    port := 12345
    // Send broadcast message every second
    addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", port))

    // Create UDP connection
    conn, _ := net.DialUDP("udp", nil, addr)
    
    for {
        message := []byte(fmt.Sprintf("%d", a))
        _, _ = conn.Write(message)
        fmt.Printf("Broadcasted message: %s\n", string(message))
        time.Sleep(time.Second)
        a++
    }
}

/*
func slave(){
    received := make(chan int)
    go listen(12435,received)
    condition := true
    deadline := time.Now().Add(5 * time.Second)
    for condition{
        select{
            case <- received:
                deadline = time.Now().Add(5 * time.Second)
            case time.Now().Before(deadline):
                condition = false
        }
    }
}
*/

func slave() {
    received := make(chan int)
    fmt.Printf("New Node started")
    

    deadline := time.Now().Add(2 * time.Second)
    go listen(12435, received)
    fmt.Printf("Listend once")
    for{
        select {
        case <-received:
            deadline = time.Now().Add(5 * time.Second)
            go listen(12435, received)
        case <-time.After(time.Until(deadline)):
            break
        }
    }
}



func main(){
    a:=0
    port := 12345
    fmt.Printf("New Node started")
    addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", port))
    // Create UDP listener
    conn, _ := net.ListenUDP("udp", addr)
    //received := make(chan int)
    for {
        conn.SetReadDeadline(time.Now().Add(5 * time.Second))

        fmt.Println("Listening for broadcasts...")
        buffer := make([]byte, 1024)
        n, _, err := conn.ReadFromUDP(buffer)

        if err != nil{
            conn.Close()
            break
        }

        fmt.Printf("Received broadcast: %s\n", string(buffer[:n]))
        a, _ = strconv.Atoi(string(buffer[:n]))

        
        fmt.Printf("Listend once")
    }
    start_node()
    
    for{
        master(a)
    }

    
}



