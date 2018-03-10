package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/skorobogatov/input"
	"net"
)

import (
	"network-labs/nw-1-protocol/lab/proto"
)

func interact(conn *net.TCPConn) {
	defer conn.Close()
	encoder, decoder := json.NewEncoder(conn), json.NewDecoder(conn)
	for {
		fmt.Printf("command = ")
		command := input.Gets()

		switch command {
		case proto.CMD_QUIT:
			send_request(encoder, proto.CMD_QUIT, nil)
			return

		case proto.CMD_COUNT:
			var circle proto.Circle
			fmt.Printf("Center X coordinate:")
			circle.Center.CoordX = input.Gets()
			fmt.Printf("Center X coordinate:")
			circle.Center.CoordY = input.Gets()
			fmt.Printf("Contour X coordinate:")
			circle.Contour.CoordX = input.Gets()
			fmt.Printf("Contour Y coordinate:")
			circle.Contour.CoordY = input.Gets()

			send_request(encoder, proto.CMD_COUNT, circle)

		default:
			fmt.Printf("error: unknown command\n")
			continue
		}

		// Получение ответа.
		var rsp proto.Response
		if err := decoder.Decode(&rsp); err != nil {
			fmt.Printf("error: %v\n", err)
			break
		}

		switch rsp.Status {
		case proto.RSP_STATUS_OK:
			fmt.Printf("ok\n")

		case proto.RSP_STATUS_FAIL:
			if rsp.Data == nil {
				fmt.Printf("error: data field is absent in response\n")

			} else {
				var errorMsg string
				if err := json.Unmarshal(*rsp.Data, &errorMsg); err != nil {
					fmt.Printf("error: malformed data field in response\n")

				} else {
					fmt.Printf("fail: %s\n", errorMsg)
				}
			}

		case proto.RSP_STATUS_SUCCESS:
			if rsp.Data == nil {
				fmt.Printf("error: data field is absent in response\n")

			} else {
				var circleSquare string
				if err := json.Unmarshal(*rsp.Data, &circleSquare); err != nil {
					fmt.Printf("error: malformed data field in response\n")

				} else {
					fmt.Printf("success: circle square is %s\n", circleSquare)
				}
			}

		default:
			fmt.Printf("error: server reports unknown status %q\n", rsp.Status)
		}
	}
}

// send_request - вспомогательная функция для передачи запроса с указанной командой
// и данными. Данные могут быть пустыми (data == nil).
func send_request(encoder *json.Encoder, command string, data interface{}) {
	var raw json.RawMessage
	raw, _ = json.Marshal(data)
	encoder.Encode(&proto.Request{command, &raw})
}

func main() {
	// Работа с командной строкой, в которой может указываться необязательный ключ -addr.
	var addrStr string
	flag.StringVar(&addrStr, "addr", "127.0.0.1:6002", "specify ip address and port")
	flag.Parse()

	// Разбор адреса, установка соединения с сервером и
	// запуск цикла взаимодействия с сервером.
	if addr, err := net.ResolveTCPAddr("tcp", addrStr); err != nil {
		fmt.Printf("error: %v\n", err)
	} else if conn, err := net.DialTCP("tcp", nil, addr); err != nil {
		fmt.Printf("error: %v\n", err)
	} else {
		interact(conn)
	}
}
