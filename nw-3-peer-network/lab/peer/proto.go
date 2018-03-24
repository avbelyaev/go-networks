package main

import "encoding/json"

const CMD_CONNECT = "connect"
const CMD_DISCONNECT = "disconnect"
const CMD_QUIT = "quit"
const CMD_MSG = "message"
const CMD_OK = "ok"


type Message struct {
	// Поле Command может принимать значения:
	// quit - прощание с сервером (после этого сервер рвёт соединение);
	// count - добавление точки к кривой;
	Command string `json:"command"`

	// Если Command == "count", то в поле Data лежит Circle
	// Если Command == "quit", то поле Data пустое
	Data *json.RawMessage `json:"data"`
}

