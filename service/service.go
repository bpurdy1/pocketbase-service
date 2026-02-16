package service

import (
	"github.com/pocketbase/pocketbase"
)

type Service struct {
	app *pocketbase.PocketBase
}

func NewService(app *pocketbase.PocketBase) *Service {
	return &Service{
		app: app,
	}
}
