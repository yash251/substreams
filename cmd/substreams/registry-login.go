package main

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var registryLoginCmd = &cobra.Command{
	Use:          "login",
	Short:        "Login to the Substreams registry",
	SilenceUsage: true,
	RunE:         runRegistryLoginE,
}

var registryTokenFilename = filepath.Join(os.Getenv("HOME"), ".config", "substreams", "registry-token")

func init() {
	registryCmd.AddCommand(registryLoginCmd)
}

func runRegistryLoginE(cmd *cobra.Command, args []string) error {
	registryURL := "https://substreams.dev"
	if newValue := os.Getenv("SUBSTREAMS_REGISTRY_ENDPOINT"); newValue != "" {
		registryURL = newValue
	}

	token, err := copyPasteTokenForm(registryURL)
	if err != nil {
		return fmt.Errorf("creating copy, paste token form %w", err)
	}

	isFileExists := checkFileExists(registryTokenFilename)
	if isFileExists {
		var confirmOverwrite bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Token already saved to registry-token").
					Value(&confirmOverwrite).
					Affirmative("Yes").
					Negative("No"),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("error running form: %w", err)
		}

		if confirmOverwrite {
			err := writeRegistryToken(token)
			if err != nil {
				return fmt.Errorf("could not write token to registry: %w", err)
			}
		} else {
			return nil
		}

	} else {
		err := writeRegistryToken(token)
		if err != nil {
			return fmt.Errorf("could not write token to registry: %w", err)
		}

	}

	fmt.Printf("All set! Token written to ~/.config/substreams/registry-token")

	return nil
}

func writeRegistryToken(token string) error {
	return os.WriteFile(registryTokenFilename, []byte(token), 0644)
}

func checkFileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}
