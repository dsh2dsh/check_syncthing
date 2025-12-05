package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dsh2dsh/go-monitoringplugin/v2"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/dsh2dsh/check_syncthing/client"
	"github.com/dsh2dsh/check_syncthing/client/api"
)

const foldersOkMsg = " syncthing folders" // N syncthing folders

var foldersCmd = cobra.Command{
	Use:   "folders",
	Short: "Check status of syncthing folders",
	Long: `Check status of syncthing folders.

Checks for any folder error and completion status of all clients.`,

	Run: func(cmd *cobra.Command, args []string) {
		NewFoldersCheck(mustAPIClient()).WithExcludeDevices(excludeDevices).
			Run().Response().OutputAndExit()
	},
}

func NewFoldersCheck(apiClient *client.Client) *FoldersCheck {
	c := &FoldersCheck{client: apiClient}
	return c.applyOptions()
}

type FoldersCheck struct {
	client *client.Client
	resp   *monitoringplugin.Response

	excludeDevices lookupDeviceId

	devices      map[string]api.DeviceConfiguration
	folders      []api.FolderConfiguration
	folderErrors []api.FileError
	completions  []*api.FolderCompletion
}

func (self *FoldersCheck) applyOptions() *FoldersCheck {
	if self.resp == nil {
		self.resp = monitoringplugin.NewResponse(foldersOkMsg)
	}
	return self
}

func (self *FoldersCheck) WithExcludeDevices(devices []string) *FoldersCheck {
	self.excludeDevices = newLookupDeviceId(devices)
	return self
}

func (self *FoldersCheck) Response() *monitoringplugin.Response {
	return self.resp
}

func (self *FoldersCheck) Run() *FoldersCheck {
	if !self.fetch(context.Background()) {
		return self
	}

	self.resp.WithDefaultOkMessage(strconv.Itoa(len(self.folders)) + foldersOkMsg)
	if self.checkFolderErrors() {
		self.checkCompletions()
		self.checkNotShared()
	}
	self.outputExcluded()
	return self
}

func (self *FoldersCheck) fetch(parentCtx context.Context) bool {
	if !self.fetchDevicesFolders(parentCtx) {
		return false
	}

	g, ctx := errgroup.WithContext(parentCtx)
	g.SetLimit(fetchProcs)

	self.fetchFolderErrors(ctx, g)
	self.fetchCompletions(ctx, g)

	self.resp.UpdateStatusOnError(g.Wait(), monitoringplugin.CRITICAL, "", true)
	return self.resp.GetStatusCode() == monitoringplugin.OK
}

func (self *FoldersCheck) fetchDevicesFolders(parentCtx context.Context) bool {
	g, ctx := errgroup.WithContext(parentCtx)
	g.SetLimit(fetchProcs)

	g.Go(func() error { return self.fetchDevices(ctx) })
	g.Go(func() error { return self.fetchFolders(ctx) })

	self.resp.UpdateStatusOnError(g.Wait(), monitoringplugin.CRITICAL, "", true)
	return self.resp.GetStatusCode() == monitoringplugin.OK
}

