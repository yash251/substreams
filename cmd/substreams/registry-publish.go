package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/tidwall/gjson"
	"go.uber.org/zap"

	"github.com/spf13/cobra"
)

func init() {
	registryPublish.PersistentFlags().String("registry", "https://api.substreams.dev", "Substreams dev endpoint")

	registryCmd.AddCommand(registryPublish)
}

var registryPublish = &cobra.Command{
	Use:   "publish [github_release_url | https_spkg_path | local_spkg_path | local_substreams_path]",
	Short: "Publish a package to the Substreams.dev registry",
	Args:  cobra.ExactArgs(1),
	RunE:  runRegistryPublish,
}

func runRegistryPublish(cmd *cobra.Command, args []string) error {
	spkgReleasePath := args[0]

	var apiKey string
	registryTokenBytes, err := os.ReadFile(registryTokenFilename)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read registry token: %w", err)
		}
	}

	substreamsRegistryToken := os.Getenv("SUBSTREAMS_REGISTRY_TOKEN")
	apiKey = string(registryTokenBytes)
	if apiKey == "" || substreamsRegistryToken != "" {
		apiKey = substreamsRegistryToken
	}

	zlog.Debug("loaded api key", zap.String("api_key", apiKey))

	/// todo: accept local spkg path, remote spkg or local_substreams_path
	org, err := getOrganizationFromGithubUrl(spkgReleasePath)
	if err != nil {
		return err
	}

	// if local -> check if valid spkg file
	// if not, return error
	// if valid, send request

	request := &publishRequest{
		OrganizationSlug: slugify(org),
		GithubUrl:        spkgReleasePath,
	}
	jsonRequest, _ := json.Marshal(request)
	requestBody := bytes.NewBuffer(jsonRequest)

	apiEndpoint, err := cmd.Flags().GetString("registry")
	if err != nil {
		return err
	}

	publishPackageEndpoint := fmt.Sprintf("%s/sf.substreams.dev.Api/PublishPackage", apiEndpoint)
	zlog.Debug("publishing package", zap.String("registry_url", publishPackageEndpoint))

	req, err := http.NewRequest("POST", publishPackageEndpoint, requestBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body")
	}

	if resp.StatusCode != http.StatusOK {
		msg := gjson.Get(string(b), "message").String()
		fmt.Println("Failed to publish package")
		fmt.Printf("\tReason: %s\n", msg)
		return nil
	}

	spkgUrlPath := gjson.Get(string(b), "spkgLink").String()

	fmt.Println("Package published successfully")
	if spkgUrlPath != "" {
		fmt.Printf("Start streaming your package with: `substreams gui %s`\n", spkgUrlPath)
	}

	return nil
}

type publishRequest struct {
	//todo: remove this, it will be the user id
	OrganizationSlug string `json:"organization_slug"`
	// change this to spkg bytes
	GithubUrl string `json:"github_url"`
}

func getOrganizationFromGithubUrl(url string) (string, error) {
	if !strings.Contains(url, "github.com") {
		return "", fmt.Errorf("invalid github url")
	}

	parts := strings.Split(url, "/")
	for i, part := range parts {
		if part == "github.com" && i < len(parts)-1 {
			return strings.ToLower(parts[i+1]), nil
		}
	}

	return "", fmt.Errorf("organization name not found in github url")
}

func slugify(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "-")
}
