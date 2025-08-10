package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
)

//go:embed "templates"
var templates embed.FS

// Mailer struct has:
// an *mail.Dialer to a SMTP server,
// sender string to store infos of the email sender
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

func New(host string, port int, username, password, sender string) Mailer {
	// Initialize an *mail.Dialer with the given SMTP server settings,
	// and set 5 sec timeout when sending an email
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// Send takes the recipient email address, the name of the file containing the templates,
// the dynamic data for the templates
func (m Mailer) Send(recipient, templateFile string, data any) error {
	tmpl, err := template.New("email").ParseFS(templates, "templates/"+templateFile+".tmpl.html")
	if err != nil {
		return err
	}

	// execute (parse) the named defined template and write the result to a bytes buffer
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	html := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(html, "html", data)
	if err != nil {
		return err
	}

	// Initializes a new *mail.Message,
	// set headers, body to plain text and an altenative to plain text body, to html
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/html", html.String())

	err = m.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}

	return nil
}
