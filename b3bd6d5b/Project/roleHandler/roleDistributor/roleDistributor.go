package roleDistributor

import (
	"Project/network/udpBroadcast/udpNetwork/localip"
	"Project/network/udpBroadcast/udpNetwork/peers"
	"bytes"
	"fmt"
	"net"
	"sort"
)

type RoleAndSortedAliveElevs struct {
	Role             string
	SortedAliveElevs []net.IP
}

type Role int

const (
	Master Role = iota // 0
	Backup             // 1
	Dummy              // 2
)

func (role Role) String() string {
	switch role {
	case Master:
		return "Master"
	case Backup:
		return "Backup"
	default:
		return "Dummy"
	}
}

func RoleDistributor(
	peerUpdateToRoleDistributorCh chan peers.PeerUpdate,
	roleAndSortedAliveElevs chan<- RoleAndSortedAliveElevs,
	masterIPCh chan net.IP,
) {
	localIPstr, err := localip.LocalIP()
	if err != nil {
		fmt.Printf("Could not get local ip: %v\n", err)
	}
	localIP := net.ParseIP(localIPstr)
	localElevInPeers := false
	for {
		p := <-peerUpdateToRoleDistributorCh
		sortedIPs := make([]net.IP, 0, len(p.Peers))
		for _, ip := range p.Peers {
			peerIP := net.ParseIP(ip)
			sortedIPs = append(sortedIPs, peerIP)
			if peerIP.Equal(localIP) {
				localElevInPeers = true
			}
		}
		if !localElevInPeers {
			break
		}
		sort.Slice(sortedIPs, func(firstIpIndex, secondIpIndex int) bool {
			return bytes.Compare(sortedIPs[firstIpIndex], sortedIPs[secondIpIndex]) < 0
		})
		checkRoles := func(sortedIPs []net.IP) string {
			for ipIndex, ip := range sortedIPs {
				var expectedRole Role
				switch ipIndex {
				case 0:
					expectedRole = Master
				case 1:
					expectedRole = Backup
				default:
					expectedRole = Dummy
				}
				if ip.Equal(localIP) {
					return expectedRole.String()
				}
			}
			return ""
		}
		newRole := ""
		if len(p.Lost) > 0 {
			newRole = checkRoles(sortedIPs)
		}
		if p.New != "" {
			newRole = checkRoles(sortedIPs)
		}
		roleAndSortedAliveElevs <- RoleAndSortedAliveElevs{newRole, sortedIPs}
		masterIPCh <- sortedIPs[int(Master)]
	}
}
