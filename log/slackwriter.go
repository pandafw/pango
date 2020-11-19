package log

import (
	"fmt"
	"os"
	"time"

	"github.com/pandafw/pango/net/slack"
)

// SlackWriter implements log Writer Interface and send log message to slack.
type SlackWriter struct {
	Webhook  string
	Username string
	Channel  string
	Timeout  time.Duration
	Subfmt   Formatter // subject formatter
	Logfmt   Formatter // log formatter
	Logfil   Filter    // log filter
}

// SetSubject set a subject formatter
func (sw *SlackWriter) SetSubject(format string) {
	sw.Subfmt = NewTextFormatter(format)
}

// SetFormat set a log formatter
func (sw *SlackWriter) SetFormat(format string) {
	sw.Logfmt = NewTextFormatter(format)
}

// SetTimeout set timeout
func (sw *SlackWriter) SetTimeout(timeout string) error {
	tmo, err := time.ParseDuration(timeout)
	if err != nil {
		return fmt.Errorf("SlackWriter - Invalid timeout: %v", err)
	}
	sw.Timeout = tmo
	return nil
}

// Write send log message to slack
func (sw *SlackWriter) Write(le *Event) {
	if sw.Logfil != nil && sw.Logfil.Reject(le) {
		return
	}
	if sw.Subfmt == nil {
		sw.Subfmt = TextFmtSubject
	}
	if sw.Logfmt == nil {
		sw.Logfmt = le.Logger.GetFormatter()
	}

	sm := &slack.Message{}
	sm.IconEmoji = getIconEmoji(le.Level)
	sm.Channel = sw.Channel
	sm.Username = sw.Username
	sm.Text = sw.Subfmt.Format(le)

	sa := &slack.Attachment{Text: sw.Logfmt.Format(le)}
	sm.AddAttachment(sa)

	err := slack.Post(sw.Webhook, sw.Timeout, sm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "SlackWriter(%q) - Post(): %v\n", sw.Webhook, err)
	}
}

// Flush implementing method. empty.
func (sw *SlackWriter) Flush() {
}

// Close implementing method. empty.
func (sw *SlackWriter) Close() {
}

func getIconEmoji(lvl int) string {
	switch lvl {
	case LevelFatal:
		return ":boom:"
	case LevelError:
		return ":fire:"
	case LevelWarn:
		return ":warning:"
	case LevelInfo:
		return ":droplet:"
	case LevelDebug:
		return ":bug:"
	case LevelTrace:
		return ":ant:"
	}
	return ":ghost:"
}
