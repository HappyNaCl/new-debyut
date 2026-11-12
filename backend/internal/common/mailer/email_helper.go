package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

//go:embed templates/*
var templates embed.FS

type EmailHelper struct {
	dialer *gomail.Dialer
	from   string
}

type WelcomeEmailData struct {
	To        string
	FirstName string
}

type OtpEmailData struct {
	To      string
	Otp     string
	Purpose string
}

type InvitationEmailData struct {
	To               string
	OrganizationName string
	InviterName      string
	Role             string
	ExpiresAt        string
	InvitationUrl    string
}

func NewEmailHelper() (*EmailHelper, error) {
	server := os.Getenv("SMTP_SERVER")
	email := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")

	portInt, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return nil, err
	}

	dialer := &gomail.Dialer{Host: server, Port: portInt}
	if password != "" {
		dialer = gomail.NewDialer(server, portInt, email, password)
	}

	return &EmailHelper{
		dialer: dialer,
		from:   email,
	}, nil
}

func (s *EmailHelper) SendWelcomeEmail(data WelcomeEmailData) error {
	return s.sendTemplated("templates/welcome.html", "Welcome to DebYut!", data.To, data)
}

func (s *EmailHelper) SendOtpEmail(data OtpEmailData) error {
	return s.sendTemplated("templates/otp.html", "Your debyut verification code", data.To, data)
}

func (s *EmailHelper) SendInvitationEmail(data InvitationEmailData) error {
	return s.sendTemplated("templates/invitation.html", "You've been invited to join "+data.OrganizationName+" on debyut", data.To, data)
}

func (s *EmailHelper) sendTemplated(templateFile, subject, to string, data any) error {
	tmpl, err := template.ParseFS(templates, templateFile)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return err
	}

	logoBytes, err := templates.ReadFile("templates/logo.png")
	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetAddressHeader("From", s.from, "noreply-debyut@gmail.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body.String())
	m.Embed("logo.png",
		gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(logoBytes)
			return err
		}),
		gomail.SetHeader(map[string][]string{
			"Content-ID": {"<logo>"},
		}),
	)

	return s.dialer.DialAndSend(m)
}
