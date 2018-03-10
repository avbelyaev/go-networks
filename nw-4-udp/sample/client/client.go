package main

import (
	"flag"
	"fmt"
	"github.com/mgutz/logxi/v1"
	"math/rand"
	"net"
	"os"
	"strconv"
)

func main() {
	var (
		serverAddrStr string
		n             uint
		helpFlag      bool
	)
	flag.StringVar(&serverAddrStr, "server", "127.0.0.1:6000", "set server IP address and port")
	flag.UintVar(&n, "n", 10, "set the number of requests")
	flag.BoolVar(&helpFlag, "help", false, "print options list")

	if flag.Parse(); helpFlag {
		fmt.Fprint(os.Stderr, "client [options]\n\nAvailable options:\n")
		flag.PrintDefaults()
	} else if serverAddr, err := net.ResolveUDPAddr("udp", serverAddrStr); err != nil {
		log.Error("resolving server address", "error", err)
	} else if conn, err := net.DialUDP("udp", nil, serverAddr); err != nil {
		log.Error("creating connection to server", "error", err)
	} else {
		defer conn.Close()

		buf := make([]byte, 32)
		for i := uint(0); i < n; i++ {
			x := rand.Intn(1000)
			if _, err := conn.Write([]byte(strconv.Itoa(x))); err != nil {
				log.Error("sending request to server", "error", err, "x", x)
			} else if bytesRead, err := conn.Read(buf); err != nil {
				log.Error("receiving answer from server", "error", err)
			} else {
				yStr := string(buf[:bytesRead])
				if y, err := strconv.Atoi(yStr); err != nil {
					log.Error("cannot parse answer", "answer", yStr, "error", err)
				} else {
					log.Info("successful interaction with server", "x", x, "y", y)
				}
			}
		}
	}
}
