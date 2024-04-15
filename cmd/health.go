package cmd

import (
	"context"
	"fmt"
	"slices"

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

Checks syncthing servers handles REST API requests, has no system errors and no
folders with errors.

In case of errors, outputs last system error and last error for every folder
with errors.`,

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

	sysErrors    []api.LogLine
	folderErrors []folderErrors
	system       *api.SystemStatus
	devices      map[string]api.DeviceConfiguration
}

type folderErrors struct {
	Id     string
	Label  string
	Errors []api.FileError
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
	if len(self.folderErrors) > 0 {
		self.checkFolderErrors(self.folderErrors)
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
	self.fetchFolderErrors(ctx, g)

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

func (self *HealthCheck) fetchFolderErrors(ctx context.Context,
	g *errgroup.Group,
) {
	folders, err := self.client.Folders(ctx)
	if err != nil || len(folders) == 0 {
		self.resp.UpdateStatusOnError(err, monitoringplugin.CRITICAL, "", true)
		return
	}

	self.folderErrors = make([]folderErrors, len(folders))
	for i := range folders {
		if ctx.Err() != nil {
			break
		}
		folder := &folders[i]
		self.folderErrors[i] = folderErrors{Id: folder.Id, Label: folder.Label}
		g.Go(func() error {
			fileErrors, err := self.client.FolderErrors(ctx, folder.Id)
			if err != nil {
				return fmt.Errorf("folder id=%q, label=%q: %w", folder.Id,
					folder.Label, err)
			} else if len(fileErrors) > 0 {
				self.folderErrors[i].Errors = fileErrors
			}
			return nil
		})
	}
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

func (self *HealthCheck) checkFolderErrors(folders []folderErrors) {
	allFoldersNum := len(folders)
	folders = slices.DeleteFunc(folders, func(folder folderErrors) bool {
		return len(folder.Errors) == 0
	})
	if len(folders) == 0 {
		return
	}

	self.resp.UpdateStatus(monitoringplugin.WARNING, fmt.Sprintf(
		"%v/%v folders with errors", len(folders), allFoldersNum))

	for i := range folders {
		folder := &folders[i]
		self.resp.UpdateStatus(monitoringplugin.WARNING, fmt.Sprintf(
			"folder: %v (%v)", folder.Id, folder.Label))
		latestError := folder.Errors[len(folder.Errors)-1]
		self.resp.UpdateStatus(monitoringplugin.WARNING, "path: "+latestError.Path)
		self.resp.UpdateStatus(monitoringplugin.WARNING, "error: "+latestError.Error)
	}
}
