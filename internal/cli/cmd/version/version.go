// This file is part of CycloneDX GoMod
//
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
// Copyright (c) OWASP Foundation. All Rights Reserved.

package version

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/CycloneDX/cyclonedx-gomod/internal/version"
	"io"
	"os"
	"text/template"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func New() *ffcli.Command {
	fs := flag.NewFlagSet("cyclonedx-gomod version", flag.ExitOnError)

	var useJSON bool
	fs.BoolVar(&useJSON, "json", false, "Output in JSON")

	return &ffcli.Command{
		Name:       "version",
		ShortHelp:  "Show version information",
		ShortUsage: "cyclonedx-gomod version",
		FlagSet:    fs,
		Exec: func(_ context.Context, _ []string) error {
			return execVersionCmd(os.Stdout, useJSON)
		},
	}
}

var outputTmpl = template.Must(template.New("").Parse(`Version:{{"\t"}}{{ .Version }}
{{ if .ModuleSum -}}
ModuleSum:{{"\t"}}{{ .ModuleSum }}
{{ end -}}
{{ if .Commit -}}
Commit:{{"\t\t"}}{{ .Commit }}
CommitDate:{{"\t"}}{{ .CommitDate }}
Modified:{{"\t"}}{{ .Modified }}
{{ end -}}
GoVersion:{{"\t"}}{{ .GoVersion }}
OS:{{"\t\t"}}{{ .OS }}
Arch:{{"\t\t"}}{{ .Arch }}
`))

func execVersionCmd(writer io.Writer, useJSON bool) error {
	if useJSON {
		enc := json.NewEncoder(writer)
		enc.SetIndent("", "  ")
		return enc.Encode(version.Info)
	}

	return outputTmpl.Execute(writer, version.Info)
}
