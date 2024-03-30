package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/caarlos0/env/v10"
	dotenv "github.com/dsh2dsh/expx-dotenv"
	"github.com/spf13/cobra"

	"github.com/dsh2dsh/check_syncthing/client"
)

const fetchProcs = 8

var (
	apiKey, baseURL string

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
	rootCmd.PersistentFlags().StringVarP(&apiKey, "key", "k", "",
		"syncthing REST API key")
	rootCmd.PersistentFlags().StringVarP(&baseURL, "url", "u", "", "server URL")
	rootCmd.AddCommand(&healthCmd)
}

func Execute() {
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
