package main

import (
	"flag"
	"fmt"
	"github.com/mgutz/logxi/v1"
	"net"
	"os"
	//"strconv"
)

/*
порядок доставки сообщений
retry если ответ не пришел за отведенное время
SYN ACK зашить в rq/rsp
 */

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
		buf := make([]byte, 32)
		for {
			if bytesRead, addr, err := conn.ReadFromUDP(buf); err != nil {
				log.Error("receiving message from client", "error", err)
			} else {
				log.Info("Idling", "read bytes", bytesRead, "from", addr)
				for {

				}
				//s := string(buf[:bytesRead])
				//if x, err := strconv.Atoi(s); err != nil {
				//	log.Error("cannot parse answer", "answer", s, "error", err)
				//} else if _, err = conn.WriteToUDP([]byte(strconv.Itoa(x*2)), addr); err != nil {
				//	log.Error("sending message to client", "error", err, "client", addr.String())
				//} else {
				//	log.Info("successful interaction with client", "x", x, "y", x*2, "client", addr.String())
				//}
			}
		}
	}
}


