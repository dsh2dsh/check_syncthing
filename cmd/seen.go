package cmd

import (
	"context"
	"strings"
	"time"

	"github.com/dsh2dsh/go-monitoringplugin/v2"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/dsh2dsh/check_syncthing/client"
	"github.com/dsh2dsh/check_syncthing/client/api"
)

const seenOkMsg = "oldest last seen: "

var warnLastSeen, critLastSeen time.Duration

var lastSeenCmd = cobra.Command{
	Use:   "last-seen",
	Short: "Check last seen time of syncthing clients",
	Long: `Check last seen time of syncthing clients.

It lookups a syncthing client with oldest last seen time and outputs warning or
critical status if it's out of given thresholds.`,

	Run: func(cmd *cobra.Command, args []string) {
		NewLastSeenCheck(mustAPIClient()).
			WithExcludeDevices(excludeDevices).
			WithThresholds(warnLastSeen, critLastSeen).
			Run().Response().OutputAndExit()
	},
}

func init() {
	lastSeenCmd.Flags().DurationVarP(&warnLastSeen, "warn", "w", 5*time.Minute,
		"warning threshold")
	lastSeenCmd.Flags().DurationVarP(&critLastSeen, "crit", "c", 15*time.Minute,
		"critical threshold")
}

func NewLastSeenCheck(apiClient *client.Client) *LastSeenCheck {
	c := &LastSeenCheck{
		client:        apiClient,
		warnThreshold: warnLastSeen,
		critThreshold: critLastSeen,
	}
	return c.applyOptions()
}

type LastSeenCheck struct {
	client *client.Client
	resp   *monitoringplugin.Response

	excludeDevices lookupDeviceId
	warnThreshold  time.Duration
	critThreshold  time.Duration

	devices map[string]api.DeviceConfiguration
	stats   map[string]api.DeviceStatistics
	system  *api.SystemStatus

	excluded []string
}

func (self *LastSeenCheck) applyOptions() *LastSeenCheck {
	if self.resp == nil {
		self.resp = monitoringplugin.NewResponse(seenOkMsg)
	}
	return self
}

func (self *LastSeenCheck) WithExcludeDevices(devices []string,
) *LastSeenCheck {
	self.excludeDevices = newLookupDeviceId(devices)
	self.excluded = make([]string, 0, len(devices))
	return self
}

func (self *LastSeenCheck) WithThresholds(warn, crit time.Duration,
) *LastSeenCheck {
	self.warnThreshold = warn
	self.critThreshold = crit
	return self
}

func (self *LastSeenCheck) Response() *monitoringplugin.Response {
	return self.resp
}

func (self *LastSeenCheck) Run() *LastSeenCheck {
	if !self.fetch(context.Background()) {
		return self
	}

	deviceId, lastSeen := self.oldest()
	if self.resp.UpdateStatusIf(lastSeen.IsZero(), monitoringplugin.WARNING,
		"never seen device "+self.deviceName(deviceId),
	) {
		return self
	}

	lastSeenSecs := time.Since(lastSeen).Truncate(time.Second)
	self.resp.WithDefaultOkMessage(seenOkMsg + lastSeenSecs.String() + " ago")

	point := monitoringplugin.NewPerformanceDataPoint(
		"last seen", lastSeenSecs.Seconds()).SetUnit("s")
	point.NewThresholds(0, warnLastSeen.Seconds(), 0, critLastSeen.Seconds())
	if err := self.resp.AddPerformanceDataPoint(point); err != nil {
		self.resp.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "", true)
	} else {
		self.output(deviceId, lastSeenSecs)
	}
	return self
}

func (self *LastSeenCheck) deviceName(id string) string {
	return deviceName(id, self.devices[id].Name)
}

func (self *LastSeenCheck) fetch(parentCtx context.Context) bool {
	g, ctx := errgroup.WithContext(parentCtx)
	g.SetLimit(fetchProcs)

	g.Go(func() error { return self.fetchSystemStatus(ctx) })
	g.Go(func() error { return self.fetchDevices(ctx) })
	g.Go(func() error { return self.fetchStats(ctx) })

	self.resp.UpdateStatusOnError(g.Wait(), monitoringplugin.CRITICAL, "", true)
	return self.resp.GetStatusCode() == monitoringplugin.OK
}

func (self *LastSeenCheck) fetchSystemStatus(ctx context.Context) error {
	system, err := self.client.SystemStatus(ctx)
	if err != nil {
		return err
	}
	self.system = system
	return nil
}

func (self *LastSeenCheck) fetchDevices(ctx context.Context) error {
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

func (self *LastSeenCheck) fetchStats(ctx context.Context) error {
	stats, err := self.client.DeviceStats(ctx)
	if err == nil {
		self.stats = stats
	}
	return err
}

func (self *LastSeenCheck) oldest() (deviceId string, seen time.Time) {
	for id, stat := range self.stats {
		if self.excludeDevices.Has(id) {
			self.excluded = append(self.excluded, id)
		} else if id != self.system.MyID {
			if seen.IsZero() || stat.LastSeen.Before(seen) {
				deviceId = id
				if stat.LastSeen.Unix() == 0 {
					break
				}
				seen = stat.LastSeen
			}
		}
	}
	return
}

func (self *LastSeenCheck) output(deviceId string, lastSeen time.Duration) {
	self.resp.UpdateStatus(self.resp.GetStatusCode(),
		"device: "+self.deviceName(deviceId))

	if self.resp.GetStatusCode() != monitoringplugin.OK {
		self.resp.UpdateStatus(self.resp.GetStatusCode(),
			"last seen: "+lastSeen.String()+" ago")
		self.outputThreshold()
	}
	self.outputExcluded()
}

func (self *LastSeenCheck) outputThreshold() {
	var s string
	if self.resp.GetStatusCode() == monitoringplugin.WARNING {
		s = self.warnThreshold.String()
	} else {
		s = self.critThreshold.String()
	}
	self.resp.UpdateStatus(self.resp.GetStatusCode(), "threshold: "+s)
}

func (self *LastSeenCheck) outputExcluded() {
	if len(self.excluded) == 0 {
		return
	}

	ex := make([]string, len(self.excluded))
	for i, id := range self.excluded {
		ex[i] = self.deviceName(id)
	}
	self.resp.UpdateStatus(
		self.resp.GetStatusCode(), "excluded: "+strings.Join(ex, ", "))
}
