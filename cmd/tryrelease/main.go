// This code is released under the MIT License
// Copyright (c) 2020 Marco Molteni and the tryrelease contributors.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"time"

	"golang.org/x/mod/semver"
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
		if err := checkGitHubVersion(out, "marco-m", "tryrelease", shortVersion); err != nil {
			fmt.Fprintln(out, err)
		}
		return 0
	}

	fmt.Printf("OS: %s\nArchitecture: %s\n", runtime.GOOS, runtime.GOARCH)

	return 0
}

// https://developer.github.com/v3/repos/releases/#get-the-latest-release
func checkGitHubVersion(out io.Writer, owner, repo, currVersion string) error {
	// API: GET /repos/:owner/:repo/releases/latest
	api_url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	human_url := fmt.Sprintf("https://github.com/%s/%s", owner, repo)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api_url, nil)
	if err != nil {
		return fmt.Errorf("create http request: %w", err)
	}
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http client Do: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		fmt.Fprintln(out, "no release found at", human_url)
		return nil
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("http reading response: %w", err)
	}

	type Response struct {
		TagName string `json:"tag_name"`
	}
	var response Response
	if err := json.Unmarshal(respBody, &response); err != nil {
		return fmt.Errorf("parsing JSON response: %w", err)
	}
	if response.TagName == "" {
		return fmt.Errorf("parsing JSON response: missing field tag_name")
	}

	if !semver.IsValid(shortVersion) {
		return fmt.Errorf("installed version is not a valid semver: %s", shortVersion)
	}
	if !semver.IsValid(response.TagName) {
		return fmt.Errorf("fetched last version is not a valid semver: %s", response.TagName)
	}

	switch semver.Compare(shortVersion, response.TagName) {
	case 0:
		fmt.Fprintf(out, "installed version %s is the same as the latest version %s\n",
			shortVersion, response.TagName)
	case -1:
		fmt.Fprintf(out, "installed version %s is older than the latest version %s\n",
			shortVersion, response.TagName)
		fmt.Fprintln(out, "To upgrade visit", human_url)
	case +1:
		fmt.Fprintf(out, "(unexpected?) installed version %s is newer than the latest version %s\n",
			shortVersion, response.TagName)
	}

	return nil
}
