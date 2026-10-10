package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// TLSMode — режим соединения с SMTP-сервером.
type TLSMode string

const (
	TLSModeNone     TLSMode = "none"     // plain SMTP (MailHog, Mailtrap на 2525)
	TLSModeSTARTTLS TLSMode = "starttls" // порт 587 (Gmail, Yandex, Mail.ru)
	TLSModeImplicit TLSMode = "implicit" // порт 465
)

type SMTPSender struct {
	host     string
	port     int
	user     string
	password string
	from     string
	mode     TLSMode
	timeout  time.Duration
}

func NewSMTPSender(host string, port int, user, password, from string, mode TLSMode) *SMTPSender {
	return &SMTPSender{
		host: host, port: port,
		user: user, password: password,
		from: from, mode: mode,
		timeout: 10 * time.Second,
	}
}

func (s *SMTPSender) SendEmailVerification(_ context.Context, to, link string) error {
	subject := "Подтверждение email — ФСП"
	body := fmt.Sprintf(`
		<p>Здравствуйте!</p>
		<p>Подтвердите ваш email, перейдя по ссылке:</p>
		<p><a href="%s">%s</a></p>
		<p>Если вы не регистрировались — просто проигнорируйте письмо.</p>`, link, link)
	return s.send(to, subject, body)
}

func (s *SMTPSender) SendPasswordReset(_ context.Context, to, link string) error {
	subject := "Сброс пароля — ФСП"
	body := fmt.Sprintf(`
		<p>Здравствуйте!</p>
		<p>Сбросить пароль можно по ссылке:</p>
		<p><a href="%s">%s</a></p>
		<p>Если это были не вы — проигнорируйте письмо.</p>`, link, link)
	return s.send(to, subject, body)
}

func (s *SMTPSender) send(to, subject, htmlBody string) error {
	var b strings.Builder
	b.WriteString("From: " + s.from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
	b.WriteString(htmlBody)
	msg := []byte(b.String())

	addr := net.JoinHostPort(s.host, fmt.Sprint(s.port))
	dialer := &net.Dialer{Timeout: s.timeout}

	var conn net.Conn
	var err error
	switch s.mode {
	case TLSModeImplicit:
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
			ServerName: s.host,
		})
	default:
		conn, err = dialer.Dial("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if s.mode == TLSModeSTARTTLS {
		if err := client.StartTLS(&tls.Config{ServerName: s.host}); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	if s.user != "" {
		auth := smtp.PlainAuth("", s.user, s.password, s.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}
	return client.Quit()
}
