package main

import (
	"github.com/spf13/cobra"
)

func init() {
	publishCmd.PersistentFlags().String("spkg-registry", "https://spkg.io", "Substreams package registry")
	publishCmd.PersistentFlags().Bool("local-development", false, "Set local development")

	rootCmd.AddCommand(publishCmd)
}

var publishCmd = &cobra.Command{
	Use:   "publish [github_release_url | https_spkg_path | local_spkg_path | local_substreams_path]",
	Short: "Publish a package to the Substreams.dev registry. Alias for `substreams registry publish`",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRegistryPublish,
}
