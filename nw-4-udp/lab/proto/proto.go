package proto

import (
	"encoding/json"
	"github.com/satori/go.uuid"
	"net"
	"fmt"
	"strconv"
	"time"
	"github.com/mgutz/logxi/v1"
	"os"
)

const CMD_COUNT = "c"
const CMD_QUIT = "q"
const CMD_OK = "ok"
const CMD_SUCCESS = "success"
const CMD_FAIL = "fail"
const CMD_UNKNOWN = "unknown"
const CMD_ACK = "ack"

type Message struct {

	Id string `json:"id"`

	Command string `json:"command"`

	Payload *json.RawMessage `json:"data"`
}


//TODO to export function from this package into another package
//function names should start with uppercase


func NewMessage(command string, payload interface{}) *Message {
	var raw json.RawMessage
	raw, _ = json.Marshal(payload)

	return &Message{
		Command: 	command,
		Payload: 	&raw,
		Id: 		uuid.Must(uuid.NewV4()).String(),
	}
}

type Point struct {
	// коорд Х, float
	CoordX string `json:"x,Number"`

	// коорд Y, float
	CoordY string `json:"y,Number"`
}

type Circle struct {
	// центр окружности (Point)
	Center Point `json:"center"`

	// точка на окржуности (Point)
	Contour Point `json:"contour"`
}

func NewCircle(centerX string, centerY string, contourX string, contourY string) *Circle {
	return &Circle{
		Center: Point{
			CoordX: centerX,
			CoordY: centerY,
		},
		Contour: Point{
			CoordX: contourX,
			CoordY: contourY,
		},
	}
}


func WriteReliably(conn *net.UDPConn,
	message *Message,
//maxRetries int,
	writeFunc func(data []byte),
	readFunc func(buffer []byte) (int, error)) bool {

	var serialized, err = json.Marshal(message)
	HandleErr(err)

	var retry = 20
	// keep sending message while we can
	for retry > 0 {
		log.Debug(fmt.Sprintf("Sending request. Retries left: %s", strconv.Itoa(retry)))
		//client.conn.Write(serialized)
		println("write ", string(serialized))
		writeFunc(serialized)

		// set timeout for nearest Read() call
		conn.SetReadDeadline(time.Now().Add(time.Duration(100) * time.Millisecond))

		// read into buffer and check if there was timeout error while reading
		var buf = make([]byte, 1000)
		//var bytesRead, readErr = client.conn.Read(buf)
		var bytesRead, readErr = readFunc(buf)

		// deserialize into regular message and check if message == ACK
		var ackBytes = buf[:bytesRead]
		println("got: ", string(ackBytes))
		var ack Message
		var unmarshalErr = json.Unmarshal(ackBytes, &ack)

		if nil == readErr && nil == unmarshalErr && CMD_ACK == ack.Command {
			// there is an ACK for the message => it was delivered for sure
			log.Debug("Ack for message has been acquired")
			//TODO if message that was received here is an actual response (lastId == id)
			// then quit this function

			return true
		}
		retry--
	}

	// message was not delivered :(
	return false
}


func HandleErr(e error) {
	if nil != e {
		log.Error("Exiting with error ", e.Error())
		println(e)
		os.Exit(1)
	}
}