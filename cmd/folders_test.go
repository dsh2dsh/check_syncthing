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
	const testId4 = "XXXXXX4-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX"

	devices := []api.DeviceConfiguration{
		{DeviceID: testId2, Name: "device2"},
		{DeviceID: testId4, Name: "device4"},
	}

	folders := []api.FolderConfiguration{
		{
			Devices: []api.FolderDeviceConfiguration{{DeviceId: testId2}},
			Id:      "default",
			Label:   "Default Folder",
			Path:    "/Default Folder",
		},
	}

	folderErrors := api.FolderErrors{Folder: "default"}
	defaultError := api.FolderErrors{
		Errors: []api.FileError{
			{
				Error: "some error",
				Path:  "/some file path",
			},
		},
		Folder: "default",
	}

	compEP := "/rest/db/completion?device=" + testId2 + "&folder=default"
	compEP2 := "/rest/db/completion?device=" + testId2 + "&folder=default2"
	completion := api.FolderCompletion{Completion: 100}

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
			name: "with devices",
			endpoints: map[string]any{
				"/rest/config/devices": devices,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "CRITICAL: ")
			},
		},
		{
			name: "with folders",
			endpoints: map[string]any{
				"/rest/config/devices": devices,
				"/rest/config/folders": folders,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "CRITICAL: ")
			},
		},
		{
			name: "with default folder errors",
			endpoints: map[string]any{
				"/rest/config/devices":               devices,
				"/rest/config/folders":               folders,
				"/rest/folder/errors?folder=default": folderErrors,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "CRITICAL: ")
			},
		},
		{
			name: "OK",
			endpoints: map[string]any{
				"/rest/config/devices":               devices,
				"/rest/config/folders":               folders,
				"/rest/folder/errors?folder=default": folderErrors,
				compEP:                               &completion,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Equal(t, "OK: 1"+foldersOkMsg, rawOutput)
			},
		},
		{
			name: "OK excluded",
			with: func(t *testing.T, check *FoldersCheck) {
				check.WithExcludeDevices([]string{newDeviceId(testId4).Short()})
			},
			endpoints: map[string]any{
				"/rest/config/devices": devices,
				"/rest/config/folders": []api.FolderConfiguration{
					{
						Devices: []api.FolderDeviceConfiguration{
							{DeviceId: testId2},
							{DeviceId: testId4},
						},
						Id:    "default",
						Label: "Default Folder",
						Path:  "/Default Folder",
					},
				},
				"/rest/folder/errors?folder=default": folderErrors,
				compEP:                               &completion,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "OK: 1"+foldersOkMsg)
				assert.Contains(t, rawOutput, "excluded: XXXXXX4 (device4)")
			},
		},
		{
			name: "with folder error",
			endpoints: map[string]any{
				"/rest/config/devices":               devices,
				"/rest/config/folders":               folders,
				"/rest/folder/errors?folder=default": defaultError,
				compEP:                               &completion,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Equal(t, `WARNING: 1/1 folders with errors
folder: default (Default Folder)
path: /some file path
error: some error`, rawOutput)
			},
		},
		{
			name: "with folder error 2",
			endpoints: map[string]any{
				"/rest/config/devices": devices,
				"/rest/config/folders": []api.FolderConfiguration{
					folders[0],
					{
						Devices: []api.FolderDeviceConfiguration{{DeviceId: testId2}},
						Id:      "default2",
						Label:   "Default Folder2",
						Path:    "/Default Folder2",
					},
				},
				"/rest/folder/errors?folder=default": defaultError,
				"/rest/folder/errors?folder=default2": api.FolderErrors{
					Folder: "default2",
				},
				compEP:  &completion,
				compEP2: &completion,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Equal(t, `WARNING: 1/2 folders with errors
folder: default (Default Folder)
path: /some file path
error: some error`, rawOutput)
			},
		},
		{
			name: "out of sync",
			endpoints: map[string]any{
				"/rest/config/devices":               devices,
				"/rest/config/folders":               folders,
				"/rest/folder/errors?folder=default": folderErrors,
				compEP: &api.FolderCompletion{
					Completion: 99,
				},
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Equal(t, `WARNING: 1/1 folders out of sync
folder: default (Default Folder)
device: XXXXXX2 (device2) - 99%`, rawOutput)
			},
		},
		{
			name: "with out of sync folder error",
			endpoints: map[string]any{
				"/rest/config/devices": devices,
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
				compEP: &api.FolderCompletion{
					Completion: 99,
				},
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Equal(t, `WARNING: 1/1 folders with errors
folder: default (Default Folder)
path: /some file path
error: some error`, rawOutput)
			},
		},
		{
			name: "not shared",
			endpoints: map[string]any{
				"/rest/config/devices": devices,
				"/rest/config/folders": []api.FolderConfiguration{
					{
						Id:    "default",
						Label: "Default Folder",
						Path:  "/Default Folder",
					},
				},
				"/rest/folder/errors?folder=default": folderErrors,
				compEP:                               &completion,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Equal(t,
					"WARNING: 1 folder not shared: default (Default Folder)",
					rawOutput)
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
