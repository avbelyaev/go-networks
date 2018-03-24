package main

import (
	"flag"
	"github.com/mgutz/logxi/v1"
	"net"
	"encoding/json"
	"os"
	"fmt"
	"github.com/skorobogatov/input"
	"time"
	"golang.org/x/crypto/openpgp/errors"
	"math/rand"
	"strconv"
)

type Client struct {
	enc      *json.Encoder
}

type Server struct {
	enabled bool
}

type Peer struct {
	name			string
	logger 			log.Logger
	client 			Client
	server			Server
	//lock			sync.Mutex
	messages		map[string]bool	// we only care about keys in this map
}


// create peer object
func newPeer(name string) *Peer {
	return &Peer{
		name:		name,
		logger: 	log.New("peer"),
		messages:	make(map[string]bool),
	}
}


// save message if this message was forwarded to peer
// if peer already contains message, ignore and return false
func (peer *Peer) saveMessageIdIfNotExists(id string) bool {
	var idAlreadyExists, _ = peer.messages[id]
	if !idAlreadyExists {
		peer.messages[id] = true
		return true

	} else {
		return false
	}
}


// only stops receiving messages as server
// sending messages to next (dead?) peer is still allowed (as client)
func (peer *Peer) stopServer()  {
	log.Info("Server is shutting down")
	peer.server.enabled = false
}


// ================ Server ================

// start listening to incoming tcp connections
func (peer *Peer) startServer(selfAddr string) {
	addr, err := net.ResolveTCPAddr("tcp", selfAddr)
	handleErr(err)

	listener, err := net.ListenTCP("tcp", addr)
	handleErr(err)
	log.Info(fmt.Sprintf("Server listening at %s", listener.Addr().String()))

	peer.server.enabled = true
	for peer.server.enabled {
		inConn, err := listener.AcceptTCP()
		handleErr(err)

		go peer.handleIncomingConnection(inConn)
	}
}


// on each incoming tcp connection to server call this function in goroutine
func (peer *Peer) handleIncomingConnection(conn *net.TCPConn) {
	defer conn.Close()

	log.Debug(fmt.Sprintf("Server handling connection (%s <- %s)", conn.LocalAddr().String(), conn.RemoteAddr().String()))
	var decoder = json.NewDecoder(conn)
	// listen for messages in connection
	for peer.server.enabled {
		var rq Message
		err := decoder.Decode(&rq)
		if nil != err {
			log.Error("Server could not decode message. Connection to peer probably lost")
			peer.stopServer()
		}

		if peer.handleRequestMessageWithExitFlag(&rq, conn) {
			break
		}
	}
}


// if request can be decoded, server handles its content with this function
// only returns false (returning true would close connection)
func (peer *Peer) handleRequestMessageWithExitFlag(msg *Message, conn *net.TCPConn) bool {
	//defer peer.lock.Unlock()

	//peer.lock.Lock()
	log.Debug(fmt.Sprintf("Server handling message with command '%s'", msg.Command))

	switch msg.Command {
	case CMD_MSG:
		if nil == msg.Payload {
			log.Debug("Server got empty data")

		} else {
			peer.tryDisplayAndForwardMessage(msg)
		}

	case CMD_EMPTY:
		log.Debug("Server connection has been lost")

	default:
		log.Debug("Server has received command", msg.Command)
	}
	return false
}


// if we do not have this message then display it, save and forward to next peer
func (peer *Peer) tryDisplayAndForwardMessage(message *Message) {
	if peer.saveMessageIdIfNotExists(message.Id) {

		var payload string
		var err = json.Unmarshal(*message.Payload, &payload)
		handleErr(err)

		log.Info("Incoming message ", message.Author, payload)

		// forward it to next peer
		log.Debug("Server forwarding message")
		time.Sleep(1 * time.Second)
		resendMessage(peer.client.enc, message)
	}
}


// ================ Client ================

func (peer *Peer) startClient(serverAddr string) {
	var outConn, err = connectToServerWithinTimeout(serverAddr, 10)
	handleErr(err)

	log.Info(fmt.Sprintf("Clt opening connection (%s -> %s)", outConn.LocalAddr().String(), outConn.RemoteAddr().String()))
	peer.interact(outConn)
}


// try acquire connection to server till period expires
func connectToServerWithinTimeout(serverAddr string, retryTimeoutSeconds int) (*net.TCPConn, error) {
	var addr, err = net.ResolveTCPAddr("tcp", serverAddr)
	handleErr(err)

	var i = 0
	for i < retryTimeoutSeconds {
		// wait for server to start properly
		time.Sleep(1 * time.Second)

		var conn, err = net.DialTCP("tcp", nil, addr)
		if nil != err {
			log.Debug(fmt.Sprintf("Clt could not connect to %s. retry %d of %d", serverAddr, i + 1, retryTimeoutSeconds))

		} else {
			return conn, nil
		}
		i++
	}
	return nil, errors.InvalidArgumentError("could not connect to server. probably incorrect addr")
}


// takes user input and sends messages upon it
func (peer *Peer) interact(conn *net.TCPConn) {
	defer conn.Close()

	var encoder, decoder = json.NewEncoder(conn), json.NewDecoder(conn)
	peer.client.enc = encoder

	for {
		fmt.Printf("command:\n")
		var cmd = input.Gets()

		switch cmd {
		case CMD_QUIT:
			peer.stopServer()
			sendMessage(encoder, CMD_QUIT, nil, peer.name)
			return

		case CMD_MSG:
			fmt.Printf("Type your message:\n")
			var msgText = input.Gets()
			var msg = newMessage(CMD_MSG, msgText, peer.name)
			// save message to already known before sending it
			peer.saveMessageIdIfNotExists(msg.Id)
			resendMessage(encoder, msg)
			continue

		default:
			log.Warn("Clt. Unknown command")
			continue
		}

		// Получение ответа.
		var rsp Message
		if err := decoder.Decode(&rsp); err != nil {
			fmt.Printf("error: %v\n", err)
			break
		}
		log.Debug("Clt response has been decoded")


		// FIXME following is probably dead code
		switch rsp.Command {
		case CMD_OK:
			if nil == rsp.Payload {
				log.Debug("Clt got empty payload")

			} else {
				var responseFromServer string
				var err = json.Unmarshal(*rsp.Payload, &responseFromServer)
				handleErr(err)

				log.Info("Incoming Message ", rsp.Author, responseFromServer)
			}

		default:
			log.Warn("Client got unknown command", rsp.Command)
		}
	}
}



// launch:
//go run peer.go message.go -name kirito -self localhost:6001 -next localhost:6002
//go run peer.go message.go -name asuna  -self localhost:6002 -next localhost:6001

// usage:
// m - message
// q - quit
func main() {
	var testAddr = "127.0.0.1:6001"
	var selfAddr, nextAddr, name string
	flag.StringVar(&name, "name", pickRandomName(), "specify name for a chat")
	flag.StringVar(&selfAddr, "self", testAddr, "specify self ip-addr:port")
	flag.StringVar(&nextAddr, "next", testAddr, "specify next ip-addr:port")
	flag.Parse()


	peer := newPeer(name)
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

func pickRandomName() string {
	rand.Seed(time.Now().Unix())
	var names = []string{
		"Asuna",
		"Lightning",
		"Yuuki",
		"Kirito",
		"Beater",
		"Black swordsman",
	}
	return names[rand.Intn(len(names))]
}
