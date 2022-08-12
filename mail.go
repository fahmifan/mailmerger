package mailmerge

import (
	"context"
	"errors"
	"io"

	"github.com/flosch/pongo2"
	"golang.org/x/sync/errgroup"
)

type MailTransporter interface {
	Send(ctx context.Context, from, to string, body []byte) error
}

type Mailer struct {
	senderEmail     string
	template        *pongo2.Template
	csvs            *Csv
	mailTransporter MailTransporter
	nworker         uint
}

func NewMailer(
	senderEmail string,
	csvs *Csv,
	mailTransporter MailTransporter,
	nworker uint,
) *Mailer {
	return &Mailer{
		senderEmail:     senderEmail,
		csvs:            csvs,
		mailTransporter: mailTransporter,
		nworker:         nworker,
	}
}

func (m *Mailer) ParseTemplate(rd io.Reader) (err error) {
	bt, err := io.ReadAll(rd)
	if err != nil {
		return
	}

	tpl, err := pongo2.FromBytes(bt)
	if err != nil {
		return
	}

	m.template = tpl
	return
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
			body, err := m.template.ExecuteBytes(pongo2.Context(row.Map()))
			if err != nil {
				return err
			}

			return m.mailTransporter.Send(ctx, m.senderEmail, row.GetCell("email"), body)
		})
	}

	return eg.Wait()
}