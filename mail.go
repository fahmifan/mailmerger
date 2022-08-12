package mailmerger

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/flosch/pongo2"
	"golang.org/x/sync/errgroup"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type MailTransporter interface {
	Send(ctx context.Context, subject, from, to string, body []byte) error
}

type Mailer struct {
	globalSubject   string
	senderEmail     string
	tmplBody        *pongo2.Template
	tmplSubject     *pongo2.Template
	csvs            *Csv
	mailTransporter MailTransporter
	nworker         uint
	tmplFunc        map[string]pongo2.FilterFunction
}

func NewMailer(
	globalSubject string,
	senderEmail string,
	csvs *Csv,
	mailTransporter MailTransporter,
	nworker uint,
) *Mailer {
	return &Mailer{
		globalSubject:   globalSubject,
		senderEmail:     senderEmail,
		csvs:            csvs,
		mailTransporter: mailTransporter,
		nworker:         nworker,
		tmplFunc: map[string]pongo2.FilterFunction{
			"title": func(in, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
				caser := cases.Title(language.English)
				return pongo2.AsValue(caser.String(in.String())), nil
			},
		},
	}
}

func (m *Mailer) ParseBodyTemplate(rd io.Reader) (err error) {
	m.tmplBody, err = m.parseTmpl(rd)
	return err
}

func (m *Mailer) ParseSubjectTemplate(rd io.Reader) (err error) {
	m.tmplSubject, err = m.parseTmpl(rd)
	return err
}

func (m *Mailer) parseTmpl(rd io.Reader) (_ *pongo2.Template, err error) {
	bt, err := io.ReadAll(rd)
	if err != nil {
		return
	}

	for key, val := range m.tmplFunc {
		pongo2.RegisterFilter(key, val)
	}

	return pongo2.FromBytes(bt)
}

func (m *Mailer) ParseCsv(rd io.Reader) (err error) {
	err = m.csvs.Parse(rd)
	const mandatoryField = "email"
	if !m.csvs.IsHeader(mandatoryField) {
		return errors.New("email field is mandatory")
	}
	return
}

// SendAll send email to all recipient from csvs
func (m *Mailer) SendAll(ctx context.Context) (err error) {
	nworker := m.nworker
	if nworker == 0 {
		nworker = 1
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(int(nworker))

	for _, row := range m.csvs.Rows() {
		row := row
		eg.Go(func() error {
			rowMap := row.Map()
			body, err := m.tmplBody.ExecuteBytes(pongo2.Context(rowMap))
			if err != nil {
				return fmt.Errorf("exec body: %w", err)
			}

			subjectBt, err := m.tmplSubject.ExecuteBytes(pongo2.Context(rowMap))
			if err != nil {
				return fmt.Errorf("exec subject: %w", err)
			}

			subjectStr := string(subjectBt)
			if subjectStr == "" {
				subjectStr = m.globalSubject
			}
			return m.mailTransporter.Send(ctx, subjectStr, m.senderEmail, row.GetCell("email"), body)
		})
	}

	return eg.Wait()
}
