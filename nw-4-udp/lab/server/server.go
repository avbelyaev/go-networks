package main

import (
	"flag"
	"fmt"
	"github.com/mgutz/logxi/v1"
	"net"
	"os"
	//"strconv"
	//"network-labs/nw-4-udp/lab/proto/neww"
	//"network-labs/nw-4-udp/lab/proto"
	"encoding/json"
	"strconv"
	"math"
	"network-labs/nw-4-udp/lab/proto"
)

/*
порядок доставки сообщений
retry если ответ не пришел за отведенное время
SYN ACK зашить в rq/rsp
 */
//

type StatelessClient struct {
	logger log.Logger    // Объект для печати логов
	conn   *net.UDPConn  // Объект TCP-соединения
	enc    *json.Encoder // Объект для кодирования и отправки сообщений
}

// NewClient - конструктор клиента, принимает в качестве параметра
// объект TCP-соединения.
func NewClient(conn *net.UDPConn) *StatelessClient {
	return &StatelessClient{
		conn:   conn,
		enc:    json.NewEncoder(conn),
	}
}

// handle - метод, в котором реализован цикл взаимодействия с клиентом.
// Подразумевается, что метод handle будет вызаваться в отдельной go-программе.
func (client *StatelessClient) handle() {
	defer client.conn.Close()
	decoder := json.NewDecoder(client.conn)
	for {
		var req proto.Message
		if err := decoder.Decode(&req); err != nil {
			client.logger.Error("cannot decode message")
			break
		} else {
			client.logger.Info("received command", "command", req.Command)
			if client.handleRequest(&req) {
				client.logger.Info("shutting down connection")
				break
			}
		}
	}
}

func countSquareDifference(coord1 string, coord2 string) (float64, error) {
	circleCenterX, err := strconv.ParseFloat(string(coord1), 64)
	circleContourX, err := strconv.ParseFloat(string(coord2), 64)
	if nil != err {
		return 0, err

	} else {
		delta := circleCenterX - circleContourX
		return delta * delta, nil
	}
}

func countCircleSquare(circle proto.Circle) (float64, error) {
	deltaXSquared, err := countSquareDifference(circle.Center.CoordX, circle.Contour.CoordX)
	if nil != err {
		return 0, err
	}

	deltaYSquared, err := countSquareDifference(circle.Center.CoordY, circle.Contour.CoordY)
	if nil != err {
		return 0, err
	}

	radius := math.Sqrt(deltaXSquared + deltaYSquared)
	return math.Pi * radius * radius, nil
}

// handleRequest - метод обработки запроса от клиента. Он возвращает true,
// если клиент передал команду "quit" и хочет завершить общение.
func (client *StatelessClient) handleRequest(req *proto.Message) bool {
	switch req.Command {
	case proto.CMD_QUIT:
		client.respond(proto.CMD_OK, nil)
		return true

	case proto.CMD_COUNT:
		errorMsg := ""
		if req.Payload == nil {
			errorMsg = "data field is absent"

		} else {
			var circle proto.Circle
			if err := json.Unmarshal(*req.Payload, &circle); err != nil {
				errorMsg = "malformed data field"

			} else {
				if circleSquare, err := countCircleSquare(circle); nil != err {
					errorMsg = "could not count circle square, invalid circle data provided"

				} else {
					circleSquareAsString := strconv.FormatFloat(circleSquare, 'f', 6, 64)
					client.logger.Info("square of circle has been counted", "value", circleSquareAsString)

					client.respond(proto.CMD_SUCCESS, circleSquareAsString)
				}
			}
		}
		if errorMsg == "" {
			client.respond(proto.CMD_OK, nil)

		} else {
			client.logger.Error("count failed", "reason", errorMsg)
			client.respond(proto.CMD_FAIL, errorMsg)
		}

	default:
		client.logger.Error("unknown command")
		client.respond("failed", "unknown command")
	}
	return false
}

// respond - вспомогательный метод для передачи ответа с указанным статусом
// и данными. Данные могут быть пустыми (data == nil).
func (client *StatelessClient) respond(status string, data interface{}) {
	client.enc.Encode(proto.NewMessage(status, data))
}

func main() {
	var (
		serverAddrStr string
		helpFlag      bool
	)
	flag.StringVar(&serverAddrStr, "addr", "127.0.0.1:6000", "set server IP address and port")
	flag.BoolVar(&helpFlag, "help", false, "print options list")

	if flag.Parse(); helpFlag {
		fmt.Fprint(os.Stderr, "server [options]\n\nAvailable options:\n")
		flag.PrintDefaults()
	} else if serverAddr, err := net.ResolveUDPAddr("udp", serverAddrStr); err != nil {
		log.Error("resolving server address", "error", err)
	} else if conn, err := net.ListenUDP("udp", serverAddr); err != nil {
		log.Error("creating listening connection", "error", err)
	} else {
		// Цикл приёма входящих соединений.
		//for {
		//	//go NewClient(conn).handle()
		//
		//}
		log.Info("server listens incoming messages from clients",
			"addr", serverAddr.String())
		buf := make([]byte, 1000)
		for {
			if bytesRead, addr, err := conn.ReadFromUDP(buf); err != nil {
				log.Error("receiving message from client", "error", err)
			} else {
				log.Info("Reading", "read bytes", bytesRead, "from", addr)

				s := string(buf[:bytesRead])
				log.Info("Got", "msg", s)

				var msg = proto.NewMessage(proto.CMD_OK, nil)
				var msgAsBytes, err = json.Marshal(msg)
				handleErr(err)

				println("responding with: ", string(msgAsBytes))

				conn.WriteToUDP(msgAsBytes, addr)
				//var enc = json.NewEncoder(conn)
				//enc.Encode(proto.NewMessage(proto.CMD_OK, nil))
				//else if _, err = conn.WriteToUDP([]byte(strconv.Itoa(x*2)), addr); err != nil {
				//	log.Error("sending message to client", "error", err, "client", addr.String())
				//} else {
				//	log.Info("successful interaction with client", "x", x, "y", x*2, "client", addr.String())
				//}
			}
		}
	}
}

func handleErr(e error) {
	if nil != e {
		log.Error("Exiting with error ", e.Error())
		println(e)
		os.Exit(1)
	}
}

