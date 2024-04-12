// +build windows

package conn


import (
    "context"
	"fmt"
	"net"
	"syscall"
)

func DialBroadcastUDP(port int) net.PacketConn {
    config := &net.ListenConfig{Control: 
        func (network, address string, conn syscall.RawConn) error {
            return conn.Control(func(descriptor uintptr) {
                syscall.SetsockoptInt(syscall.Handle(descriptor), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
                syscall.SetsockoptInt(syscall.Handle(descriptor), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
            })
        },
    }

	conn, err := config.ListenPacket(context.Background(), "udp4", fmt.Sprintf(":%d", port)) 
	if err != nil { fmt.Println("Error: net.ListenConfig.ListenPacket:", err) }

	return conn
}
