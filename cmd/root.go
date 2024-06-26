package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	dotenv "github.com/dsh2dsh/expx-dotenv"
	"github.com/spf13/cobra"

	"github.com/dsh2dsh/check_syncthing/client"
	"github.com/dsh2dsh/check_syncthing/client/api"
)

const fetchProcs = 8

var (
	apiKey, baseURL string
	excludeDevices  []string

	rootCmd = cobra.Command{
		Use:   "check_syncthing",
		Short: "Monitoring plugin for syncthing daemon.",
		Long: `This plugin monitors syncthing daemon by using its REST API.

Requires server URL and API key using flags or environment variables
SYNCTHING_API_KEY and SYNCTHING_URL. Environment variables can be configured
inside .env file in current dir.`,

		PersistentPreRunE: persistentPreRunE,
	}
)

func init() {
	rootCmd.PersistentFlags().StringArrayVarP(&excludeDevices, "exclude", "x",
		[]string{}, "short IDs of devices to exclude")
	rootCmd.PersistentFlags().StringVarP(&apiKey, "key", "k", "",
		"syncthing REST API key")
	rootCmd.PersistentFlags().StringVarP(&baseURL, "url", "u", "", "server URL")
	rootCmd.AddCommand(&foldersCmd)
	rootCmd.AddCommand(&healthCmd)
	rootCmd.AddCommand(&lastSeenCmd)
}

func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func persistentPreRunE(cmd *cobra.Command, args []string) error {
	// Don't show usage on app errors.
	// https://github.com/spf13/cobra/issues/340#issuecomment-378726225
	cmd.SilenceUsage = true

	if err := loadEnvs(); err != nil {
		return err
	}
	return rootValidate()
}

func loadEnvs() error {
	cfg := struct {
		Key string `env:"SYNCTHING_API_KEY"`
		URL string `env:"SYNCTHING_URL"`
	}{}
	err := dotenv.New().WithDepth(1).Load(func() error { return env.Parse(&cfg) })
	if err != nil {
		return fmt.Errorf("load .env: %w", err)
	}

	if apiKey == "" {
		apiKey = cfg.Key
	}
	if baseURL == "" {
		baseURL = cfg.URL
	}
	return nil
}

func rootValidate() error {
	if baseURL == "" {
		return errors.New("empty server URL")
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("validate %q: %w", baseURL, err)
	} else if !u.IsAbs() {
		return fmt.Errorf("url %q not absolute", u.String())
	}
	baseURL = u.String()
	return nil
}

func mustAPIClient() *client.Client {
	c, err := client.New(baseURL)
	if err != nil {
		cobra.CheckErr(fmt.Errorf("create syncthing REST API client: %w", err))
	}
	return c.WithKey(apiKey)
}

// --------------------------------------------------

func deviceName(id, name string) string {
	return newDeviceId(id).Short() + " (" + name + ")"
}

func newDeviceId(id string) deviceId {
	return deviceId(id)
}

type deviceId string

func (self deviceId) Short() string {
	shortId, _, _ := strings.Cut(string(self), "-")
	return shortId
}

// --------------------------------------------------

func newLookupDeviceId(ids []string) lookupDeviceId {
	l := lookupDeviceId{devices: make(map[string]bool, len(ids))}
	for _, id := range ids {
		l.Add(id)
	}
	return l
}

type lookupDeviceId struct {
	devices  map[string]bool
	excluded []string
}

func (self *lookupDeviceId) Add(id string) {
	self.devices[self.shortId(id)] = false
}

func (self *lookupDeviceId) shortId(id string) string {
	return newDeviceId(id).Short()
}

func (self *lookupDeviceId) Has(id string) bool {
	shortId := self.shortId(id)
	seen, ok := self.devices[shortId]
	if ok && !seen {
		self.excluded = append(self.excluded, id)
		self.devices[shortId] = true
	}
	return ok
}

func (self *lookupDeviceId) Excluded() bool {
	return len(self.excluded) > 0
}

func (self *lookupDeviceId) ExcludedString(
	devices map[string]api.DeviceConfiguration,
) string {
	excluded := make([]string, len(self.excluded))
	for i, id := range self.excluded {
		excluded[i] = deviceName(id, devices[id].Name)
	}
	return strings.Join(excluded, ", ")
}