func (self *FoldersCheck) fetchDevices(ctx context.Context) error {
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

func (self *FoldersCheck) fetchFolders(ctx context.Context) error {
	folders, err := self.client.Folders(ctx)
	if err != nil {
		return err
	}
	self.folders = folders
	return nil
}

func (self *FoldersCheck) fetchFolderErrors(ctx context.Context,
	g *errgroup.Group,
) {
	self.folderErrors = make([]api.FileError, len(self.folders))
	for i := range self.folders {
		if ctx.Err() != nil {
			break
		}
		folder := &self.folders[i]
		g.Go(func() error {
			folderError, err := self.folderError(ctx, folder.Id)
			self.folderErrors[i] = folderError
			return err
		})
	}
}

func (self *FoldersCheck) folderError(ctx context.Context, folderId string,
) (folderError api.FileError, err error) {
	fileErrors, err := self.client.FolderErrors(ctx, folderId)
	if err != nil {
		err = fmt.Errorf("folder id=%q: %w", folderId, err)
	} else if len(fileErrors) > 0 {
		folderError = fileErrors[len(fileErrors)-1]
	}
	return folderError, err
}

func (self *FoldersCheck) fetchCompletions(ctx context.Context,
	g *errgroup.Group,
) {
	self.completions = make([]*api.FolderCompletion, self.numCompletions())
	var nextIdx int
	for i := range self.folders {
		folder := &self.folders[i]
		for j := range folder.Devices {
			deviceId := folder.Devices[j].DeviceId
			if ctx.Err() != nil {
				return
			} else if self.excludeDevices.Has(deviceId) {
				continue
			}
			idx := nextIdx + j
			g.Go(func() error {
				comp, err := self.completion(ctx, folder, folder.Devices[j].DeviceId)
				self.completions[idx] = comp
				return err
			})
		}
		nextIdx += len(folder.Devices)
	}
}

func (self *FoldersCheck) numCompletions() int {
	var needComps int
	for i := range self.folders {
		f := &self.folders[i]
		needComps += len(f.Devices)
	}
	return needComps
}

func (self *FoldersCheck) completion(ctx context.Context,
	folder *api.FolderConfiguration, deviceId string,
) (*api.FolderCompletion, error) {
	comp, err := self.client.Completion(ctx, folder.Id, deviceId)
	if err != nil {
		err = fmt.Errorf("completion folder=%q, device=%q: %w",
			folderName(folder), self.deviceName(deviceId), err)
	}
	return comp, err
}

func (self *FoldersCheck) deviceName(id string) string {
	return deviceName(id, self.devices[id].Name)
}

func (self *FoldersCheck) checkFolderErrors() bool {
	var numErrors int
	for i := range self.folderErrors {
		if self.folderErrors[i].Error != "" {
			numErrors++
		}
	}
	if numErrors == 0 {
		return true
	}

	self.resp.UpdateStatus(monitoringplugin.WARNING, fmt.Sprintf(
		"%v/%v folders with errors", numErrors, len(self.folders)))
	for i := range self.folders {
		self.outputFolderError(&self.folders[i], &self.folderErrors[i])
	}
	return false
}

func (self *FoldersCheck) outputFolderError(folder *api.FolderConfiguration,
	folderError *api.FileError,
) {
	if folderError.Error == "" {
		return
	}

	self.resp.UpdateStatus(monitoringplugin.WARNING,
		"folder: "+folderName(folder))
	self.resp.UpdateStatus(monitoringplugin.WARNING,
		"path: "+folderError.Path)
	self.resp.UpdateStatus(monitoringplugin.WARNING,
		"error: "+folderError.Error)
}

func folderName(folder *api.FolderConfiguration) string {
	return folder.Id + " (" + folder.Label + ")"
}

func (self *FoldersCheck) checkCompletions() {
	foldersCnt, outOfSync := self.outOfSyncFolders()
	if foldersCnt == 0 {
		return
	}

	self.resp.UpdateStatus(monitoringplugin.WARNING, fmt.Sprintf(
		"%v/%v folders out of sync", foldersCnt, len(self.folders)))
	for i := range self.folders {
		if devices := outOfSync[i]; devices != "" {
			self.outputOutOfSync(&self.folders[i], devices)
		}
	}
}

func (self *FoldersCheck) outOfSyncFolders() (int, []string) {
	outOfSync := make([]string, len(self.folders))
	var idx, cnt int
	for i := range self.folders {
		folder := &self.folders[i]
		syncDevices := self.outOfSyncDevices(folder.Devices,
			self.completions[idx:idx+len(folder.Devices)])
		if len(syncDevices) > 0 {
			outOfSync[i] = strings.Join(syncDevices, ", ")
			cnt++
		}
		idx += len(folder.Devices)
	}
	return cnt, outOfSync
}

func (self *FoldersCheck) outOfSyncDevices(
	devices []api.FolderDeviceConfiguration, completions []*api.FolderCompletion,
) []string {
	outOfSync := make([]string, 0, len(devices))
	for i := range devices {
		comp := completions[i]
		if comp != nil {
			if pct := int(comp.Completion); pct < 100 {
				outOfSync = append(outOfSync, self.deviceName(devices[i].DeviceId)+
					" - "+strconv.Itoa(pct)+"%")
			}
		}
	}
	return outOfSync
}

func (self *FoldersCheck) outputOutOfSync(folder *api.FolderConfiguration,
	devices string,
) {
	self.resp.UpdateStatus(monitoringplugin.WARNING,
		"folder: "+folderName(folder))
	self.resp.UpdateStatus(monitoringplugin.WARNING, "device: "+devices)
}

func (self *FoldersCheck) checkNotShared() {
	notShared := make([]string, 0, len(self.folders))
	for i := range self.folders {
		folder := &self.folders[i]
		if len(folder.Devices) == 0 {
			notShared = append(notShared, folderName(folder))
		}
	}

	if len(notShared) > 0 {
		self.resp.UpdateStatus(monitoringplugin.WARNING,
			strconv.Itoa(len(notShared))+" folder not shared: "+
				strings.Join(notShared, ", "))
	}
}

func (self *FoldersCheck) outputExcluded() {
	if self.excludeDevices.Excluded() {
		self.resp.UpdateStatus(self.resp.GetStatusCode(),
			"excluded: "+self.excludeDevices.ExcludedString(self.devices))
	}
}
