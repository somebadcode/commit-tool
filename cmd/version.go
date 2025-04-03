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
