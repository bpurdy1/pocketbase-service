package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"pocketbase-server/server"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
}

func main() {
	srv, err := server.New()
	if err != nil {
		panic(err)
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down PocketBase...")
		os.Exit(0)
	}()

	log.Println("Starting PocketBase server...")
	if err := srv.Start(); err != nil {
		log.Fatalf("PocketBase start error: %v", err)
	}
}
