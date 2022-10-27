// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag at
// 2019-04-25 15:46:40.818502 +0800 CST m=+0.118716811

package docs

import (
	"bytes"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "license": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/v1/addr": {
            "post": {
                "produces": [
                    "application/json"
                ],
                "summary": "生成地址",
                "parameters": [
                    {
                        "description": "data",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "{\"code\":0,\"message\":\"ok\",\"data\":{},\"hash\":\"8978608dad8f150ea142e1c076f6564e\"}",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/v1/create": {
            "post": {
                "produces": [
                    "application/json"
                ],
                "summary": "创建签名模板",
                "parameters": [
                    {
                        "description": "sign tpl",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "{\"code\":0,\"message\":\"ok\",\"data\":{},\"hash\":\"8978608dad8f150ea142e1c076f6564e\"}",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/v1/push": {
            "post": {
                "produces": [
                    "application/json"
                ],
                "summary": "广播交易",
                "parameters": [
                    {
                        "description": "push hex",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "{\"code\":0,\"message\":\"ok\",\"data\":{},\"hash\":\"8978608dad8f150ea142e1c076f6564e\"}",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/v1/sign": {
            "post": {
                "produces": [
                    "application/json"
                ],
                "summary": "模板签名",
                "parameters": [
                    {
                        "description": "sign json",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "{\"code\":0,\"message\":\"ok\",\"data\":{},\"hash\":\"8978608dad8f150ea142e1c076f6564e\"}",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo swaggerInfo

type s struct{}

func (s *s) ReadDoc() string {
	t, err := template.New("swagger_info").Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, SwaggerInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
