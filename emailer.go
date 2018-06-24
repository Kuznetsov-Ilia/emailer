package emailer

import (
	"crypto/tls"
	"errors"
	"log"
	"os"
	"time"

	"gopkg.in/gomail.v2"
)

var (
	ch       = make(chan *gomail.Message)
	host     = os.Getenv("SMTP_HOST")
	port     = 465
	username = os.Getenv("SMTP_USER")
	password = os.Getenv("SMTP_PASS")
)

func init() {
	go func() {
		log.Print("host: " + host)
		log.Print("username: " + username)
		d := gomail.NewDialer(host, port, username, password)
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

		var s gomail.SendCloser
		var err error
		open := false
		for {
			select {
			case m, ok := <-ch:
				if !ok {
					return
				}
				if !open {
					if s, err = d.Dial(); err != nil {
						panic(err)
					}
					open = true
				}
				if err := gomail.Send(s, m); err != nil {
					log.Print(err)
				}
				// Close the connection to the SMTP server if no email was sent in
				// the last 30 seconds.
			case <-time.After(30 * time.Second):
				if open {
					if err := s.Close(); err != nil {
						panic(err)
					}
					open = false
				}
			}
		}
	}()
}

func NewEmailMessage(from string, to string, subject string, body string) *gomail.Message {
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	return m
}

func SendChan(m *gomail.Message) {
	ch <- m
}

func Send(m *gomail.Message) error {
	if m == nil {
		return errors.New("Message can not be nil")
	}
	d := gomail.NewDialer(host, port, username, password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	return d.DialAndSend(m)
}

func Shutdown() {
	close(ch)
}
