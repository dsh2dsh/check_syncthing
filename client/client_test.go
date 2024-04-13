package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsh2dsh/check_syncthing/client/api"
)

func TestNew(t *testing.T) {
	const baseURL = "/"
	var callCount int

	c, err := New(baseURL, func(*Client) error {
		callCount++
		return nil
	})
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, baseURL, c.baseURL)
	assert.NotNil(t, c.apiClient)
	assert.Equal(t, 1, callCount)

	wantErr := errors.New("test error")
	_, err = New(baseURL, func(*Client) error { return wantErr })
	require.ErrorIs(t, err, wantErr)

	apiClient := c.apiClient
	c, err = New(baseURL, func(self *Client) error {
		err := self.NewClientWithResponses()
		apiClient = self.apiClient
		return err
	})
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Same(t, apiClient, c.apiClient)
}

func TestClient_NewClientWithResponses(t *testing.T) {
	c := Client{}
	wantErr := errors.New("test error")
	err := c.NewClientWithResponses(func(*api.Client) error { return wantErr })
	require.ErrorIs(t, err, wantErr)
}

func TestClient_WithKey(t *testing.T) {
	const key = "foobar"
	c, err := New("/")
	require.NoError(t, err)
	assert.Same(t, c, c.WithKey(key))
	assert.Equal(t, key, c.apiKey)
}

func TestClient_withAPIKey(t *testing.T) {
	const key = "foobar"
	httpClient := testHttpDoer{
		func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, key, req.Header.Get(authHeader))
			r := httptest.NewRecorder()
			r.Header().Set("Content-Type", "application/json")
			_, err := r.WriteString(`{ "status": "OK" }`)
			require.NoError(t, err)
			return r.Result(), nil
		},
	}

	c, err := New("/", func(self *Client) error {
		return self.NewClientWithResponses(api.WithHTTPClient(&httpClient))
	})
	require.NoError(t, err)
	require.NotNil(t, c)

	require.NoError(t, c.WithKey(key).Health(context.Background()))
}

func TestMakeAPIError(t *testing.T) {
	err := makeAPIError(nil, "404 not found", []byte("some body\n"))
	t.Log(err)
	require.ErrorContains(t, err,
		"unexpected syncthing response: 404 not found (some body)")

	jsonErr := api.Error{Error: "some error"}
	err = makeAPIError(&jsonErr, "", nil)
	t.Log(err)
	require.ErrorContains(t, err, "unexpected syncthing error: some error")
}

type testHttpDoer struct {
	do func(req *http.Request) (*http.Response, error)
}

func (self *testHttpDoer) Do(req *http.Request) (*http.Response, error) {
	return self.do(req)
}

