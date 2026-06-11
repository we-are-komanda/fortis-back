package middleware

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

const swaggerJsonLocalFile = "./var/swagger.json"
const swaggerJsonPath = "/_/swagger"
const swaggerUIPath = "/_/swaggerui"

type Swagger struct {
}

func NewSwagger() *Swagger {
	return &Swagger{}
}

//go:cover off
func (*Swagger) Process(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(c *fasthttp.RequestCtx) {
		if string(c.Path()) == swaggerJsonPath {
			c.SendFile(swaggerJsonLocalFile)
			return
		}
		if string(c.Path()) == swaggerUIPath {
			handler := fasthttpadaptor.NewFastHTTPHandler(middleware.SwaggerUI(middleware.SwaggerUIOpts{
				Path:    swaggerUIPath,
				SpecURL: swaggerJsonPath,
			}, nil))
			handler(c)
			return
		}
		next(c)
	}
}
