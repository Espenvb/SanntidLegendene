package distribution

import (
	"fmt"
	"project/communication"
	pt "project/project_types"
	"reflect"
	"time"
)

var _retryTimeout = 10 * 100 * time.Millisecond
var _timeout = 10 * 700 * time.Millisecond

// `systemOrderChan` and `aliveNodes` shall be used toghether to start the sending process
// if not successful, do nothing in primary service, in order to fulfil the button light contract
func BackupDistribution(
	systemOrderChan <-chan pt.SystemOrder, // TODO: either combine this to new type
	aliveNodesChan <-chan pt.NodeUpdate, //         or untangle mega primary to OrderService and Request/ButtonService
	ackChan <-chan pt.SystemOrderAck,
	successful chan<- bool,
) {
	// bool value does not matter, but delete not possible for array
	waitingForAck := map[pt.Node]bool{}
	systemOrder := pt.SystemOrder{}
	retryTimer := time.NewTimer(_retryTimeout)
	timeoutTimer := time.NewTimer(_timeout)
	for {
		switch {
		case len(waitingForAck) == 0:
			// fmt.Println("BackupDistribution: len(waitingForAck) ==", len(waitingForAck))
			select {
			case systemOrder = <-systemOrderChan:
				// wait for both channels before start sending
				fmt.Println("BackupDistr: received systemOrder", systemOrder)
				nodes := (<-aliveNodesChan).AliveNodes
				fmt.Println("BackupDistr: received alive nodes", nodes)

				waitingForAck = make(map[pt.Node]bool, len(nodes))
				for _, node := range nodes {
					communication.Send(node, systemOrder)
					waitingForAck[node] = true
				}
				if !retryTimer.Stop() {
					<-retryTimer.C
				}
				retryTimer.Reset(_retryTimeout)
				if !timeoutTimer.Stop() {
					<-timeoutTimer.C
				}
				timeoutTimer.Reset(_timeout)
			case <-ackChan:
			}
		case len(waitingForAck) != 0:
			// fmt.Println("BackupDistribution: len(waitingForAck) ==", len(waitingForAck))
			select {
			case ack := <-ackChan:
				fmt.Println("BackupDistr: ack =", ack)
				if reflect.DeepEqual(ack.SystemOrder, systemOrder) {
					delete(waitingForAck, ack.Node)
				}
				if len(waitingForAck) == 0 {
					successful <- true
				}
			case <-retryTimer.C:
				for node, waiting := range waitingForAck {
					if waiting {
						communication.Send(node, systemOrder)
					}
				}
				retryTimer.Reset(_retryTimeout)
			case <-timeoutTimer.C:
				fmt.Println("BackupDistribution not successful")
				waitingForAck = map[pt.Node]bool{}
				systemOrder = pt.SystemOrder{} // not necessary to reset as always
				successful <- false
			case <-systemOrderChan:
			case <-aliveNodesChan:
			}
		}
	}
}

// here no alive list input nessesary because implicitly given by the map of`orderChan`
// if not successful, reassign orders in order to fullfil button light contract
func OrderDistribution(
	ordersChan <-chan map[pt.Node]pt.Order,
	ackChan <-chan pt.OrderAck,
	successful chan<- bool,
) {
	// bool value does not matter, but delete not possible for array
	waitingForAck := map[pt.Node]bool{}
	orders := map[pt.Node]pt.Order{}
	retryTimer := time.NewTimer(_retryTimeout)
	timeoutTimer := time.NewTimer(_timeout)
	for {
		switch {
		case len(waitingForAck) == 0:
			select {
			case orders = <-ordersChan:
				if len(orders) == 0 { // TODO: should actually not be neccesarry
					successful <- true
					break
				}
				waitingForAck = make(map[pt.Node]bool, len(orders))
				for node, order := range orders {
					communication.Send(node, order)
					waitingForAck[node] = true // TODO: redesign when to sent/get acknowledgments
				}
				if !retryTimer.Stop() {
					<-retryTimer.C
				}
				retryTimer = time.NewTimer(_retryTimeout)
				if !timeoutTimer.Stop() {
					<-timeoutTimer.C
				}
				timeoutTimer = time.NewTimer(_timeout)
			case <-ackChan:
			}

		case len(waitingForAck) != 0:
			select {
			case ack := <-ackChan:
				if reflect.DeepEqual(ack.Order, orders[ack.Node]) {
					delete(waitingForAck, ack.Node)
				}
				if len(waitingForAck) == 0 {
					successful <- true
				}
			case <-retryTimer.C:
				for node := range waitingForAck {
					communication.Send(node, orders[node])
				}
				retryTimer.Reset(_retryTimeout)
			case <-timeoutTimer.C:
				waitingForAck = map[pt.Node]bool{}
				orders = map[pt.Node]pt.Order{}
				successful <- false
			}
		}
	}
}

