package proto

import "encoding/json"

/*
Var 4
Протокол вычисления длины ломаной линии на
плоскости, заданной последовательностью точек.
 */

const CMD_QUIT = "quit"
const CMD_ADD = "add"
const CMD_COUNT = "count"
const RSP_STATUS_OK = "ok"
const RSP_STATUS_FAILED = "failed"
const RSP_STATUS_RESULT = "result"


type Request struct {
	// Поле Command может принимать значения:
	// quit - прощание с сервером (после этого сервер рвёт соединение);
	// add - добавление точки к кривой;
	// count - подсчет длины кривой
	Command string `json:"command"`

	// Если Command == "count", то поле Data пустое
	// Если Command == "quit", то поле Data пустое
	// Если Command == "add", то поле Data содержит Point
	Data *json.RawMessage `json:"data"`
}

type Response struct {
	// Поле Status может принимать три значения:
	// * "ok" - успешное выполнение команды "quit" или "add";
	// * "failed" - в процессе выполнения команды произошла ошибка;
	// * "result" - успешное выполнение команды "count";
	Status string `json:"status"`

	// Если Status == "ok", то поле Data пустое
	// Если Status == "failed", то в поле Data находится сообщение об ошибке.
	// Если Status == "result", то в поле Data лежит длина кривой ломаной
	Data *json.RawMessage `json:"data"`
}

type Point struct {
	X string `json:"x,Number"`

	Y string `json:"y,Number"`
}