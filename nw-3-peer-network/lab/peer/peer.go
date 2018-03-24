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
	logger 	log.Logger
	client 	Client
	server 	Server
	lock	sync.Mutex
	enabled	bool
}


// create peer object
func newPeer() *Peer {
	return &Peer{
		logger: 	log.New("peer"),
		enabled:	true,
	}
}


// setup client for peer
func (peer *Peer) setupConnForClientPart(conn *net.TCPConn) {
	peer.client = Client{
		conn: conn,
		enc:  json.NewEncoder(conn),
	}
}


// setup server for peer
func (peer *Peer) setupConnForServerPart(conn *net.TCPConn) {
	peer.server = Server{
		conn: conn,
		enc:  json.NewEncoder(conn),
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

		//peer.wg.Add(1)
		go peer.handleIncomingConnection(inConn)
	}
}

func (peer *Peer) stopServer()  {
	logS("shutting down server")
	peer.enabled = false
}


func (peer *Peer) startClient(nextAddr string) {
	addr, err := net.ResolveTCPAddr("tcp", nextAddr)
	handleErr(err)

	// wait for server to start properly
	time.Sleep(1 * time.Second)
	outConn, err := net.DialTCP("tcp", nil, addr)
	handleErr(err)

	logC("connecting [", outConn.LocalAddr().String(), " -> ", outConn.RemoteAddr().String(), "]")
	peer.setupConnForClientPart(outConn)
	peer.interact(outConn)
}


func (peer *Peer) interact(conn *net.TCPConn) {
	logC("interacting")

	defer conn.Close()
	var encoder, decoder = json.NewEncoder(conn), json.NewDecoder(conn)
	for {
		fmt.Printf("command = ")
		var rq = input.Gets()

		switch rq {
		case CMD_CONNECT:
			sendMessage(encoder, CMD_CONNECT, nil)

		case CMD_STOP:
			peer.stopServer()
			return

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
			if nil == rsp.Data {
				logC("empty response data")

			} else {
				var responseFromServer string
				var err = json.Unmarshal(*rsp.Data, &responseFromServer)
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
func (peer *Peer) handleRequestMessageWithExitFlag(rq *Message, conn *net.TCPConn) bool {
	defer func() {
		logS("unlocking peer")
		peer.lock.Unlock()
	}()

	logS("locking on peer")
	peer.lock.Lock()
	logS("handling message with command '", rq.Command, "'")

	var encoder = json.NewEncoder(conn)
	switch rq.Command {
	case CMD_CONNECT:
		logS("smb has connected. responding")
		sendMessage(encoder, CMD_OK, "Fuck you!")

	case CMD_STOP:
		logS("peer's server can be shut down only by peer's client")

	default:
		logS("server has received command '", rq.Command, "'!")
	}
	return false
}


func sendMessage(encoder *json.Encoder, command string, data interface{}) {
	var raw json.RawMessage
	raw, _ = json.Marshal(data)
	encoder.Encode(Message{command, &raw})
}


func main() {
	// Работа с командной строкой, в которой может указываться необязательный ключ -addr.
	var addrStr string
	flag.StringVar(&addrStr, "addr", "127.0.0.1:6002", "specify ip address and port")
	flag.Parse()

	peer := newPeer()
	go peer.startServer("127.0.0.1:6001")
	peer.startClient("127.0.0.1:6001")
}


func handleErr(e error) {
	if nil != e {
		log.Error("Exiting with error ", e.Error())
		println(e)
		os.Exit(1)
	}
}

func logS(args ...string)  {
	print("Server: ")
	fmt.Print(args)
	println()
}

func logC(args ...string)  {
	print("-Client: ")
	fmt.Print(args)
	println()
}
