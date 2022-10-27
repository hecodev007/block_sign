package api

import (
	"encoding/json"
	"github.com/bamzi/jobrunner"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"io/ioutil"
)

//定时任务页面可视化基本配置

func JobJson(c *gin.Context) {
	// returns a map[string]interface{} that can be marshalled as JSON
	c.JSON(200, jobrunner.StatusJson())
}

func JobHtml(c *gin.Context) {
	// Returns the template data pre-parsed
	c.HTML(200, "status.html", jobrunner.StatusPage())
}

//移除指定id的定时任务
func JobRemove(c *gin.Context) {
	var data map[string]interface{}
	body, _ := ioutil.ReadAll(c.Request.Body)
	json.Unmarshal(body, &data)
	cornId := cron.EntryID(data["id"].(float64))
	jobrunner.Remove(cornId)
	c.JSON(200, "success")
}
