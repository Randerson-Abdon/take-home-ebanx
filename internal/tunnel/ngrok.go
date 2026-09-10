// Package tunnel exposes the local HTTP server through ngrok.
package tunnel

import (
	"context"

	"golang.ngrok.com/ngrok/v2"
)

// Forwarder represents an active public tunnel.
type Forwarder struct {
	endpoint ngrok.EndpointForwarder
}

// Start creates an ngrok forwarder for the local application port.
func Start(ctx context.Context, port string) (*Forwarder, error) {
	endpoint, err := ngrok.Forward(
		ctx,
		ngrok.WithUpstream("http://localhost:"+port),
	)
	if err != nil {
		return nil, err
	}

	return &Forwarder{endpoint: endpoint}, nil
}

// URL returns the public tunnel URL.
func (f *Forwarder) URL() string {
	return f.endpoint.URL().String()
}

// Close stops the public tunnel.
func (f *Forwarder) Close() error {
	return f.endpoint.Close()
}
