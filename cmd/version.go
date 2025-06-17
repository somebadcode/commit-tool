/*
 * This file is part of commit-tool which is released under EUPL 1.2.
 * See the file LICENSE in the repository root for full license details.
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package cmd

import (
	"context"
	"errors"
	"log/slog"
	"runtime/debug"
	"strconv"
)

var (
	ErrNoBuildInformation = errors.New("no build information available, can't obtain version")
)

type VersionCommand struct{}

func (cmd *VersionCommand) Run(ctx context.Context, l *slog.Logger) error {
	bi, hasBuildInfo := debug.ReadBuildInfo()
	if !hasBuildInfo {
		return ErrNoBuildInformation
	}

	l.LogAttrs(ctx, slog.LevelInfo, "version information",
		slog.String("version", bi.Main.Version),
	)

	if !l.Enabled(ctx, slog.LevelDebug) {
		return nil
	}

	settings := buildSettingsToMap(bi.Settings)

	l.LogAttrs(ctx, slog.LevelDebug, "build information",
		slog.String("goversion", bi.GoVersion),
		slog.String("module", bi.Path),
		slog.Group("vcs",
			slog.String("type", settings["vcs"]),
			slog.String("revision", settings["vcs.revision"]),
			slog.String("time", settings["vcs.time"]),
			slog.Bool("modified", toBool(settings["vcs.modified"], true)),
		),
	)

	return nil
}

func buildSettingsToMap(bs []debug.BuildSetting) map[string]string {
	settings := make(map[string]string, len(bs))

	for _, setting := range bs {
		settings[setting.Key] = setting.Value
	}

	return settings
}

func toBool(v string, whenEmpty bool) bool {
	if v == "" {
		return whenEmpty
	}

	b, err := strconv.ParseBool(v)
	if err != nil {
		panic(err)
	}

	return b
}
