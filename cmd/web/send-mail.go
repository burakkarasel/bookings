package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/burakkarasel/bookings/internal/models"
	mail "github.com/xhit/go-simple-mail/v2"
)

// listenForMail runs asynchronously while our program runs, if mail channel receives a mail it sends it
func listenForMail() {
	go func() {
		for {
			msg := <-app.MailChan
			sendMsg(msg)
		}
	}()
}

// sendMsg creates a server, and connects the client to this server then uses a mail template, and fill it with the information
// that received from reservation, and recevives the details of the mail from mail channel
func sendMsg(m models.MailData) {
	server := mail.NewSMTPClient()
	server.Host = "localhost"
	server.Port = 1025
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	client, err := server.Connect()

	if err != nil {
		errorLog.Println(err)
	}

	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)
	if m.Template == "" {
		email.SetBody(mail.TextHTML, m.Content)
	} else {
		data, err := ioutil.ReadFile(fmt.Sprintf("./email-templates/%s", m.Template))

		if err != nil {
			app.ErrorLog.Println(err)
		}

		mailTemplate := string(data)
		msgToSend := strings.Replace(mailTemplate, "[%body%]", m.Content, 1)
		email.SetBody(mail.TextHTML, msgToSend)
	}

	err = email.Send(client)

	if err != nil {
		errorLog.Println(err)
	}
}
