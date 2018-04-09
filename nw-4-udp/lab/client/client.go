package main

import (
	"flag"
	"github.com/mgutz/logxi/v1"
	"net"
	"os"
	"fmt"
	"encoding/json"
	"github.com/skorobogatov/input"
	//"time"
	"network-labs/nw-4-udp/lab/proto"
	"time"
	"strconv"
)

const (
	MAX_WRITE_RETRIES = 50
	MAX_READ_RETRIES = 50

	UDP_READ_DEADLINE_MILLIS = 50

	BUFFER_SIZE_BYTES = 1000
)

type DurableClient struct {
	conn           	*net.UDPConn
	addr           	*net.UDPAddr
	messages		map[string]bool
	lastMessageId	string
}

func NewDurableClient(conn *net.UDPConn, addr *net.UDPAddr, retryTimeoutMillis int) *DurableClient {
	return &DurableClient{
		conn:           conn,
		addr:           addr,
		messages:		make(map[string]bool),
	}
}



func (client *DurableClient) interact() {
	defer client.conn.Close()

	var message *proto.Message
	for {
		fmt.Printf("command: ")
		var command = input.Gets()

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



		reSend:
		// update message Id, save message Id as latest
		message.UpdateId()
		client.lastMessageId = message.Id

		// send and make sure we got an ACK for it. otherwise resend
		var ack = client.sendReliably(message)
		if nil == ack {
			log.Error("Message could not be delivered :( Resending")
			goto reSend
		}


		// read answer for message. Resend it if ans cannot be read
		var rsp = client.readReliably()
		if nil == rsp {
			log.Error("Answer could not be received :( Resending")
			goto reSend
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
			fmt.Sprintf("Unknown command %s. Resending", rsp.Command)
			goto reSend
		}
	}
}


// sends message and receives an ACK for it. no ACK -> message was not received -> ret nil
func (client *DurableClient) sendReliably(message *proto.Message) *proto.Message {
	// serialize message
	var serialized, err = json.Marshal(message)
	handleErr(err)

	var ack *proto.Message
	var writeRetries = MAX_WRITE_RETRIES

	for writeRetries > 0 {
		log.Debug(fmt.Sprintf("Sending. Retries left: %s, %s", strconv.Itoa(writeRetries), message.Id))
		client.conn.Write(serialized)


		ack = client.readConditionally(func(msg *proto.Message) bool {
			// message was delivered if ACK that we got == ACK we were looking for
			return client.lastMessageId == msg.Id && proto.CMD_ACK == msg.Command
		})
		if nil != ack {
			log.Debug("Ack for message has been received")
			return ack
		}

		writeRetries--
	}

	return nil
}


// reads message that is not an ACK. 
func (client *DurableClient) readReliably() *proto.Message {
	var rsp *proto.Message

	var readRetries = MAX_READ_RETRIES
	for readRetries > 0 {

		rsp = client.readConditionally(func(msg *proto.Message) bool {
			// answer is a message that we were looking for (by id). ACKs are ignored
			return client.lastMessageId == msg.Id && proto.CMD_ACK != msg.Command
		})
		if nil != rsp {
			log.Debug("Answer has been received")
			return rsp
		}

		readRetries--
	}

	return nil
}


func (client *DurableClient) readConditionally(targetMsgCondition func(msg *proto.Message) bool) *proto.Message {
	// set timeout for nearest Read() call
	client.conn.SetReadDeadline(time.Now().Add(time.Duration(UDP_READ_DEADLINE_MILLIS) * time.Millisecond))

	// read into buffer and check if there was timeout error while reading
	var buf = make([]byte, BUFFER_SIZE_BYTES)
	var bytesRead, readErr = client.conn.Read(buf)

	// deserialize into regular message
	var msgBytes = buf[:bytesRead]
	var rsp proto.Message
	var deserializeErr = json.Unmarshal(msgBytes, &rsp)
	// println("read ", string(msgBytes))

	if nil == readErr && nil == deserializeErr && targetMsgCondition(&rsp) {
		log.Debug("Traget message acquired")
		return &rsp
	
	} else {
		return nil
	}
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