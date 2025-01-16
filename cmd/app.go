package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"

	"github.com/urfave/cli/v2"

	"github.com/somebadcode/commit-tool/cmd/lint"
	"github.com/somebadcode/commit-tool/internal/replaceattr"
)

type StatusCode = int

const (
	StatusOK StatusCode = iota
	StatusFailure
)

type AppInfo struct {
	Version string
}

func getAppInfo() AppInfo {
	info, hasBuildInfo := debug.ReadBuildInfo()
	if !hasBuildInfo {
		panic("build information is missing")
	}

	return AppInfo{
		Version: info.Main.Version,
	}
}

func Execute() StatusCode {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	level := &slog.LevelVar{}
	level.Set(slog.LevelInfo)

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource:   false,
		Level:       level,
		ReplaceAttr: replaceattr.ReplaceAttr,
	})

	errWriter := slog.NewLogLogger(handler, slog.LevelError)

	lintCommand := &lint.Command{
		Logger: slog.New(handler),
	}

	app := &cli.App{
		Name:           filepath.Base(os.Args[0]),
		Usage:          "A git commit tool",
		DefaultCommand: "lint",
		Reader:         os.Stdin,
		Writer:         os.Stdout,
		Authors: []*cli.Author{
			{
				Name:  "Tobias Dahlberg",
				Email: "lokskada@live.se",
			},
		},
		Version:              getAppInfo().Version,
		EnableBashCompletion: true,
		ErrWriter:            errWriter.Writer(),
		ExitErrHandler: func(c *cli.Context, err error) {
			if err == nil {
				return
			}

			lintCommand.Logger.Error("application error",
				slog.Any("error", err),
			)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "log-level",
				Usage: "",
				Value: "info",
				Action: func(c *cli.Context, s string) error {
					return level.UnmarshalText([]byte(s))
				},
			},
		},
		Commands: []*cli.Command{
			{
				Name:        "lint",
				Aliases:     []string{"linter"},
				UsageText:   "lint [command options]",
				Description: "lint git repository",
				Action: func(c *cli.Context) error {
					return lintCommand.Execute(c.Context)
				},
				Flags:                  lintCommand.Flags(),
				UseShortOptionHandling: true,
			},
		},
	}

	err := app.RunContext(ctx, os.Args)
	if err != nil {
		return StatusFailure
	}

	return StatusOK
}
