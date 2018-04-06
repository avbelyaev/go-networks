package proto

import "encoding/json"

const CMD_COUNT = "c"
const CMD_QUIT = "q"
const CMD_OK = "ok"
const CMD_SUCCESS = "success"
const CMD_FAIL = "fail"

type Message struct {

	Command string `json:"command"`

	Data *json.RawMessage `json:"data"`
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