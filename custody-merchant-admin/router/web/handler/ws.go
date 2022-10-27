package handler

func WsHandler(c *Context) error {

	c.Set("tmpl", "web/ws")
	c.Set("data", map[string]interface{}{
		"title": "Web Socket",
	})
	return nil
}
func HealthHandler(c *Context) error {
	res := NewSuccess()
	return res.ResultOk(c)
}
