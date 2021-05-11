// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag

package docs

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "termsOfService": "https://rrl360.com/index.html",
        "contact": {
            "name": "伍晓飞",
            "email": "wuxiaofei@rechaintech.com"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/pub/stock/aggre_info": {
            "get": {
                "description": "获取共识美股价格 苹果代码  AAPL  ,苹果代码 TSLA",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "default"
                ],
                "summary": "获取共识美股价格:",
                "operationId": "StockAggreHandler",
                "parameters": [
                    {
                        "type": "string",
                        "default": "AAPL",
                        "description": "美股代码",
                        "name": "code",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "default": 1620383144,
                        "description": "unix 秒数",
                        "name": "timestamp",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "stock info",
                        "schema": {
                            "$ref": "#/definitions/services.StockData"
                        },
                        "headers": {
                            "sign": {
                                "type": "string",
                                "description": "签名信息"
                            }
                        }
                    },
                    "500": {
                        "description": "失败时，有相应测试日志输出",
                        "schema": {
                            "$ref": "#/definitions/main.ApiErr"
                        }
                    }
                }
            }
        },
        "/pub/stock/info": {
            "get": {
                "description": "获取美股价格 苹果代码  AAPL  ,苹果代码 TSLA",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "default"
                ],
                "summary": "获取美股价格:",
                "operationId": "StockInfoHandler",
                "parameters": [
                    {
                        "type": "string",
                        "default": "AAPL",
                        "description": "美股代码",
                        "name": "code",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "default": 1620383144,
                        "description": "unix 秒数",
                        "name": "timestamp",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "stock info",
                        "schema": {
                            "$ref": "#/definitions/services.StockNode"
                        },
                        "headers": {
                            "sign": {
                                "type": "string",
                                "description": "签名信息"
                            }
                        }
                    },
                    "500": {
                        "description": "失败时，有相应测试日志输出",
                        "schema": {
                            "$ref": "#/definitions/main.ApiErr"
                        }
                    }
                }
            }
        },
        "/pub/stock/node_wallet": {
            "get": {
                "description": "当前节点钱包地址",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "default"
                ],
                "summary": "当前节点钱包地址:",
                "operationId": "NodeWalletAddreHandler",
                "responses": {
                    "200": {
                        "description": "stock info",
                        "schema": {
                            "type": "string"
                        },
                        "headers": {
                            "sign": {
                                "type": "string",
                                "description": "签名信息"
                            }
                        }
                    },
                    "500": {
                        "description": "失败时，有相应测试日志输出",
                        "schema": {
                            "$ref": "#/definitions/main.ApiErr"
                        }
                    }
                }
            }
        },
        "/pub/stock/node_wallets": {
            "get": {
                "description": "所有节点钱包地址列表",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "default"
                ],
                "summary": "所有节点钱包地址列表:",
                "operationId": "AllWalletAddreHandler",
                "responses": {
                    "200": {
                        "description": "stock info",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/main.NodeAddre"
                            }
                        }
                    },
                    "500": {
                        "description": "失败时，有相应测试日志输出",
                        "schema": {
                            "$ref": "#/definitions/main.ApiErr"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "main.ApiErr": {
            "type": "object",
            "properties": {
                "Error": {
                    "type": "string"
                }
            }
        },
        "main.NodeAddre": {
            "type": "object",
            "properties": {
                "node": {
                    "type": "string"
                },
                "walletAddre": {
                    "type": "string"
                }
            }
        },
        "services.StockData": {
            "type": "object",
            "properties": {
                "Price": {
                    "description": "平均价",
                    "type": "number"
                },
                "Timestamp": {
                    "description": "unix 秒数",
                    "type": "integer"
                },
                "code": {
                    "type": "string"
                },
                "sign": {
                    "description": "计算平均价格的节点的签名",
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "signs": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/services.StockNode"
                    }
                }
            }
        },
        "services.StockNode": {
            "type": "object",
            "properties": {
                "Timestamp": {
                    "description": "unix 秒数",
                    "type": "integer"
                },
                "code": {
                    "type": "string"
                },
                "node": {
                    "description": "节点名字",
                    "type": "string"
                },
                "price": {
                    "description": "新价",
                    "type": "number"
                },
                "sign": {
                    "type": "array",
                    "items": {
                        "type": "integer"
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
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "1.0",
	Host:        "",
	BasePath:    "/",
	Schemes:     []string{},
	Title:       "stock-info-api",
	Description: "stock-info-api接口文档.",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
