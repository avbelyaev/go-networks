package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mgutz/logxi/v1"
	"math"
	"net"
)

import (
	"network-labs/nw-1-protocol/lab/proto"
	"strconv"
)


type StatelessClient struct {
	logger log.Logger    // Объект для печати логов
	conn   *net.TCPConn  // Объект TCP-соединения
	enc    *json.Encoder // Объект для кодирования и отправки сообщений
}

// NewClient - конструктор клиента, принимает в качестве параметра
// объект TCP-соединения.
func NewClient(conn *net.TCPConn) *StatelessClient {
	return &StatelessClient{
		logger: log.New(fmt.Sprintf("client %s", conn.RemoteAddr().String())),
		conn:   conn,
		enc:    json.NewEncoder(conn),
	}
}

// handle - метод, в котором реализован цикл взаимодействия с клиентом.
// Подразумевается, что метод handle будет вызаваться в отдельной go-программе.
func (client *StatelessClient) handle() {
	defer client.conn.Close()
	decoder := json.NewDecoder(client.conn)
	for {
		var req proto.Request
		if err := decoder.Decode(&req); err != nil {
			client.logger.Error("cannot decode message")
			break
		} else {
			client.logger.Info("received command", "command", req.Command)
			if client.handleRequest(&req) {
				client.logger.Info("shutting down connection")
				break
			}
		}
	}
}

func countSquareDifference(coord1 string, coord2 string) (float64, error) {
	circleCenterX, err1 := strconv.ParseFloat(string(coord1), 64)
	circleContourX, err2 := strconv.ParseFloat(string(coord2), 64)
	if nil != err1 {
		return 0, err1

	} else if nil != err2 {
		return 0, err2

	} else {
		delta := circleCenterX - circleContourX
		return delta * delta, nil
	}
}

func countCircleSquare(circle proto.Circle) (float64, error) {
	deltaXSquared, err := countSquareDifference(circle.Center.CoordX, circle.Contour.CoordX)
	if nil != err {
		return 0, err
	}

	deltaYSquared, err := countSquareDifference(circle.Center.CoordY, circle.Contour.CoordY)
	if nil != err {
		return 0, err
	}

	radius := math.Sqrt(deltaXSquared + deltaYSquared)
	return math.Pi * radius * radius, nil
}

// handleRequest - метод обработки запроса от клиента. Он возвращает true,
// если клиент передал команду "quit" и хочет завершить общение.
func (client *StatelessClient) handleRequest(req *proto.Request) bool {
	switch req.Command {
	case proto.CMD_QUIT:
		client.respond(proto.RSP_STATUS_OK, nil)
		return true

	case proto.CMD_COUNT:
		errorMsg := ""
		if req.Data == nil {
			errorMsg = "data field is absent"

		} else {
			var circle proto.Circle
			if err := json.Unmarshal(*req.Data, &circle); err != nil {
				errorMsg = "malformed data field"

			} else {
				if circleSquare, err := countCircleSquare(circle); nil != err {
					errorMsg = "could not count circle square, invalid circle data provided"

				} else {
					circleSquareAsString := strconv.FormatFloat(circleSquare, 'f', 6, 64)
					client.logger.Info("square of circle has been counted", "value", circleSquareAsString)

					client.respond(proto.RSP_STATUS_SUCCESS, circleSquareAsString)
				}
			}
		}
		if errorMsg == "" {
			// client.respond(proto.RSP_STATUS_OK, nil)

		} else {
			client.logger.Error("count failed", "reason", errorMsg)
			client.respond(proto.RSP_STATUS_FAIL, errorMsg)
		}

	default:
		client.logger.Error("unknown command")
		client.respond("failed", "unknown command")
	}
	return false
}

// respond - вспомогательный метод для передачи ответа с указанным статусом
// и данными. Данные могут быть пустыми (data == nil).
func (client *StatelessClient) respond(status string, data interface{}) {
	var raw json.RawMessage
	raw, _ = json.Marshal(data)
	client.enc.Encode(&proto.Response{status, &raw})
}

func main() {
	// Работа с командной строкой, в которой может указываться необязательный ключ -addr.
	var addrStr string
	flag.StringVar(&addrStr, "addr", "127.0.0.1:6002", "specify ip address and port")
	flag.Parse()

	// Разбор адреса, строковое представление которого находится в переменной addrStr.
	if addr, err := net.ResolveTCPAddr("tcp", addrStr); err != nil {
		log.Error("address resolution failed", "address", addrStr)
	} else {
		log.Info("resolved TCP address", "address", addr.String())

		// Инициация слушания сети на заданном адресе.
		if listener, err := net.ListenTCP("tcp", addr); err != nil {
			log.Error("listening failed", "reason", err)
		} else {
			// Цикл приёма входящих соединений.
			for {
				if conn, err := listener.AcceptTCP(); err != nil {
					log.Error("cannot accept connection", "reason", err)
				} else {
					log.Info("accepted connection", "address", conn.RemoteAddr().String())

					// Запуск go-программы для обслуживания клиентов.
					go NewClient(conn).handle()
				}
			}
		}
	}
}
