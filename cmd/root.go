package cmd

import (
	"runtime/debug"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/somebadcode/commit-tool/internal/zapatomicflag"
	"github.com/somebadcode/commit-tool/internal/zapctx"
)

var RootCommand *cobra.Command

func init() {
	atomicLogLevel := zap.NewAtomicLevelAt(zap.InfoLevel)

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		panic("build information is missing")
	}

	RootCommand = &cobra.Command{
		Use:               "commit-tool",
		ValidArgsFunction: cobra.NoFileCompletions,
		Args:              cobra.NoArgs,
		Version:           buildInfo.Main.Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			core := zapcore.NewCore(
				zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
				zapcore.AddSync(cmd.ErrOrStderr()),
				atomicLogLevel,
			)

			ctx := zapctx.WithContext(cmd.Context(), zap.New(core))
			cmd.SetContext(ctx)
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			l := zapctx.Value(cmd.Context())
			_ = l.Sync()
		},
		TraverseChildren:   true,
		Hidden:             true,
		DisableSuggestions: true,
	}

	flagSet := RootCommand.PersistentFlags()
	flagSet.Var(zapatomicflag.New(atomicLogLevel), "log-level", "log level")
}
