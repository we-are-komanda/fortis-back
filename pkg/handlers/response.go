package handlers

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
)

type ResponseBody struct {
	// Статус ответа
	// Example: ok
	Status string  `json:"status"`
	Errors []Error `json:"errors,omitempty"`
}

type Error struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func ErrorHandler(ctx *fasthttp.RequestCtx, name, value string, responseBody *ResponseBody, status int) {
	responseBody.Status = "error"
	responseBody.Errors = append(responseBody.Errors, Error{
		Name:  name,
		Value: value,
	})

	responseJson, _ := json.Marshal(responseBody)
	ctx.SetStatusCode(status)
	ctx.SetBody(responseJson)
}
