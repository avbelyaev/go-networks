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

type Client struct {
	logger   log.Logger    // Объект для печати логов
	conn     *net.TCPConn  // Объект TCP-соединения
	enc      *json.Encoder // Объект для кодирования и отправки сообщений
	point    proto.Point   // Текущая конечная точка прямой
	len      float64       // Текущая длина кривой
	pointNum int           // Кол-во точек
}

// NewClient - конструктор клиента, принимает в качестве параметра
// объект TCP-соединения.
func NewClient(conn *net.TCPConn) *Client {
	return &Client{
		logger:   log.New(fmt.Sprintf("client %s", conn.RemoteAddr().String())),
		conn:     conn,
		enc:      json.NewEncoder(conn),
		point:    proto.Point{},
		len:      0,
		pointNum: 0,
	}
}

// serve - метод, в котором реализован цикл взаимодействия с клиентом.
// Подразумевается, что метод serve будет вызаваться в отдельной go-программе.
func (client *Client) serve() {
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

func squareDifference(coord1 string, coord2 string) (float64, error) {
	circleCenterX, err := strconv.ParseFloat(string(coord1), 64)
	circleContourX, err := strconv.ParseFloat(string(coord2), 64)
	if err != nil {
		return 0, err

	} else {
		delta := circleCenterX - circleContourX
		return delta * delta, nil
	}
}

func distBetweenPoints(pointA proto.Point, pointB proto.Point) (float64, error) {
	// delta_x ^ 2
	dX2, err := squareDifference(pointA.X, pointB.X)
	if err != nil {
		return 0, err
	}

	dY2, err := squareDifference(pointA.Y, pointB.Y)
	if err != nil {
		return 0, err
	}

	return math.Sqrt(dX2 + dY2), nil
}

func (client *Client) addPoint(newPoint proto.Point) bool {
	if client.pointNum == 0 {
		// if its the first point, just report ok
		client.point = newPoint
		client.pointNum += 1
		return true

	} else {
		// count distance between line's end point and new point
		currentEndPoint := client.point
		if increasedLen, err := distBetweenPoints(currentEndPoint, newPoint); err != nil {
			// if distance could not be counted, report false
			return false

		} else {
			// otherwise update line info
			client.len += increasedLen
			client.point = newPoint
			client.pointNum += 1
			return true
		}
	}
}

// handleRequest - метод обработки запроса от клиента. Он возвращает true,
// если клиент передал команду "quit" и хочет завершить общение.
func (client *Client) handleRequest(req *proto.Request) bool {
	switch req.Command {
	case proto.CMD_QUIT:
		client.respond(proto.RSP_STATUS_OK, nil)
		return true

	case proto.CMD_ADD:
		errorMsg := ""
		var point proto.Point
		// try to deserialize from json
		if err := json.Unmarshal(*req.Data, &point); err != nil {
			errorMsg = "malformed data field"

		} else {
			// check if point is valid and can be added
			if client.addPoint(point) {
				client.logger.Info("point has been added")
				client.respond(proto.RSP_STATUS_OK, nil)

			} else {
				errorMsg = "point cannot be added"
			}
		}
		if errorMsg != "" {
			client.logger.Error("addition failed", "reason", errorMsg)
			client.respond(proto.RSP_STATUS_FAILED, errorMsg)
		}

	case proto.CMD_COUNT:
		errorMsg := ""
		print(client.pointNum)
		if client.pointNum >= 2 {
			lineLen := strconv.FormatFloat(client.len, 'f', 6, 64)
			client.logger.Info("calculation succeeded", "length", lineLen)
			client.respond(proto.RSP_STATUS_RESULT, lineLen)

		} else {
			errorMsg = "could not count length. not enough points"
			client.logger.Error("calculation failed", "reason", errorMsg)
			client.respond(proto.RSP_STATUS_FAILED, errorMsg)
		}

	default:
		client.logger.Error("unknown command")
		client.respond("failed", "unknown command")
	}
	return false
}

// respond - вспомогательный метод для передачи ответа с указанным статусом
// и данными. Данные могут быть пустыми (data == nil).
func (client *Client) respond(status string, data interface{}) {
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
					go NewClient(conn).serve()
				}
			}
		}
	}
}
