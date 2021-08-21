package mailer

import (
	"fmt"

	"gopkg.in/gomail.v2"

	"secnds-server/env"
)

const (
	email_port     = "EMAIL_PORT"
	email_host     = "EMAIL_HOST"
	email_username = "EMAIL_USERNAME"
	email_password = "EMAIL_PASSWORD"
)

var emailPort int = env.GetInt(email_port)
var emailHost string = env.Get(email_host)
var emailUsername string = env.Get(email_username)
var emailPassword string = env.Get(email_password)

func SendMail(name string, to string, subject string) {

	gomailDialer := gomail.NewDialer(emailHost, emailPort, emailUsername, emailPassword)

	m := gomail.NewMessage()
	m.SetHeader("To", to)
	m.SetHeader("From", emailUsername)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", fmt.Sprintf("Hello %s !", name))

	if err := gomailDialer.DialAndSend(m); err != nil {
		panic(err)
	}
}
