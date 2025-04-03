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

// Package cmd contains the command-line interface code.
package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/go-git/go-git/v5"

	"github.com/somebadcode/commit-tool/internal/replaceattr"
	"github.com/somebadcode/commit-tool/kongmappings"
)

type StatusCode = int

const (
	StatusOK StatusCode = iota
	StatusInternalError
	StatusBadArguments
	StatusFailure
)

type CLI struct {
	LogLevel  slog.LevelVar  `kong:"default='info',optional,placeholder='level',help='log level'"`
	LogJSON   bool           `kong:"optional,help='log using JSON'"`
	LogSource bool           `kong:"optional,hidden,help='log source (filename and line)'"`
	Lint      LintCommand    `kong:"cmd,help='lint'"`
	Version   VersionCommand `kong:"cmd,default='1',help='version'"`
}

func Run() StatusCode {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var cli CLI

	command := kong.Parse(&cli,
		kong.TypeMapper(reflect.TypeOf((*git.Repository)(nil)), kongmappings.Repository()),
	)

	handlerOpts := &slog.HandlerOptions{
		AddSource:   cli.LogSource,
		Level:       &cli.LogLevel,
		ReplaceAttr: replaceattr.ReplaceAttr,
	}

	var handler slog.Handler

	switch {
	case cli.LogJSON:
		handler = slog.NewJSONHandler(os.Stderr, handlerOpts)
	default:
		handler = slog.NewTextHandler(os.Stderr, handlerOpts)
	}

	logger := slog.New(handler)

	if err := command.Error; err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "failed to parse arguments",
			slog.String("error", err.Error()),
		)

		return StatusBadArguments
	}

	if err := command.ApplyDefaults(); err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "failed to set default values",
			slog.String("error", err.Error()),
		)

		return StatusInternalError
	}

	command.BindTo(ctx, (*context.Context)(nil))
	command.Bind(logger)

	err := command.Run()
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "command failed",
			slog.String("command", command.Command()),
			slog.String("error", err.Error()),
		)

		return StatusFailure
	}

	return StatusOK
}
