package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/streamingfast/cli"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage substreams registry",
	Long: cli.Dedent(`
		Login, publish and list packages from the Substreams registry
	`),
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(registryCmd)
}

func getSubstreamsRegistryEndpoint() string {
	endpoint := "https://substreams.dev"
	if newValue := os.Getenv("SUBSTREAMS_REGISTRY_ENDPOINT"); newValue != "" {
		fmt.Println("Using registry endpoint: " + newValue)
		endpoint = newValue
	}
	return endpoint
}

func getSubstreamsDownloadEndpoint() string {
	endpoint := "https://spkg.io"
	if newValue := os.Getenv("SUBSTREAMS_DOWNLOAD_ENDPOINT"); newValue != "" {
		fmt.Println("Using download endpoint: " + newValue)
		endpoint = newValue
	}
	return endpoint
}

func getSubstreamsCodegenEndpoint() string {
	endpoint := "https://codegen.substreams.dev"
	if newValue := os.Getenv("SUBSTREAMS_CODEGEN_ENDPOINT"); newValue != "" {
		fmt.Println("Using codegen endpoint: " + newValue)
		endpoint = newValue
	}
	return endpoint
}
