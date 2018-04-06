package main

import (
	"flag"
	"github.com/mgutz/logxi/v1"
	"net"
	"os"
	"fmt"
	"encoding/json"
	"github.com/skorobogatov/input"
	//"network-labs/nw-4-udp/lab/proto"
	"github.com/golang-collections/collections/stack"
	//"time"
	"network-labs/nw-4-udp/lab/proto"
)

type DurableClient struct {
	conn				*net.UDPConn
	retryTimeoutMillis 	int         // if there is no response within this time, message will be sent again
	messagesStack      	stack.Stack // ordered list of ids of messages that were sent
}

func NewDurableClient(conn *net.UDPConn, retryTimeoutMillis int) *DurableClient {
	return &DurableClient{
		conn: 				conn,
		retryTimeoutMillis: retryTimeoutMillis,
		messagesStack:      stack.Stack{},
	}
}

func (client *DurableClient) interact() {
	defer client.conn.Close()

	//var encoder, decoder = json.NewEncoder(client.conn), json.NewDecoder(client.conn)
	//var decoder = json.NewDecoder(client.conn)
	for {
		fmt.Printf("command: ")
		var command = input.Gets()
		var message *proto.Message

		switch command {
		case proto.CMD_QUIT:
			//send_request(encoder, proto.CMD_QUIT, nil)
			message = proto.NewMessage(proto.CMD_QUIT, nil)
			return

		case proto.CMD_COUNT:
			var circle proto.Circle
			fmt.Printf("Center X coordinate:")
			circle.Center.CoordX = input.Gets()
			fmt.Printf("Center X coordinate:")
			circle.Center.CoordY = "1"//input.Gets()
			fmt.Printf("Contour X coordinate:")
			circle.Contour.CoordX = "2"//input.Gets()
			fmt.Printf("Contour Y coordinate:")
			circle.Contour.CoordY = "2"//input.Gets()

			message = proto.NewMessage(proto.CMD_COUNT, circle)
			//send_request(encoder, proto.CMD_COUNT, circle)

		default:
			fmt.Printf("error: unknown command\n")
			continue
		}

		// handle outcoming message
		send_request(json.NewEncoder(client.conn), message)

		// Получение ответа.
		println("handling response")

		var buf = make([]byte, 1000)
		var bytesRead, _, readErr = client.conn.ReadFromUDP(buf)
		handleErr(readErr)

		var rspBytes = buf[:bytesRead]
		println("response: ", string(rspBytes))

		var rsp proto.Message
		var unmarshalErr = json.Unmarshal(rspBytes, &rsp)
		handleErr(unmarshalErr)
		//var rsp proto.Message
		//if err := decoder.Decode(s); err != nil {
		//	fmt.Printf("error: %v\n", err)
		//	break
		//}

		println("siwtching")
		switch rsp.Command {
		case proto.CMD_OK:
			fmt.Printf("ok\n")

		case proto.CMD_FAIL:
			if rsp.Payload == nil {
				fmt.Printf("error: data field is absent in response\n")

			} else {
				var errorMsg string
				if err := json.Unmarshal(*rsp.Payload, &errorMsg); err != nil {
					fmt.Printf("error: malformed data field in response\n")

				} else {
					fmt.Printf("fail: %s\n", errorMsg)
				}
			}

		case proto.CMD_SUCCESS:
			if rsp.Payload == nil {
				fmt.Printf("error: data field is absent in response\n")

			} else {
				var circleSquare string
				if err := json.Unmarshal(*rsp.Payload, &circleSquare); err != nil {
					fmt.Printf("error: malformed data field in response\n")

				} else {
					fmt.Printf("success: circle square is %s\n", circleSquare)
				}
			}

		default:
			fmt.Printf("error: server reports unknown status %q\n", rsp.Command)
		}
	}
}


// send_request - вспомогательная функция для передачи запроса с указанной командой
// и данными. Данные могут быть пустыми (data == nil).
func send_request(encoder *json.Encoder, message *proto.Message) {
	//var raw json.RawMessage
	//raw, _ = json.Marshal(data)
	encoder.Encode(message)
}


func acquireAnswerWithinTimeout() bool {
	return true
}


func main() {
	var serverAddrStr string
	flag.StringVar(&serverAddrStr, "server", "127.0.0.1:6000", "set server IP address and port")

	var serverAddr, addrErr = net.ResolveUDPAddr("udp", serverAddrStr)
	handleErr(addrErr)

	var conn, err = net.DialUDP("udp", nil, serverAddr)
	handleErr(err)

	//conn.SetReadDeadline(1 * time.Second.Seconds())
	var client = NewDurableClient(conn, 5000)
	client.interact()
	//conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	//interact(conn)
	//{


		//buf := make([]byte, 32)
		//for i := uint(0); i < n; i++ {
		//	x := rand.Intn(1000)
		//	if _, err := conn.Write([]byte(strconv.Itoa(x))); err != nil {
		//		log.Error("sending request to server", "error", err, "x", x)
		//
		//	} else if bytesRead, err := conn.Read(buf); err != nil {
		//		log.Error("receiving answer from server", "error", err)
		//
		//	} else {
		//		yStr := string(buf[:bytesRead])
		//		if y, err := strconv.Atoi(yStr); err != nil {
		//			log.Error("cannot parse answer", "answer", yStr, "error", err)
		//
		//		} else {
		//			log.Info("successful interaction with server", "x", x, "y", y)
		//		}
		//	}
		//}
	//}
}


func handleErr(e error) {
	if nil != e {
		log.Error("Exiting with error ", e.Error())
		println(e)
		os.Exit(1)
	}
}