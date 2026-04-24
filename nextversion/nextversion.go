/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package nextversion

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"

	"codeberg.org/somebadcode/commit-tool/commitparser"
	"codeberg.org/somebadcode/commit-tool/semver"
)

var (
	ErrRepositoryRequired = errors.New("repository is required")
	ErrRevIsAlreadyTagged = errors.New("selected revision already has a tag")
)

type NextVersion struct {
	Logger     *slog.Logger
	Repository *git.Repository
	Revision   plumbing.Revision
	Prerelease []string
	Metadata   []string
	Writer     io.Writer
	VSuffix    bool
}

func (nv *NextVersion) Validate() error {
	if nv.Repository == nil {
		return ErrRepositoryRequired
	}

	if nv.Revision == "" {
		nv.Revision = plumbing.Revision(plumbing.HEAD)
	}

	if nv.Logger == nil {
		nv.Logger = slog.New(slog.DiscardHandler)
	}

	if nv.Writer == nil {
		nv.Writer = os.Stdout
	}

	return nil
}

func (nv *NextVersion) Run(ctx context.Context) error {
	if err := nv.Validate(); err != nil {
		return err
	}

	hash, err := nv.Repository.ResolveRevision(nv.Revision)
	if err != nil {
		return fmt.Errorf("failed to resolve revision %q: %w", nv.Revision, err)
	}

	var iter object.CommitIter

	iter, err = nv.Repository.Log(&git.LogOptions{
		From:  *hash,
		Order: git.LogOrderBSF,
	})

	if err != nil {
		return fmt.Errorf("could not iterate over commits: %w", err)
	}

	var tags map[plumbing.Hash]semver.Version

	tags, err = findTags(ctx, nv.Repository)
	if err != nil {
		return fmt.Errorf("could not find tags: %w", err)
	}

	// There's no next version if the selected revision already has a version tag.
	if _, exists := tags[*hash]; exists {
		return ErrRevIsAlreadyTagged
	}

	if len(tags) > 0 && nv.Logger.Enabled(ctx, slog.LevelDebug) {
		attrs := make([]slog.Attr, 0, len(tags))

		for h, v := range tags {
			attrs = append(attrs, slog.String(v.String(), h.String()))
		}

		slices.SortFunc(attrs, func(a, b slog.Attr) int {
			var v1, v2 semver.Version

			v1, err = semver.Parse(a.Key)
			if err != nil {
				panic(err)
			}

			v2, err = semver.Parse(b.Key)
			if err != nil {
				panic(err)
			}

			return semver.Compare(v1, v2)
		})

		nv.Logger.LogAttrs(ctx, slog.LevelDebug, "discovered tags",
			slog.Any("tags", slog.GroupValue(attrs...)),
		)
	}

	var major, minor, patch bool

	var version semver.Version

	err = iter.ForEach(func(commit *object.Commit) error {
		if ctx.Err() != nil {
			if nv.Logger.Enabled(ctx, slog.LevelDebug) {
				nv.Logger.LogAttrs(ctx, slog.LevelDebug, "cancelling finding tags",
					slog.String("hash", commit.Hash.String()),
					slog.String("cause", ctx.Err().Error()),
				)
			}

			return ctx.Err()
		}

		if version = tags[commit.Hash]; !version.IsZero() {
			return storer.ErrStop
		}

		var msg commitparser.CommitMessage

		msg, err = commitparser.Parse(commit.Message)
		if err != nil {
			return fmt.Errorf("could not parse commit message: %w", err)
		}

		if msg.Breaking {
			major = true

			return nil
		}

		switch msg.Type {
		case "feat":
			minor = true
		case "fix", "sec":
			patch = true
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to calculate next version: %w", err)
	}

	// If the major version is zero then breaking changes are acceptable and should not increment major version.
	if version.Major == 0 {
		major = false
	}

	switch {
	case major:
		version.IncreaseMajor()
	case minor:
		version.IncreaseMinor()
	case patch:
		version.IncreasePatch()
	}

	if nv.Prerelease != nil {
		err = version.SetPrerelease(nv.Prerelease...)
		if err != nil {
			return fmt.Errorf("could not set prerelease version: %w", err)
		}
	}

	if nv.Metadata != nil {
		err = version.SetBuildMetadata(nv.Metadata...)
		if err != nil {
			return fmt.Errorf("could not set metadata version: %w", err)
		}
	}

	if nv.VSuffix {
		_, err = nv.Writer.Write([]byte{'v'})
		if err != nil {
			return fmt.Errorf("could not write version suffix: %w", err)
		}
	}

	_, err = nv.Writer.Write([]byte(version.String()))
	if err != nil {
		return fmt.Errorf("could not write next version: %w", err)
	}

	return nil
}

func findTags(ctx context.Context, repo *git.Repository) (map[plumbing.Hash]semver.Version, error) {
	iter, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("could not get tags: %w", err)
	}

	defer iter.Close()

	tags := make(map[plumbing.Hash]semver.Version)

	err = iter.ForEach(func(ref *plumbing.Reference) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		var v semver.Version
		v, err = semver.Parse(ref.Name().Short())
		if err != nil {
			return nil
		}

		tags[ref.Hash()] = v

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not iterate tags: %w", err)
	}

	return tags, nil
}
