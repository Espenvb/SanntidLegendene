package project_types

import (
	"fmt"
	"strconv"
	"strings"
)

func EmptyNode() Node {
	return Node{}
}

// string representation of node struct nessary when using actually wanting to use
// `map[Node]Something`, because only string map keys are a JSON encodable type.
// function attached to struct Node; call with myNode.String()
func (node Node) String() string {
	return fmt.Sprintf("%d:%s:%d", node.GroupId, node.IP, node.Port)
}

func StringToNode(id string) Node {
	var groupId int
	var ip string
	var port int
	idParts := strings.Split(id, ":")
	// groupId = strconv.ParseInt(idParts[0], 10)
	groupId, _ = strconv.Atoi(idParts[0])
	// n, _ := fmt.Sscanf(idParts[0], "%d", groupId)
	// fmt.Println(groupId)
	// if n != 1 {
	// 	panic("IdToNode: id format could not be matched")
	// }
	ip = idParts[1]
	// n, _ = fmt.Sscanf(idParts[2], "%d", port)
	// if n != 1 {
	// 	panic("IdToNode: id format could not be matched")
	// }
	port, _ = strconv.Atoi(idParts[2])
	return Node{
		GroupId: groupId,
		IP:      ip,
		Port:    port,
	}
}

func EmptyHallRequests(m_numFloors int) [][2]bool {
	return make([][2]bool, m_numFloors)
}

func EmptyCabRequests(m_numFloors int) []bool {
	return make([]bool, m_numFloors)
}

func EmptySystemOrder(m_numFloors int) SystemOrder {
	return SystemOrder{
		HallRequests: EmptyHallRequests(m_numFloors),
		CabRequests:  make(map[string][]bool),
	}
}

func LessThan(left Node, right Node) bool {
	if left.IP == right.IP {
		return left.Port < right.Port
	}
	return left.IP < right.IP
}
