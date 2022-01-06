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
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

// GetModuleVersion attempts to detect a given module's version.
//
// If no Git repository is found in moduleDir, directories will be traversed
// upwards until the root directory is reached. This is done to accommodate
// for multi-module repositories, where modules are not placed in the repo root.
func GetModuleVersion(moduleDir string) (string, error) {
	log.Debug().
		Str("moduleDir", moduleDir).
		Msg("detecting module version")

	repoDir, err := filepath.Abs(moduleDir)
	if err != nil {
		return "", err
	}

	for {
		if tagVersion, err := GetVersionFromTag(repoDir); err == nil {
			return tagVersion, nil
		} else {
			if strings.Contains(err.Error(), "no git repository found") {
				if strings.HasSuffix(repoDir, string(filepath.Separator)) {
					// filepath.Abs and filepath.Dir both return paths
					// that do not end with separators, UNLESS it's the
					// root dir. We can't move up any further.
					return "", fmt.Errorf("no git repository found")
				}
				repoDir = filepath.Dir(repoDir) // Move to the parent dir
				continue
			}

			return "", err
		}
	}
}

// GetVersionFromTag checks if the HEAD commit is annotated with a tag and if it is, returns that tag's name.
// If the HEAD commit is not tagged, a pseudo version will be generated and returned instead.
func GetVersionFromTag(moduleDir string) (string, error) {
	headCommit, err := doExec("sh", []string{"-c", "git show-ref --heads | cut -d' ' -f1"}, moduleDir, nil)
	if err != nil {
		return "", err
	}

	latestTag, err := doExec("sh", []string{"-c", "git describe --tags $(git rev-list --tags --max-count=1)"}, moduleDir, nil)
	if err != nil {
		return "", err
	}

	latestTagCommit, err := doExec("sh", []string{"-c", fmt.Sprintf("git show-ref -s %s", string(latestTag))}, moduleDir, nil)
	if err != nil {
		return "", err
	}

	if string(latestTagCommit) == string(headCommit) {
		return string(latestTag), nil
	}

	when, err := doExec("git", []string{"show", "--show-signature", strings.TrimSpace(string(latestTagCommit)), "--format=%cd", "--date=format:%Y-%m-%d %H:%M:%S"}, moduleDir, nil)
	if err != nil {
		return "", err
	}

	if strings.Contains(string(when), "gpg: Can't check signature: No public key") {
		when, err = doExec("sh", []string{"-c", "tail -n +4 | head -n1"}, moduleDir, bytes.NewBuffer(when))
		if err != nil {
			return "", err
		}
	}

	suffix := strings.TrimSpace(string(when))
	pseudoTime, err := time.Parse("2006-01-02 15:04:05", suffix)
	if err != nil {
		return "", err
	}

	return module.PseudoVersion(
		semver.Major(strings.TrimSpace(string(latestTag))),
		strings.TrimSpace(string(latestTag)),
		pseudoTime,
		strings.TrimSpace(string(latestTagCommit))[:12],
	), nil
}

func doExec(cmd string, args []string, workDir string, stdin io.Reader) ([]byte, error) {
	c := exec.Command(cmd, args...)
	c.Dir = workDir
	if stdin != nil {
		c.Stdin = stdin
	}
	return c.CombinedOutput()
}

//type tag struct {
//	name   string
//	commit *object.Commit
//}

//// GetLatestTag determines the latest tag relative to HEAD.
//// Only tags with valid semver are considered.
//func GetLatestTag(repo *git.Repository, headCommit *object.Commit) (*tag, error) {
//	log.Debug().
//		Str("headCommit", headCommit.Hash.String()).
//		Msg("getting latest tag for head commit")
//
//	tagRefs, err := repo.Tags()
//	if err != nil {
//		return nil, err
//	}
//
//	var latestTag tag
//
//	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
//		if semver.IsValid(ref.Name().Short()) {
//			rev := plumbing.Revision(ref.Name().String())
//
//			commitHash, err := repo.ResolveRevision(rev)
//			if err != nil {
//				return err
//			}
//
//			commit, err := repo.CommitObject(*commitHash)
//			if err != nil {
//				return err
//			}
//
//			isBeforeOrAtHead := commit.Committer.When.Before(headCommit.Author.When) ||
//				commit.Committer.When.Equal(headCommit.Committer.When)
//
//			if isBeforeOrAtHead && (latestTag.commit == nil || commit.Committer.When.After(latestTag.commit.Committer.When)) {
//				latestTag.name = ref.Name().Short()
//				latestTag.commit = commit
//			}
//		} else {
//			log.Debug().
//				Str("tag", ref.Name().Short()).
//				Str("hash", ref.Hash().String()).
//				Str("reason", "not a valid semver").
//				Msg("skipping tag")
//		}
//
//		return nil
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	if latestTag.commit == nil {
//		return nil, plumbing.ErrObjectNotFound
//	}
//
//	return &latestTag, nil
//}
