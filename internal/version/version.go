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
	"golang.org/x/mod/module"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"
)

var Info = struct {
	Version    string
	ModuleSum  string     `json:",omitempty"`
	Commit     string     `json:",omitempty"`
	CommitDate *time.Time `json:",omitempty"`
	Modified   *bool      `json:",omitempty"`
	GoVersion  string
	OS         string
	Arch       string
}{
	Version: "v0.0.0-unknown",
}

func init() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	Info.Commit = buildSetting(bi, "vcs.revision")
	if vcsDate, err := time.Parse(time.RFC3339, buildSetting(bi, "vcs.time")); err == nil {
		Info.CommitDate = &vcsDate
	}
	if vcsModified, err := strconv.ParseBool(buildSetting(bi, "vcs.modified")); err == nil {
		Info.Modified = &vcsModified
	}

	Info.GoVersion = bi.GoVersion
	Info.OS = runtime.GOOS
	Info.Arch = runtime.GOARCH

	if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		Info.Version = bi.Main.Version
	} else if Info.Commit != "" && Info.CommitDate != nil {
		Info.Version = module.PseudoVersion("", "", *Info.CommitDate, Info.Commit[:12])
	}
	Info.ModuleSum = bi.Main.Sum
}

func buildSetting(bi *debug.BuildInfo, key string) (val string) {
	for _, setting := range bi.Settings {
		if setting.Key == key {
			val = setting.Value
			break
		}
	}

	return
}
