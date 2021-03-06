package slack

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestSlackError(t *testing.T) {
	e := &Error{Status: "429 Too Many Requests", StatusCode: 429, RetryAfter: 30}
	fmt.Printf("%v\n", e)
	fmt.Printf("%s\n", e)
}

func postSlack(t *testing.T, sm *Message) {
	url := os.Getenv("SLACK_WEBHOOK")
	if len(url) < 1 {
		t.Skip("SLACK_WEBHOOK not set")
		return
	}
	err := Post(url, time.Second*5, sm)
	if err != nil {
		t.Error(err)
	}
}

// Test post slack message
func TestSlackPostText(t *testing.T) {
	sm := &Message{Text: "TestSlackPost"}
	postSlack(t, sm)
}

// Test post slack message with name
func TestSlackPostWithName(t *testing.T) {
	sm := &Message{Username: "go-test-name", Text: "TestSlackPostWithName"}
	postSlack(t, sm)
}

// Test post slack message with icon
func TestSlackPostWithIcon(t *testing.T) {
	sm := &Message{Username: "go-test-icon", IconEmoji: ":bug:", Text: "TestSlackPostWithIcon"}
	postSlack(t, sm)
}

// Test post slack message with attach
func TestSlackPostWithAttach(t *testing.T) {
	sm := &Message{Username: "go-test-attach", IconEmoji: ":fire:", Text: "TestSlackPostWithAttach"}
	sm.AddAttachment(&Attachment{Text: "attachment text"})
	postSlack(t, sm)
}
