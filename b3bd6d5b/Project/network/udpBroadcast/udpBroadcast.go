package udpBroadcast

import (
	"Project/network/udpBroadcast/udpNetwork/bcast"
	"Project/network/udpBroadcast/udpNetwork/localip"
	"Project/network/udpBroadcast/udpNetwork/peers"
	"flag"
	"fmt"
)

type HelloMsg struct {
	Message string
	Iter    int
}

func StartPeerBroadcasting(peerUpdateToPrimaryHandlerCh chan peers.PeerUpdate, peerTxEnable chan bool) {
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = string(localIP)
	}
	peerUpdateCh := make(chan peers.PeerUpdate)
	go peers.Transmitter(15645, id, peerTxEnable)
	go peers.Receiver(15645, peerUpdateCh)
	helloTx := make(chan HelloMsg)
	helloRx := make(chan HelloMsg)
	go bcast.Transmitter(16565, helloTx)
	go bcast.Receiver(16565, helloRx)
	for {
		newPeerUpd := <-peerUpdateCh
		if localIPInPeers(newPeerUpd) {
			peerUpdateToPrimaryHandlerCh <- newPeerUpd
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", newPeerUpd.Peers)
			fmt.Printf("  New:      %q\n", newPeerUpd.New)
			fmt.Printf("  Lost:     %q\n", newPeerUpd.Lost)
		}
	}
}

func localIPInPeers(newPeerUpd peers.PeerUpdate) bool {
	localIP, _ := localip.LocalIP()
	for _, peerIP := range newPeerUpd.Peers {
		if localIP == peerIP {
			return true
		}
	}
	return false
}
