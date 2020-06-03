// This code is released under the MIT License
// Copyright (c) 2020 Marco Molteni and the tryrelease contributors.

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"
	"time"
)

var (
	// Filled by the linker.
	fullVersion = "unknown" // example: v0.0.9-8-g941583d027-dirty
	semver      = "unknown" // example: v0.0.9
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
		checkGitHubVersion(out, "marco-m", "tryrelease", semver)
		return 0
	}

	fmt.Printf("OS: %s\nArchitecture: %s\n", runtime.GOOS, runtime.GOARCH)

	return 0
}

// https://developer.github.com/v3/repos/releases/#get-the-latest-release
func checkGitHubVersion(out io.Writer, owner, repo, currVersion string) error {
	// ==> vagrant: A new version of Vagrant is available: 2.2.9 (installed version: 2.2.8)!
	// ==> vagrant: To upgrade visit: https://www.vagrantup.com/downloads.html
	fmt.Fprintln(out, semver)

	// API: GET /repos/:owner/:repo/releases/latest
	url := "https://api.github.com" + path.Join("/repos", owner, repo, "releases/latest")

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create http request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http client Do: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("http reading response: %w", err)
	}

	fmt.Fprintln(out, string(respBody))

	return nil
}
