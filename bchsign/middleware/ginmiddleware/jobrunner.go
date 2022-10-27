package ginmiddleware

import (
	"fmt"
	"github.com/bamzi/jobrunner"
	"github.com/gin-gonic/gin"
)

//添加job 路由状态页面显示
//fix: 要注意的是没有任务的时候调用是会直接异常的
//todo: 修复空队列任务异常
func AddJobRunner(routes *gin.Engine) {
	// Resource to return the JSON data
	routes.GET("/jobrunner/json", JobJson)

	// Load template file location relative to the current working directory
	routes.LoadHTMLGlob("middleware/ginmiddleware/jobrunner/views/Status.html")

	// Returns html page at given endpoint based on the loaded
	// template from above
	routes.GET("/jobrunner/html", JobHtml)
}

func JobJson(c *gin.Context) {
	// returns a map[string]interface{} that can be marshalled as JSON
	c.JSON(200, jobrunner.StatusJson())
}

func JobHtml(c *gin.Context) {
	// Returns the template data pre-parsed
	c.HTML(200, "Status.html", jobrunner.StatusPage())
}

//===================DEMO======================
func JobRunnerDemo() {
	jobrunner.Start() // optional: jobrunner.Start(pool int, concurrent int) (10, 1)
	jobrunner.Schedule("@every 5s", ReminderEmails{})
}

// Job Specific Functions
type ReminderEmails struct {
	// filtered
}

// ReminderEmails.Run() will get triggered automatically.
func (e ReminderEmails) Run() {
	// Queries the DB
	// Sends some email
	fmt.Printf("Every 5 sec send reminder emails \n")
}

//===================DEMO======================
