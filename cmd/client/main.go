package main

import (
	"context"
	"encoding/json"
	"fmt"

	authclient "pocketbase-server/client"
)

func main() {

	ctx := context.Background()

	client := authclient.New(
		authclient.WithBaseURL("http://localhost:8090"),
	)

	// Register
	client.Register(ctx, &authclient.RegisterRequest{
		Email:           "user@example.com",
		Password:        "password123",
		PasswordConfirm: "password123",
	})

	// Login
	auth, err := client.Login(ctx, "user@example.com", "password123")
	if err != nil {
		panic(err)
	}

	b, err := json.MarshalIndent(auth, "", "   ")
	fmt.Println(string(b), err)

}
