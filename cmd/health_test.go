package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dsh2dsh/go-monitoringplugin/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsh2dsh/check_syncthing/client"
	"github.com/dsh2dsh/check_syncthing/client/api"
)

func TestNewHealthCheck(t *testing.T) {
	check := testNewHealthCheck(t, nil)
	resp := check.Response()
	require.NotNil(t, resp)
	assert.Equal(t, "OK: "+healthOkMsg, resp.GetInfo().RawOutput)
}

func testNewHealthCheck(t *testing.T, endpoints map[string]string,
) *HealthCheck {
	check := NewHealthCheck(newTestClient(t, fakeAPI(t, endpoints)))
	require.NotNil(t, check)
	return check
}

func newTestClient(t *testing.T, httpDoer testHttpDoer) *client.Client {
	c, err := client.New("/", func(c *client.Client) error {
		return c.NewClientWithResponses(api.WithHTTPClient(httpDoer))
	})
	require.NoError(t, err)
	return c
}

type testHttpDoer func(req *http.Request) (*http.Response, error)

func (self testHttpDoer) Do(req *http.Request) (*http.Response, error) {
	return self(req)
}

func fakeAPI(t *testing.T, endpoints map[string]string,
) func(req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		require.NotNil(t, req.URL)
		t.Log(req.URL)
		json, ok := endpoints[req.URL.String()]
		r := httptest.NewRecorder()
		if !ok {
			r.WriteHeader(http.StatusNotFound)
			return r.Result(), nil
		}
		r.Header().Set("Content-Type", "application/json")
		r.WriteHeader(http.StatusOK)
		_, err := r.WriteString(json)
		require.NoError(t, err)
		return r.Result(), nil
	}
}

func TestHealthCheck_applyOptionsWithResp(t *testing.T) {
	resp := monitoringplugin.NewResponse("def ok msg")
	check := &HealthCheck{resp: resp}
	require.Same(t, check, check.applyOptions())
	assert.Same(t, resp, check.Response())
}

func TestHealthCheck_Run(t *testing.T) {
	const healthJSON = `{ "status": "OK" }`
	const devicesJSON = `
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
]`
	const foldersJSON = `
[
  {
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
  }
]`
	const systemStatusJSON = `
{
 "alloc": 35216400,
  "connectionServiceStatus": {
    "quic://0.0.0.0:22000": {
      "error": null,
      "lanAddresses": [
        "quic://0.0.0.0:22000",
        "quic://127.0.0.1:22000"
      ],
      "wanAddresses": [
        "quic://0.0.0.0:22000",
        "quic://10.0.0.1:59574"
      ]
    },
    "tcp://0.0.0.0:22000": {
      "error": null,
      "lanAddresses": [
        "tcp://0.0.0.0:22000",
        "tcp://127.0.0.1:22000"
      ],
      "wanAddresses": [
        "tcp://0.0.0.0:0",
        "tcp://0.0.0.0:22000"
      ]
    }
  },
  "cpuPercent": 0,
  "discoveryEnabled": true,
  "discoveryErrors": {
    "IPv6 local": "listen udp6: socket: protocol not supported"
  },
  "discoveryMethods": 2,
  "discoveryStatus": {
    "IPv4 local": {
      "error": null
    },
    "IPv6 local": {
      "error": "listen udp6: socket: protocol not supported"
    }
  },
  "goroutines": 151,
  "guiAddressOverridden": false,
  "guiAddressUsed": "127.0.0.1:8384",
  "lastDialStatus": {},
  "MyID": "XXXXXX1-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXX1",
  "pathSeparator": "/",
  "startTime": "2024-03-29T06:02:42+01:00",
  "sys": 137501992,
  "tilde": "/",
  "uptime": 1333881,
  "urVersionMax": 3
}`
	const systemErrorsJSON = `{ "errors": null }`
	const folderErrorJSON = `
{
  "errors": null,
  "folder": "default",
  "page": 1,
  "perpage": 65536
}`

	tests := []struct {
		name         string
		endpoints    map[string]string
		assertOutput func(t *testing.T, rawOutput string)
	}{
		{
			name: "crit health",
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "CRITICAL: health:")
			},
		},
		{
			name: "crit with health",
			endpoints: map[string]string{
				"/rest/noauth/health": healthJSON,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "CRITICAL: ")
				assert.Contains(t, rawOutput, "404 Not Found")
			},
		},
		{
			name: "crit with health and folders",
			endpoints: map[string]string{
				"/rest/noauth/health":  healthJSON,
				"/rest/config/folders": foldersJSON,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "CRITICAL: ")
				assert.Contains(t, rawOutput, "404 Not Found")
			},
		},
		{
			name: "crit without devices and system status",
			endpoints: map[string]string{
				"/rest/noauth/health":  healthJSON,
				"/rest/config/folders": foldersJSON,
				"/rest/system/error":   systemErrorsJSON,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "CRITICAL: ")
				assert.Contains(t, rawOutput, "404 Not Found")
			},
		},
		{
			name: "crit without system status",
			endpoints: map[string]string{
				"/rest/noauth/health":  healthJSON,
				"/rest/config/devices": devicesJSON,
				"/rest/config/folders": foldersJSON,
				"/rest/system/error":   systemErrorsJSON,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "CRITICAL: ")
				assert.Contains(t, rawOutput, "404 Not Found")
			},
		},
		{
			name: "crit without folder error",
			endpoints: map[string]string{
				"/rest/noauth/health":  healthJSON,
				"/rest/config/devices": devicesJSON,
				"/rest/config/folders": foldersJSON,
				"/rest/system/status":  systemStatusJSON,
				"/rest/system/error":   systemErrorsJSON,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput,
					`CRITICAL: folder id="default", label="Default Folder": folder errors:`)
			},
		},
		{
			name: "OK alive",
			endpoints: map[string]string{
				"/rest/noauth/health":                healthJSON,
				"/rest/config/devices":               devicesJSON,
				"/rest/config/folders":               foldersJSON,
				"/rest/system/status":                systemStatusJSON,
				"/rest/system/error":                 systemErrorsJSON,
				"/rest/folder/errors?folder=default": folderErrorJSON,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput,
					"OK: syncthing server alive: XXXXXX1 (some device)")
			},
		},
		{
			name: "with default folder error",
			endpoints: map[string]string{
				"/rest/noauth/health":  healthJSON,
				"/rest/config/devices": devicesJSON,
				"/rest/config/folders": foldersJSON,
				"/rest/system/status":  systemStatusJSON,
				"/rest/system/error":   systemErrorsJSON,
				"/rest/folder/errors?folder=default": `
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
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, `WARNING: 1/1 folders with errors
folder: default (Default Folder)
path: /some file path
error: some error`)
			},
		},
		{
			name: "with system error",
			endpoints: map[string]string{
				"/rest/noauth/health":  healthJSON,
				"/rest/config/devices": devicesJSON,
				"/rest/config/folders": foldersJSON,
				"/rest/system/status":  systemStatusJSON,
				"/rest/system/error": `
{
  "errors": [
    {
      "level": 0,
      "message": "some error",
      "when": "2024-03-28T20:15:11+01:00"
    }
  ]
}`,
				"/rest/folder/errors?folder=default": folderErrorJSON,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, `WARNING: 1 system error(s): some error
last error at: 2024-03-28 20:15:11`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := testNewHealthCheck(t, tt.endpoints)
			require.Same(t, check, check.Run())
			resp := check.Response()
			require.NotNil(t, resp)
			t.Log(resp.GetInfo().RawOutput)
			tt.assertOutput(t, resp.GetInfo().RawOutput)
		})
	}
}
