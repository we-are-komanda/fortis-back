package ui

import (
	"github.com/flosch/pongo2"
	"github.com/valyala/fasthttp"
	"log/slog"
)

type WebController struct {
}

func NewWebController() *WebController {
	return &WebController{}
}

func (controller *WebController) Main(ctx *fasthttp.RequestCtx) {
	data := pongo2.Context{
		"data": pongo2.Context{
			"title":   "Card Title",
			"context": "This is some card content.",
		},
	}

	renderTemplate(ctx, "html/templates/base.html", data)
}

func renderTemplate(ctx *fasthttp.RequestCtx, tpl string, data pongo2.Context) {
	template := pongo2.Must(pongo2.FromFile(tpl))

	out, err := template.ExecuteBytes(data)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		slog.Error(err.Error())
		return
	}
	ctx.SetBody(out)
}
