package middleware

import "github.com/valyala/fasthttp"

type Middleware interface {
	Process(next fasthttp.RequestHandler) fasthttp.RequestHandler
}
