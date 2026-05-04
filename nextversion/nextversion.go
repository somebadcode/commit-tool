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

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"

	"codeberg.org/somebadcode/commit-tool/classifier"
	"codeberg.org/somebadcode/commit-tool/commitparser"
	"codeberg.org/somebadcode/commit-tool/semver"
)

var (
	ErrRepositoryRequired = errors.New("repository is required")
	ErrRevIsAlreadyTagged = errors.New("selected revision already has a tag")
)

type NextVersion struct {
	Logger           *slog.Logger
	Repository       *git.Repository
	Revision         plumbing.Revision
	Prerelease       []string
	Metadata         []string
	Writer           io.Writer
	CommitClassifier *classifier.Classifier
	VSuffix          bool
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

	if nv.CommitClassifier == nil {
		nv.CommitClassifier, _ = classifier.New(nil)
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

	var tags map[plumbing.Hash]semver.Version

	tags, err = tagsAsMap(ctx, nv.Repository)
	if err != nil {
		return fmt.Errorf("could not find tags: %w", err)
	}

	// There's no next version if the selected revision already has a version tag.
	if _, exists := tags[*hash]; exists {
		return ErrRevIsAlreadyTagged
	}

	if len(tags) > 0 && nv.Logger.Enabled(ctx, slog.LevelDebug) {
		attrs := make([]slog.Attr, 0, len(tags))

		for k, v := range tags {
			attrs = append(attrs, slog.String(v.String(), k.String()))
		}

		nv.Logger.LogAttrs(ctx, slog.LevelDebug, "discovered tags",
			slog.Any("tags", slog.GroupValue(attrs...)),
		)
	}

	var version *semver.Version

	version, err = nv.ComputeNextVersion(ctx, tags)
	if err != nil {
		return err
	}

	_, err = nv.Writer.Write([]byte(version.String()))
	if err != nil {
		return fmt.Errorf("could not write next version: %w", err)
	}

	return nil
}

func (nv *NextVersion) ComputeNextVersion(ctx context.Context, tags map[plumbing.Hash]semver.Version) (*semver.Version, error) {
	hash, err := nv.Repository.ResolveRevision(nv.Revision)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve revision %q: %w", nv.Revision, err)
	}

	// There's no next version if the selected revision already has a version tag.
	if _, exists := tags[*hash]; exists {
		return nil, ErrRevIsAlreadyTagged
	}

	var iter object.CommitIter

	iter, err = nv.Repository.Log(&git.LogOptions{
		From:  *hash,
		Order: git.LogOrderBSF,
	})

	if err != nil {
		return nil, fmt.Errorf("could not iterate over commits: %w", err)
	}

	var classification classifier.Classification

	var version semver.Version

	err = iter.ForEach(func(commit *object.Commit) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Proceed
		}

		if version = tags[commit.Hash]; !version.IsZero() {
			return storer.ErrStop
		}

		var msg commitparser.CommitMessage

		msg, err = commitparser.Parse(commit.Message)
		if err != nil {
			return fmt.Errorf("could not parse commit message: %w", err)
		}

		class := nv.CommitClassifier.Classify(&msg)
		classification = max(classification, class)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to calculate next version: %w", err)
	}

	version = DetermineNextVersion(&version, classification)

	if nv.Prerelease != nil {
		err = version.SetPrerelease(nv.Prerelease...)
		if err != nil {
			return nil, fmt.Errorf("could not set prerelease version: %w", err)
		}
	}

	if nv.Metadata != nil {
		err = version.SetBuildMetadata(nv.Metadata...)
		if err != nil {
			return nil, fmt.Errorf("could not set metadata version: %w", err)
		}
	}

	return &version, nil
}

func DetermineNextVersion(v *semver.Version, classification classifier.Classification) semver.Version {
	version := v.Clone()

	if len(version.Prerelease) > 0 {
		isPreMajor := version.Major >= 1 && version.Minor == 0 && version.Patch == 0

		version.Finalize()

		if classification == classifier.Major && !isPreMajor && version.Major >= 1 {
			version.IncreaseMajor()
		}

		return version
	}

	switch classification {
	case classifier.Major:
		// If the major version is zero then breaking changes are acceptable and should not increment major version.
		if version.Major == 0 {
			version.IncreaseMinor()
			break
		}

		version.IncreaseMajor()

	case classifier.Minor:
		version.IncreaseMinor()

	case classifier.Patch:
		version.IncreasePatch()

	case classifier.None:
		// Do nothing.
	}

	return version
}

func tagsAsMap(ctx context.Context, repo *git.Repository) (map[plumbing.Hash]semver.Version, error) {
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

		if tag, _ := repo.TagObject(ref.Hash()); tag != nil {
			var v semver.Version
			v, err = semver.Parse(tag.Name)
			if err != nil {
				return nil
			}

			tags[tag.Target] = v

			return nil
		}

		var v semver.Version
		v, err = semver.Parse(ref.Name().Short())
		if err != nil {
			return nil
		}

		prev := tags[ref.Hash()]
		if prev.IsZero() {
			tags[ref.Hash()] = v

			return nil
		}

		i := semver.Compare(v, prev)
		if i > 0 {
			tags[ref.Hash()] = v
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not iterate tags: %w", err)
	}

	return tags, nil
}
