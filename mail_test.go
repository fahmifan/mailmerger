package mailmerge

import (
	"context"
	"strings"
	"testing"

	"github.com/fahmifan/mailmerge/tests/mock_mailmerge"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

var pongoCsv = `email,name,token
john@doe.com,john doe,token1
jean@doe.com,jean doe,token2
han@doe.com,han doe,token3
`

var pongoBodyTmpl = `
	Selamat pagi {{ name | title }}
	Berikut adalah token registrasi yang dapat dipakai {{ token }}
`

var pongoSubjectTmpl = `
	Token Registrasi {{ name | title }}
`

func TestMailer_SendAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mailTransporterMock := mock_mailmerge.NewMockMailTransporter(ctrl)

	sender := "test@mail.com"
	csv := Csv{}
	mailer := NewMailer("Token Registrasi", sender, &csv, mailTransporterMock, 2)

	err := mailer.ParseCsv(strings.NewReader(pongoCsv))
	require.NoError(t, err)

	err = mailer.ParseBodyTemplate(strings.NewReader(pongoBodyTmpl))
	require.NoError(t, err)

	err = mailer.ParseSubjectTemplate(strings.NewReader(pongoSubjectTmpl))
	require.NoError(t, err)

	ctx := context.TODO()
	mailTransporterMock.EXPECT().Send(gomock.Any(), gomock.Any(), sender, gomock.Any(), gomock.Any()).AnyTimes().
		DoAndReturn(func(ctx context.Context, subject, from, to string, body []byte) error {
			bodyStr := string(body)
			if to == "john@doe.com" {
				require.Contains(t, bodyStr, `Selamat pagi John Doe`)
			}
			if to == "jean@doe.com" {
				require.Contains(t, subject, `Token Registrasi Jean Doe`)
			}
			if to == "han@doe.com" {
				require.Contains(t, bodyStr, `Berikut adalah token registrasi yang dapat dipakai token3`)
			}
			return nil
		})

	err = mailer.SendAll(ctx)
	require.NoError(t, err)
}