func TestClient_Health(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		contentType string
		body        string
		assertErr   func(t *testing.T, err error)
	}{
		{
			name:        "status OK",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			body:        `{ "status": "OK" }`,
		},
		{
			name:        "status NOT OK",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			body:        `{ "status": "NOT OK" }`,
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "health: unexpected status:")
			},
		},
		{
			name:        "empty body",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			assertErr: func(t *testing.T, err error) {
				var syntaxErr *json.SyntaxError
				assert.ErrorAs(t, err, &syntaxErr)
			},
		},
		{
			name:       "wrong content type",
			statusCode: http.StatusOK,
			body:       `{ "status": "OK" }`,
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "health: unexpected syncthing response:")
			},
		},
		{
			name:        "has error",
			statusCode:  http.StatusInternalServerError,
			contentType: "application/json",
			body:        `{ "error": "some error message" }`,
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err,
					"health: unexpected syncthing error: some error message")
			},
		},
		{
			name:       "wrong status code",
			statusCode: http.StatusNotFound,
			body:       "something not found",
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err,
					"health: unexpected syncthing response: 404 Not Found (something not found)")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, tt.contentType, tt.statusCode, tt.body)
			err := c.Health(context.Background())
			if tt.assertErr != nil {
				t.Log(err)
				require.Error(t, err)
				tt.assertErr(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func newTestClient(t *testing.T, contentType string, statusCode int,
	body string,
) *Client {
	httpClient := testHttpDoer{
		func(req *http.Request) (*http.Response, error) {
			r := httptest.NewRecorder()
			r.Header().Set("Content-Type", contentType)
			r.WriteHeader(statusCode)
			_, err := r.WriteString(body)
			require.NoError(t, err)
			return r.Result(), nil
		},
	}

	c, err := New("/", func(self *Client) error {
		return self.NewClientWithResponses(api.WithHTTPClient(&httpClient))
	})
	require.NoError(t, err)
	require.NotNil(t, c)
	return c
}

func TestClient_Connections(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		contentType string
		body        string
		assertErr   func(t *testing.T, err error)
	}{
		{
			name:        "OK",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			body: `
{
  "connections": {
    "XXXXXX1-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXX1": {
      "at": "2024-03-22T21:14:32+01:00",
      "inBytesTotal": 10539,
      "outBytesTotal": 877500,
      "startedAt": "2024-03-22T14:34:54+01:00",
      "connected": true,
      "paused": false,
      "clientVersion": "v0.0.4",
      "address": "127.0.1.2:22000",
      "type": "tcp-server",
      "isLocal": true,
      "crypto": "TLS1.3-TLS_AES_128_GCM_SHA256",
      "primary": {
        "at": "2024-03-22T21:14:32+01:00",
        "inBytesTotal": 10539,
        "outBytesTotal": 877500,
        "startedAt": "2024-03-22T14:34:54+01:00",
        "address": "127.0.1.2:22000",
        "type": "tcp-server",
        "isLocal": true,
        "crypto": "TLS1.3-TLS_AES_128_GCM_SHA256"
      }
    },
    "XXXXXX2-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXX2": {
      "at": "0001-01-01T00:00:00Z",
      "inBytesTotal": 0,
      "outBytesTotal": 0,
      "startedAt": "0001-01-01T00:00:00Z",
      "connected": false,
      "paused": false,
      "clientVersion": "",
      "address": "",
      "type": "",
      "isLocal": false,
      "crypto": "",
      "primary": {
        "at": "0001-01-01T00:00:00Z",
        "inBytesTotal": 0,
        "outBytesTotal": 0,
        "startedAt": "0001-01-01T00:00:00Z",
        "address": "",
        "type": "",
        "isLocal": false,
        "crypto": ""
      }
    }
  },

  "total": {
    "at": "2024-03-22T21:14:32+01:00",
    "inBytesTotal": 163746446,
    "outBytesTotal": 427863355
  }
}`,
		},
		{
			name:        "empty body",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			assertErr: func(t *testing.T, err error) {
				var syntaxErr *json.SyntaxError
				assert.ErrorAs(t, err, &syntaxErr)
			},
		},
		{
			name:        "has error",
			statusCode:  http.StatusInternalServerError,
			contentType: "application/json",
			body:        `{ "error": "some error message" }`,
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err,
					"connections: unexpected syncthing error: some error message")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, tt.contentType, tt.statusCode, tt.body)
			conn, err := c.Connections(context.Background())
			if tt.assertErr != nil {
				t.Log(err)
				require.Error(t, err)
				tt.assertErr(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, conn)
				if tt.body != "" {
					var want *api.Connections
					require.NoError(t, json.Unmarshal([]byte(tt.body), &want))
					assert.Equal(t, want, conn)
				}
			}
		})
	}
}

func TestClient_Folders(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		contentType string
		body        string
		assertErr   func(t *testing.T, err error)
	}{
		{
			name:        "OK",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			body: `
[{
	"id": "default",
	"label": "Default Folder",
	"filesystemType": "basic",
	"path": "/Default Folder",
	"type": "sendreceive",
	"devices": [
		{
			"deviceID": "XXXXXX1-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXX1",
			"introducedBy": "",
			"encryptionPassword": ""
		}
	],
	"rescanIntervalS": 3600,
	"fsWatcherEnabled": true,
	"fsWatcherDelayS": 10,
	"ignorePerms": false,
	"autoNormalize": true,
	"minDiskFree": {
		"value": 0,
		"unit": ""
	},
	"versioning": {
		"type": "",
		"params": {},
		"cleanupIntervalS": 3600,
		"fsPath": "",
		"fsType": "basic"
	},
	"copiers": 0,
	"pullerMaxPendingKiB": 0,
	"hashers": 0,
	"order": "random",
	"ignoreDelete": false,
	"scanProgressIntervalS": 0,
	"pullerPauseS": 0,
	"maxConflicts": -1,
	"disableSparseFiles": false,
	"disableTempIndexes": false,
	"paused": false,
	"weakHashThresholdPct": 25,
	"markerName": ".stfolder",
	"copyOwnershipFromParent": false,
	"modTimeWindowS": 0,
	"maxConcurrentWrites": 2,
	"disableFsync": false,
	"blockPullOrder": "standard",
	"copyRangeMethod": "standard",
	"caseSensitiveFS": false,
	"junctionsAsDirs": false,
	"syncOwnership": false,
	"sendOwnership": false,
	"syncXattrs": false,
	"sendXattrs": false,
	"xattrFilter": {
		"entries": [],
		"maxSingleEntrySize": 0,
		"maxTotalSize": 0
	}
}]`,
		},
		{
			name:        "empty body",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			assertErr: func(t *testing.T, err error) {
				var syntaxErr *json.SyntaxError
				assert.ErrorAs(t, err, &syntaxErr)
			},
		},
		{
			name:        "has error",
			statusCode:  http.StatusInternalServerError,
			contentType: "application/json",
			body:        `{ "error": "some error message" }`,
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err,
					"folders: unexpected syncthing error: some error message")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, tt.contentType, tt.statusCode, tt.body)
			folders, err := c.Folders(context.Background())
			if tt.assertErr != nil {
				t.Log(err)
				require.Error(t, err)
				tt.assertErr(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, folders)
				if tt.body != "" {
					var want []api.FolderConfiguration
					require.NoError(t, json.Unmarshal([]byte(tt.body), &want))
					assert.Equal(t, want, folders)
				}
			}
		})
	}
}

