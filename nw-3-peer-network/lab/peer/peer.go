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
	wg		sync.WaitGroup
	lock	sync.Mutex
}


// create peer object
func newPeer() *Peer {
	return &Peer{
		logger: log.New("peer"),
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
	defer peer.wg.Done()
	addr, err := net.ResolveTCPAddr("tcp", selfAddr)
	handleErr(err)

	listener, err := net.ListenTCP("tcp", addr)
	handleErr(err)
	println("listening at ", listener.Addr().String())
	for {
		inConn, err := listener.AcceptTCP()
		handleErr(err)

		//peer.setupConnForServerPart(inConn)
		peer.wg.Add(1)
		go peer.handleIncomingConnection(inConn)
	}
}


func (peer *Peer) startClient(nextAddr string) {
	addr, err := net.ResolveTCPAddr("tcp", nextAddr)
	handleErr(err)

	outConn, err := net.DialTCP("tcp", nil, addr)
	handleErr(err)

	println("connecting [", outConn.LocalAddr().String(), " -> ", outConn.RemoteAddr().String(), "]")
	peer.setupConnForClientPart(outConn)
	peer.interact(outConn)
}


func (peer *Peer) interact(conn *net.TCPConn) {
	println("interact")

	defer conn.Close()
	encoder, decoder := json.NewEncoder(conn), json.NewDecoder(conn)
	for {
		fmt.Printf("command = ")
		rq := input.Gets()

		switch rq {
		case CMD_CONNECT:
			sendMessage(encoder, CMD_CONNECT, nil)

		case CMD_QUIT:
			sendMessage(encoder, CMD_QUIT, nil)
			return

		default:
			fmt.Printf("error: unknown command\n")
			continue
		}

		// Получение ответа.
		var rsp Message
		if err := decoder.Decode(&rsp); err != nil {
			fmt.Printf("error: %v\n", err)
			break
		}
		println("response has been decoded")

		switch rsp.Command {
		case CMD_OK:
			println("response ok has been received")

		default:
			fmt.Printf("error: server reports unknown command %q\n", rsp.Command)
		}
	}
}


// on each incoming tcp connection to server call this function in goroutine
func (peer *Peer) handleIncomingConnection(conn *net.TCPConn) {
	defer conn.Close()
	defer peer.wg.Done()

	println("handling connection [", conn.LocalAddr().String(), " <- ",
		conn.RemoteAddr().String(), "]")
	decoder := json.NewDecoder(conn)
	// listen for messages in connection
	for {
		var rq Message
		err := decoder.Decode(&rq)
		handleErr(err)

		if peer.handleRequestMessageAndExit(&rq, conn) {
			break
		}
	}
}


// if request can be decoded, server handles its content with this function
func (peer *Peer) handleRequestMessageAndExit(rq *Message, conn *net.TCPConn) bool {
	defer func() {
		println("unlocking peer")
		peer.lock.Unlock()
	}()

	println("locking on peer")
	peer.lock.Lock()
	println("handling message with command '", rq.Command, "'")

	var encoder = json.NewEncoder(conn)
	switch rq.Command {
	case CMD_CONNECT:
		println("smb has connected")
		sendMessage(encoder, CMD_OK, &Message{
			Command: "fuck you!",
			Data:    nil,
		})

	case CMD_QUIT:
		print("quit")
		return true

	default:
		println("server has received command '", rq.Command, "'!")
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
	peer.wg.Add(1)
	go peer.startServer("127.0.0.1:6001")
	peer.startClient("127.0.0.1:6001")

	println("waiting to finish")
	peer.wg.Wait()
}


func handleErr(e error) {
	if nil != e {
		log.Error("Exiting with error ", e.Error())
		println(e)
		os.Exit(1)
	}
}
