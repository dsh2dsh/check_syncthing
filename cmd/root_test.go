package cmd

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsh2dsh/check_syncthing/client/api"
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
	t.Setenv("SYNCTHING_API_KEY", "foobaz")
	t.Setenv("SYNCTHING_URL", "/syncthing/")
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

func TestDeviceName(t *testing.T) {
	name := deviceName(
		"XXXXXX2-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXX1",
		"system name")
	assert.Equal(t, "XXXXXX2 (system name)", name)
}

func TestNewDeviceId(t *testing.T) {
	id := newDeviceId(
		"XXXXXX2-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXX1")
	assert.Equal(t, "XXXXXX2", id.Short())
}

func TestNewLookupDeviceId(t *testing.T) {
	const testId2 = "XXXXXX2-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXX1"
	const testEx2 = "XXXXXX2 (device 1)"
	const testId4 = "XXXXXX4-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXX1"

	devices := map[string]api.DeviceConfiguration{
		testId2: {DeviceID: testId2, Name: "device 1"},
		testId4: {DeviceID: testId4, Name: "device 2"},
	}

	l := lookupDeviceId{}
	assert.False(t, l.Has(testId2))
	assert.False(t, l.Excluded())
	assert.Zero(t, l.ExcludedString(devices))

	l = newLookupDeviceId([]string{})
	assert.False(t, l.Has(testId2))
	assert.False(t, l.Excluded())
	assert.Zero(t, l.ExcludedString(devices))

	l = newLookupDeviceId([]string{"XXXXXX1", testId2, "XXXXXX3"})
	assert.False(t, l.Excluded())
	assert.Zero(t, l.ExcludedString(devices))
	assert.True(t, l.Has(testId2))

	assert.True(t, l.Excluded())
	assert.Equal(t, testEx2, l.ExcludedString(devices))

	assert.True(t, l.Has(testId2))
	assert.Equal(t, testEx2, l.ExcludedString(devices))

	assert.False(t, l.Has(testId4))
	assert.Equal(t, testEx2, l.ExcludedString(devices))

	l.Add(testId4)
	assert.True(t, l.Has(testId4))
	assert.Equal(t, testEx2+", XXXXXX4 (device 2)", l.ExcludedString(devices))
}
