package main

import "github.com/spf13/cobra"

func init() {
	publishCmd.PersistentFlags().String("registry", "https://api.substreams.dev", "Substreams dev endpoint")

	rootCmd.AddCommand(publishCmd)
}

var publishCmd = &cobra.Command{
	Use:   "publish <github_release_url>",
	Short: "Publish a package to the Substreams.dev registry. Alias for `substreams registry publish`",
	Args:  cobra.ExactArgs(1),
	RunE:  runREgistryPublish,
}
