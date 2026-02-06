package router

import (
	"github.com/pocketbase/pocketbase"
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

}
