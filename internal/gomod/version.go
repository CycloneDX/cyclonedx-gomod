package gomod

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// GetPseudoVersion constructs a pseudo version for a Go module at the given path.
// Note that this is only possible when path points to a Git repository.
// See https://golang.org/ref/mod#pseudo-versions
func GetPseudoVersion(path string) (string, error) {
	repo, err := git.PlainOpen(path)
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

// GetVersionFromTag checks if the current commit is annotated with a tag and if yes, returns that tag's name.
// Note that this is only possible when path points to a Git repository.
func GetVersionFromTag(path string) (string, error) {
	repo, err := git.PlainOpen(path)
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

	tagName := ""
	err = tags.ForEach(func(reference *plumbing.Reference) error {
		if reference.Hash() == headRef.Hash() && strings.Index(reference.Name().String(), "refs/tags/v") == 0 {
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
