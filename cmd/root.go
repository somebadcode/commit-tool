package cmd

import (
	"context"
	"log/slog"
	"runtime/debug"

	"github.com/somebadcode/commit-tool/commitslogvalue"
	"github.com/spf13/cobra"

	"github.com/somebadcode/commit-tool/slogctx"
)

const (
	FlagLogLevel  = "log-level"
	FlagLogSource = "log-source"
)

var command = &cobra.Command{
	Use:               "commit-tool",
	ValidArgsFunction: cobra.NoFileCompletions,
	Args:              cobra.NoArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		levelFlagValue := cmd.Flags().Lookup(FlagLogLevel).Value.String()

		level := &slog.LevelVar{}
		levelError := level.UnmarshalText([]byte(levelFlagValue))

		addSource, _ := cmd.Flags().GetBool(FlagLogSource)

		opts := &slog.HandlerOptions{
			AddSource:   addSource,
			Level:       level,
			ReplaceAttr: commitslogvalue.ReplaceAttr,
		}

		handler := slog.NewJSONHandler(cmd.ErrOrStderr(), opts)
		logger := slog.New(handler)
		if levelError != nil {
			logger.Error("bad error level",
				slog.String("value", levelFlagValue),
			)
		}

		ctx := slogctx.Context(cmd.Context(), logger)

		cmd.SetContext(ctx)

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if syncer, ok := cmd.ErrOrStderr().(interface{ Sync() error }); ok {
			_ = syncer.Sync()
		}
	},
	TraverseChildren:   true,
	Hidden:             true,
	DisableSuggestions: true,
}

func init() {
	buildInfo, hasBuildInfo := debug.ReadBuildInfo()
	if hasBuildInfo {
		command.Version = buildInfo.Main.Version
	}

	flagSet := command.PersistentFlags()
	flagSet.String(FlagLogLevel, slog.LevelInfo.String(), "log level")
	flagSet.Bool(FlagLogSource, false, "show source in log messages")
}

func ExecuteContext(ctx context.Context) error {
	return command.ExecuteContext(ctx)
}

func AddCommand(cmds ...*cobra.Command) {
	command.AddCommand(cmds...)
}
