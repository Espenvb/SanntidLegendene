package main

import(
	"net"
	"fmt"
	"time"
	"os/exec"
)


func broadcast(message []byte, port int) {
    // Resolve UDP broadcast address
    addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", port))

    // Create UDP connection
    conn, _ := net.DialUDP("udp", nil, addr)
    defer conn.Close()

    // Send broadcast message every second
    for {
        _, _ = conn.Write(message)
        fmt.Printf("Broadcasted message: %s\n", string(message))
        time.Sleep(time.Second)
    }
}

func start_node() *exec.Cmd {
    //Starts a new node
    fmt.Println("Starting new node")

    cmd := exec.Command("gnome-terminal", "--", "go", "run", "listener.go")
    err := cmd.Run()
    if err != nil {
        fmt.Println("Error:", err)
        return cmd
    }
    fmt.Println("Created new node")
    return cmd
}



func main(){

	start_node()
	go broadcast([]byte("Hello UDP broadcast!"), 12345)

	select{}


}