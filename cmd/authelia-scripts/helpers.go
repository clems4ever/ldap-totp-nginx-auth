package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/authelia/authelia/internal/utils"
)

func getXFlags(branch, build, extra string) (flags []string, err error) {
	if branch == "" {
		out, _, err := utils.RunCommandAndReturnOutput("git rev-parse --abbrev-ref HEAD")
		if err != nil {
			return flags, fmt.Errorf("error getting branch with git rev-parse: %w", err)
		}

		if out == "" {
			branch = "master"
		} else {
			branch = out
		}
	}

	gitTagCommit, _, err := utils.RunCommandAndReturnOutput("git rev-list --tags --max-count=1")
	if err != nil {
		return flags, fmt.Errorf("error getting tag commit with git rev-list: %w", err)
	}

	tag, _, err := utils.RunCommandAndReturnOutput(fmt.Sprintf("git describe --tags --abbrev=0 %s", gitTagCommit))
	if err != nil {
		return flags, fmt.Errorf("error getting tag with git describe: %w", err)
	}

	commit, _, err := utils.RunCommandAndReturnOutput("git rev-parse HEAD")
	if err != nil {
		return flags, fmt.Errorf("error getting commit with git rev-parse: %w", err)
	}

	var states []string

	if gitTagCommit == commit {
		states = append(states, "tagged")
	} else {
		states = append(states, "untagged")
	}

	if _, exitCode, _ := utils.RunCommandAndReturnOutput("git diff --quiet"); exitCode != 0 {
		states = append(states, "dirty")
	} else {
		states = append(states, "clean")
	}

	if build == "" {
		build = "manual"
	}

	return []string{
		fmt.Sprintf(fmtLDFLAGSX, "BuildBranch", branch),
		fmt.Sprintf(fmtLDFLAGSX, "BuildTag", tag),
		fmt.Sprintf(fmtLDFLAGSX, "BuildCommit", commit),
		fmt.Sprintf(fmtLDFLAGSX, "BuildDate", time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700")),
		fmt.Sprintf(fmtLDFLAGSX, "BuildState", strings.Join(states, " ")),
		fmt.Sprintf(fmtLDFLAGSX, "BuildExtra", extra),
		fmt.Sprintf(fmtLDFLAGSX, "BuildNumber", build),
	}, nil
}
