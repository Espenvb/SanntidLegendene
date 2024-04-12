// Provides functions to observe alive messages of nodes.
package alive

import (
	pt "project/project_types"
	"sort"
	"time"
)

// const Interval = 500 * time.Millisecond // slow for debug
const Interval = 4 * 15 * time.Millisecond // WONTFIX: consider merging primary-reconfig and alive package

// const Timeout = 3000 * time.Millisecond // slow for debug
const Timeout = 4 * 500 * time.Millisecond

var _initialized = false
var _node pt.Node

func Init(node pt.Node) {
	if _initialized {
		panic("trying to re-initialize alive package!")
	}
	_initialized = true
	_node = node
}

func Transmitter(
	obstructedChan <-chan bool,
	// initialAliveMessage pt.AliveMessage,
	// nodeTypeUpdate <-chan pt.NodeType, // TODO:obstructedChan
	aliveSendChan chan<- pt.AliveMessage,
) {
	if !_initialized {
		panic("trying to start un-initialized alive package!")
	}
	aliveMessage := pt.AliveMessage(_node)
	obstructed := false
	for {
		select {
		case obstructed = <-obstructedChan: // TODO check if time.After is still triggered
		case <-time.After(Interval):
			switch {
			case !obstructed:
				aliveSendChan <- aliveMessage
			}
		}
	}
}

// Waits new alive messages on `aliveReceiveChan` and, if reqired,
// sends node updates (aka peer updates) and `nodeUpdateChan`.
// If the type of a node (`node.Type`) changes an update is triggered
// with `.New` is set to the node with new type and an other update,
// where the node with the old type is put into `.Lost` list.
func Receiver(
	aliveReceiveChan <-chan pt.AliveMessage,
	nodeUpdateChan chan<- pt.NodeUpdate, // to reconfig
) {
	if !_initialized {
		panic("trying to start un-initialized alive package!")
	}
	var receivedAliveMessage pt.AliveMessage
	var nodeUpdate pt.NodeUpdate
	lastSeen := make(map[pt.Node]time.Time)
	updated := false
	timer := time.NewTimer(Interval)

	for {
		switch {
		case !updated:
			select {
			case receivedAliveMessage = <-aliveReceiveChan:
				nodeUpdate, lastSeen, updated = addNode(receivedAliveMessage, nodeUpdate, lastSeen)
			case <-timer.C:
				nodeUpdate, lastSeen, updated = removeDead(nodeUpdate, lastSeen)
				timer.Reset(Interval)
			}
		case updated:
			nodeUpdate, lastSeen = normalize(nodeUpdate, lastSeen)
			nodeUpdateChan <- nodeUpdate

			// reset states
			nodeUpdate.New = pt.EmptyNode()
			updated = false
		}
	}
}

// disclaimer: has side-effects on parameter oldNodeUpdate and oldLastSeen as beeing of pointer type
func addNode(
	oldAliveMessage pt.AliveMessage,
	oldNodeUpdate pt.NodeUpdate,
	oldLastSeen map[pt.Node]time.Time,
) (
	nodeUpdate pt.NodeUpdate,
	lastSeen map[pt.Node]time.Time,
	updated bool,
) {
	nodeUpdate.AliveNodes = oldNodeUpdate.AliveNodes
	lastSeen = oldLastSeen
	updated = false

	if !isValidAliveMessage(oldAliveMessage) {
		return nodeUpdate, lastSeen, false
	}

	// Adding new connection
	if _, nodeExists := oldLastSeen[pt.Node(oldAliveMessage)]; !nodeExists {
		nodeUpdate.New = pt.Node(oldAliveMessage)
		updated = true
	}

	// add/touch node in lastSeen
	lastSeen[pt.Node(oldAliveMessage)] = time.Now()

	return nodeUpdate, lastSeen, updated
}

func isValidAliveMessage(aliveMessage pt.AliveMessage) bool {
	if aliveMessage.GroupId != _node.GroupId {
		return false
	}
	// t := aliveMessage.Type
	// return (t == pt.NT_Primary || t == pt.NT_Backup || t == pt.NT_Undefined)
	return true
}

func removeDead(
	oldNodeUpdate pt.NodeUpdate,
	oldLastSeen map[pt.Node]time.Time,
) (
	nodeUpdate pt.NodeUpdate,
	lastSeen map[pt.Node]time.Time,
	updated bool,
) {
	// TODO: use copy instead
	nodeUpdate = oldNodeUpdate
	lastSeen = oldLastSeen
	updated = false

	for k, v := range lastSeen {
		if time.Since(v) > Timeout {
			nodeUpdate.Lost = append(nodeUpdate.Lost, k)
			delete(lastSeen, k)
			updated = true
		}
	}

	return nodeUpdate, lastSeen, updated
}

// disclaimer: has side-effects on parameter oldNodeUpdate and oldLastSeen as beeing of pointer type
func normalize(
	oldNodeUpdate pt.NodeUpdate,
	oldLastSeen map[pt.Node]time.Time,
) (
	nodeUpdate pt.NodeUpdate,
	lastSeen map[pt.Node]time.Time,
) {
	nodeUpdate = oldNodeUpdate
	lastSeen = oldLastSeen

	nodeUpdate.AliveNodes = make([]pt.Node, 0, len(lastSeen))
	//for k, _ := range lastSeen {
	for k := range lastSeen {
		nodeUpdate.AliveNodes = append(nodeUpdate.AliveNodes, k)
	}

	sort.Slice(nodeUpdate.AliveNodes, func(i, j int) bool {
		return pt.LessThan(nodeUpdate.AliveNodes[i], nodeUpdate.AliveNodes[j])
	})
	sort.Slice(nodeUpdate.Lost, func(i, j int) bool {
		return pt.LessThan(nodeUpdate.Lost[i], nodeUpdate.Lost[j])
	})

	return nodeUpdate, lastSeen
}
