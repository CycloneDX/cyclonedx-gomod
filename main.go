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

package main

import (
	"context"
	"fmt"
	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
	"github.com/rs/zerolog/log"
	"os"

	"github.com/CycloneDX/cyclonedx-gomod/internal/cli"
)

func main() {
	err := cli.New().ParseAndRun(context.Background(), os.Args[1:])
	if err != nil {
		if _, ok := err.(*options.ValidationError); ok {
			_, _ = fmt.Fprintln(os.Stderr, err)
		} else {
			log.Err(err).Msg("")
		}
		os.Exit(1)
	}
}
