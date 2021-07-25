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
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// GetModuleVersion attempts to detect a given module's version by first
// calling GetVersionFromTag and if that fails, GetPseudoVersion on it.
//
// If no Git repository is found in moduleDir, directories will be traversed
// upwards until the root directory is reached. This is done to accommodate
// for multi-module repositories, where modules are not placed in the repo root.
func GetModuleVersion(moduleDir string) (string, error) {
	repoDir, err := filepath.Abs(moduleDir)
	if err != nil {
		return "", err
	}

	for {
		if tagVersion, err := GetVersionFromTag(repoDir); err == nil {
			return tagVersion, nil
		} else {
			if errors.Is(err, git.ErrRepositoryNotExists) {
				if strings.HasSuffix(repoDir, string(filepath.Separator)) {
					// filepath.Abs and filepath.Dir both return paths
					// that do not end with separators, UNLESS it's the
					// root dir. We can't move up any further.
					return "", fmt.Errorf("no git repository found")
				}
				repoDir = filepath.Dir(repoDir) // Move to the parent dir
				continue
			} else if errors.Is(err, plumbing.ErrObjectNotFound) {
				// It's a Git repo, but there's no tag pointing at HEAD.
				// Construct a pseudo version instead.
				pseudoVersion, err := GetPseudoVersion(repoDir)
				if err != nil {
					return "", fmt.Errorf("constructing pseudo version failed: %w", err)
				}
				return pseudoVersion, nil
			}
			return "", err
		}
	}
}

// GetPseudoVersion constructs a pseudo version for a Go module at a given path.
// See https://golang.org/ref/mod#pseudo-versions
func GetPseudoVersion(moduleDir string) (string, error) {
	repo, err := git.PlainOpen(moduleDir)
	if err != nil {
		return "", err
	}

	headRef, err := repo.Head()
	if err != nil {
		return "", err
	}

	headCommit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		return "", err
	}

	commitHash := headCommit.Hash.String()[:12]
	commitDate := headCommit.Author.When.Format("20060102150405")

	return fmt.Sprintf("v0.0.0-%s-%s", commitDate, commitHash), nil
}

// GetVersionFromTag checks if the current commit is annotated with a tag and if it is, returns that tag's name.
func GetVersionFromTag(moduleDir string) (string, error) {
	repo, err := git.PlainOpen(moduleDir)
	if err != nil {
		return "", err
	}

	headRef, err := repo.Head()
	if err != nil {
		return "", err
	}

	tags, err := repo.Tags()
	if err != nil {
		return "", err
	}

	var tagName string
	err = tags.ForEach(func(reference *plumbing.Reference) error {
		if reference.Hash() == headRef.Hash() && strings.HasPrefix(reference.Name().String(), "refs/tags/v") {
			tagName = strings.TrimPrefix(reference.Name().String(), "refs/tags/")
			return storer.ErrStop // break
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if tagName == "" {
		return "", plumbing.ErrObjectNotFound
	}

	return tagName, nil
}
