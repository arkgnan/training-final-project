package helpers

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-mail/mail/v2"
	"github.com/stretchr/testify/assert"
)

// fakeDialer implements Dialer and sends serialized message text to a channel.
type fakeDialer struct {
	ch chan string
}

// Use variadic param to match the Dialer interface wrapping mail.Dialer.
func (f *fakeDialer) DialAndSend(ms ...*mail.Message) error {
	if len(ms) == 0 {
		return nil
	}

	var buf bytes.Buffer
	// Serialize the full MIME message into buf. mail.Message typically exposes WriteTo.
	// If your version has WriteTo signature (io.Writer) error or (int64, error), this will work.
	// We attempt the common signature below.
	if _, err := ms[0].WriteTo(&buf); err == nil {
		f.ch <- buf.String()
	} else {
		// If WriteTo has a different signature returning (int64, error), try capturing that as well.
		// Attempt the alternate call via type assertion to interface with different WriteTo signature.
		type writeToInt64 interface {
			WriteTo(w *bytes.Buffer) (int64, error)
		}
		if a, ok := any(ms[0]).(writeToInt64); ok {
			if _, err2 := a.WriteTo(&buf); err2 == nil {
				f.ch <- buf.String()
			}
		}
	}
	return nil
}

func TestSendWelcomeEmail_SerializesMessageAndContainsBody(t *testing.T) {
	t.Parallel()

	// Set env vars used by SendWelcomeEmail
	os.Setenv("MAIL_FROM_ADDRESS", "no-reply@mygram.test")
	os.Setenv("MAIL_HOST", "smtp.test")
	os.Setenv("MAIL_USERNAME", "user")
	os.Setenv("MAIL_PASSWORD", "pass")
	os.Setenv("MAIL_PORT", "25")

	ch := make(chan string, 1)

	// Save original factories and restore after test
	origDialerFactory := NewDialerFactory
	origMessageFactory := NewMessageFactory
	defer func() {
		NewDialerFactory = origDialerFactory
		NewMessageFactory = origMessageFactory
	}()

	// Inject fake dialer returning serialized message on channel
	NewDialerFactory = func(host string, port int, user, pass string) Dialer {
		return &fakeDialer{ch: ch}
	}
	// Use a real mail.Message in production; NewMessageFactory returns *mail.Message
	NewMessageFactory = func() *mail.Message { return mail.NewMessage() }

	recipient := "alice@example.test"
	username := "alice"

	// Call the async function (it runs the send in a goroutine)
	SendWelcomeEmail(recipient, username)

	select {
	case serialized := <-ch:
		// The serialized message contains headers and body. Check headers exist.
		assert.True(t, strings.Contains(serialized, "To: "+recipient), "serialized message should contain To header")
		assert.True(t, strings.Contains(serialized, "From: no-reply@mygram.test"), "serialized message should contain From header")
		assert.True(t, strings.Contains(serialized, "Subject: Welcome to MyGram!"), "serialized message should contain Subject header")

		// Check body includes the username (body will be somewhere in the serialized MIME)
		assert.True(t, strings.Contains(serialized, username), "serialized message should contain username in body")
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for fake dialer to be invoked")
	}
}
