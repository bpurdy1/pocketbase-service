package router

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type Router struct {
	app *pocketbase.PocketBase
}

func NewRouter(app *pocketbase.PocketBase) *Router {
	return &Router{
		app: app,
	}
}

func (r *Router) RegisterBind() {
	r.app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		r.bindInviteRoutes(e)
		r.bindAdminRoutes(e)
		return e.Next()
	})
}
