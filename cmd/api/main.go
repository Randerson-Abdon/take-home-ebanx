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

	memoryStore := store.NewMemoryStore()
	accountService := account.NewService(memoryStore)
	handler := httpapi.NewHandler(accountService)

	if os.Getenv("NGROK_AUTHTOKEN") != "" {
		serveErrors := make(chan error, 1)
		go func() {
			serveErrors <- http.Serve(listener, handler)
		}()
		log.Printf("server listening on http://localhost:%s", port)
		log.Printf("ngrok token loaded; starting tunnel")

		forwarder, err := tunnel.Start(context.Background(), port)
		if err != nil {
			log.Fatalf("start ngrok tunnel: %v", err)
		}
		defer func() {
			if err := forwarder.Close(); err != nil {
				log.Printf("close ngrok tunnel: %v", err)
			}
		}()
		log.Printf("public URL: %s", forwarder.URL())

		if err := <-serveErrors; err != nil {
			log.Fatal(err)
		}
		return
	}

	log.Printf("NGROK_AUTHTOKEN is not configured; ngrok tunnel disabled")
	log.Printf("server listening on http://localhost:%s", port)
	if err := http.Serve(listener, handler); err != nil {
		log.Fatal(err)
	}
}