func TestClient_Devices(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		contentType string
		body        string
		assertErr   func(t *testing.T, err error)
	}{
		{
			name:        "OK",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			body: `
[
  {
    "deviceID": "XXXXXX1-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXX1",
    "name": "some device",
    "addresses": [
      "dynamic"
    ],
    "compression": "metadata",
    "certName": "",
    "introducer": false,
    "skipIntroductionRemovals": false,
    "introducedBy": "",
    "paused": false,
    "allowedNetworks": [],
    "autoAcceptFolders": false,
    "maxSendKbps": 0,
    "maxRecvKbps": 0,
    "ignoredFolders": [],
    "maxRequestKiB": 0,
    "untrusted": false,
    "remoteGUIPort": 0,
    "numConnections": 0
  }
]`,
		},
		{
			name:        "empty body",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			assertErr: func(t *testing.T, err error) {
				var syntaxErr *json.SyntaxError
				assert.ErrorAs(t, err, &syntaxErr)
			},
		},
		{
			name:        "has error",
			statusCode:  http.StatusInternalServerError,
			contentType: "application/json",
			body:        `{ "error": "some error message" }`,
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err,
					"devices: unexpected syncthing error: some error message")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, tt.contentType, tt.statusCode, tt.body)
			devices, err := c.Devices(context.Background())
			if tt.assertErr != nil {
				t.Log(err)
				require.Error(t, err)
				tt.assertErr(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, devices)
				if tt.body != "" {
					var want []api.DeviceConfiguration
					require.NoError(t, json.Unmarshal([]byte(tt.body), &want))
					assert.Equal(t, want, devices)
				}
			}
		})
	}
}

func TestClient_DeviceStats(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		contentType string
		body        string
		assertErr   func(t *testing.T, err error)
	}{
		{
			name:        "OK",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			body: `
{
  "XXXXXX1-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXX1": {
    "lastSeen": "2024-03-28T20:15:11+01:00",
    "lastConnectionDurationS": 451.813963185
  }
}`,
		},
		{
			name:        "empty body",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			assertErr: func(t *testing.T, err error) {
				var syntaxErr *json.SyntaxError
				assert.ErrorAs(t, err, &syntaxErr)
			},
		},
		{
			name:        "has error",
			statusCode:  http.StatusInternalServerError,
			contentType: "application/json",
			body:        `{ "error": "some error message" }`,
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err,
					"device stats: unexpected syncthing error: some error message")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, tt.contentType, tt.statusCode, tt.body)
			stats, err := c.DeviceStats(context.Background())
			if tt.assertErr != nil {
				t.Log(err)
				require.Error(t, err)
				tt.assertErr(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, stats)
				if tt.body != "" {
					var want map[string]api.DeviceStatistics
					require.NoError(t, json.Unmarshal([]byte(tt.body), &want))
					assert.Equal(t, want, stats)
				}
			}
		})
	}
}

