package webhooks

import "net/http"

type Event struct {
	Provider string
	Type     string
	Payload  []byte
	Headers  http.Header
}

type Handler func(event Event) error

type Provider interface {
	Name() string
	Verify(r *http.Request) error
	Parse(r *http.Request) (*Event, error)
}
