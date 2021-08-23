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
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/require"
)

func TestGetLatestTag(t *testing.T) {
	repo, err := git.PlainOpen("../../")
	require.NoError(t, err)

	headCommit, err := repo.CommitObject(plumbing.NewHash("a078e094ad2504c6b16d8d9eeccc86e519f5c1da"))
	require.NoError(t, err)

	tag, err := GetLatestTag(repo, headCommit)
	require.NoError(t, err)
	require.NotNil(t, tag)

	require.Equal(t, "v0.9.0", tag.name)
	require.Equal(t, "a078e094ad2504c6b16d8d9eeccc86e519f5c1da", tag.commit.Hash.String())
}
