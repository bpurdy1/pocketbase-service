package service

import (
	"github.com/pocketbase/pocketbase"

	"pocketbase-server/internal/collections"
)

type Service struct {
	app *pocketbase.PocketBase
}

func NewService(app *pocketbase.PocketBase) *Service {
	collections.NewCollectionManager(app)
	return &Service{
		app: app,
	}
}
