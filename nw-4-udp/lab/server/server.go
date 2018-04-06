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
	logger 	log.Logger    // Объект для печати логов
	conn   	*net.UDPConn  // Объект TCP-соединения
	addr 	*net.UDPAddr
}


func NewClient(conn *net.UDPConn, addr *net.UDPAddr) *StatelessClient {
	return &StatelessClient{
		conn:   	conn,
		addr:		addr,
		logger:		log.New(fmt.Sprintf("Client %s", addr)),
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


func (client *StatelessClient) handleRequest(req *proto.Message) {
	var message *proto.Message

	switch req.Command {
	case proto.CMD_QUIT:
		message = proto.NewMessage(proto.CMD_OK, nil)

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

					message = proto.NewMessage(proto.CMD_SUCCESS, circleSquareAsString)
				}
			}
		}
		if errorMsg == "" {
			//message = proto.NewMessage(proto.CMD_OK, nil)

		} else {
			client.logger.Error("count failed", "reason", errorMsg)
			message = proto.NewMessage(proto.CMD_FAIL, errorMsg)
		}

	default:
		client.logger.Error("unknown command")
		message = proto.NewMessage(proto.CMD_UNKNOWN, "unknown command")
	}

	for {

	}
	client.respond(message)
}


func (client *StatelessClient) respond(message *proto.Message) {
	var msgBytes, err = json.Marshal(message)
	handleErr(err)

	client.conn.WriteToUDP(msgBytes, client.addr)
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
		log.Info("server listens incoming messages from clients",
			"addr", serverAddr.String())
		buf := make([]byte, 1000)
		for {
			if bytesRead, addr, err := conn.ReadFromUDP(buf); err != nil {
				log.Error("receiving message from client", "error", err)
			} else {
				log.Info("Reading", "read bytes", bytesRead, "from", addr)

				var rqBytes = buf[:bytesRead]
				log.Info("Got", "msg", string(rqBytes))

				var rq proto.Message
				var err = json.Unmarshal(rqBytes, &rq)
				handleErr(err)

				go NewClient(conn, addr).handleRequest(&rq)
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

