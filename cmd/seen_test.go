package cmd

import (
	"fmt"
	"testing"
	"time"

	"github.com/dsh2dsh/go-monitoringplugin/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsh2dsh/check_syncthing/client/api"
)

func TestNewLastSeenCheck(t *testing.T) {
	check := testNewLastSeenCheck(t, nil)
	resp := check.Response()
	require.NotNil(t, resp)
	assert.Equal(t, "OK: "+seenOkMsg, resp.GetInfo().RawOutput)
	assert.Equal(t, warnLastSeen, check.warnThreshold)
	assert.Equal(t, critLastSeen, check.critThreshold)
}

func testNewLastSeenCheck(t *testing.T, endpoints map[string]any,
) *LastSeenCheck {
	check := NewLastSeenCheck(newTestClient(t, fakeAPI(t, endpoints)))
	require.NotNil(t, check)
	return check
}

func TestLastSeenCheck_applyOptionsWithResp(t *testing.T) {
	resp := monitoringplugin.NewResponse("def ok msg")
	check := &LastSeenCheck{resp: resp}
	require.Same(t, check, check.applyOptions())
	assert.Same(t, resp, check.Response())
}

func TestLastSeenCheck_WithExcludeDevices(t *testing.T) {
	check := testNewLastSeenCheck(t, nil)
	assert.Same(t, check, check.WithExcludeDevices([]string{"A", "B"}))
}

func TestLastSeenCheck_WithThresholds(t *testing.T) {
	check := testNewLastSeenCheck(t, nil)
	assert.Same(t, check, check.WithThresholds(time.Second, time.Minute))
	assert.Equal(t, time.Second, check.warnThreshold)
	assert.Equal(t, time.Minute, check.critThreshold)
}

func TestLastSeenCheck_Run(t *testing.T) {
	const testId2 = "XXXXXX2-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX"
	const testId3 = "XXXXXX3-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX"

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
  "myID": "XXXXXX1-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX",
  "pathSeparator": "/",
  "startTime": "2024-03-29T06:02:42+01:00",
  "sys": 137501992,
  "tilde": "/",
  "uptime": 1333881,
  "urVersionMax": 3
}`

	const devicesJSON = `
