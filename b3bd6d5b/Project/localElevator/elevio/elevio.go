package elevio

import (
	"net"
	"sync"
	"time"
)

const _pollRate = 20 * time.Millisecond

var _initialized bool = false
var _numFloors int = 4
var _mtx sync.Mutex
var _conn net.Conn

const (
	N_FLOORS  = 4
	N_BUTTONS = 3
)

type MotorDirection int

const (
	Down MotorDirection = -1
	Stop MotorDirection = 0
	Up   MotorDirection = 1
)

type ButtonType int

const (
	HallUp   ButtonType = 0
	HallDown ButtonType = 1
	Cab      ButtonType = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

func Init(addr string, numFloors int) {
	if _initialized {
		return
	}
	_numFloors = numFloors
	_mtx = sync.Mutex{}

	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
}

func SetButtonLamp(floor int, btn ButtonType, onOff bool) {
	write([4]byte{2, byte(btn), byte(floor), toByte(onOff)})
}

func SetMotorDirection(dir MotorDirection) {
	write([4]byte{1, byte(dir), 0, 0})
}

func SetFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(onOff bool) {
	write([4]byte{4, toByte(onOff), 0, 0})
}

func SetStopLamp(onOff bool) {
	write([4]byte{5, toByte(onOff), 0, 0})
}

func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(_pollRate)
		floor := GetFloor()
		if floor != prev && floor != -1 {
			receiver <- floor
		}
		prev = floor
	}
}

func PollRequestButtons(receiver chan<- ButtonEvent) {
	prev := make([][3]bool, _numFloors)
	for {
		time.Sleep(_pollRate)
		for floor := 0; floor < _numFloors; floor++ {
			for btnType := ButtonType(0); btnType < 3; btnType++ {
				btn := getButton(btnType, floor)
				if btn != prev[floor][btnType] && btn {
					receiver <- ButtonEvent{floor, ButtonType(btnType)}
				}
				prev[floor][btnType] = btn
			}
		}
	}
}

func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		onOff := getStop()
		if onOff != prev {
			receiver <- onOff
		}
		prev = onOff
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		onOff := getObstruction()
		if onOff != prev {
			receiver <- onOff
		}
		prev = onOff
	}
}

func getButton(button ButtonType, floor int) bool {
	val := read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(val[1])
}

func GetFloor() int {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{7, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	if buf[1] != 0 {
		return int(buf[2])
	} else {
		return -1
	}
}

func getStop() bool {
	val := read([4]byte{8, 0, 0, 0})
	return toBool(val[1])
}

func getObstruction() bool {
	val := read([4]byte{9, 0, 0, 0})
	return toBool(val[1])
}

func read(in [4]byte) [4]byte {
	_mtx.Lock()
	defer _mtx.Unlock()
	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
	var out [4]byte
	_, err = _conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
	return out
}

func write(in [4]byte) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

func toByte(boolVal bool) byte {
	var byteVal byte = 0
	if boolVal {
		byteVal = 1
	}
	return byteVal
}

func toBool(byteVal byte) bool {
	var boolVal bool = false
	if byteVal != 0 {
		boolVal = true
	}
	return boolVal
}
