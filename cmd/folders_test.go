package cmd

import (
	"testing"

	"github.com/dsh2dsh/go-monitoringplugin/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsh2dsh/check_syncthing/client/api"
)

func TestNewFoldersCheck(t *testing.T) {
	check := testNewFoldersCheck(t, nil)
	resp := check.Response()
	require.NotNil(t, resp)
	assert.Equal(t, "OK: "+foldersOkMsg, resp.GetInfo().RawOutput)
}

func testNewFoldersCheck(t *testing.T, endpoints map[string]any,
) *FoldersCheck {
	check := NewFoldersCheck(newTestClient(t, fakeAPI(t, endpoints)))
	require.NotNil(t, check)
	return check
}

func TestFoldersCheck_applyOptionsWithResp(t *testing.T) {
	resp := monitoringplugin.NewResponse("def ok msg")
	check := &FoldersCheck{resp: resp}
	require.Same(t, check, check.applyOptions())
	assert.Same(t, resp, check.Response())
}

func TestFoldersCheck_WithExcludeDevices(t *testing.T) {
	check := testNewFoldersCheck(t, nil)
	assert.Same(t, check, check.WithExcludeDevices([]string{"A", "B"}))
}

func TestFoldersCheck_Run(t *testing.T) {
	const testId2 = "XXXXXX2-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX"

	folders := []api.FolderConfiguration{
		{
			Devices: []api.FolderDeviceConfiguration{{DeviceId: testId2}},
			Id:      "default",
			Label:   "Default Folder",
			Path:    "/Default Folder",
		},
	}

	tests := []struct {
		name         string
		with         func(t *testing.T, check *FoldersCheck)
		endpoints    map[string]any
		assertOutput func(t *testing.T, rawOutput string)
	}{
		{
			name: "without endpoints",
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "CRITICAL: ")
			},
		},
		{
			name: "with folders",
			endpoints: map[string]any{
				"/rest/config/folders": folders,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput,
					`CRITICAL: folder id="default", label="Default Folder":`)
			},
		},
		{
			name: "OK",
			endpoints: map[string]any{
				"/rest/config/folders": folders,
				"/rest/folder/errors?folder=default": api.FolderErrors{
					Folder: "default",
				},
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Equal(t, "OK: 1"+foldersOkMsg, rawOutput)
			},
		},
		{
			name: "with folder error",
			endpoints: map[string]any{
				"/rest/config/folders": folders,
				"/rest/folder/errors?folder=default": api.FolderErrors{
					Errors: []api.FileError{
						{
							Error: "some error",
							Path:  "/some file path",
						},
					},
					Folder: "default",
				},
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Equal(t, `WARNING: 1/1 folders with errors
folder: default (Default Folder)
path: /some file path
error: some error`, rawOutput)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := testNewFoldersCheck(t, tt.endpoints)
			if tt.with != nil {
				tt.with(t, check)
			}
			require.Same(t, check, check.Run())
			resp := check.Response()
			require.NotNil(t, resp)
			t.Log(resp.GetInfo().RawOutput)
			tt.assertOutput(t, resp.GetInfo().RawOutput)
		})
	}
}
