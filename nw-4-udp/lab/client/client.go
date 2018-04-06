package main

import (
	"flag"
	"github.com/mgutz/logxi/v1"
	"net"
	"os"
	"fmt"
	"encoding/json"
	"github.com/skorobogatov/input"
	"network-labs/nw-4-udp/lab/proto"
	"github.com/golang-collections/collections/stack"
	"time"
)

type Dialogue struct {

	// SetReadDeadline(retry...)
	retryTimeoutMillis int		// if there is no response within this time, message will be sent again
	sentMessageIds stack.Stack	// ordered list of ids of messages that were sent
}

func interact(conn *net.UDPConn) {
	defer conn.Close()

	var encoder, decoder = json.NewEncoder(conn), json.NewDecoder(conn)
	for {
		fmt.Printf("command: ")
		var command = input.Gets()

		switch command {
		case proto.CMD_QUIT:
			send_request(encoder, proto.CMD_QUIT, nil)
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

			send_request(encoder, proto.CMD_COUNT, circle)

		default:
			fmt.Printf("error: unknown command\n")
			continue
		}

		// Получение ответа.
		var rsp proto.Message
		if err := decoder.Decode(&rsp); err != nil {
			fmt.Printf("error: %v\n", err)
			break
		}

		switch rsp.Command {
		case proto.CMD_OK:
			fmt.Printf("ok\n")

		case proto.CMD_FAIL:
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

		case proto.CMD_SUCCESS:
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
			fmt.Printf("error: server reports unknown status %q\n", rsp.Command)
		}
	}
}


// send_request - вспомогательная функция для передачи запроса с указанной командой
// и данными. Данные могут быть пустыми (data == nil).
func send_request(encoder *json.Encoder, command string, data interface{}) {
	var raw json.RawMessage
	raw, _ = json.Marshal(data)
	encoder.Encode(&proto.Message{command, &raw})
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
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	interact(conn)
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