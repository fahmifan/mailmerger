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

var pongoTmpl = `
	Selamat pagi {{ name }}
	Berikut adalah token yang dapat dipakai {{ token }}
`

var _ MailTransporter = (*MailTransporterMock)(nil)

type MailTransporterMock struct {
}

func (m *MailTransporterMock) Send(ctx context.Context, from, to string, body []byte) error {
	return nil
}

func TestMailer_SendAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mailTransporterMock := mock_mailmerge.NewMockMailTransporter(ctrl)

	sender := "test@mail.com"
	csv := Csv{}
	mailer := NewMailer(sender, &csv, mailTransporterMock, 2)

	err := mailer.ParseCsv(strings.NewReader(pongoCsv))
	require.NoError(t, err)

	err = mailer.ParseTemplate(strings.NewReader(pongoTmpl))
	require.NoError(t, err)

	ctx := context.TODO()
	mailTransporterMock.EXPECT().Send(gomock.Any(), sender, gomock.Any(), gomock.Any()).AnyTimes().
		DoAndReturn(func(ctx context.Context, from, to string, body []byte) error {
			bodyStr := string(body)
			if to == "john@doe.com" {
				require.Contains(t, bodyStr, `Selamat pagi john doe`)
			}
			if to == "han@doe.com" {
				require.Contains(t, bodyStr, `Berikut adalah token yang dapat dipakai token3`)
			}
			return nil
		})

	err = mailer.SendAll(ctx)
	require.NoError(t, err)
}
