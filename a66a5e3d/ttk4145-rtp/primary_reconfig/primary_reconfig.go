package primary_reconfig

import (
	"fmt"
	"math/rand"
	"project/alive"
	pt "project/project_types"
	"time"
)

type Role string

const (
	r_unknown        Role = "unknown"
	r_determineRole  Role = "determineRole"
	r_setupPrimary   Role = "setupPrimary"
	r_runningPrimary Role = "runningPrimary"
	r_setupBackup    Role = "setupBackup"
	r_runningBackup  Role = "runningBackup"
)

var _initialized = false
var _node pt.Node

func Init(node pt.Node) {
	if _initialized {
		panic("trying to re-initialize primary_reconfig package!")
	}
	_initialized = true
	_node = node
}

func ReconfigurationService(
	// input channels:
	nodeUpdateChanFromAlive <-chan pt.NodeUpdate,
	primaryAnnounceChanRx <-chan pt.PrimaryAnnounce,
	// output channels:
	nodeUpdateChanToPrim chan<- pt.NodeUpdate,
	primaryAnnounceChanTx chan<- pt.PrimaryAnnounce,
	pausePrimary chan<- bool,
	primaryChangeChan chan<- pt.Node,
) {
	if !_initialized {
		panic("primary_reconfig package not initialized!")
	}
	// reconfig := ReconfigState{
	// 	mode: s_unknown,
	// 	// nodeUpdate initially empty
	// 	ownNode: initialNode,
	// 	// primary initially empty
	// }
	var role = r_unknown
	fmt.Println(role)
	var nodeUpdate pt.NodeUpdate
	primary := pt.EmptyNode()
	var primaryLastSeen time.Time

	// if you want to take primary reconfig a bit more easy
	// primaryAnnounceInterval := 40 * alive.Interval
	// primaryTimout := 40 * alive.Timeout
	primaryAnnounceInterval := alive.Interval
	primaryTimout := alive.Timeout

	unknownWaitingTime := time.Duration(1e09*rand.Float32()) + 2*time.Second // wait about 1s to 2s

	primAnnounceTimer := time.NewTimer(primaryAnnounceInterval)
	unknownTimer := time.NewTimer(unknownWaitingTime)

	for {
		switch role {
		case r_unknown:
			select {
			case nodeUpdate = <-nodeUpdateChanFromAlive:
				fmt.Println(nodeUpdate)
			case primaryAnnouncement := <-primaryAnnounceChanRx:
				primary = pt.Node(primaryAnnouncement)
			case <-unknownTimer.C:
				role, primary = determineRole(nodeUpdate, _node, primary)
				fmt.Println(role)
				// role = r_determineRole
			}
		// case r_determineRole:
		// 	fmt.Println(role)
		case r_setupPrimary:
			primaryChangeChan <- primary
			pausePrimary <- false
			nodeUpdateChanToPrim <- nodeUpdate
			primAnnounceTimer.Reset(primaryAnnounceInterval)
			role = r_runningPrimary
			fmt.Println(role)
		case r_runningPrimary:
			select {
			case <-primAnnounceTimer.C:
				primaryAnnounceChanTx <- pt.PrimaryAnnounce(_node)
				primAnnounceTimer.Reset(primaryAnnounceInterval)
			case primaryAnnoucement := <-primaryAnnounceChanRx:
				if pt.Node(primaryAnnoucement) == _node {
					break // ignore my own primary annoucements
				}
				fmt.Println("ROLE=primary: a (malicious/errornous) node anounces to be primary :/ - going to reconfiguration")
				pausePrimary <- true
				unknownTimer.Reset(unknownWaitingTime)
				primary = pt.EmptyNode()
				role = r_unknown
				fmt.Println(role)
			case nodeUpdate = <-nodeUpdateChanFromAlive:
				// if len(nodeUpdate.AliveNodes) == 0 { // meaning no backups available // TODO: check if neccessary
				// 	pausePrimary <- true
				// 	role = r_unknown
				// 	break
				// }
				fmt.Println(nodeUpdate)
				nodeUpdateChanToPrim <- nodeUpdate // in primary mode just forward nodeUpdate to primary
			}
		case r_setupBackup:
			primaryChangeChan <- primary
			pausePrimary <- true
			role = r_runningBackup
			fmt.Println(role)
		case r_runningBackup:
			select {
			case nodeUpdate = <-nodeUpdateChanFromAlive: // in backup mode just save nodeUpdate until required
				fmt.Println(nodeUpdate)
			case primAnnounce := <-primaryAnnounceChanRx:
				primaryLastSeen = time.Now()
				if pt.Node(primAnnounce) != primary {
					primary = pt.Node(primAnnounce)
					primaryChangeChan <- primary
				}
			case <-time.After(alive.Interval):
				if time.Since(primaryLastSeen) > primaryTimout {
					primary = pt.EmptyNode()
					unknownTimer.Reset(unknownWaitingTime)
					role = r_unknown
					fmt.Println(role)
				}
			}
		}
	}
}

func determineRole(
	nodeUpdate pt.NodeUpdate,
	ownNode pt.Node,
	primary pt.Node,
) (
	role Role,
	newPrimary pt.Node,
) {
	if iShallBePrimary(nodeUpdate, ownNode, primary) {
		return r_setupPrimary, ownNode
	}
	return r_setupBackup, primary
}

func iShallBePrimary(nodeUpdate pt.NodeUpdate, ownNode pt.Node, primary pt.Node) bool {
	if alone(nodeUpdate, ownNode) {
		return true
	}
	if primary != pt.EmptyNode() {
		return false
	}
	if ownNode == minIPAndPort(nodeUpdate.AliveNodes) {
		return true
	}
	return false
}

func alone(nodeUpdate pt.NodeUpdate, ownNode pt.Node) bool {
	if len(nodeUpdate.AliveNodes) == 1 && nodeUpdate.AliveNodes[0] == ownNode {
		return true
	}
	return false
}

// Returns the node with the lowest IP+Port.
func minIPAndPort(nodes []pt.Node) pt.Node {
	minNode := nodes[0] // should be the min already, but recompute, just to be sure :)
	for _, node := range nodes {
		if pt.LessThan(node, minNode) {
			minNode = node
		}
	}
	return minNode
}
