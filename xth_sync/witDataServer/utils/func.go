package utils

import "encoding/json"

func String(d interface{}) string{
	str,_:=json.Marshal(d)
	return string(str)
}