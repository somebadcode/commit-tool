package commitlinter

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/somebadcode/commit-tool/commitparser"
)

type FilterFunc func(message commitparser.CommitMessage, commit *object.Commit, err error) error

type Filters []FilterFunc

const (
	initialCommit = "initial commit"
)

func (filters Filters) Filter(msg commitparser.CommitMessage, commit *object.Commit, err error) error {
	for _, filter := range filters {
		if filter(msg, commit, err) == nil {
			return nil
		}
	}

	return err
}

func FilterInitialCommit(_ commitparser.CommitMessage, commit *object.Commit, err error) error {
	// If the commit has one or more parents then there's nothing to do and the error should be returned.
	if commit.NumParents() > 0 {
		return err
	}

	if !strings.EqualFold(commit.Message, initialCommit) {
		return fmt.Errorf("expected commit message %q: %w", initialCommit, err)
	}

	return nil
}
