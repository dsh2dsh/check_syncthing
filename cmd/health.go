package cmd

import (
	"context"
	"fmt"

	"github.com/dsh2dsh/go-monitoringplugin/v2"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/dsh2dsh/check_syncthing/client"
	"github.com/dsh2dsh/check_syncthing/client/api"
)

const healthOkMsg = "syncthing server alive: "

var healthCmd = cobra.Command{
	Use:   "health",
	Short: "Check health of syncthing server",
	Long: `Check health of syncthing server.

Checks syncthing server handles REST API requests and has no system errors. In
case of errors, outputs last system error.`,

	Run: func(cmd *cobra.Command, args []string) {
		NewHealthCheck(mustAPIClient()).Run().Response().OutputAndExit()
	},
}

func NewHealthCheck(apiClient *client.Client) *HealthCheck {
	c := &HealthCheck{client: apiClient}
	return c.applyOptions()
}

type HealthCheck struct {
	client *client.Client
	resp   *monitoringplugin.Response

	sysErrors []api.LogLine
	system    *api.SystemStatus
	devices   map[string]api.DeviceConfiguration
}

func (self *HealthCheck) applyOptions() *HealthCheck {
	if self.resp == nil {
		self.resp = monitoringplugin.NewResponse(healthOkMsg)
	}
	return self
}

func (self *HealthCheck) Response() *monitoringplugin.Response {
	return self.resp
}

func (self *HealthCheck) Run() *HealthCheck {
	ctx := context.Background()
	if !self.checkHealth(ctx) || !self.fetch(ctx) {
		return self
	}
	self.resp.WithDefaultOkMessage(healthOkMsg + self.systemName())

	if len(self.sysErrors) > 0 {
		self.checkSysErrors(self.sysErrors)
	}
	return self
}

func (self *HealthCheck) checkHealth(ctx context.Context) bool {
	return !self.resp.UpdateStatusOnError(self.client.Health(ctx),
		monitoringplugin.CRITICAL, "", true)
}

func (self *HealthCheck) fetch(parentCtx context.Context) bool {
	g, ctx := errgroup.WithContext(parentCtx)
	g.SetLimit(fetchProcs)

	g.Go(func() error { return self.fetchSysErrors(ctx) })
	g.Go(func() error { return self.fetchSystemStatus(ctx) })
	g.Go(func() error { return self.fetchDevices(ctx) })

	self.resp.UpdateStatusOnError(g.Wait(), monitoringplugin.CRITICAL, "", true)
	return self.resp.GetStatusCode() == monitoringplugin.OK
}

func (self *HealthCheck) fetchSysErrors(ctx context.Context) error {
	sysErrors, err := self.client.SystemErrors(ctx)
	if err != nil {
		return err
	} else if len(sysErrors) > 0 {
		self.sysErrors = sysErrors
	}
	return nil
}

func (self *HealthCheck) fetchSystemStatus(ctx context.Context) error {
	system, err := self.client.SystemStatus(ctx)
	if err != nil {
		return err
	}
	self.system = system
	return nil
}

func (self *HealthCheck) fetchDevices(ctx context.Context) error {
	devices, err := self.client.Devices(ctx)
	if err != nil {
		return err
	}

	self.devices = make(map[string]api.DeviceConfiguration, len(devices))
	for i := range devices {
		d := &devices[i]
		self.devices[d.DeviceID] = devices[i]
	}
	return nil
}

func (self *HealthCheck) systemName() string {
	return deviceName(self.system.MyID, self.devices[self.system.MyID].Name)
}

func (self *HealthCheck) checkSysErrors(sysErrors []api.LogLine) {
	lastError := sysErrors[len(sysErrors)-1]
	self.resp.UpdateStatus(monitoringplugin.WARNING,
		fmt.Sprintf("%v system error(s): %v", len(sysErrors),
			lastError.Message))
	self.resp.UpdateStatus(monitoringplugin.WARNING,
		"last error at: "+lastError.When.String())
}
