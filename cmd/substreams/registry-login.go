package main

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
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

	linkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	fmt.Printf("Login to the Substreams registry.")
	fmt.Println()
	fmt.Println()
	fmt.Println("Navigate to: ")
	fmt.Println()
	fmt.Println("    " + linkStyle.Render(fmt.Sprintf("%s/me", registryURL)))
	fmt.Println("")

	var token string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				EchoMode(huh.EchoModePassword).
				Title("Paste the token here:").
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

	isFileExists := checkFileExists(registryTokenFilename)
	if isFileExists {
		var confirmOverwrite bool
		form = huh.NewForm(
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
