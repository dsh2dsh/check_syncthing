package cmd

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/dsh2dsh/go-monitoringplugin/v2"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/dsh2dsh/check_syncthing/client"
	"github.com/dsh2dsh/check_syncthing/client/api"
)

const foldersOkMsg = " syncthing folders"

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
	excluded       []string

	folders      []api.FolderConfiguration
	folderErrors []folderErrors
}

type folderErrors struct {
	Id     string
	Label  string
	Errors []api.FileError
}

func (self *FoldersCheck) applyOptions() *FoldersCheck {
	if self.resp == nil {
		self.resp = monitoringplugin.NewResponse(foldersOkMsg)
	}
	return self
}

func (self *FoldersCheck) WithExcludeDevices(devices []string) *FoldersCheck {
	self.excludeDevices = newLookupDeviceId(devices)
	self.excluded = make([]string, 0, len(devices))
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
	if len(self.folderErrors) > 0 {
		self.checkFolderErrors(self.folderErrors)
	}
	return self
}

func (self *FoldersCheck) fetch(parentCtx context.Context) bool {
	g, ctx := errgroup.WithContext(parentCtx)
	g.SetLimit(fetchProcs)

	self.fetchFolderErrors(ctx, g)

	self.resp.UpdateStatusOnError(g.Wait(), monitoringplugin.CRITICAL, "", true)
	return self.resp.GetStatusCode() == monitoringplugin.OK
}

func (self *FoldersCheck) fetchFolderErrors(ctx context.Context,
	g *errgroup.Group,
) {
	folders, err := self.client.Folders(ctx)
	if err != nil || len(folders) == 0 {
		self.resp.UpdateStatusOnError(err, monitoringplugin.CRITICAL, "", true)
		return
	}
	self.folders = folders

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

func (self *FoldersCheck) checkFolderErrors(folders []folderErrors) {
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
		self.resp.UpdateStatus(monitoringplugin.WARNING,
			"folder: "+folderName(folder.Id, folder.Label))
		latestError := folder.Errors[len(folder.Errors)-1]
		self.resp.UpdateStatus(monitoringplugin.WARNING, "path: "+latestError.Path)
		self.resp.UpdateStatus(monitoringplugin.WARNING, "error: "+latestError.Error)
	}
}

func folderName(id, name string) string {
	return id + " (" + name + ")"
}
