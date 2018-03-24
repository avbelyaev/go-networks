package main

import (
	"encoding/json"
	"github.com/satori/go.uuid"
)

const CMD_QUIT = "q"
const CMD_MSG = "m"
const CMD_OK = "ok"
const CMD_EMPTY = ""


type Message struct {
	// id сообщения. нужен чтобы пир мог сказать, знает он это сообщение или нет
	Id string `json:"id"`

	// автор сообщения
	Author string `json:"author"`

	// команда
	Command string `json:"command"`

	//полезная нагрузка сообщения (текст)
	Payload *json.RawMessage `json:"data"`
}


// compose message with UUID
func newMessage(command string, payload interface{}, author string) *Message {
	var raw json.RawMessage
	raw, _ = json.Marshal(payload)
	// additionally provide uuid
	// otherwise peer wont know if message that he has just received is his own (end of peer-ring)
	// or from another peer (he should display it and forward to next peer)
	return &Message{
		Author: author,
		Command: command,
		Payload: &raw,
		Id: uuid.Must(uuid.NewV4()).String(),
	}
}


// compose message and send it
func sendMessage(encoder *json.Encoder, command string, payload interface{}, author string) {
	var msg = newMessage(command, payload, author)
	resendMessage(encoder, msg)
}


// send message that is already composed
func resendMessage(encoder *json.Encoder, message *Message) {
	encoder.Encode(message)
}