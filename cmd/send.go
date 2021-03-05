package cmd

import (
	"github.com/jordan-wright/email"
	"github.com/sirupsen/logrus"
	"net/smtp"
	"os"
)

func sendAndRemove(emails []string) {
	for _, eml := range emails {
		log := logrus.WithField("email", eml)

		log.Debugf("email send")
		err := sendEmail(eml)
		if err != nil {
			log.WithError(err).Errorf("send failed")
			continue
		}
		log.Info("email sent")

		_ = os.Remove(eml)
	}
}

func sendEmail(eml string) (err error) {
	file, err := os.Open(eml)
	if err != nil {
		return
	}
	defer file.Close()

	e, err := email.NewEmailFromReader(file)
	if err != nil {
		return
	}

	e.From = send.GetFrom()
	e.To = []string{send.GetTo()}

	return e.Send(send.Addr, send.GetAuth())
}

type Send struct {
	From string
	To   string
	Addr string
	Auth struct {
		Identity string
		Username string
		Password string
		Host     string
	}
	auth smtp.Auth
}

func (s *Send) GetFrom() string {
	if *argFrom != "" {
		return *argFrom
	}
	return s.From
}
func (s *Send) GetTo() string {
	if *argTo != "" {
		return *argTo
	}
	return s.To
}
func (s *Send) GetAuth() smtp.Auth {
	if s.auth == nil {
		s.auth = smtp.PlainAuth(send.Auth.Identity, send.Auth.Username, send.Auth.Password, send.Auth.Host)
	}
	return s.auth
}
