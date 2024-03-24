package client

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_applyOpts(t *testing.T) {
	c := Client{}

	var callCount int
	require.NoError(t, c.applyOpts([]Option{
		func(*Client) error {
			callCount++
			return nil
		},
	}))
	assert.Equal(t, 1, callCount)
	assert.NotNil(t, c.apiClient)

	c = Client{}
	wantErr := errors.New("test error")
	require.ErrorIs(t, c.applyOpts([]Option{
		func(*Client) error { return wantErr },
	}), wantErr)
	assert.Nil(t, c.apiClient)

	c = Client{}
	require.NoError(t, c.withDefaultClient())
	apiClient := c.apiClient
	require.NoError(t, c.applyOpts([]Option{
		func(*Client) error {
			c.apiClient = apiClient
			return nil
		},
	}))
	assert.Same(t, c.apiClient, apiClient)
}
