package main

import (
	"flag"
	"fmt"
	"github.com/mgutz/logxi/v1"
	"net"
	"os"
	//"strconv"
	"encoding/json"
	"strconv"
	"math"
	"network-labs/nw-4-udp/lab/proto"
	//"time"
	"time"
)


var MESSAGES = make(map[string]bool)


type StatelessClient struct {
	logger 			log.Logger    // Объект для печати логов
	conn   			*net.UDPConn  // Объект TCP-соединения
	addr 			*net.UDPAddr
	lastMessageId 	string
}


func NewClient(conn *net.UDPConn, addr *net.UDPAddr) *StatelessClient {
	return &StatelessClient{
		conn:   	conn,
		addr:		addr,
		logger:		log.New(fmt.Sprintf("Client %s", addr)),
	}
}


func messageNotExists(id string) bool {
	var idAlreadyExists, _ = MESSAGES[id]
	if !idAlreadyExists {
		MESSAGES[id] = true
		return true

	} else {
		return false
	}
}


func countSquareDifference(coord1 string, coord2 string) (float64, error) {
	circleCenterX, err1 := strconv.ParseFloat(string(coord1), 64)
	circleContourX, err2 := strconv.ParseFloat(string(coord2), 64)
	if nil != err1 {
		return 0, err1

	} else if nil != err2 {
		return 0, err2

	} else {
		delta := circleCenterX - circleContourX
		return delta * delta, nil
	}
}

func countCircleSquare(circle proto.Circle) (float64, error) {
	//println("center: (", circle.Center.CoordX, ", ", circle.Center.CoordY, ")")
	//println("cntour: (", circle.Contour.CoordX, ", ", circle.Contour.CoordY, ")")
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

	var ack = proto.NewMessage(proto.CMD_ACK, nil)
	ack.Id = req.Id
	log.Debug(fmt.Sprintf("Sending ACK %s", ack.Id))
	client.sendReliably(ack)

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

	time.Sleep(1 * time.Second)

	message.Id = req.Id
	client.lastMessageId = message.Id
	log.Debug(fmt.Sprintf("Sending DATA %s", message.Id))
	client.sendReliably(message)

	//client.conn.Close()
	//client.respond(message)
	MESSAGES[message.Id] = true
}


func (client *StatelessClient) respond(message *proto.Message) {
	var msgBytes, err = json.Marshal(message)
	handleErr(err)

	client.conn.WriteToUDP(msgBytes, client.addr)
}


func (client *StatelessClient) sendReliably(message *proto.Message) *proto.Message {
	var serialized, err = json.Marshal(message)
	handleErr(err)

	var retriesLeft = 20
	for retriesLeft > 0 {

		time.Sleep(50 * time.Millisecond)
		log.Debug(fmt.Sprintf("Out. Retries left: %s, %s", strconv.Itoa(retriesLeft), message.Id))
		client.conn.WriteToUDP(serialized, client.addr)

		retriesLeft--
	}

	// message was not delivered :(
	return nil
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
				//log.Error("receiving message from client error")

			} else {
				var rqBytes = buf[:bytesRead]
				log.Info("Got", "msg", string(rqBytes))

				var rq proto.Message
				var err = json.Unmarshal(rqBytes, &rq)
				handleErr(err)

				if messageNotExists(rq.Id) {
					NewClient(conn, addr).handleRequest(&rq)

				} else {
					log.Debug(fmt.Sprintf("Ignored duplicate %s", rq.Id))
				}
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

