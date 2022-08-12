package mailmerger

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/flosch/pongo2"
	"golang.org/x/sync/errgroup"
)

type Transporter interface {
	Send(ctx context.Context, subject, from, to string, body []byte) error
}

type MailerConfig struct {
	SenderEmail string
	// DefaultSubject use when the subject template is empty
	DefaultSubject  string
	CsvSrc          io.Reader
	BodyTemplate    io.Reader
	SubjectTemplate io.Reader
	Concurrency     uint
	Transporter     Transporter
}

type Mailer struct {
	tmplBody    *pongo2.Template
	tmplSubject *pongo2.Template
	csv         *CSV
	*MailerConfig
}

// RegisterFilter expose pongo2.RegisterFilter
func RegisterFilter(name string, fn pongo2.FilterFunction) error {
	return pongo2.RegisterFilter(name, fn)
}

func NewMailer(cfg *MailerConfig) *Mailer {
	return &Mailer{
		csv:          &CSV{},
		MailerConfig: cfg,
	}
}

// Parse parse csv & templates
func (m *Mailer) Parse() (err error) {
	if err = m.parseCsv(m.CsvSrc); err != nil {
		return
	}
	if err = m.parseBodyTemplate(m.BodyTemplate); err != nil {
		return
	}
	if err = m.parseSubjectTemplate(m.SubjectTemplate); err != nil {
		return
	}
	return
}

func (m *Mailer) parseBodyTemplate(rd io.Reader) (err error) {
	m.tmplBody, err = m.parseTmpl(rd)
	return err
}

func (m *Mailer) parseSubjectTemplate(rd io.Reader) (err error) {
	m.tmplSubject, err = m.parseTmpl(rd)
	return err
}

func (m *Mailer) parseTmpl(rd io.Reader) (_ *pongo2.Template, err error) {
	bt, err := io.ReadAll(rd)
	if err != nil {
		return
	}

	return pongo2.FromBytes(bt)
}

func (m *Mailer) parseCsv(rd io.Reader) (err error) {
	err = m.csv.Parse(rd)
	const mandatoryField = "email"
	if !m.csv.IsHeader(mandatoryField) {
		return errors.New("email field is mandatory")
	}
	return
}

// SendAll send email to all recipient from csv
func (m *Mailer) SendAll(ctx context.Context) (err error) {
	conc := m.Concurrency
	if conc == 0 {
		conc = 1
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(int(conc))

	for _, row := range m.csv.Rows() {
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
				subjectStr = m.DefaultSubject
			}
			return m.Transporter.Send(ctx, subjectStr, m.SenderEmail, row.GetCell("email"), body)
		})
	}

	return eg.Wait()
}
