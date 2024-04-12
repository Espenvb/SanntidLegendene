// Provides functions dealing with sockets and IP addresses.
package connection

import (
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
)

// Returns the first non IPv6 and non local only IP address from `net.InterfaceAddrs()`.
// If none found, returns an empty string.
func GetLocalIP() string {
	addresses, _ := net.InterfaceAddrs()

	for _, address := range addresses {
		ipnet, ok := address.(*net.IPNet) // kind of a cast
		//fmt.Println(address.String(), ipnet.Network(), ipnet.IP.IsLoopback(), ipnet.IP.DefaultMask(), ipnet.IP.IsGlobalUnicast())

		if ok && // check type to be *net.IPNet
			ipnet.IP.DefaultMask() != nil && // filter IPv6 addresses
			!ipnet.IP.IsLoopback() { // filter local only addresses
			return strings.Split(address.String(), "/")[0] // ignore sub network
		}
	}

	return ""
}

// Custom dialing to broadcast UDP with socket reuse option and binding `port`
// to the socket. Taken from Network-go `conn` package.
func DialBroadcastUDP(port int) net.PacketConn {
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		panic(fmt.Sprintln("Error: Socket:", err))
	}

	err = syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		panic(fmt.Sprintln("Error: SetSockOpt REUSEADDR:", err))
	}

	err = syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	if err != nil {
		panic(fmt.Sprintln("Error: SetSockOpt BROADCAST:", err))
	}

	err = syscall.Bind(s, &syscall.SockaddrInet4{Port: port})
	if err != nil {
		panic(fmt.Sprintln("Error: Bind:", err))
	}

	f := os.NewFile(uintptr(s), "")
	defer f.Close()

	conn, err := net.FilePacketConn(f)
	if err != nil {
		panic(fmt.Sprintln("Error: FilePacketConn:", err))
	}

	return conn
}

// Custom dialing to a UDP with socket reuse where `local_port` is binded to.
func DialUDP(local_port int) net.PacketConn {
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		panic(fmt.Sprintln("Error: Socket:", err))
	}

	err = syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		panic(fmt.Sprintln("Error: SetSockOpt REUSEADDR:", err))
	}

	err = syscall.Bind(s, &syscall.SockaddrInet4{Port: local_port})
	if err != nil {
		panic(fmt.Sprintln("Error: Bind:", err))
	}

	f := os.NewFile(uintptr(s), "")
	defer f.Close()

	conn, err := net.FilePacketConn(f)
	if err != nil {
		panic(fmt.Sprintln("Error: FilePacketConn:", err))
	}

	return conn
}
