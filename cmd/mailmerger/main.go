package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fahmifan/mailmerger"
	"github.com/fahmifan/mailmerger/pkg/smtp"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
		fmt.Println(err)
	}
}

func run() error {
	cmd := runBlastEmail()
	return cmd.Execute()
}

func runBlastEmail() *cobra.Command {
	cmd := cobra.Command{
		Use: "",
	}

	var cfgFile string
	cmd.Flags().StringVar(&cfgFile, "config", "config.json", `--config=my_config.json`)

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		cfgBuf, err := os.ReadFile(cfgFile)
		if err != nil {
			return
		}

		cfg := config{}
		if err = json.Unmarshal(cfgBuf, &cfg); err != nil {
			return
		}

		mailerCfg, err := cfg.mailerConfig()
		if err != nil {
			return
		}

		mailer := mailmerger.NewMailer(mailerCfg)
		if err = mailer.Parse(); err != nil {
			return
		}

		fmt.Println(">> start sending email")
		if err = mailer.SendAll(cmd.Context()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return nil
		}

		fmt.Println(">> success")
		return nil
	}
	return &cmd
}

var validate = validator.New()

type template struct {
	Subject string `json:"subject" validate:"required"`
	Body    string `json:"body" validate:"required"`
}

type config struct {
	Sender         string      `json:"sender" validate:"email,required"`
	CsvList        string      `json:"csv_list" validate:"required"`
	Template       template    `json:"template"`
	DefaultSubject string      `json:"default_subject"`
	Concurrency    int         `json:"concurrency" validate:"required"`
	SMTP           smtp.Config `json:"smtp"`
}

func (c *config) validate() error {
	return validate.Struct(c)
}

func (c *config) mailerConfig() (_ *mailmerger.MailerConfig, err error) {
	if err = c.validate(); err != nil {
		return
	}

	var (
		csvSrc  *os.File
		subject *os.File
		body    *os.File
	)

	if csvSrc, err = os.Open(c.CsvList); err != nil {
		return
	}

	if subject, err = os.Open(c.Template.Subject); err != nil {
		return
	}

	if body, err = os.Open(c.Template.Body); err != nil {
		return
	}

	smtClient, err := smtp.NewSmtpClient(&c.SMTP)
	if err != nil {
		return
	}

	mmrCfg := &mailmerger.MailerConfig{
		SenderEmail:     c.Sender,
		DefaultSubject:  c.DefaultSubject,
		CsvSrc:          csvSrc,
		BodyTemplate:    body,
		SubjectTemplate: subject,
		Concurrency:     uint(c.Concurrency),
		Transporter:     smtClient,
	}

	return mmrCfg, nil
}
