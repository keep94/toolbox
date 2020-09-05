// Package mailer sends emails via gmail asynchronously.
package mailer

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

// Email represents a single email.
type Email struct {
	To      []string
	Subject string
	Body    string
}

func (e *Email) toAddresses() string {
	return strings.Join(e.To, ", ")
}

// Mailer sends emails asynchronously via gmail.
type Mailer struct {
	emailCh  chan Email
	emailId  string
	password string
}

// New creates a new instance. emailId and password are the gmail
// sender address and password respectively.
func New(emailId, password string) *Mailer {
	result := &Mailer{
		emailCh:  make(chan Email, 100),
		emailId:  emailId,
		password: password,
	}
	go result.loop()
	return result
}

// Send sends one email asynchronously returning immediately. When it
// eventually sends the email, it reports any errors to stderr.
func (m *Mailer) Send(email Email) {
	m.emailCh <- email
}

func (m *Mailer) loop() {
	auth := smtp.PlainAuth("", m.emailId, m.password, "smtp.gmail.com")
	for {
		email := <-m.emailCh
		msgTemplate := "From: %s\n" +
			"To: %s\n" +
			"Subject: %s\n\n%s"
		msg := fmt.Sprintf(
			msgTemplate,
			m.emailId,
			email.toAddresses(),
			email.Subject,
			email.Body)
		err := smtp.SendMail(
			"smtp.gmail.com:587", auth, m.emailId, email.To, []byte(msg))
		if err != nil {
			log.Println(err)
		}
	}
}
