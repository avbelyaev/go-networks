package main

import (
	"flag"
	"github.com/mgutz/logxi/v1"
	"net"
	"encoding/json"
	"fmt"
	//"strings"
	//"strconv"
	"os"
	//"debug/pe"
)

type Peer struct {
	logger 		log.Logger    		// Объект для печати логов
	conn   		*net.TCPConn  		// Объект TCP-соединения
	enc    		*json.Encoder 		// Объект для кодирования и отправки сообщений
	selfAddr	string		  		// свой адрес, строка вида ip:port
	nextAddr	string   			// строка вида ip:port следующего пира
}

// create peer object
func newPeer(conn *net.TCPConn) *Peer {
	return &Peer {
		logger: log.New(fmt.Sprintf("client %s", conn.RemoteAddr().String())),
		conn:   conn,
		enc:    json.NewEncoder(conn),
	}
}

// start listening to incoming tcp connections
func startServer(serverAddr string)  {
	addr, err := net.ResolveTCPAddr("tcp", serverAddr)
	handleErr(err)

	listener, err := net.ListenTCP("tcp", addr)
	handleErr(err)
	for {
		incomingConnection, err := listener.A()
		handleErr(err)
		peer := newPeer(incomingConnection)
		print(peer)
		//peer.connectToNextPeer()

		//go peer.handleIncomingConnection()
	}
}

// on each incoming tcp connection call this function in go routine
func (peer *Peer) handleIncomingConnection()  {
	defer peer.conn.Close()
	decoder := json.NewDecoder(peer.conn)
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

	startServer(addrStr)
}


func handleErr(e error)  {
	if nil != e {
		log.Error("Exiting with error")
		println(e)
		os.Exit(1)
	}
}