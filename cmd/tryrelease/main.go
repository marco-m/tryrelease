// This code is released under the MIT License
// Copyright (c) 2020 Marco Molteni and the tryrelease contributors.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/marco-m/tryrelease/pkg/release"
)

var (
	// Filled by the linker.
	fullVersion  = "unknown" // example: v0.0.9-8-g941583d027-dirty
	shortVersion = "unknown" // example: v0.0.9
)

func main() {
	os.Exit(run("tryrelease", os.Args[1:], os.Stderr))
}

func run(progname string, args []string, out io.Writer) int {
	flag.CommandLine.SetOutput(out)
	flagSet := flag.NewFlagSet(progname, flag.ContinueOnError)
	flagSet.Usage = func() {
		fmt.Fprintf(out, "%s some experiments with GitHub release API\n\n", progname)
		fmt.Fprintf(out, "Usage: %s\n\n", progname)
		fmt.Fprintf(out, "Options:\n")
		flagSet.PrintDefaults()
	}

	var (
		showVersion  = flagSet.Bool("version", false, "show version")
		checkVersion = flagSet.Bool("check-version", false, "check online if new version is available")
	)
	if flagSet.Parse(args) != nil {
		return 2
	}
	if *showVersion {
		fmt.Fprintln(out, "tryrelease version", fullVersion)
		return 0
	}
	if *checkVersion {
		human_url := fmt.Sprintf("https://github.com/%s/%s", "marco-m", "tryrelease")
		latestVersion, err := release.GitHubLatest("marco-m", "tryrelease")
		if err != nil {
			fmt.Fprintln(out, err)
			return 1
		}
		result, err := release.Compare(shortVersion, latestVersion)
		if err != nil {
			fmt.Fprintln(out, err)
			return 1
		}
		switch result {
		case 0:
			fmt.Fprintf(out, "installed version %s is the same as the latest version %s\n",
				shortVersion, latestVersion)
		case -1:
			fmt.Fprintf(out, "installed version %s is older than the latest version %s\n",
				shortVersion, latestVersion)
			fmt.Fprintln(out, "To upgrade visit", human_url)
		case +1:
			fmt.Fprintf(out, "(unexpected?) installed version %s is newer than the latest version %s\n",
				shortVersion, latestVersion)
		}
		return 0
	}

	fmt.Printf("OS: %s\nArchitecture: %s\n", runtime.GOOS, runtime.GOARCH)

	return 0
}
