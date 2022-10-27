package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

type requestFormat struct {
	Path        string      `json:"path"`
	Method      string      `json:"method"`
	ContentType string      `json:"contentType"`
	Body        interface{} `json:"body"`
}

//生产环境，不建议打开，测试环境使用,只输出在控制台,中控服务
//======================中间件:参数打印======================
func GinPrintParams() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rf := &requestFormat{
			Path:        ctx.Request.RequestURI,
			Method:      ctx.Request.Method,
			ContentType: ctx.ContentType(),
			//Body:        json.RawMessage(bodyByte),
		}
		bodyByte, _ := ioutil.ReadAll(ctx.Request.Body)
		fmt.Println("request body:", string(bodyByte))
		//if err == nil {
		//	if "application/json" == ctx.ContentType() {
		//		rf.Body = json.RawMessage(bodyByte)
		//	}
		//}
		rf.Body = json.RawMessage(bodyByte)
		if len(json.RawMessage(bodyByte)) == 0 {
			rf.Body = string(bodyByte)
		}
		data, _ := json.Marshal(rf)
		var str bytes.Buffer
		err := json.Indent(&str, data, "", "    ")
		if err == nil {
			//控制台打印
			fmt.Println(str.String())
		}
		ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyByte))
		ctx.Next()
	}
}

//======================中间件:参数打印======================