func TestClient_Completion(t *testing.T) {
	const folder = "default"
	const device = "XXXXXX1-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXX1"

	tests := []struct {
		name           string
		folder, device string
		statusCode     int
		contentType    string
		body           string
		assertErr      func(t *testing.T, err error)
	}{
		{
			name:        "OK",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			body: `
{
  "completion": 100,
  "globalBytes": 10574239170,
  "globalItems": 2808,
  "needBytes": 0,
  "needDeletes": 0,
  "needItems": 0,
  "remoteState": "unknown",
  "sequence": 0
}`,
		},
		{
			name:        "OK with folder",
			folder:      folder,
			statusCode:  http.StatusOK,
			contentType: "application/json",
			body: `
{
  "completion": 100,
  "globalBytes": 26996726,
  "globalItems": 39,
  "needBytes": 0,
  "needDeletes": 0,
  "needItems": 0,
  "remoteState": "unknown",
  "sequence": 145
}`,
		},
		{
			name:        "OK with folder and device",
			folder:      folder,
			device:      device,
			statusCode:  http.StatusOK,
			contentType: "application/json",
			body: `
{
  "completion": 100,
  "globalBytes": 26996726,
  "globalItems": 39,
  "needBytes": 0,
  "needDeletes": 0,
  "needItems": 0,
  "remoteState": "valid",
  "sequence": 805
}`,
		},
		{
			name:        "OK device",
			device:      device,
			statusCode:  http.StatusOK,
			contentType: "application/json",
			body: `
{
  "completion": 100,
  "globalBytes": 567088438,
  "globalItems": 1003,
  "needBytes": 0,
  "needDeletes": 0,
  "needItems": 0,
  "remoteState": "unknown",
  "sequence": 0
}`,
		},
		{
			name:        "empty body",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			assertErr: func(t *testing.T, err error) {
				var syntaxErr *json.SyntaxError
				assert.ErrorAs(t, err, &syntaxErr)
			},
		},
		{
			name:        "has error",
			statusCode:  http.StatusInternalServerError,
			contentType: "application/json",
			body:        `{ "error": "some error message" }`,
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err,
					"completion: unexpected syncthing error: some error message")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, tt.contentType, tt.statusCode, tt.body)
			comp, err := c.Completion(context.Background(), tt.folder, tt.device)
			if tt.assertErr != nil {
				t.Log(err)
				require.Error(t, err)
				tt.assertErr(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, comp)
				if tt.body != "" {
					var want api.FolderCompletion
					require.NoError(t, json.Unmarshal([]byte(tt.body), &want))
					assert.Equal(t, &want, comp)
				}
			}
		})
	}
}

func TestClient_SystemErrors(t *testing.T) {
	tests := []struct {
		name           string
		folder, device string
		statusCode     int
		contentType    string
		body           string
		assertErr      func(t *testing.T, err error)
	}{
		{
			name:        "OK no system errors",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			body: `
{
  "errors": null
}`,
		},
		{
			name:        "OK with system errors",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			body: `
{
  "errors": [
    {
      "level": 0,
      "message": "some error",
      "when": "2024-03-28T20:15:11+01:00"
    }
  ]
}`,
		},
		{
			name:        "empty body",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			assertErr: func(t *testing.T, err error) {
				var syntaxErr *json.SyntaxError
				assert.ErrorAs(t, err, &syntaxErr)
			},
		},
		{
			name:        "has error",
			statusCode:  http.StatusInternalServerError,
			contentType: "application/json",
			body:        `{ "error": "some error message" }`,
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err,
					"system errors: unexpected syncthing error: some error message")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, tt.contentType, tt.statusCode, tt.body)
			sysErrors, err := c.SystemErrors(context.Background())
			if tt.assertErr != nil {
				t.Log(err)
				require.Error(t, err)
				tt.assertErr(t, err)
			} else {
				require.NoError(t, err)
				if tt.body != "" {
					var want api.SystemErrors
					require.NoError(t, json.Unmarshal([]byte(tt.body), &want))
					assert.Equal(t, want.Errors, sysErrors)
				}
			}
		})
	}
}

func TestClient_FolderErrors(t *testing.T) {
	tests := []struct {
		name           string
		folder, device string
		statusCode     int
		contentType    string
		body           string
		assertErr      func(t *testing.T, err error)
	}{
		{
			name:        "OK no folder errors",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			body: `
{
  "errors": null,
  "folder": "default",
  "page": 1,
  "perpage": 65536
}`,
		},
		{
			name:        "OK with folder errors",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			body: `
{
  "errors": [
    {
      "error": "some error",
      "path": "/some file path"
    }
  ],
  "folder": "default",
  "page": 1,
  "perpage": 65536
}`,
		},
		{
			name:        "empty body",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			assertErr: func(t *testing.T, err error) {
				var syntaxErr *json.SyntaxError
				assert.ErrorAs(t, err, &syntaxErr)
			},
		},
		{
			name:        "has error",
			statusCode:  http.StatusInternalServerError,
			contentType: "application/json",
			body:        `{ "error": "some error message" }`,
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err,
					"folder errors: unexpected syncthing error: some error message")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, tt.contentType, tt.statusCode, tt.body)
			folderErrors, err := c.FolderErrors(context.Background(), "default")
			if tt.assertErr != nil {
				t.Log(err)
				require.Error(t, err)
				tt.assertErr(t, err)
			} else {
				require.NoError(t, err)
				if tt.body != "" {
					var want api.FolderErrors
					require.NoError(t, json.Unmarshal([]byte(tt.body), &want))
					assert.Equal(t, want.Errors, folderErrors)
				}
			}
		})
	}
}
