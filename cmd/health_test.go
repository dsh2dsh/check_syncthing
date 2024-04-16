package cmd

import (
	"encoding/json"
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

func testNewHealthCheck(t *testing.T, endpoints map[string]any,
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

func fakeAPI(t *testing.T, endpoints map[string]any,
) func(req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		require.NotNil(t, req.URL)
		t.Log(req.URL)
		v, ok := endpoints[req.URL.String()]
		r := httptest.NewRecorder()
		if !ok {
			r.WriteHeader(http.StatusNotFound)
			return r.Result(), nil
		}
		r.Header().Set("Content-Type", "application/json")
		r.WriteHeader(http.StatusOK)
		var b []byte
		switch v := v.(type) {
		case string:
			b = []byte(v)
		default:
			b2, err := json.Marshal(v)
			require.NoError(t, err)
			b = b2
		}
		_, err := r.Write(b)
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

	tests := []struct {
		name         string
		endpoints    map[string]any
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
			endpoints: map[string]any{
				"/rest/noauth/health": healthJSON,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "CRITICAL: ")
				assert.Contains(t, rawOutput, "404 Not Found")
			},
		},
		{
			name: "crit without devices and system status",
			endpoints: map[string]any{
				"/rest/noauth/health": healthJSON,
				"/rest/system/error":  systemErrorsJSON,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "CRITICAL: ")
				assert.Contains(t, rawOutput, "404 Not Found")
			},
		},
		{
			name: "crit without system status",
			endpoints: map[string]any{
				"/rest/noauth/health":  healthJSON,
				"/rest/config/devices": devicesJSON,
				"/rest/system/error":   systemErrorsJSON,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "CRITICAL: ")
				assert.Contains(t, rawOutput, "404 Not Found")
			},
		},
		{
			name: "OK alive",
			endpoints: map[string]any{
				"/rest/noauth/health":  healthJSON,
				"/rest/config/devices": devicesJSON,
				"/rest/system/status":  systemStatusJSON,
				"/rest/system/error":   systemErrorsJSON,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput,
					"OK: syncthing server alive: XXXXXX1 (some device)")
			},
		},
		{
			name: "with system error",
			endpoints: map[string]any{
				"/rest/noauth/health":  healthJSON,
				"/rest/config/devices": devicesJSON,
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
