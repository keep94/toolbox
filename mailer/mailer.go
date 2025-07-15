// Package mailer sends emails via gmail asynchronously.
package mailer

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"
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
	emailCh  chan *emailJob
	emailId  string
	password string
	pause    time.Duration
	done     chan struct{}
}

// New creates a new instance. emailId and password are the gmail
// sender address and password respectively.
func New(emailId, password string) *Mailer {
	result := &Mailer{
		emailCh:  make(chan *emailJob, 100),
		emailId:  emailId,
		password: password,
		pause:    time.Second,
		done:     make(chan struct{}),
	}
	go result.loop()
	return result
}

// Send sends one email asynchronously returning immediately. When it
// eventually sends the email, it reports any errors to stderr.
func (m *Mailer) Send(email Email) {
	responseCh := m.SendFuture(email)
	go func() {
		err := <-responseCh
		if err != nil {
			log.Println(err)
		}
	}()
}

// SendFuture sends one email asynchronously returning immediately. Caller
// must use returned channel to get the result of the send.
func (m *Mailer) SendFuture(email Email) <-chan error {
	emailJob := &emailJob{Email: email, Response: make(chan error, 1)}
	m.emailCh <- emailJob
	return emailJob.Response
}

// Shutdown shuts down this mailer. Shutdown waits to return until all
// pending emails have been sent. It is an error to call Send or SendFuture
// after calling Shutdown.
func (m *Mailer) Shutdown() {
	close(m.emailCh)
	<-m.done
}

func (m *Mailer) loop() {
	auth := smtp.PlainAuth("", m.emailId, m.password, "smtp.gmail.com")
	for emailJob := range m.emailCh {
		msgTemplate := "From: %s\n" +
			"To: %s\n" +
			"Subject: %s\n\n%s"
		msg := fmt.Sprintf(
			msgTemplate,
			m.emailId,
			emailJob.toAddresses(),
			emailJob.Subject,
			emailJob.Body)
		err := smtp.SendMail(
			"smtp.gmail.com:587", auth, m.emailId, emailJob.To, []byte(msg))
		emailJob.SetResponse(err)
		time.Sleep(m.pause)
	}
	close(m.done)
}

type emailJob struct {
	Email
	Response chan error
}

func (e *emailJob) SetResponse(err error) {
	e.Response <- err
	close(e.Response)
}
