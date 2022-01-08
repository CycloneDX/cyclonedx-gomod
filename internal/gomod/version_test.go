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

package gomod

import (
	"io"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestGetLatestTag(t *testing.T) {
	repo, err := git.PlainClone(t.TempDir(), false, &git.CloneOptions{
		URL: "https://github.com/CycloneDX/cyclonedx-go.git",
	})
	require.NoError(t, err)

	headCommit, err := repo.CommitObject(plumbing.NewHash("a20be9f00d406e7b792973ee1826e637e58a23d7"))
	require.NoError(t, err)

	tag, err := GetLatestTag(zerolog.New(io.Discard), repo, headCommit)
	require.NoError(t, err)
	require.NotNil(t, tag)

	require.Equal(t, "v0.3.0", tag.name)
	require.Equal(t, "a20be9f00d406e7b792973ee1826e637e58a23d7", tag.commit.Hash.String())
}
