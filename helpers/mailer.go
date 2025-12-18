package helpers

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-mail/mail/v2"
)

// Dialer is an interface used by the mailer so we can inject fakes in tests.
type Dialer interface {
	DialAndSend(...*mail.Message) error
}

// NewDialerFactory is a function that returns a Dialer. By default it wraps mail.NewDialer.
var NewDialerFactory = func(host string, port int, user, pass string) Dialer {
	return mail.NewDialer(host, port, user, pass)
}

// NewMessageFactory returns a pointer to mail.Message. Can be overridden in tests.
var NewMessageFactory = func() *mail.Message {
	return mail.NewMessage()
}

// SendWelcomeEmail sends an email to the new user in a goroutine
func SendWelcomeEmail(recipientEmail, username string) {
	// Run email sending in a goroutine
	go func() {
		m := mail.NewMessage()
		m.SetHeader("From", os.Getenv("MAIL_FROM_ADDRESS"))
		m.SetHeader("To", recipientEmail)
		m.SetHeader("Subject", "Welcome to MyGram!")

		body := fmt.Sprintf("Hello %s,\n\nWelcome to MyGram! We're excited to have you.", username)
		m.SetBody("text/plain", body)

		host := os.Getenv("MAIL_HOST")
		user := os.Getenv("MAIL_USERNAME")
		pass := os.Getenv("MAIL_PASSWORD")
		port, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
		d := NewDialerFactory(host, port, user, pass)

		// This log helps debug the goroutine process
		if err := d.DialAndSend(m); err != nil {
			log.Printf("Could not send welcome email to %s: %v", recipientEmail, err)
		} else {
			log.Printf("Successfully sent welcome email to %s", recipientEmail)
		}
	}()
}
