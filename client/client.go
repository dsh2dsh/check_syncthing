//go:generate oapi-codegen -config api/http.gen.yaml api/syncthing.yaml
//go:generate oapi-codegen -config api/models.gen.yaml api/syncthing.yaml
package client

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/dsh2dsh/check_syncthing/client/api"
)

const authHeader = "X-API-Key"

func New(baseURL string, opts ...Option) (*Client, error) {
	c := &Client{baseURL: baseURL}
	return c, c.applyOpts(opts)
}

type Client struct {
	apiClient api.ClientWithResponsesInterface
	apiKey    string
	baseURL   string
}

func (self *Client) WithKey(apiKey string) *Client {
	self.apiKey = apiKey
	return self
}

func (self *Client) NewClientWithResponses(moreOpts ...api.ClientOption) error {
	opts := [...]api.ClientOption{
		api.WithRequestEditorFn(self.withAPIKey),
	}

	apiClient, err := api.NewClientWithResponses(self.baseURL,
		slices.Concat(opts[:], moreOpts)...)
	if err != nil {
		return fmt.Errorf("new client for %q: %w", self.baseURL, err)
	}
	self.apiClient = apiClient
	return nil
}

func (self *Client) withAPIKey(ctx context.Context, req *http.Request) error {
	req.Header.Set(authHeader, self.apiKey)
	return nil
}

// --------------------------------------------------

func (self *Client) Health(ctx context.Context) error {
	r, err := self.apiClient.HealthWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("health request: %w", err)
	}

	if r.JSON200 == nil {
		return fmt.Errorf("health: %w", makeAPIError(r.JSONDefault,
			r.Status(), r.Body))
	} else if r.JSON200.Status != "OK" {
		return fmt.Errorf("health: unexpected status: %v", r.JSON200.Status)
	}
	return nil
}

func makeAPIError(jsonErr *api.Error, status string, body []byte) error {
	if jsonErr != nil {
		return fmt.Errorf("unexpected syncthing error: %v", jsonErr.Error)
	}
	return fmt.Errorf("unexpected syncthing response: %v (%v)", status,
		string(body))
}

func (self *Client) Connections(ctx context.Context) (*api.Connections, error) {
	r, err := self.apiClient.ConnectionsWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("connections request: %w", err)
	}

	if r.JSON200 == nil {
		return nil, fmt.Errorf("connections: %w", makeAPIError(r.JSONDefault,
			r.Status(), r.Body))
	}
	return r.JSON200, nil
}

func (self *Client) Folders(ctx context.Context) ([]api.FolderConfiguration,
	error,
) {
	r, err := self.apiClient.FoldersWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("folders request: %w", err)
	}

	if r.JSON200 == nil {
		return nil, fmt.Errorf("folders: %w", makeAPIError(r.JSONDefault,
			r.Status(), r.Body))
	}
	return *r.JSON200, nil
}

func (self *Client) Devices(ctx context.Context) ([]api.DeviceConfiguration,
	error,
) {
	r, err := self.apiClient.DevicesWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("devices request: %w", err)
	}

	if r.JSON200 == nil {
		return nil, fmt.Errorf("devices: %w", makeAPIError(r.JSONDefault,
			r.Status(), r.Body))
	}
	return *r.JSON200, nil
}

func (self *Client) DeviceStats(ctx context.Context) (
	map[string]api.DeviceStatistics, error,
) {
	r, err := self.apiClient.DeviceStatsWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("device stats request: %w", err)
	}

	if r.JSON200 == nil {
		return nil, fmt.Errorf("device stats: %w", makeAPIError(r.JSONDefault,
			r.Status(), r.Body))
	}
	return *r.JSON200, nil
}

func (self *Client) Completion(ctx context.Context, folder, device string,
) (*api.FolderCompletion, error) {
	params := api.CompletionParams{Folder: folder, Device: device}
	r, err := self.apiClient.CompletionWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("completion request: %w", err)
	}

	if r.JSON200 == nil {
		return nil, fmt.Errorf("completion: %w", makeAPIError(r.JSONDefault,
			r.Status(), r.Body))
	}
	return r.JSON200, nil
}
