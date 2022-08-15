package smtp

import (
	"context"
	_ "embed"

	"github.com/fahmifan/mailmerger"
	"gopkg.in/gomail.v2"
)

var _ mailmerger.Transporter = (*SMTP)(nil)

type Config struct {
	Host     string `json:"host" validate:"required"`
	Port     int    `json:"port" validate:"required"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type SMTP struct {
	mail *gomail.Dialer
	cfg  *Config
}

func NewSmtpClient(cfg *Config) (smtp *SMTP, err error) {
	smtp = &SMTP{
		cfg: cfg,
		mail: gomail.NewDialer(
			cfg.Host,
			cfg.Port,
			cfg.Username,
			cfg.Password,
		),
	}

	closer, err := smtp.mail.Dial()
	if err != nil {
		return nil, err
	}
	closer.Close()

	return smtp, nil
}

func (m *SMTP) Send(ctx context.Context, subject, from, to string, body []byte) (err error) {
	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", string(body))

	return m.mail.DialAndSend(msg)
}
