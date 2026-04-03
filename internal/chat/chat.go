package chat

import (
	"time"

	"github.com/pocketbase/pocketbase/core"

	"pocketbase-server/internal/logging"
)

type Message struct {
	Id      string    `json:"id"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`

	Chat     string                 `json:"chat"`     // The ID of the chat group
	Sender   string                 `json:"sender"`   // User ID of the sender
	Text     string                 `json:"text"`     // The actual message
	Metadata map[string]interface{} `json:"metadata"` // For "read" status or extra info
}

func FromRecord(record *core.Record) *Message {
	// 1. Handle Metadata correctly using the built-in Unmarshal helper
	var metadata map[string]any
	if err := record.UnmarshalJSONField("metadata", &metadata); err != nil {
		logging.Error(err, "failed to UnmarshalJSONField")
	}

	return &Message{
		Id: record.Id, // Accessible via embedded BaseModel

		// 2. Use GetDateTime followed by .Time()
		Created: record.GetDateTime("created").Time(),
		Updated: record.GetDateTime("updated").Time(),

		// 3. Standard Getters
		Chat:   record.GetString("chat"),
		Sender: record.GetString("sender"),
		Text:   record.GetString("text"),

		Metadata: metadata,
	}
}

type Client interface {
	SaveMessage(msg Message) (*Message, error)
}

type chatClient struct {
	app core.App
}

func NewChatClient(app core.App) Client {
	return &chatClient{
		app: app,
	}
}

func (c *chatClient) SaveMessage(msg Message) (*Message, error) {
	collection, err := c.app.FindCollectionByNameOrId("messages")
	if err != nil {
		return nil, err
	}

	record := core.NewRecord(collection)

	// Set fields from our struct
	record.Set("chat", msg.Chat)
	record.Set("sender", msg.Sender)
	record.Set("text", msg.Text)
	record.Set("metadata", msg.Metadata)

	if err := c.app.Save(record); err != nil {
		return nil, err
	}

	// Return the hydrated struct (including the new ID and Timestamps)
	return FromRecord(record), nil
}
