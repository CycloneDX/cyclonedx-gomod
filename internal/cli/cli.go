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

package cli

import (
	"context"
	"flag"

	appcmd "github.com/CycloneDX/cyclonedx-gomod/internal/cli/cmd/app"
	bincmd "github.com/CycloneDX/cyclonedx-gomod/internal/cli/cmd/bin"
	modcmd "github.com/CycloneDX/cyclonedx-gomod/internal/cli/cmd/mod"
	versioncmd "github.com/CycloneDX/cyclonedx-gomod/internal/cli/cmd/version"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func NewRootCmd() *ffcli.Command {
	return &ffcli.Command{
		Name:       "cyclonedx-gomod",
		ShortUsage: "cyclonedx-gomod <SUBCOMMAND> [FLAGS...] [<ARG>...]",
		Subcommands: []*ffcli.Command{
			appcmd.New(),
			bincmd.New(),
			modcmd.New(),
			versioncmd.New(),
		},
		Exec: func(_ context.Context, _ []string) error {
			return execRootCmd()
		},
	}
}

func execRootCmd() error {
	return flag.ErrHelp
}
