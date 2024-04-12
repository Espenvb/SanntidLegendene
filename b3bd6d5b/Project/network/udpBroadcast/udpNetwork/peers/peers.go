package peers

import (
	"Project/network/udpBroadcast/udpNetwork/conn"
	"fmt"
	"net"
	"sort"
	"time"
)

type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}

const interval = 15 * time.Millisecond
const timeout = 500 * time.Millisecond

func Transmitter(port int, id string, transmitEnable <-chan bool) {
	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))
	enable := true
	for {
		select {
		case enable = <-transmitEnable:
		case <-time.After(interval):
		}
		if enable {
			conn.WriteTo([]byte(id), addr)
		}
	}
}

func Receiver(port int, peerUpdateCh chan<- PeerUpdate) {
	var buf [1024]byte
	var peerUpd PeerUpdate
	lastSeen := make(map[string]time.Time)
	conn := conn.DialBroadcastUDP(port)
	for {
		updated := false
		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])
		id := string(buf[:n])
		peerUpd.New = ""
		if id != "" {
			if _, idExists := lastSeen[id]; !idExists {
				peerUpd.New = id
				updated = true
			}

			lastSeen[id] = time.Now()
		}
		peerUpd.Lost = make([]string, 0)
		for index, value := range lastSeen {
			if time.Now().Sub(value) > timeout {
				updated = true
				peerUpd.Lost = append(peerUpd.Lost, index)
				delete(lastSeen, index)
			}
		}
		if updated {
			peerUpd.Peers = make([]string, 0, len(lastSeen))
			for index, _ := range lastSeen {
				peerUpd.Peers = append(peerUpd.Peers, index)
			}
			sort.Strings(peerUpd.Peers)
			sort.Strings(peerUpd.Lost)
			peerUpdateCh <- peerUpd
		}
	}
}
