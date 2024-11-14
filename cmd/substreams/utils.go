package main

import (
	"fmt"
	"github.com/spf13/cobra"
	pbsubstreams "github.com/streamingfast/substreams/pb/sf/substreams/v1"
	"strconv"
	"strings"
)

func readStartBlockFlag(cmd *cobra.Command, flagName string) (int64, bool, error) {
	val, err := cmd.Flags().GetString(flagName)
	if err != nil {
		panic(fmt.Sprintf("flags: couldn't find flag %q", flagName))
	}
	if val == "" {
		return 0, true, nil
	}

	startBlock, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, false, fmt.Errorf("start block is invalid: %w", err)
	}

	return startBlock, false, nil
}

func readStopBlockFlag(cmd *cobra.Command, startBlock int64, flagName string, withCursor bool) (uint64, error) {
	val, err := cmd.Flags().GetString(flagName)
	if err != nil {
		panic(fmt.Sprintf("flags: couldn't find flag %q", flagName))
	}

	isRelative := strings.HasPrefix(val, "+")
	if isRelative {
		if withCursor {
			return 0, fmt.Errorf("relative stop block is not supported with a cursor")
		}

		if startBlock < 0 {
			return 0, fmt.Errorf("relative end block is supported only with an absolute start block")
		}

		val = strings.TrimPrefix(val, "+")
	}

	endBlock, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("end block is invalid: %w", err)
	}

	if isRelative {
		return uint64(startBlock) + endBlock, nil
	}

	return endBlock, nil
}

func warnIncompletePackage(spkg *pbsubstreams.Package) {
	if len(spkg.PackageMeta) > 0 {
		if spkg.PackageMeta[0].Doc == "" {
			fmt.Println(warningStyle.Render("Warning: README not found for this package."))
		}

		if spkg.PackageMeta[0].Url == "" {
			fmt.Println(warningStyle.Render("Warning: URL is not set for this package."))
		}

		if spkg.PackageMeta[0].Description == "" {
			fmt.Println(warningStyle.Render("Warning: Description is not set for this package."))
		}
	}

	if spkg.Network == "" {
		fmt.Println(warningStyle.Render("Warning: Network is not set for this package."))
	}

	if spkg.Image == nil {
		fmt.Println(warningStyle.Render("Warning: Image is not set for this package."))
	}

	fmt.Println()
}

func printPackageDetails(spkg *pbsubstreams.Package) {
	fmt.Println()
	fmt.Println(headerStyle.Render("Package Details"))
	fmt.Printf("%s: %s\n", purpleStyle.Render("Name"), spkg.PackageMeta[0].Name)
	fmt.Printf("%s: %s\n", purpleStyle.Render("Version"), spkg.PackageMeta[0].Version)
	fmt.Printf("%s: %s\n", purpleStyle.Render("URL"), spkg.PackageMeta[0].Url)
	fmt.Println()
}