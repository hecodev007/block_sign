package controller

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/module/files"
	"custody-merchant-admin/router/web/handler"
	"fmt"
)

// UpLoadFiles 增加业务线
func UpLoadFiles(c *handler.Context) error {
	var (
		fileResponse = ""
		path         = ""
	)
	host := fmt.Sprintf("%s://%s", c.Scheme(), c.Request().Host)
	tag := c.FormValue("type")
	domains := c.FormValue("domains")
	userId := c.FormValue("userId")
	if domains != "" {
		path = "/" + domains
	} else {
		return global.WarnMsgError(fmt.Sprintf(global.DataNoHaveErr, "domains"))
	}
	if userId != "" {
		path += "/" + userId
	}
	switch tag {
	case "1": // 单图
		f, err := c.FormFile("file")
		if err != nil {
			return err
		}
		fileResponse, err = files.SingleFile(f, path, host)
		if err != nil {
			return err
		}
		break
	case "2": // 多图
		form, err := c.MultipartForm()
		if err != nil {
			return err
		}
		fileResponse, err = files.MultipleFile(form, path, host)
		if err != nil {
			return err
		}
		break
	default:
		f, err := c.FormFile("file")
		if err != nil {
			return err
		}
		fileResponse, err = files.SingleFile(f, path, host)
		if err != nil {
			return err
		}
		break
	}
	res := handler.NewSuccess()
	res.AddData("path", fileResponse)
	return res.ResultOk(c)

}
