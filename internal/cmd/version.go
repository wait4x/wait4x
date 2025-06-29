// Copyright 2019-2025 The Wait4X Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cmd provides the command-line interface for the Wait4X application.
package cmd

import (
	"bytes"
	"fmt"
	"runtime"
	"text/template"

	"github.com/spf13/cobra"
)

var versionTemplate = `Version:           {{.AppVersion}}
Go version:        {{.GoVersion}}
Git commit:        {{.GitCommit}}
Built:             {{.BuildTime}}
OS/Arch:           {{.GoOs}}/{{.GoArch}}`

var (
	// AppVersion represents Wait4X version
	AppVersion = "$Format:%(describe:tags=true)$"
	// GitCommit represents Wait4X commit sha1 hash from git, output of $(git rev-parse HEAD)
	GitCommit = "$Format:%H$"
	// BuildTime represents Wait4X build time in ISO8601 format, output of $(date -u '+%FT%TZ')
	BuildTime = "1970-01-01T00:00:00Z"
)

// Version represents some information which useful in version sub-command
type Version struct {
	AppVersion string
	GoVersion  string
	GoOs       string
	GoArch     string
	GitCommit  string
	BuildTime  string
}

// NewVersionCommand creates a new version sub-command
func NewVersionCommand() *cobra.Command {
	versionCommand := &cobra.Command{
		Use:   "version",
		Short: "Show Wait4X version information",
		Long:  "Display detailed version information about the Wait4X application",
		RunE:  runVersion,
	}

	return versionCommand
}

// runVersion runs the version command
func runVersion(_ *cobra.Command, _ []string) error {
	versionValues := Version{
		AppVersion: AppVersion,
		GoVersion:  runtime.Version(),
		GoOs:       runtime.GOOS,
		GoArch:     runtime.GOARCH,
		GitCommit:  GitCommit,
		BuildTime:  BuildTime,
	}
	var tmplBytes bytes.Buffer

	t := template.Must(template.New("version").Parse(versionTemplate))
	err := t.Execute(&tmplBytes, versionValues)
	if err != nil {
		return fmt.Errorf("unable to parse version template: %w", err)
	}

	fmt.Println(tmplBytes.String())
	return nil
}
