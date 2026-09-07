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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/module"
)

func TestGetVersionFromTagPseudoVersionUsesHeadRevision(t *testing.T) {
	repositoryDir, repository, worktree := initTestRepository(t)

	taggedAt := time.Date(2025, time.January, 2, 3, 4, 5, 0, time.UTC)
	taggedHash := commitTestFile(t, repositoryDir, worktree, "tagged", taggedAt)
	_, err := repository.CreateTag("v1.2.3", taggedHash, nil)
	require.NoError(t, err)

	headAt := taggedAt.Add(time.Hour)
	headHash := commitTestFile(t, repositoryDir, worktree, "head", headAt)

	version, err := GetVersionFromTag(zerolog.Nop(), repositoryDir)
	require.NoError(t, err)
	require.Equal(t, module.PseudoVersion("v1", "v1.2.3", headAt, headHash.String()[:12]), version)
}

func TestGetVersionFromTagIgnoresUnreachableTags(t *testing.T) {
	repositoryDir, repository, worktree := initTestRepository(t)

	baseAt := time.Date(2025, time.January, 2, 3, 4, 5, 0, time.UTC)
	baseHash := commitTestFile(t, repositoryDir, worktree, "base", baseAt)
	taggedHash := commitTestFile(t, repositoryDir, worktree, "tagged", baseAt.Add(time.Hour))
	_, err := repository.CreateTag("v1.2.3", taggedHash, nil)
	require.NoError(t, err)

	require.NoError(t, worktree.Reset(&git.ResetOptions{Mode: git.HardReset, Commit: baseHash}))
	headAt := baseAt.Add(2 * time.Hour)
	headHash := commitTestFile(t, repositoryDir, worktree, "head", headAt)

	version, err := GetVersionFromTag(zerolog.Nop(), repositoryDir)
	require.NoError(t, err)
	require.Equal(t, module.PseudoVersion("v0", "", headAt, headHash.String()[:12]), version)
}

func TestGetLatestTag(t *testing.T) {
	repo, err := git.PlainClone(t.TempDir(), false, &git.CloneOptions{
		URL: "https://github.com/CycloneDX/cyclonedx-go.git",
	})
	require.NoError(t, err)

	headCommit, err := repo.CommitObject(plumbing.NewHash("a20be9f00d406e7b792973ee1826e637e58a23d7"))
	require.NoError(t, err)

	tag, err := GetLatestTag(zerolog.Nop(), repo, headCommit)
	require.NoError(t, err)
	require.NotNil(t, tag)

	require.Equal(t, "v0.3.0", tag.name)
	require.Equal(t, "a20be9f00d406e7b792973ee1826e637e58a23d7", tag.commit.Hash.String())
}

func initTestRepository(t *testing.T) (string, *git.Repository, *git.Worktree) {
	t.Helper()

	repositoryDir := t.TempDir()
	repository, err := git.PlainInit(repositoryDir, false)
	require.NoError(t, err)

	worktree, err := repository.Worktree()
	require.NoError(t, err)
	return repositoryDir, repository, worktree
}

func commitTestFile(t *testing.T, repositoryDir string, worktree *git.Worktree, contents string, when time.Time) plumbing.Hash {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Join(repositoryDir, "module.go"), []byte(contents), 0o600))
	_, err := worktree.Add("module.go")
	require.NoError(t, err)

	signature := &object.Signature{Name: "Test", Email: "test@example.com", When: when}
	hash, err := worktree.Commit(contents, &git.CommitOptions{Author: signature, Committer: signature})
	require.NoError(t, err)
	return hash
}
