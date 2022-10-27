package handler

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"net/http"
)

type ResultData struct {
	Code int                    `json:"code"`
	Msg  string                 `json:"msg"`
	Data map[string]interface{} `json:"data,omitempty"`
}

func NewError(c *Context, msg string) error {
	return c.JSON(http.StatusOK, echo.Map{
		"code": 416,
		"msg":  msg,
		"data": nil,
	})
}

func NewCodeError(c *Context, code int, msg string) error {
	return c.JSON(http.StatusOK, echo.Map{
		"code": code,
		"msg":  msg,
		"data": nil,
	})
}

func NewResult(code int, msg string) *ResultData {
	return &ResultData{
		Code: code,
		Msg:  msg,
		Data: make(map[string]interface{}),
	}
}

func NewSuccess() *ResultData {
	return &ResultData{
		Code: 200,
		Msg:  "success",
		Data: make(map[string]interface{}),
	}
}

func (res *ResultData) ResultOk(c *Context) error {
	return c.JSON(http.StatusOK, res)
}

func (res *ResultData) AddData(key string, v interface{}) {
	res.Data[key] = v
}

func NewSuccessByStruct(s interface{}) *ResultData {
	data := make(map[string]interface{})
	b, _ := json.Marshal(s)
	json.Unmarshal(b, &data)
	return &ResultData{
		Code: 200,
		Msg:  "success",
		Data: data,
	}
}

func OutCodeError(c *Context, code int, msg string) error {
	return c.JSON(http.StatusOK, echo.Map{
		"code":    code,
		"message": msg,
		"data":    nil,
	})
}

func OutResult(c *Context, code int, msg string, data interface{}) error {
	return c.JSON(http.StatusOK, echo.Map{
		"code":    code,
		"message": msg,
		"data":    data,
	})
}
