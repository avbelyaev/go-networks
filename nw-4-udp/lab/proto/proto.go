package proto

import (
	"encoding/json"
	"github.com/satori/go.uuid"
)

const CMD_COUNT = "c"
const CMD_QUIT = "q"
const CMD_OK = "ok"
const CMD_SUCCESS = "success"
const CMD_FAIL = "fail"
const CMD_UNKNOWN = "unknown"
const CMD_ACK = "a"

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