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

	appCmd "github.com/CycloneDX/cyclonedx-gomod/internal/cli/cmd/app"
	binCmd "github.com/CycloneDX/cyclonedx-gomod/internal/cli/cmd/bin"
	modCmd "github.com/CycloneDX/cyclonedx-gomod/internal/cli/cmd/mod"
	versionCmd "github.com/CycloneDX/cyclonedx-gomod/internal/cli/cmd/version"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func New() *ffcli.Command {
	return &ffcli.Command{
		Name:       "cyclonedx-gomod",
		ShortUsage: "cyclonedx-gomod <SUBCOMMAND> [FLAGS...] [<ARG>...]",
		LongHelp: `cyclonedx-gomod creates CycloneDX Software Bill of Materials (SBOM) from Go modules.

Multiple subcommands are offered, each targeting different use cases:

- SBOMs generated with "app" include only those modules that the target application
  actually depends on. Modules required by tests or packages that are not imported
  by the application are not included. Build constraints are evaluated, which enables
  a very detailed view of what's really compiled into an application's binary.
  
- SBOMs generated with "mod" include the aggregate of modules required by all 
  packages in the target module. This optionally includes modules required by
  tests and test packages. Build constraints are NOT evaluated, allowing for 
  a "whole picture" view on the target module's dependencies.

- "bin" offers support of generating rudimentary SBOMs from binaries built with Go modules.

Distributors of applications will typically use "app" and provide the resulting SBOMs
alongside their application's binaries. This enables users to only consume SBOMs for
artifacts that they actually use. For example, a Go module may include "server" and
"client" applications, of which only the "client" is distributed to users. 
Additionally, modules included in "client" may differ, depending on which platform 
it was compiled for.

Vendors or maintainers may choose to use "mod" for internal use, where it's too
cumbersome to deal with many SBOMs for the same product. Possible use cases are: 
- Tracking of component inventory
- Tracking of third party component licenses
- Continuous monitoring for vulnerabilities
"mod" may also be used to generate SBOMs for libraries.`,
		Subcommands: []*ffcli.Command{
			appCmd.New(),
			binCmd.New(),
			modCmd.New(),
			versionCmd.New(),
		},
		Exec: func(_ context.Context, _ []string) error {
			return execRootCmd()
		},
	}
}

func execRootCmd() error {
	return flag.ErrHelp
}
