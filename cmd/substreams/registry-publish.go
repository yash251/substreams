package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/streamingfast/cli/sflags"
	"github.com/streamingfast/substreams/manifest"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func init() {
	registryPublish.PersistentFlags().String("spkg-registry", "https://spkg.io", "Substreams package registry")
	registryPublish.PersistentFlags().Bool("local-development", false, "Set local development")

	registryCmd.AddCommand(registryPublish)
}

var registryPublish = &cobra.Command{
	Use:   "publish [github_release_url | https_spkg_path | local_spkg_path | local_substreams_path]",
	Short: "Publish a package to the Substreams.dev registry",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRegistryPublish,
}

// FLOW:
// - The user get an API_KEY (registry token) on substreams.dev
// - Set API_KEY :
// 	- If the user doesn't have the API_KEY SET FOR REGISTRY, let's redirect him to `substreams.dev` and grab a registry token
// 	- If it has one already, use it
// - SET UP Publish request :
// 	- If the user does the command on a manifest, pack it first
//  - If the user does provide an spkg, use it as is
//  - If the user does provide a github release url, download the spkg and pack it

func runRegistryPublish(cmd *cobra.Command, args []string) error {
	apiEndpoint := "https://substreams.dev"
	isLocal := sflags.MustGetBool(cmd, "local-development")
	if isLocal {
		apiEndpoint = "http://localhost:9000"
	}

	var apiKey string
	registryTokenBytes, err := os.ReadFile(registryTokenFilename)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read registry token: %w", err)
		}
	}

	substreamsRegistryToken := os.Getenv("SUBSTREAMS_REGISTRY_TOKEN")
	apiKey = string(registryTokenBytes)
	if apiKey == "" {
		if substreamsRegistryToken != "" {
			apiKey = substreamsRegistryToken
		} else {
			linkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
			//Let's redirect the user to substreams.dev to get a registry token and let them paste in the terminal
			fmt.Println("`SUBSTREAMS_REGISTRY_TOKEN` env variable is missing...")
			fmt.Println("You can get a token using the following link: ")
			fmt.Println()
			fmt.Println("    " + linkStyle.Render(fmt.Sprintf("%s/me", apiEndpoint)))
			fmt.Println("")

			var token string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						EchoMode(huh.EchoModePassword).
						Title("After retrieving your registry token, paste it here:").
						Inline(true).
						Value(&token).
						Validate(func(s string) error {
							if s == "" {
								return errors.New("token cannot be empty")
							}
							return nil
						}),
				),
			)

			if err := form.Run(); err != nil {
				return fmt.Errorf("error running form: %w", err)
			}

			// Set the API_KEY using the input token
			apiKey = token
		}
	}


	zlog.Debug("loaded api key", zap.String("api_key", apiKey))

	var manifestPath string
	switch len(args) {
	case 0:
		manifestPath, err = resolveManifestFile("")
		if err != nil {
			return fmt.Errorf("resolving manifest: %w", err)
		}
	case 1:
		manifestPath = args[0]
	}

	readerOptions := []manifest.Option{
		manifest.WithRegistryURL(sflags.MustGetString(cmd, "spkg-registry")),
	}

	manifestReader, err := manifest.NewReader(manifestPath, readerOptions...)
	if err != nil {
		return fmt.Errorf("manifest reader: %w", err)
	}

	pkgBundle, err := manifestReader.Read()
	if err != nil {
		return fmt.Errorf("read manifest %q: %w", manifestPath, err)
	}

	spkg := pkgBundle.Package

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Create form file to get it read from the `substreams.dev`  server

	formFile, err := writer.CreateFormFile("file", "substreams_package")
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	data, err := proto.Marshal(spkg)
	if err != nil {
		return fmt.Errorf("marshalling substreams package: %w", err)
	}

	_, err = formFile.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write file content: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	publishPackageEndpoint := fmt.Sprintf("%s/sf.substreams.dev.Api/PublishPackage", apiEndpoint)

	zlog.Debug("publishing package", zap.String("registry_url", publishPackageEndpoint))

	req, err := http.NewRequest("POST", publishPackageEndpoint, &requestBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
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
		linkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		fmt.Println("")
		fmt.Println(linkStyle.Render("Failed to publish package")+ "\n")
		fmt.Println("Reason:" + string(b))
		return nil
	}

	spkgUrlPath := gjson.Get(string(b), "spkgLink").String()

	fmt.Println("Package published successfully")
	if spkgUrlPath != "" {
		fmt.Printf("Start streaming your package with: `substreams gui %s`\n", spkgUrlPath)
	}

	return nil
}

