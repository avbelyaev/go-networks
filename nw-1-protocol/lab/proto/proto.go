package proto

import "encoding/json"

/*
Var 7
Протокол вычисления площадей кругов на плоскости,
заданных центром и точкой на окружности.
 */

const CMD_QUIT = "quit"
const CMD_COUNT = "count"
const RSP_STATUS_OK = "ok"
const RSP_STATUS_FAIL = "fail"
const RSP_STATUS_SUCCESS = "success"


type Request struct {
	// Поле Command может принимать значения:
	// quit - прощание с сервером (после этого сервер рвёт соединение);
	// count - добавление точки к кривой;
	Command string `json:"command"`

	// Если Command == "count", то в поле Data лежит Circle
	// Если Command == "quit", то поле Data пустое
	Data *json.RawMessage `json:"data"`
}

type Response struct {
	// Поле Status может принимать три значения:
	// * "ok" - успешное выполнение команды "quit";
	// * "fail" - в процессе выполнения команды произошла ошибка;
	// * "success" - успешное выполнение команды "count";
	Status string `json:"status"`

	// Если Status == "fail", то в поле Data находится сообщение об ошибке.
	// Если Status == "ok", то поле Data пустое
	// Если Status == "success", то в поле Data лежит знаковый float
	Data *json.RawMessage `json:"data"`
}

type Point struct {
	// коорд Х, знаковый float
	CoordX string `json:"x,Number"`

	// коорд Y, знаковый float
	CoordY string `json:"y,Number"`
}

type Circle struct {
	// центр окружности (Point)
	Center Point `json:"center"`

	// точка на окржуности (Point)
	Contour Point `json:"contour"`
}