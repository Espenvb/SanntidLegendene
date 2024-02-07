package main

import(
	"net"
	"fmt"


)


func listen(port int) {
    // Resolve UDP address to listen on
    addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", port))

    // Create UDP listener
    conn, _ := net.ListenUDP("udp", addr)
    defer conn.Close()

    fmt.Println("Listening for broadcasts...")

    // Continuously listen for incoming UDP packets
    deadline := time.Now().Add(n * time.Second)
    for time.Now().Before(deadline) {
        buffer := make([]byte, 1024)
        n, _, _ := conn.ReadFromUDP(buffer)
        fmt.Printf("Received broadcast: %s\n", string(buffer[:n]))
    }
}


func main() {

go listen(12345)

select{}


}