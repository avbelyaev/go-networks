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
	"time"
	"strconv"
)

type DurableClient struct {
	conn           	*net.UDPConn
	addr           	*net.UDPAddr
	timeoutSeconds 	int         // if there is no response within this time, message will be sent again
	messagesStack  	stack.Stack // ordered list of ids of messages that were sent
	maxRetries		int
	messages		map[string]bool
	lastMessageId	string
}

func NewDurableClient(conn *net.UDPConn, addr *net.UDPAddr, retryTimeoutMillis int) *DurableClient {
	return &DurableClient{
		conn:           conn,
		addr:           addr,
		timeoutSeconds: retryTimeoutMillis,
		messagesStack:  stack.Stack{},
		maxRetries:		10,
		messages:		make(map[string]bool),
	}
}


//func handleMessageDrop(conn *net.UDPConn, timeoutSeconds int) {
//
//}


func (client *DurableClient) interact() {
	defer client.conn.Close()

	for {
		fmt.Printf("command: ")
		var command = proto.CMD_COUNT//input.Gets()
		var message *proto.Message

		switch command {
		case proto.CMD_QUIT:
			message = proto.NewMessage(proto.CMD_QUIT, nil)
			return

		case proto.CMD_COUNT:
			fmt.Printf("Center X coordinate:")
			var x1 = input.Gets()
			fmt.Printf("Center X coordinate:")
			var y1 = "1"//input.Gets()
			fmt.Printf("Contour X coordinate:")
			var x2 = "2"//input.Gets()
			fmt.Printf("Contour Y coordinate:")
			var y2 = "2"//input.Gets()

			var circle = proto.NewCircle(x1, y1, x2, y2)
			message = proto.NewMessage(proto.CMD_COUNT, circle)

		default:
			fmt.Printf("error: unknown command\n")
			continue
		}

		// handle outcoming message
		client.lastMessageId = message.Id
		var rsp = client.sendReliably(message)
		if nil == rsp {
			log.Error("Message could not be delivered :(")
			continue
		}

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


func (client *DurableClient) sendReliably(message *proto.Message) *proto.Message {
	var serialized, err = json.Marshal(message)
	handleErr(err)

	var retriesLeft = 50
	for retriesLeft > 0 {
		log.Debug(fmt.Sprintf("Sending message. Retries left: %s", strconv.Itoa(retriesLeft)))
		client.conn.Write(serialized)

		// set timeout for nearest Read() call
		client.conn.SetReadDeadline(time.Now().Add(time.Duration(100) * time.Millisecond))

		// read into buffer and check if there was timeout error while reading
		var buf = make([]byte, 1000)
		var bytesRead, readErr = client.conn.Read(buf)

		// deserialize into regular message and check if message == ACK
		var ackBytes = buf[:bytesRead]
		//println("got ", string(ackBytes))
		var ack proto.Message
		var deserializeErr = json.Unmarshal(ackBytes, &ack)

		if nil == readErr && nil == deserializeErr && client.lastMessageId == ack.Id {
			// there is an ACK for the message => it was delivered for sure
			log.Debug("Ack for message has been acquired")
			return &ack
		}
		retriesLeft--
	}

	// message was not delivered :(
	return nil
}


func main() {
	var serverAddrStr string
	flag.StringVar(&serverAddrStr, "server", "127.0.0.1:5000", "set server IP address and port")

	var serverAddr, addrErr = net.ResolveUDPAddr("udp", serverAddrStr)
	handleErr(addrErr)

	var conn, err = net.DialUDP("udp", nil, serverAddr)
	handleErr(err)

	var client = NewDurableClient(conn, serverAddr, 3)
	client.interact()
}


func handleErr(e error) {
	if nil != e {
		log.Error("Exiting with error ", e.Error())
		println(e)
		os.Exit(1)
	}
}