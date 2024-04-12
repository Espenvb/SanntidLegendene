// Providing routines for transmitting and receiving values from/into Go channels.
// Adapted bcast package, code from https://github.com/TTK4145/Network-go
package communication

import (
	"encoding/json"
	"fmt"
	"net"
	"project/connection"
	pt "project/project_types"
	"reflect"
)

// Wrapper for data sent/received over network, to be able to unpack every
// received network packet the same way in the first step (Unmarshal with
// interface typeTaggedJSON) and then do the Unmarshal with the found type.
// This is how every valid message can be sorted to the correct channel.
type typeTaggedJSON struct {
	TypeId string
	JSON   []byte
}

const bufSize = 1024

var _initialized = false
var _localPort int
var _broadcastPort int
var _connUnicastFoSend net.PacketConn

func Init(communicationPort int, broadcastPort int) {
	if _initialized {
		panic("trying to re-initialize communication package!")
	}
	_initialized = true
	_localPort = communicationPort
	_broadcastPort = broadcastPort
	_connUnicastFoSend = connection.DialUDP(_localPort + 2)
}

// Encodes value `payload` intp type-tagged JSON and sends it to the
// specified `node` via UDP Unicast.
// `payload` must be a JSON encodable type.
func Send(node pt.Node, payload interface{}) {
	if !_initialized {
		panic("communication not initialized!")
	}
	checkTypeRecursive(reflect.TypeOf(payload), []int{0})

	resolvedRemoteAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", node.IP, node.Port))
	if err != nil {
		panic(fmt.Sprintln("Error: ResolveUDPAddr:", err))
	}

	payloadJson, _ := json.Marshal(payload)
	ttjJson, _ := json.Marshal(typeTaggedJSON{
		TypeId: reflect.TypeOf(payload).String(),
		JSON:   payloadJson,
	})
	if len(ttjJson) > bufSize {
		panic(errorStringTooLargePayload(ttjJson, bufSize))
	}
	_connUnicastFoSend.WriteTo(ttjJson, resolvedRemoteAddr)
}

// waits initially wait for
// Encodes received values from `channels` into type-tagged JSON, sends it
// to the specified `addr` and `port` via UDP from the given `local_port`.
// `channels` must only contain channels of JSON encodable types
func TransmitterUnicast(
	primaryChangeChan <-chan pt.Node,
	channels ...interface{},
) {
	if !_initialized {
		panic("communication not initialized!")
	}
	checkArgs(channels...)

	var primary pt.Node
	var err error
	var resolvedRemoteAddr *net.UDPAddr                   // == nil
	primaChangeChanReflectHotFix := make(chan pt.Node, 1) // MUST be non-blocking to feed and immediatly after that listen on it

	typeNames := make([]string, len(channels))
	selectCases := make([]reflect.SelectCase, len(channels)+1) // +1 to be able to add primaryChangeChan
	for i, channel := range channels {
		typeNames[i] = reflect.TypeOf(channel).Elem().String()
		selectCases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(channel),
		}
	}

	// add selectCase waiting for primaryChangeChan
	indexPrimChangeChan := len(selectCases) - 1
	selectCases[indexPrimChangeChan] = reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(primaryChangeChan),
	}

	conn := connection.DialUDP(_localPort + 1)
	defer conn.Close()

	primary = <-primaryChangeChan
	resolvedRemoteAddr, err = net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", primary.IP, primary.Port))
	if err != nil {
		panic(fmt.Sprintln("Error: ResolveUDPAddr:", err))
	}
	for {
		indexChosenCase, value, _ := reflect.Select(selectCases)
		switch indexChosenCase {

		case indexPrimChangeChan: // send to new primary on next input to ...channels
			reflect.Select([]reflect.SelectCase{{
				Dir:  reflect.SelectSend,
				Chan: reflect.ValueOf(primaChangeChanReflectHotFix),
				Send: reflect.Indirect(value),
			}})
			primary = <-primaChangeChanReflectHotFix

			resolvedRemoteAddr, err = net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", primary.IP, primary.Port))
			if err != nil {
				panic(fmt.Sprintln("Error: ResolveUDPAddr:", err))
			}

		default:
			jsonBytes, _ := json.Marshal(value.Interface())
			typeTaggedJsonBytes, _ := json.Marshal(typeTaggedJSON{
				TypeId: typeNames[indexChosenCase],
				JSON:   jsonBytes,
			})
			if len(typeTaggedJsonBytes) > bufSize {
				panic(errorStringTooLargePayload(typeTaggedJsonBytes, bufSize))
			}
			conn.WriteTo(typeTaggedJsonBytes, resolvedRemoteAddr)
			// n, err := conn.WriteTo(typeTaggedJsonBytes, resolvedRemoteAddr)
			// fmt.Println("TxUnicast: wrote", n, "bytes of", value, "remoteAddr", *resolvedRemoteAddr, "with error", err)
		}
	}
}

