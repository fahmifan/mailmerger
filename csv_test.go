package mailmerger

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var csvExample = `id,name,email
123,john doe,john@doe.com
456,jean doe,jean@doe.com
789,han doe,han@doe.com
`

func TestCsv_Parse(t *testing.T) {
	cp := CSV{}

	err := cp.Parse(strings.NewReader(csvExample))
	require.NoError(t, err)

	require.Equal(t, 0, cp.mapHeaderIndex["id"])
	require.Equal(t, 1, cp.mapHeaderIndex["name"])
	require.Equal(t, 2, cp.mapHeaderIndex["email"])

	rows := cp.Rows()
	require.Equal(t, "123", rows[0].GetCell("id"))
	require.Equal(t, "jean doe", rows[1].GetCell("name"))
	require.Equal(t, "han@doe.com", rows[2].GetCell("email"))
}
