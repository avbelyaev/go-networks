package main

import (
	"encoding/json"
	"github.com/satori/go.uuid"
)

const CMD_CONNECT = "c"
const CMD_DISCONNECT = "d"
const CMD_QUIT = "q"
const CMD_MSG = "m"
const CMD_OK = "ok"


type Message struct {
	// Поле Command может принимать значения:
	// quit - прощание с сервером (после этого сервер рвёт соединение);
	// count - добавление точки к кривой;
	Command string `json:"command"`

	// Если Command == "count", то в поле Data лежит Circle
	// Если Command == "quit", то поле Data пустое
	Payload *json.RawMessage `json:"data"`

	// uuid сообщения
	Id string `json:"id"`
}


// compose message with UUID
func newMessage(command string, payload interface{}) *Message {
	var raw json.RawMessage
	raw, _ = json.Marshal(payload)
	// additionally provide uuid
	// otherwise peer wont know if message that he has just received is his own (end of peer-ring)
	// or from another peer (he should display it and forward to next peer)
	return &Message{
		Command: command,
		Payload: &raw,
		Id: uuid.Must(uuid.NewV4()).String(),
	}
}


// compose message and send it
func sendMessage(encoder *json.Encoder, command string, payload interface{}) {
	var msg = newMessage(command, payload)
	resendMessage(encoder, msg)
}


// send message that is already composed
func resendMessage(encoder *json.Encoder, message *Message) {
	encoder.Encode(message)
}