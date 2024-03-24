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

	"github.com/dsh2dsh/check_syncthing/api"
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

func TestCache_WithKey(t *testing.T) {
	const key = "foobar"
	c, err := New("/")
	require.NoError(t, err)
	assert.Same(t, c, c.WithKey(key))
	assert.Equal(t, key, c.apiKey)
}

func TestCache_withAPIKey(t *testing.T) {
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

type testHttpDoer struct {
	do func(req *http.Request) (*http.Response, error)
}

func (self *testHttpDoer) Do(req *http.Request) (*http.Response, error) {
	return self.do(req)
}

func TestCache_Health(t *testing.T) {
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
			httpClient := testHttpDoer{
				func(req *http.Request) (*http.Response, error) {
					r := httptest.NewRecorder()
					r.Header().Set("Content-Type", tt.contentType)
					r.WriteHeader(tt.statusCode)
					_, err := r.WriteString(tt.body)
					require.NoError(t, err)
					return r.Result(), nil
				},
			}

			c, err := New("/", func(self *Client) error {
				return self.NewClientWithResponses(api.WithHTTPClient(&httpClient))
			})
			require.NoError(t, err)
			require.NotNil(t, c)

			err = c.Health(context.Background())
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

func TestCache_Connections(t *testing.T) {
	const conn1 = "XXXXXX1-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXX1"
	const conn2 = "XXXXXX2-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXXX-XXXXXX2"

	tests := []struct {
		name        string
		statusCode  int
		contentType string
		body        string
		assertErr   func(t *testing.T, err error)
		assert      func(t *testing.T, conn *api.Connections)
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
			assert: func(t *testing.T, conn *api.Connections) {
				require.NotEmpty(t, conn.Connections)
				assert.Len(t, conn.Connections, 2)
				assert.Contains(t, conn.Connections, conn1)
				assert.NotZero(t, conn.Connections[conn1])
				assert.Contains(t, conn.Connections, conn2)
				assert.Zero(t, conn.Connections[conn2])
				assert.NotZero(t, conn.Total)
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
			httpClient := testHttpDoer{
				func(req *http.Request) (*http.Response, error) {
					r := httptest.NewRecorder()
					r.Header().Set("Content-Type", tt.contentType)
					r.WriteHeader(tt.statusCode)
					_, err := r.WriteString(tt.body)
					require.NoError(t, err)
					return r.Result(), nil
				},
			}

			c, err := New("/", func(self *Client) error {
				return self.NewClientWithResponses(api.WithHTTPClient(&httpClient))
			})
			require.NoError(t, err)
			require.NotNil(t, c)

			conn, err := c.Connections(context.Background())
			if tt.assertErr != nil {
				t.Log(err)
				require.Error(t, err)
				tt.assertErr(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, conn)
				if tt.assert != nil {
					tt.assert(t, conn)
				}
			}
		})
	}
}
