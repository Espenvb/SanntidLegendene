package peers

import (
	"Heis/network/conn"
	"Heis/network/localip"
	"flag"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"time"
)

type PeerUpdate struct {
	Peers  []string
	New    string
	Lost   []string
	Master string
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
	var p PeerUpdate
	lastSeen := make(map[string]time.Time)

	conn := conn.DialBroadcastUDP(port)

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])

		id := string(buf[:n])

		// Adding new connection
		p.New = ""
		if id != "" {
			if _, idExists := lastSeen[id]; !idExists {
				p.New = id
				updated = true
			}

			lastSeen[id] = time.Now()
		}

		// Removing dead connection
		p.Lost = make([]string, 0)
		for k, v := range lastSeen {
			if time.Now().Sub(v) > timeout {
				updated = true
				p.Lost = append(p.Lost, k)
				delete(lastSeen, k)
			}
		}

		// Adding new Master
		if id != "" {
			if _, idExists := lastSeen[id]; !idExists {
				p.Master = id
				updated = true
			}

			lastSeen[id] = time.Now()
		}

		// Sending update
		if updated {
			p.Peers = make([]string, 0, len(lastSeen))

			for k, _ := range lastSeen {
				p.Peers = append(p.Peers, k)
			}

			sort.Strings(p.Peers)
			sort.Strings(p.Lost)
			p.Master = determineMaster(p.Peers)
			peerUpdateCh <- p
		}
	}
}


func PeerUpdates(id string, peerCh1 chan PeerUpdate, peerCh2 chan PeerUpdate, isMasterCh1 chan bool, isMasterCh2 chan bool, sendMasterIdToReceive chan string, sendMasterIdToNotify chan string) {
	for {
		p := <-peerCh1
		PrintUpdatedPeers(p)
		if p.Master != "" {
			masterId := p.Master
			sendMasterIdToReceive <- masterId
			sendMasterIdToNotify <- masterId
			if id == masterId {
				isMasterCh1 <- true
				isMasterCh2 <- true
			} else {
				isMasterCh1 <- false
				isMasterCh2 <- false
			}
		}
		peerCh2 <- p
	}
}


func determineMaster(peers []string) string {
	for len(peers) > 0 {
		return peers[0]
	}
	return ""
}

func PrintUpdatedPeers(p PeerUpdate) {
	fmt.Printf("Peer update:\n")
	fmt.Printf("  Peers:    %q\n", p.Peers)
	fmt.Printf("  New:      %q\n", p.New)
	fmt.Printf("  Lost:     %q\n", p.Lost)
	fmt.Printf("  Master:     %q\n", p.Master)
}

func MakeId() string {
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}
	return id
}


func ExtractIpFromPeer(peer string) string {
	data := strings.Split(peer, "-")
	return data[1]
}


