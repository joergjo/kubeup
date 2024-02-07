package webhook

import (
	"bytes"
	"html/template"

	"github.com/joergjo/go-samples/kubeup/internal/templates"
)

// Message represents a message.
type Message struct {
	Source    string
	PlainText string
	HTML      string
}

// MessageBuilder builds messages for a specific Azure Event Grid event type.
type MessageBuilder[T ContainerServiceEvent] struct {
	tmpl *template.Template
}

// NewMessageBuilder creates a new MessageBuilder with the given template filename.
func NewMessageBuilder[T ContainerServiceEvent](filename string) MessageBuilder[T] {
	tmpl := template.Must(template.ParseFS(templates.FS, filename))
	return MessageBuilder[T]{tmpl: tmpl}
}

// Build creates a new Message from the given event type and event source.
func (m MessageBuilder[T]) Build(e T, src string) (Message, error) {
	var msg Message
	var buf bytes.Buffer
	data := struct {
		Source string
		Event  T
	}{
		Source: src,
		Event:  e,
	}
	if err := m.tmpl.Execute(&buf, data); err != nil {
		return msg, err
	}
	msg.HTML = buf.String()
	msg.PlainText = e.String()
	msg.Source = src
	return msg, nil
}
