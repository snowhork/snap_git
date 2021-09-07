package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
)

func collectChangedFile(status git.Status) ([]string, error) {
	var targetPaths []string

	for path, f := range status {
		if f.Staging == git.Unmodified && f.Worktree == git.Unmodified {
			continue
		}

		if status.IsUntracked(path) {
			return nil, errors.Wrapf(UntrackedFileExistsError, "%s is untracked", path)
		}

		targetPaths = append(targetPaths, path)
	}

	return targetPaths, nil
}

func addChangedFile(w *git.Worktree, targetPaths []string) error {
	for _, path := range targetPaths {

		if _, err := w.Add(path); err != nil {
			return err
		}
	}
	return nil
}

var NoChangedError = errors.New("No changed")
var UntrackedFileExistsError = errors.New("Untracked File exists")

const snapShotBranchPrefix = "__snapshot"

func doGit(repoPath string, now time.Time) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return errors.WithStack(err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return errors.WithStack(err)
	}

	status, err := w.Status()
	if err != nil {
		return errors.WithStack(err)
	}

	targetPaths, err := collectChangedFile(status)
	if err != nil {
		return errors.WithStack(err)
	}
	if len(targetPaths) == 0 {
		return NoChangedError
	}

	headRef, err := repo.Head()

	if !strings.HasPrefix(headRef.Name().Short(), snapShotBranchPrefix) {
		branchName := fmt.Sprintf("__snapshot_%d", now.Unix())
		if err := w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branchName),
			Create: true,
			Keep:   true,
		}); err != nil {
			return errors.WithStack(err)
		}
	}

	if err := addChangedFile(w, targetPaths); err != nil {
		return err
	}

	if _, err := w.Commit(fmt.Sprintf("%d", now.Unix()), &git.CommitOptions{}); err != nil {
		return err
	}

	if err := repo.Push(&git.PushOptions{}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func gitExists(repoPath string) (bool, error) {
	_, err := git.PlainOpen(repoPath)
	if err != nil {
		return false, err
	}
	return true, nil
}
