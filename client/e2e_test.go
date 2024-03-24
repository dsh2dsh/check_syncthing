//go:build e2e

package client

import (
	"context"
	"testing"

	"github.com/caarlos0/env/v10"
	dotenv "github.com/dsh2dsh/expx-dotenv"
	"github.com/stretchr/testify/suite"
)

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

type ClientTestSuite struct {
	suite.Suite
	client *Client
}

func (self *ClientTestSuite) SetupSuite() {
	cfg := struct {
		Key string `env:"SYNCTHING_API_KEY,notEmpty"`
		URL string `env:"SYNCTHING_URL,notEmpty"`
	}{}
	self.Require().NoError(dotenv.Load(func() error { return env.Parse(&cfg) }))

	client, err := New(cfg.URL)
	self.Require().NoError(err)
	self.Require().NotNil(client)
	self.client = client.WithKey(cfg.Key)
}

func (self *ClientTestSuite) TestHealth() {
	self.Require().NoError(self.client.Health(context.Background()))
}

func (self *ClientTestSuite) TestConnections() {
	conns, err := self.client.Connections(context.Background())
	self.Require().NoError(err)
	self.Require().NotEmpty(conns)
}

func (self *ClientTestSuite) TestFolders() {
	folders, err := self.client.Folders(context.Background())
	self.Require().NoError(err)
	self.NotEmpty(folders)
}

func (self *ClientTestSuite) TestDevices() {
	devices, err := self.client.Devices(context.Background())
	self.Require().NoError(err)
	self.NotEmpty(devices)
}

func (self *ClientTestSuite) TestDeviceStats() {
	stats, err := self.client.DeviceStats(context.Background())
	self.Require().NoError(err)
	self.NotEmpty(stats)
}

func (self *ClientTestSuite) TestCompletion() {
	comp, err := self.client.Completion(context.Background(), "", "")
	self.Require().NoError(err)
	self.NotEmpty(comp)
}
