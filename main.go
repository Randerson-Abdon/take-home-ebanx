package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"golang.ngrok.com/ngrok/v2"
)

const defaultPort = "8085"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	startServer(port, newHandler())

	if os.Getenv("NGROK_AUTHTOKEN") != "" {
		connectNgrok(port)
	}

	select {}
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "ok")
	})

	return mux
}

func startServer(port string, handler http.Handler) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := http.Serve(listener, handler); err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("server listening on http://localhost:%s", port)
}

func connectNgrok(port string) {
	forwarder, err := ngrok.Forward(
		context.Background(),
		ngrok.WithUpstream("http://localhost:"+port),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("public URL: %s", forwarder.URL())
}
