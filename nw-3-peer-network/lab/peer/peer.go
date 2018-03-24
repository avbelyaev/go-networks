package main

import (
	"flag"
	"github.com/mgutz/logxi/v1"
	"net"
	"encoding/json"
	//"fmt"
	//"strings"
	//"strconv"
	"os"
	//"debug/pe"
	"fmt"
	//"github.com/skorobogatov/input"
	"sync"
	//"time"
	//"time"
	"github.com/skorobogatov/input"
	//"time"
	"time"
	"golang.org/x/crypto/openpgp/errors"
	//"strconv"
	//"golang.org/x/text/message"
)

type Client struct {
	conn     *net.TCPConn
	enc      *json.Encoder
	nextAddr string
}

type Server struct {
	conn     *net.TCPConn
	enc      *json.Encoder
	selfAddr string
}

type Peer struct {
	logger 		log.Logger
	client 		Client
	server 		Server
	lock		sync.Mutex
	enabled		bool
	messages	map[string]bool	// we only care about keys in this map
}


// create peer object
func newPeer() *Peer {
	return &Peer{
		logger: 	log.New("peer"),
		enabled:	true,
		messages:	make(map[string]bool),
	}
}

func (peer *Peer) saveMessageIdIfNotExists(id string) bool {
	var idAlreadyExists, _ = peer.messages[id]
	if !idAlreadyExists {
		peer.messages[id] = true
		return true

	} else {
		return false
	}
}


// start listening to incoming tcp connections
func (peer *Peer) startServer(selfAddr string) {
	addr, err := net.ResolveTCPAddr("tcp", selfAddr)
	handleErr(err)

	listener, err := net.ListenTCP("tcp", addr)
	handleErr(err)
	logS("listening at ", listener.Addr().String())

	for peer.enabled {
		inConn, err := listener.AcceptTCP()
		handleErr(err)

		go peer.handleIncomingConnection(inConn)
	}
}

func (peer *Peer) stopServer()  {
	logS("shutting down server")
	peer.enabled = false
}


func (peer *Peer) startClient(serverAddr string) {
	var outConn, err = connectToServerWithinTimeout(serverAddr, 5)
	handleErr(err)

	logC("connecting [", outConn.LocalAddr().String(), " -> ", outConn.RemoteAddr().String(), "]")
	peer.interact(outConn)
}


func connectToServerWithinTimeout(serverAddr string, retryTimeoutSeconds int) (*net.TCPConn, error) {
	var addr, err = net.ResolveTCPAddr("tcp", serverAddr)
	handleErr(err)

	var i = 0
	for i < retryTimeoutSeconds {
		// wait for server to start properly
		time.Sleep(1 * time.Second)

		var conn, err = net.DialTCP("tcp", nil, addr)
		if nil != err {
			logC("could not connect to ", serverAddr, ". retry ", i + 1, " of ", retryTimeoutSeconds)

		} else {
			return conn, nil
		}
		i++
	}
	return nil, errors.InvalidArgumentError("could not connect to server. probably incorrect addr")
}


func (peer *Peer) interact(conn *net.TCPConn) {
	logC("interacting")

	defer conn.Close()
	var encoder, decoder = json.NewEncoder(conn), json.NewDecoder(conn)
	peer.client.enc = encoder

	for {
		fmt.Printf("command = ")
		var rq = input.Gets()

		switch rq {
		case CMD_CONNECT:
			sendMessage(encoder, CMD_CONNECT, nil)

		case CMD_QUIT:
			peer.stopServer()
			return

		case CMD_MSG:
			fmt.Printf("Type your message: ")
			var msgText = input.Gets()
			var msg = newMessage(CMD_MSG, msgText)
			peer.saveMessageIdIfNotExists(msg.Id)
			resendMessage(encoder, msg)
			continue

		default:
			logC("unknown command")
			continue
		}

		// Получение ответа.
		var rsp Message
		if err := decoder.Decode(&rsp); err != nil {
			fmt.Printf("error: %v\n", err)
			break
		}
		logC("response has been decoded")

		switch rsp.Command {
		case CMD_OK:
			logC("response with command ok has been received")
			if nil == rsp.Payload {
				logC("empty response data")

			} else {
				var responseFromServer string
				var err = json.Unmarshal(*rsp.Payload, &responseFromServer)
				handleErr(err)

				logC("response data: ", responseFromServer)
			}

		default:
			logC("unknown command ", rsp.Command)
		}
	}
}


// on each incoming tcp connection to server call this function in goroutine
func (peer *Peer) handleIncomingConnection(conn *net.TCPConn) {
	defer conn.Close()

	logS("handling connection [", conn.LocalAddr().String(), " <- ",
		conn.RemoteAddr().String(), "]")
	var decoder = json.NewDecoder(conn)
	// listen for messages in connection
	for {
		var rq Message
		err := decoder.Decode(&rq)
		handleErr(err)

		if peer.handleRequestMessageWithExitFlag(&rq, conn) {
			break
		}
	}
}


// if request can be decoded, server handles its content with this function
func (peer *Peer) handleRequestMessageWithExitFlag(msg *Message, conn *net.TCPConn) bool {
	defer func() {
		logS("unlocking peer")
		peer.lock.Unlock()
	}()

	logS("locking on peer")
	peer.lock.Lock()
	logS("handling message with command '", msg.Command, "'")

	var encoder = json.NewEncoder(conn)
	switch msg.Command {
	case CMD_CONNECT:
		logS("smb has connected. responding")
		sendMessage(encoder, CMD_OK, nil)

	case CMD_QUIT:
		logS("peer's server can be shut down only by peer's client")

	case CMD_MSG:
		if nil == msg.Payload {
			logS("empty data")

		} else {
			peer.tryDisplayAndForwardMessage(msg)
		}

	default:
		logS("server has received command '", msg.Command, "'!")
	}
	return false
}


// if we do not have this message then display it, save and forward to next peer
func (peer *Peer) tryDisplayAndForwardMessage(message *Message) {
	if peer.saveMessageIdIfNotExists(message.Id) {

		var payload string
		var err = json.Unmarshal(*message.Payload, &payload)
		handleErr(err)

		logS("payload: ", payload)

		// forward it to next peer
		logS("forwarding message")
		time.Sleep(1 * time.Second)
		resendMessage(peer.client.enc, message)
	}
}


func main() {
	var testAddr = "127.0.0.1:6001"
	var selfAddr, nextAddr string
	flag.StringVar(&selfAddr, "self", testAddr, "specify self ip-addr:port")
	flag.StringVar(&nextAddr, "next", testAddr, "specify next ip-addr:port")
	flag.Parse()

	peer := newPeer()
	go peer.startServer(selfAddr)
	peer.startClient(nextAddr)
}


func handleErr(e error) {
	if nil != e {
		log.Error("Exiting with error ", e.Error())
		println(e)
		os.Exit(1)
	}
}

func logS(args ...interface{})  {
	print("Server: ")
	fmt.Print(args)
	println()
}

func logC(args ...interface{})  {
	print("-Client: ")
	fmt.Print(args)
	println()
}
