package cbwebcommon

import (
	"github.com/codingbeard/cbweb"
	"github.com/fasthttp/router"
)

func (m *Module) SetRoutes(routes *router.Router) {
	routes.GET("/css/{filepath:*}", m.FileServer)
	routes.GET("/js/{filepath:*}", m.FileServer)
	routes.GET("/img/{filepath:*}", m.FileServer)
	routes.GET("/assets/{filepath:*}", m.FileServer)
	routes.GET("/manifest.json", m.FileServer)
	routes.GET("/404", cbweb.MiddlewareHandler{}.
		AddMiddleware(cbweb.HtmlMiddleware).
		SetFinal(m.FourOFourError).
		Handle,
	)
	routes.GET("/500", cbweb.MiddlewareHandler{}.
		AddMiddleware(cbweb.HtmlMiddleware).
		SetFinal(m.FiveHundredError).
		Handle,
	)
	routes.NotFound = cbweb.MiddlewareHandler{}.
		AddMiddleware(cbweb.HtmlMiddleware).
		SetFinal(m.FourOFourError).
		Handle
}
