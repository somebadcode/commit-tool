/*
 * Copyright 2025 Tobias Dahlberg
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package nextversion

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"

	"github.com/somebadcode/commit-tool/commitparser"
)

var (
	ErrRepositoryRequired = errors.New("repository is required")
)

type NextVersion struct {
	Repository *git.Repository
	Revision   plumbing.Revision
	Writer     io.Writer
	Logger     *slog.Logger
	Prerelease string
	Metadata   string
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

	var tags map[plumbing.Hash]*semver.Version

	tags, err = findTags(ctx, nv.Repository)
	if err != nil {
		return fmt.Errorf("could not find tags: %w", err)
	}

	if len(tags) > 0 && nv.Logger.Enabled(ctx, slog.LevelDebug) {
		attrs := make([]slog.Attr, len(tags))

		for h, v := range tags {
			attrs = append(attrs, slog.String(v.String(), h.String()))
		}

		nv.Logger.LogAttrs(ctx, slog.LevelDebug, "discovered tags",
			slog.Any("tags", slog.GroupValue(attrs...)),
		)
	}

	var major, minor, patch bool

	var version *semver.Version

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

		if version = tags[commit.Hash]; version != nil {
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

	if version == nil {
		version = semver.MustParse("0.0.0")
	}

	switch {
	case major && version.Major() > 0:
		*version = version.IncMajor()
	case minor || (major && version.Major() > 0):
		*version = version.IncMinor()
	case patch:
		*version = version.IncPatch()
	}

	*version, err = version.SetPrerelease(nv.Prerelease)
	if err != nil {
		return fmt.Errorf("could not set prerelease version: %w", err)
	}

	*version, err = version.SetMetadata(nv.Metadata)
	if err != nil {
		return fmt.Errorf("could not set metadata version: %w", err)
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

func findTags(ctx context.Context, repo *git.Repository) (map[plumbing.Hash]*semver.Version, error) {
	iter, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("could not get tags: %w", err)
	}

	defer iter.Close()

	tags := make(map[plumbing.Hash]*semver.Version)

	err = iter.ForEach(func(ref *plumbing.Reference) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		var v *semver.Version
		v, err = semver.NewVersion(ref.Name().Short())
		if err != nil {
			return nil
		}

		tags[ref.Hash()] = v

		return nil
	})

	if err != nil {
		return nil, err
	}

	return tags, nil
}
