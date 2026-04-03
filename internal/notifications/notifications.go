package notifications

import (
	"github.com/pocketbase/pocketbase/core"
)

// Notification types
const (
	TypeInfo    = "info"
	TypeWarning = "warning"
	TypeError   = "error"
	TypeInvite  = "invite"
	TypeSystem  = "system"
)

// Notification matches your PocketBase collection schema
type Notification struct {
	Id           string         `json:"id"`
	Created      string         `json:"created"`
	Updated      string         `json:"updated"`
	Recipient    string         `json:"recipient"`
	Owner        string         `json:"owner"`
	Organization string         `json:"organization"`
	Type         string         `json:"type"`
	Title        string         `json:"title"`
	Message      string         `json:"message"`
	Dismissed    bool           `json:"dismissed"`
	Data         map[string]any `json:"data"`
}

// FromRecord populates the struct from a PocketBase record
func (n *Notification) FromRecord(record *core.Record) {
	n.Id = record.Id

	// GetDateTime returns a types.DateTime object
	n.Created = record.GetDateTime("created").String()
	n.Updated = record.GetDateTime("updated").String()

	n.Recipient = record.GetString("recipient")
	n.Owner = record.GetString("owner")
	n.Organization = record.GetString("organization")
	n.Type = record.GetString("type")
	n.Title = record.GetString("title")
	n.Message = record.GetString("message")
	n.Dismissed = record.GetBool("dismissed")

	if val := record.Get("data"); val != nil {
		if m, ok := val.(map[string]any); ok {
			n.Data = m
		}
	}
}

type NotificationOpts struct {
	Recipient    string
	Owner        string
	Organization string
	Type         string
	Title        string
	Message      string
	Data         map[string]any
}

// NotificationClient defines the interface for sending notifications manually.
type NotificationClient interface {
	Send(opts NotificationOpts) (*Notification, error)
}

type notificationService struct {
	app core.App
}

// NewClient initializes a new notification service.
func NewClient(app core.App) NotificationClient {
	return &notificationService{app: app}
}

// Send creates a record and returns the typed Notification struct
func (s *notificationService) Send(opts NotificationOpts) (*Notification, error) {
	col, err := s.app.FindCollectionByNameOrId("notifications")
	if err != nil {
		return nil, err
	}

	record := core.NewRecord(col)
	record.Set("recipient", opts.Recipient)
	record.Set("type", opts.Type)
	record.Set("title", opts.Title)
	record.Set("message", opts.Message)
	record.Set("owner", opts.Owner)
	record.Set("organization", opts.Organization)
	record.Set("data", opts.Data)
	record.Set("dismissed", false)

	if err := s.app.Save(record); err != nil {
		return nil, err
	}

	n := &Notification{}
	n.FromRecord(record)
	return n, nil
}
