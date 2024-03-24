package client

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dsh2dsh/check_syncthing/api"
)

const httpTimeout = 15

type Option func(c *Client) error

func (self *Client) applyOpts(opts []Option) error {
	for _, fn := range opts {
		if err := fn(self); err != nil {
			return fmt.Errorf("apply option: %w", err)
		}
	}

	if self.apiClient == nil {
		if err := self.withDefaultClient(); err != nil {
			return err
		}
	}
	return nil
}

func (self *Client) withDefaultClient() error {
	httpClient := &http.Client{Timeout: httpTimeout * time.Second}
	err := self.NewClientWithResponses(api.WithHTTPClient(httpClient))
	if err != nil {
		return err
	}
	return nil
}
