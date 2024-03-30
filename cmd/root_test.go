package cmd

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersistentPreRunE(t *testing.T) {
	origBaseURL := baseURL
	silenceUsage := rootCmd.SilenceUsage
	t.Cleanup(func() {
		baseURL = origBaseURL
		rootCmd.SilenceUsage = silenceUsage
	})

	baseURL = "http://127.0.0.1"
	require.NoError(t, persistentPreRunE(&rootCmd, []string{}))
	assert.True(t, rootCmd.SilenceUsage)
}

func TestLoadEnvs(t *testing.T) {
	origKey, origBaseURL := apiKey, baseURL
	t.Cleanup(func() {
		apiKey = origKey
		baseURL = origBaseURL
	})
	apiKey, baseURL = "foobar", "/"

	require.NoError(t, loadEnvs())
	assert.Equal(t, "foobar", apiKey)
	assert.Equal(t, "/", baseURL)

	apiKey, baseURL = "", ""
	require.NoError(t, os.Setenv("SYNCTHING_API_KEY", "foobaz"))
	require.NoError(t, os.Setenv("SYNCTHING_URL", "/syncthing/"))
	require.NoError(t, loadEnvs())
	assert.Equal(t, "foobaz", apiKey)
	assert.Equal(t, "/syncthing/", baseURL)
}

func TestRootValidate(t *testing.T) {
	origBaseURL := baseURL
	t.Cleanup(func() { baseURL = origBaseURL })
	require.ErrorContains(t, rootValidate(), "empty server URL")

	baseURL = ":/"
	require.ErrorContains(t, rootValidate(), "validate ")

	baseURL = "/"
	require.ErrorContains(t, rootValidate(), "not absolute")

	baseURL = "http://127.0.0.1"
	require.NoError(t, rootValidate())
}

func TestMustAPIClient(t *testing.T) {
	if os.Getenv("EXECUTE_TEST") == t.Name() {
		mustAPIClient()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^"+t.Name())
	cmd.Env = append(os.Environ(), "EXECUTE_TEST="+t.Name())
	require.NoError(t, cmd.Run())
}
