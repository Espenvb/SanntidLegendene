package project_types

import (
	"project/elevio"
)

// type NodeType string

// const (
// 	NT_Primary   NodeType = "primary"
// 	NT_Backup    NodeType = "backup"
// 	NT_Undefined NodeType = "undefined"
// )

type Node struct {
	GroupId int
	IP      string
	Port    int
	// Type    NodeType
}

// Alias required to differentiate between an alive message and a primary announcement on receiver side.
type PrimaryAnnounce Node

// Alias for `node.Node` to be used to send over network.
type AliveMessage Node

type NodeUpdate struct {
	AliveNodes []Node
	New        Node
	Lost       []Node
}

type ElevatorBehaviour string

const (
	EB_Idle     ElevatorBehaviour = "idle"
	EB_Moving   ElevatorBehaviour = "moving"
	EB_DoorOpen ElevatorBehaviour = "doorOpen"
)

// TODO implement converter from elevio.MotorDirection to string

// Order for a single elevator. Which floor/cab requests to service.
type Order struct {
	HallRequests [][2]bool //hallRequests[f][dir], dir==0==elevio.BT_HallUp -> up; dir==1==elevio.BT_HallDown -> down
	CabRequests  []bool
}

type OrderAck struct {
	Order Order // the order to acknowledge
	Node  Node  // itself
}

// Defines which lights shall be ON/OFF. Renaming neccecarry as the receiver must know
type LightOrder Order

type LightOrderAck struct {
	LightOrder LightOrder
	Node       Node
}

type Elevator struct {
	Floor    int                   // NonNegativeInteger
	Dir      elevio.MotorDirection // up==1, down==-1, stop==0
	Behavior ElevatorBehaviour     // idle | moving | doorOpen
	Order    Order                 // only to be used by execution
	Node     Node
}

type SystemOrder struct {
	HallRequests [][2]bool
	CabRequests  map[string][]bool
}

type SystemOrderAck struct {
	SystemOrder SystemOrder
	Node        Node
}

type ButtonPress struct {
	Button elevio.ButtonEvent
	Node   Node
}

type FloorServiced struct {
	Floor elevio.ButtonEvent
	Node  Node
}
