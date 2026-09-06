package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/testutil"
)

// executeRoot runs the whole command tree with args and returns its combined
// output. Errors are returned (the root silences its own error printing).
func executeRoot(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestNewRootCmdWiring(t *testing.T) {
	root := NewRootCmd()
	require.Equal(t, "foreman", root.Name())

	sub := map[string]bool{}
	for _, c := range root.Commands() {
		sub[c.Name()] = true
	}
	require.True(t, sub["task"], "task command is registered")
	require.True(t, sub["message"], "message command is registered")
}

func TestEnvOr(t *testing.T) {
	const key = "FOREMAN_TEST_ENVOR"
	value := testutil.RandomText()
	t.Setenv(key, value)

	require.Equal(t, value, envOr(key, "fallback"))
	require.Equal(t, "fallback", envOr("FOREMAN_TEST_ENVOR_MISSING", "fallback"))
}

func TestGlobalFlagsClient(t *testing.T) {
	_, err := (&globalFlags{apiURL: "http://x", apiKey: ""}).client()
	require.Error(t, err)

	c, err := (&globalFlags{apiURL: "http://x", apiKey: testutil.RandomAPIKey()}).client()
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestRender(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	textRan := false
	require.NoError(t, render(cmd, &globalFlags{}, nil, func() { textRan = true }))
	require.True(t, textRan)

	buf.Reset()
	err := render(cmd, &globalFlags{json: true}, map[string]string{"k": "v"}, func() {
		t.Fatal("text renderer must not run in json mode")
	})
	require.NoError(t, err)
	require.JSONEq(t, `{"k":"v"}`, buf.String())
}
