package helpers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"

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

type EmailTemplateData struct {
	AppName   string
	AppDomain string
	Subject   string
	Name      string
	Email     string
	// add more fields if templates need them
}

type EmailTemplateService struct {
	templates map[string]*template.Template
}

var (
	emailTemplateService *EmailTemplateService
	emailInitOnce        sync.Once
)

// InitEmailTemplates initializes and parses templates (idempotent)
func InitEmailTemplates() error {
	var initErr error
	emailInitOnce.Do(func() {
		ets := &EmailTemplateService{
			templates: make(map[string]*template.Template),
		}

		templatesDir := "templates/emails"
		layoutPath := filepath.Join(templatesDir, "layouts", "base.html")

		// Map template key -> file path
		templateFiles := map[string]string{
			"welcome": filepath.Join(templatesDir, "welcome-email.html"),
			// add other templates here
		}

		for name, path := range templateFiles {
			tmpl, err := template.ParseFiles(layoutPath, path)
			if err != nil {
				initErr = fmt.Errorf("failed to parse template %s (%s): %v", name, path, err)
				return
			}
			ets.templates[name] = tmpl
		}

		emailTemplateService = ets
	})
	return initErr
}

// RenderEmailTemplate renders a parsed template with given data
func RenderEmailTemplate(templateName string, data EmailTemplateData) (string, error) {
	if emailTemplateService == nil {
		return "", fmt.Errorf("email templates not initialized")
	}
	tmpl, ok := emailTemplateService.templates[templateName]
	if !ok {
		return "", fmt.Errorf("template %s not found", templateName)
	}

	var buf bytes.Buffer
	// in ParseFiles the layout filename is base.html (basename), so execute that
	if err := tmpl.ExecuteTemplate(&buf, "base.html", data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %v", templateName, err)
	}
	return buf.String(), nil
}

// SendTemplatedEmail sends an email rendered from a template (synchronous)
func SendTemplatedEmail(toEmail, subject, templateName string, data EmailTemplateData) error {
	body, err := RenderEmailTemplate(templateName, data)
	if err != nil {
		return fmt.Errorf("render template error: %v", err)
	}

	m := NewMessageFactory()

	fromEmail := os.Getenv("MAIL_FROM_ADDRESS")
	fromName := os.Getenv("MAIL_FROM_NAME")
	if fromName != "" {
		m.SetHeader("From", fmt.Sprintf("%s <%s>", fromName, fromEmail))
	} else {
		m.SetHeader("From", fromEmail)
	}

	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)

	// optional headers
	m.SetHeader("X-MJ-TrackOpen", "0")
	m.SetHeader("X-MJ-TrackClick", "0")

	m.SetBody("text/html", body)

	host := os.Getenv("MAIL_HOST")
	user := os.Getenv("MAIL_USERNAME")
	pass := os.Getenv("MAIL_PASSWORD")
	port, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
	d := NewDialerFactory(host, port, user, pass)

	// send
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send templated email: %v", err)
	}
	return nil
}

func sendPlainWelcomeEmailSync(recipientEmail, username string) error {
	m := NewMessageFactory()
	from := os.Getenv("MAIL_FROM_ADDRESS")
	if from == "" {
		return fmt.Errorf("MAIL_FROM_ADDRESS is not set")
	}
	m.SetHeader("From", from)
	m.SetHeader("To", recipientEmail)
	// Keep subject identical to original test expectation
	subject := "Welcome to MyGram!"
	m.SetHeader("Subject", subject)

	body := fmt.Sprintf("Hello %s,\n\nWelcome to MyGram! We're excited to have you.", username)
	m.SetBody("text/plain", body)

	host := os.Getenv("MAIL_HOST")
	user := os.Getenv("MAIL_USERNAME")
	pass := os.Getenv("MAIL_PASSWORD")
	port, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
	d := NewDialerFactory(host, port, user, pass)

	return d.DialAndSend(m)
}

func SendWelcomeEmailWithTemplate(recipientEmail, username string) error {
	// Attempt to init templates; if it fails -> fallback immediately to plain-text
	if emailTemplateService == nil {
		log.Printf("Email templates not initialized; falling back to plain-text welcome email.")
		return sendPlainWelcomeEmailSync(recipientEmail, username)
	}

	appDomain := os.Getenv("APP_DOMAIN")
	appName := os.Getenv("MAIL_FROM_NAME")
	if appName == "" {
		appName = "MyGram"
	}
	subject := "Welcome to MyGram!"

	data := EmailTemplateData{
		AppName:   appName,
		AppDomain: appDomain,
		Subject:   subject,
		Name:      username,
		Email:     recipientEmail,
	}

	// Try to send templated; on error fallback to plain-text
	if err := SendTemplatedEmail(recipientEmail, subject, "welcome", data); err != nil {
		log.Printf("Failed to send templated welcome email: %v. Falling back to plain-text.", err)
		return sendPlainWelcomeEmailSync(recipientEmail, username)
	}

	return nil
}

// SendWelcomeEmail sends an email to the new user in a goroutine
func SendWelcomeEmail(recipientEmail, username string) {
	go func() {
		if err := SendWelcomeEmailWithTemplate(recipientEmail, username); err != nil {
			log.Printf("Could not send welcome email to %s: %v", recipientEmail, err)
		} else {
			log.Printf("Successfully sent welcome email to %s", recipientEmail)
		}
	}()
}
