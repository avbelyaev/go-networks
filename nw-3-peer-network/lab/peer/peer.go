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
)

type Client struct {
	conn 		*net.TCPConn
	enc			*json.Encoder
	nextAddr	string
}

type Server struct {
	conn 		*net.TCPConn
	enc			*json.Encoder
	selfAddr	string
}

type Peer struct {
	logger 		log.Logger	// Объект для печати логов
	client		Client
	server 		Server
}

// create peer object
func newPeer() *Peer {
	return &Peer {
		logger: 	log.New("peer"),
	}
}

// setup client for peer
func (peer *Peer) setupConnectionFromClient(conn *net.TCPConn) {
	peer.client = Client{
		conn: 		conn,
		enc: 		json.NewEncoder(conn),
	}
}

// setup server for peer
func (peer *Peer) setupConnectionToServer(conn *net.TCPConn) {
	peer.server = Server{
		conn: 		conn,
		enc: 		json.NewEncoder(conn),
	}
}

// start listening to incoming tcp connections
func (peer *Peer) startServer(selfAddr string)  {
	addr, err := net.ResolveTCPAddr("tcp", selfAddr)
	handleErr(err)

	listener, err := net.ListenTCP("tcp", addr)
	handleErr(err)
	for {
		inConn, err := listener.AcceptTCP()
		handleErr(err)

		peer.setupConnectionToServer(inConn)

		go peer.handleIncomingConnection()
	}
}


func (peer *Peer) startClient(nextAddr string)  {
	addr, err := net.ResolveTCPAddr("tcp", nextAddr)
	handleErr(err)

	outConn, err := net.DialTCP("tcp", nil, addr)
	handleErr(err)

	peer.setupConnectionFromClient(outConn)
	peer.interact(outConn)
}


func (peer *Peer) interact(conn *net.TCPConn) {

}

// on each incoming tcp connection to server call this function in go routine
func (peer *Peer) handleIncomingConnection()  {
	defer peer.server.conn.Close()
	decoder := json.NewDecoder(peer.client.conn)
	for {
		var rq Request
		if err := decoder.Decode(&rq); err != nil {
			peer.logger.Error("cannot decode message")
			break

		} else {
			peer.logger.Info("received command", rq.Command)
			peer.handleRequest(&rq)
		}
	}
}

// if request can be decoded, handle its content with this function
func (peer *Peer) handleRequest(rq *Request) {

}


func main() {
	// Работа с командной строкой, в которой может указываться необязательный ключ -addr.
	var addrStr string
	flag.StringVar(&addrStr, "addr", "127.0.0.1:6002", "specify ip address and port")
	flag.Parse()

	peer := newPeer()
	peer.startServer("127.0.0.1:6001")
	peer.startClient("127.0.0.1:6002")
}


func handleErr(e error)  {
	if nil != e {
		log.Error("Exiting with error")
		println(e)
		os.Exit(1)
	}
}