[
  {
    "deviceID": "XXXXXX1-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX",
    "name": "server",
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
  },
  {
    "deviceID": "XXXXXX2-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX",
    "name": "client",
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
  },
  {
    "deviceID": "XXXXXX3-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX",
    "name": "client",
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

	const deviceStatsJSON = `
{
  "XXXXXX1-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX": {
    "lastSeen": "2024-03-28T20:15:11+01:00",
    "lastConnectionDurationS": 451.813963185
  },
  "XXXXXX2-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX": {
    "lastSeen": "2024-03-28T20:15:11+01:00",
    "lastConnectionDurationS": 451.813963185
  }
}`

	tests := []struct {
		name         string
		with         func(t *testing.T, check *LastSeenCheck)
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
			name: "with system status",
			endpoints: map[string]any{
				"/rest/system/status": systemStatusJSON,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "CRITICAL: ")
			},
		},
		{
			name: "with devices",
			endpoints: map[string]any{
				"/rest/system/status":  systemStatusJSON,
				"/rest/config/devices": devicesJSON,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "CRITICAL: ")
			},
		},
		{
			name: "never seen",
			endpoints: map[string]any{
				"/rest/system/status":  systemStatusJSON,
				"/rest/config/devices": devicesJSON,
				"/rest/stats/device": `
{
  "XXXXXX2-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX": {}
}`,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Equal(t, "WARNING: never seen device XXXXXX2 (client)", rawOutput)
			},
		},
		{
			name: "never seen by unix",
			endpoints: map[string]any{
				"/rest/system/status":  systemStatusJSON,
				"/rest/config/devices": devicesJSON,
				"/rest/stats/device": `
{
  "XXXXXX2-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX": {
    "lastSeen": "1970-01-01T01:00:00+01:00"
  }
}`,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Equal(t, "WARNING: never seen device XXXXXX2 (client)", rawOutput)
			},
		},
		{
			name: "warning threshold",
			endpoints: map[string]any{
				"/rest/system/status":  systemStatusJSON,
				"/rest/config/devices": devicesJSON,
				"/rest/stats/device": map[string]api.DeviceStatistics{
					testId2: {LastSeen: time.Now().Add(-warnLastSeen - time.Second)},
				},
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "WARNING: last seen is outside")
				assert.Contains(t, rawOutput, "device: XXXXXX2 (client)")
				assert.Contains(t, rawOutput, fmt.Sprintf("threshold: %v", warnLastSeen))
			},
		},
		{
			name: "critical threshold",
			endpoints: map[string]any{
				"/rest/system/status":  systemStatusJSON,
				"/rest/config/devices": devicesJSON,
				"/rest/stats/device":   deviceStatsJSON,
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "CRITICAL: last seen is outside")
				assert.Contains(t, rawOutput, "device: XXXXXX2 (client)")
				assert.Contains(t, rawOutput, fmt.Sprintf("threshold: %v", critLastSeen))
			},
		},
		{
			name: "OK",
			endpoints: map[string]any{
				"/rest/system/status":  systemStatusJSON,
				"/rest/config/devices": devicesJSON,
				"/rest/stats/device": map[string]api.DeviceStatistics{
					testId2: {LastSeen: time.Now()},
				},
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "OK: oldest last seen: 0s ago")
				assert.Contains(t, rawOutput, fmt.Sprintf(
					"device: XXXXXX2 (client) | 'last seen'=0s;%v;%v;;",
					warnLastSeen.Seconds(), critLastSeen.Seconds()))
			},
		},
		{
			name: "OK with ExcludeDevices",
			with: func(t *testing.T, check *LastSeenCheck) {
				check.WithExcludeDevices([]string{"XXX"})
			},
			endpoints: map[string]any{
				"/rest/system/status":  systemStatusJSON,
				"/rest/config/devices": devicesJSON,
				"/rest/stats/device": map[string]api.DeviceStatistics{
					testId2: {LastSeen: time.Now()},
				},
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "OK: oldest last seen: 0s ago")
				assert.Contains(t, rawOutput, fmt.Sprintf(
					"device: XXXXXX2 (client) | 'last seen'=0s;%v;%v;;",
					warnLastSeen.Seconds(), critLastSeen.Seconds()))
				assert.NotContains(t, rawOutput, "excluded:")
			},
		},
		{
			name: "OK with excluded",
			with: func(t *testing.T, check *LastSeenCheck) {
				check.WithExcludeDevices([]string{testId3})
			},
			endpoints: map[string]any{
				"/rest/system/status":  systemStatusJSON,
				"/rest/config/devices": devicesJSON,
				"/rest/stats/device": map[string]api.DeviceStatistics{
					testId2: {LastSeen: time.Now()},
					testId3: {LastSeen: time.Now()},
				},
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput, "OK: oldest last seen: 0s ago")
				assert.Contains(t, rawOutput, "device: XXXXXX2 (client)")
				assert.Contains(t, rawOutput, fmt.Sprintf(
					"excluded: XXXXXX3 (client) | 'last seen'=0s;%v;%v;;",
					warnLastSeen.Seconds(), critLastSeen.Seconds()))
			},
		},
		{
			name: "data point error",
			with: func(t *testing.T, check *LastSeenCheck) {
				point := monitoringplugin.NewPerformanceDataPoint("last seen", 0).
					SetUnit("s")
				require.NoError(t, check.Response().AddPerformanceDataPoint(point))
			},
			endpoints: map[string]any{
				"/rest/system/status":  systemStatusJSON,
				"/rest/config/devices": devicesJSON,
				"/rest/stats/device": map[string]api.DeviceStatistics{
					testId2: {LastSeen: time.Now()},
				},
			},
			assertOutput: func(t *testing.T, rawOutput string) {
				assert.Contains(t, rawOutput,
					"UNKNOWN: failed to add performance data point:")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := testNewLastSeenCheck(t, tt.endpoints)
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
