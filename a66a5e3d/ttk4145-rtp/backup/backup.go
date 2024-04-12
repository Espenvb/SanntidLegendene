package backup

import (
	pt "project/project_types"
	"time"
)

var _initialized = false
var _node pt.Node
var _m_numFloors int

func Init(node pt.Node, numFloors int) {
	if _initialized {
		panic("trying to re-initialize backup package!")
	}
	_initialized = true
	_node = node
	_m_numFloors = numFloors
}

// `systemOrderRecoveryChan` MUST be a non-blocking channel in order for
// the primary to always get the latest update
func BackupService(
	systemOrderChan <-chan pt.SystemOrder,
	ackChan chan<- pt.SystemOrderAck,
	systemOrderRecoveryChan chan<- pt.SystemOrder,
) {
	if !_initialized {
		panic("trying to start un-initialized backup package!")
	}
	systemOrderBackup := pt.EmptySystemOrder(_m_numFloors)
	for {
		select {
		case systemOrderBackup = <-systemOrderChan:
			ackChan <- pt.SystemOrderAck{
				SystemOrder: systemOrderBackup,
				Node:        _node,
			}
		case <-time.After(time.Second): // continously sent; primary will pick up, when un-paused by reconfig
			systemOrderRecoveryChan <- systemOrderBackup
		}
	}
}
