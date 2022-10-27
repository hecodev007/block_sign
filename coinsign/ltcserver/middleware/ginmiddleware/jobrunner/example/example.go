package main

import (
	"fmt"
	"github.com/bamzi/jobrunner"
	"github.com/gin-gonic/gin"
)

// Example of GIN micro framework
func main() {

	jobrunner.Start() // optional: jobrunner.Start(pool int, concurrent int) (10, 1)
	jobrunner.Schedule("@every 5s", ReminderEmails{})

	routes := gin.Default()

	// Resource to return the JSON data
	routes.GET("/jobrunner/json", JobJson)

	// Load template file location relative to the current working directory
	routes.LoadHTMLGlob("../github.com/bamzi/jobrunner/views/Status.html")
	//routes.LoadHTMLFiles("./views/Status.html")

	// Returns html page at given endpoint based on the loaded
	// template from above
	routes.GET("/jobrunner/html", JobHtml)

	routes.Run(":8081")
}

func JobJson(c *gin.Context) {
	// returns a map[string]interface{} that can be marshalled as JSON
	c.JSON(200, jobrunner.StatusJson())
}

func JobHtml(c *gin.Context) {
	// Returns the template data pre-parsed
	c.HTML(200, "Status.html", jobrunner.StatusPage())

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