// Encodes received values from `channel` into type-tagged JSON, and broadcasts
// it via UDP to `port`. The local port is also `port`.
// `channels` must only contain channels of JSON encodable types
func TransmitterBroadcast(channels ...interface{}) {
	if !_initialized {
		panic("communication not initialized!")
	}
	checkArgs(channels...)

	broadcastAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", "255.255.255.255", _broadcastPort))
	if err != nil {
		panic(fmt.Sprintln("Error: ResolveUDPAddr:", err))
	}

	typeNames := make([]string, len(channels))
	selectCases := make([]reflect.SelectCase, len(channels))
	for i, channel := range channels {
		typeNames[i] = reflect.TypeOf(channel).Elem().String()
		selectCases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(channel),
		}
	}

	conn := connection.DialBroadcastUDP(_broadcastPort)
	defer conn.Close()

	for {
		indexChosenCase, value, _ := reflect.Select(selectCases)
		jsonBytes, _ := json.Marshal(value.Interface())
		typeTaggedJsonBytes, _ := json.Marshal(typeTaggedJSON{
			TypeId: typeNames[indexChosenCase],
			JSON:   jsonBytes,
		})
		if len(typeTaggedJsonBytes) > bufSize {
			panic(errorStringTooLargePayload(typeTaggedJsonBytes, bufSize))
		}
		conn.WriteTo(typeTaggedJsonBytes, broadcastAddr)
	}
}

// Matches type-tagged JSON received on `port` to element types of `channels`,
// then sends the decoded value on the corresponding channel.
// `channels` must only contain channels of JSON encodable types
func Receiver(port int, channels ...interface{}) {
	if !_initialized {
		panic("communication not initialized!")
	}
	checkArgs(channels...)
	chansMap := make(map[string]interface{})
	for _, channel := range channels {
		chansMap[reflect.TypeOf(channel).Elem().String()] = channel
	}

	var buf [bufSize]byte

	// setting up broadcast socket not required for receiving broadcast messages
	//conn := connection.DialBroadcastUDP(port)
	conn := connection.DialUDP(port)
	defer conn.Close()

	for {
		n, _, err := conn.ReadFrom(buf[0:])
		if err != nil {
			panic(fmt.Sprintf("communication.Receiver(%d, ...):ReadFrom() failed: \"%+v\"\n", port, err))
		}

		var typeTaggedJSON typeTaggedJSON
		err = json.Unmarshal(buf[:n], &typeTaggedJSON)
		if err != nil {
			continue // bytes not matching expected typeTaggedJSON marshaled format
		}

		channel, ok := chansMap[typeTaggedJSON.TypeId]
		if !ok {
			continue // not listening on given type
		}

		value := reflect.New(reflect.TypeOf(channel).Elem())
		err = json.Unmarshal(typeTaggedJSON.JSON, value.Interface())
		if err != nil {
			continue // decoded value's type not matching channel value's type
		}

		reflect.Select([]reflect.SelectCase{{
			Dir:  reflect.SelectSend,
			Chan: reflect.ValueOf(channel),
			Send: reflect.Indirect(value),
		}})
	}
}

// Checks that args to Tx'er/Rx'er are valid:
//
//	All args must be channels
//	Element types of channels must be encodable with JSON
//	No element types are repeated
//
// Implementation note:
//   - Why there is no `isMarshalable()` function in encoding/json is a mystery,
//     so the tests on element type are hand-copied from `encoding/json/encode.go`
//
// Copied from Network-go bcast
func checkArgs(chans ...interface{}) {
	n := 0
	for range chans {
		n++
	}
	elemTypes := make([]reflect.Type, n)

	for i, ch := range chans {
		// Must be a channel
		if reflect.ValueOf(ch).Kind() != reflect.Chan {
			panic(fmt.Sprintf(
				"Argument must be a channel, got '%s' instead (arg# %d)",
				reflect.TypeOf(ch).String(), i+1))
		}

		elemType := reflect.TypeOf(ch).Elem()

		// Element type must not be repeated
		for j, e := range elemTypes {
			if e == elemType {
				panic(fmt.Sprintf(
					"All channels must have mutually different element types, arg# %d and arg# %d both have element type '%s'",
					j+1, i+1, e.String()))
			}
		}
		elemTypes[i] = elemType

		// Element type must be encodable with JSON
		checkTypeRecursive(elemType, []int{i + 1})

	}
}

// Panics if the type of `val` is recursive, otherwise does nothing.
// `offsets` is for improved error location output.
// Copied from Network-go bcast, they state to have taken it from `encoding/json/encode.go`
func checkTypeRecursive(val reflect.Type, offsets []int) {
	switch val.Kind() {
	case reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		panic(fmt.Sprintf(
			"Channel element type must be supported by JSON, got '%s' instead (nested arg# %v)",
			val.String(), offsets))
	case reflect.Map:
		if val.Key().Kind() != reflect.String {
			panic(fmt.Sprintf(
				"Channel element type must be supported by JSON, got '%s' instead (map keys must be 'string') (nested arg# %v)",
				val.String(), offsets))
		}
		checkTypeRecursive(val.Elem(), offsets)
	case reflect.Array, reflect.Ptr, reflect.Slice:
		checkTypeRecursive(val.Elem(), offsets)
	case reflect.Struct:
		for idx := 0; idx < val.NumField(); idx++ {
			checkTypeRecursive(val.Field(idx).Type, append(offsets, idx+1))
		}
	}
}

func errorStringTooLargePayload(typeTaggedJsonBytes []byte, bufSize int) string {
	return fmt.Sprintf(
		"Tried to send a message longer than the buffer size (length: %d, buffer size: %d)\n\t'%s'\n"+
			"Either send smaller packets, or go to network/bcast/bcast.go and increase the buffer size",
		len(typeTaggedJsonBytes), bufSize, string(typeTaggedJsonBytes))
}
