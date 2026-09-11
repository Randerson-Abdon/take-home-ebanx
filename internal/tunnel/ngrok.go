// Package tunnel exposes the local HTTP server through ngrok.
package tunnel

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"golang.ngrok.com/ngrok/v2"
)

// Forwarder represents an active public tunnel.
type Forwarder struct {
	endpoint ngrok.EndpointForwarder
	agent    ngrok.Agent
}

// Start creates an ngrok forwarder for the local application port.
func Start(ctx context.Context, port string) (*Forwarder, error) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})).With("component", "ngrok")

	agent, err := ngrok.NewAgent(
		ngrok.WithAuthtoken(os.Getenv("NGROK_AUTHTOKEN")),
		ngrok.WithLogger(logger),
		ngrok.WithEventHandler(func(event ngrok.Event) {
			switch event := event.(type) {
			case *ngrok.EventAgentConnectSucceeded:
				logger.Info("agent connected and authenticated")
			case *ngrok.EventAgentDisconnected:
				logger.Warn("agent disconnected", "error", event.Error)
			}
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create ngrok agent: %w", err)
	}

	logger.Info("connecting agent to ngrok")
	if err := agent.Connect(ctx); err != nil {
		return nil, fmt.Errorf("connect ngrok agent: %w", err)
	}

	logger.Info("creating public endpoint", "upstream", "http://localhost:"+port)
	endpoint, err := agent.Forward(
		ctx,
		ngrok.WithUpstream("http://localhost:"+port),
	)
	if err != nil {
		_ = agent.Disconnect()
		return nil, err
	}

	return &Forwarder{endpoint: endpoint, agent: agent}, nil
}

// URL returns the public tunnel URL.
func (f *Forwarder) URL() string {
	return f.endpoint.URL().String()
}

// Close stops the public tunnel.
func (f *Forwarder) Close() error {
	if err := f.endpoint.Close(); err != nil {
		return err
	}
	return f.agent.Disconnect()
}
