//go:build dev

package main

import (
	"fmt"

	"github.com/joho/godotenv"

	"pocketbase-server/internal/logging"
)

func init() {
	logging.Warn("loading .env file")

	if err := godotenv.Load(); err != nil {
		fmt.Println(err)
	}
}
