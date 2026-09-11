package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/Randerson-Abdon/take-home-ebanx/internal/account"
	"github.com/Randerson-Abdon/take-home-ebanx/internal/config"
	"github.com/Randerson-Abdon/take-home-ebanx/internal/httpapi"
	"github.com/Randerson-Abdon/take-home-ebanx/internal/store"
	"github.com/Randerson-Abdon/take-home-ebanx/internal/tunnel"
)

const defaultPort = "8085"

func main() {
	if err := config.LoadEnv(".env"); err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("server listening on http://localhost:%s", port)

	if os.Getenv("NGROK_AUTHTOKEN") != "" {
		forwarder, err := tunnel.Start(context.Background(), port)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := forwarder.Close(); err != nil {
				log.Printf("close ngrok tunnel: %v", err)
			}
		}()
		log.Printf("public URL: %s", forwarder.URL())
	}

	memoryStore := store.NewMemoryStore()
	accountService := account.NewService(memoryStore)

	if err := http.Serve(listener, httpapi.NewHandler(accountService)); err != nil {
		log.Fatal(err)
	}
}
