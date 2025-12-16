package event

import (
	"errors"
	"testing"
)

func TestPublisherErrors(t *testing.T) {
	err1 := errors.New("publisher 1 failed")
	err2 := errors.New("publisher 2 failed")

	p := &Publisher{
		publisherFns: []PublisherFunc{
			func(Message) error { return err1 },
			func(Message) error { return err2 },
		},
	}

	err := p.Publish(Message{Source: "src", PlainText: "plain", HTML: "<p>info</p>"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, err1) {
		t.Fatalf("expected joined error to match err1")
	}
	if !errors.Is(err, err2) {
		t.Fatalf("expected joined error to match err2")
	}
}

func TestPublisherSuccess(t *testing.T) {
	p := &Publisher{
		publisherFns: []PublisherFunc{
			func(Message) error { return nil },
			func(Message) error { return nil },
		},
	}

	if err := p.Publish(Message{}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
