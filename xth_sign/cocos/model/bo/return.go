package bo

/*code 0 为返回正常
其它值为返回错误
*/
type HttpReturn struct {
	Code  int         `json:"code"`
	Body  string      `json:"body"`
	Error interface{} `json:"error"`
}
