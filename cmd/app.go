/*
 * This file is part of commit-tool which is released under EUPL 1.2.
 * See the file LICENSE in the repository root for full license details.
 *
 * SPDX-License-Identifier: EUPL-1.2
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

	"github.com/somebadcode/commit-tool/cpuprofiler"
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
	LogLevel   slog.LevelVar         `kong:"optional,group='logging',env='LOG_LEVEL',default='info',placeholder='level',help='log level'"`
	LogJSON    bool                  `kong:"optional,group='logging',env='LOG_JSON',negatable,help='log using JSON'"`
	LogSource  bool                  `kong:"optional,group='logging',env='LOG_SOURCE',hidden,help='log source (filename and line number)'"`
	CPUProfile *cpuprofiler.Profiler `kong:"optional,hidden,help='profile cpu destination',placeholder='FILENAME'"`

	// Commands:
	Lint        LintCommand        `kong:"cmd,default='',help='lint the commit messages in a git repository'"`
	NextVersion NextVersionCommand `kong:"cmd,help='get next version (lint is recommended prior to running this)'"`
	Version     VersionCommand     `kong:"cmd,help='show program version'"`
}

func Run() StatusCode {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var cli CLI

	command := kong.Parse(&cli,
		kong.TypeMapper(reflect.TypeOf((*git.Repository)(nil)), kongmappings.Repository()),
		kong.TypeMapper(reflect.TypeOf((*cpuprofiler.Profiler)(nil)), cpuprofiler.KongMapperFunc()),
		kong.Description("Commit Tool - a helpful tool for when you're working with git repositories"),
		kong.ExplicitGroups([]kong.Group{
			{
				Key:         "logging",
				Title:       "Logging",
				Description: "Logging configuration",
			},
		}),
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

	if cli.CPUProfile != nil {
		stopCPUProfiler, err := cli.CPUProfile.Start()
		if err != nil {
			logger.LogAttrs(ctx, slog.LevelError, "failed to start CPU profile",
				slog.String("error", err.Error()),
			)

			return StatusFailure
		}

		defer stopCPUProfiler()
	}

	command.BindTo(ctx, (*context.Context)(nil))
	command.Bind(logger.With("logger", command.Selected().Name))

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