// `systemOrderChan` and `aliveNodes` shall be used toghether to start the sending process
// waiting for successful is only nessesarry when switching off lights after floor service
// in order to fulfill the button-light-contract, in such case, if not successful then use
// of `panic()` in primary service is recommended.
func LightDistribution(
	systemOrderChan <-chan pt.SystemOrder,
	aliveNodesChan <-chan pt.NodeUpdate,
	ackChan <-chan pt.LightOrderAck,
	successful chan<- bool,
) {
	// bool value does not matter, but delete not possible for array
	waitingForAck := map[pt.Node]bool{}
	lightOrders := map[pt.Node]pt.LightOrder{}
	retryTimer := time.NewTimer(_retryTimeout)
	timeoutTimer := time.NewTimer(_timeout)
	for {
		switch {
		case len(waitingForAck) == 0:
			select {
			case systemOrder := <-systemOrderChan:
				// wait for both channels before start sending
				nodes := (<-aliveNodesChan).AliveNodes

				waitingForAck = make(map[pt.Node]bool, len(nodes))
				lightOrders = make(map[pt.Node]pt.LightOrder, len(nodes))
				// startSendingTime := time.Now()
				for _, node := range nodes {
					lightOrders[node] = pt.LightOrder{
						HallRequests: systemOrder.HallRequests,
						CabRequests:  systemOrder.CabRequests[node.String()],
					} // TODO make sure only valid pt.EmptyCabRequests(_m_numFloors) are used if key not ok

					communication.Send(node, lightOrders[node])
					waitingForAck[node] = true
				}
				// fmt.Println("LightDistribtution: sending took:", time.Since(startSendingTime).Seconds())
				if !retryTimer.Stop() {
					<-retryTimer.C
				}
				retryTimer.Reset(_retryTimeout)
				if !timeoutTimer.Stop() {
					<-timeoutTimer.C
				}
				timeoutTimer.Reset(_timeout)
			case ack := <-ackChan:
				fmt.Println("LightDistr throghing away ack", ack)
			}
		case len(waitingForAck) != 0:
			select {
			case ack := <-ackChan:
				fmt.Println("LightDistr: received ack", ack)
				if reflect.DeepEqual(ack.LightOrder, lightOrders[ack.Node]) {
					delete(waitingForAck, ack.Node)
				}
				if len(waitingForAck) == 0 {
					successful <- true
				}
			case <-retryTimer.C:
				fmt.Println("LightDistr: retryTimer timed out")
				// startSendingTime := time.Now()
				for node, waiting := range waitingForAck {
					if waiting {
						fmt.Println("retry send for", node, "lightOrder", lightOrders[node])
						communication.Send(node, lightOrders[node])
					}
				}
				// fmt.Println("LightDistribtution: sending took:", time.Since(startSendingTime).Seconds())
				retryTimer.Reset(_retryTimeout)
			case <-timeoutTimer.C:
				fmt.Println("LightDistr: timedout")
				waitingForAck = map[pt.Node]bool{}
				lightOrders = map[pt.Node]pt.LightOrder{}
				successful <- false
			}
		}
	}
}
