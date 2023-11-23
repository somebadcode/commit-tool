package lint

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"

	"github.com/somebadcode/commit-tool/cmd"
	"github.com/somebadcode/commit-tool/internal/commitlinter"
	"github.com/somebadcode/commit-tool/internal/commitlinter/defaultlinter"
	"github.com/somebadcode/commit-tool/internal/revisionflag"
	"github.com/somebadcode/commit-tool/internal/zapctx"
)

func init() {
	var linter commitlinter.CommitLinter

	command := &cobra.Command{
		Use:     "lint",
		Aliases: []string{"linter"},
		Args:    cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var repoPath string
			if len(args) == 1 {
				repoPath = args[0]
			} else {
				wd, wdErr := os.Getwd()
				if wdErr != nil {
					return fmt.Errorf("repository path required")
				}

				repoPath = wd
			}

			var err error

			linter.Repo, err = git.PlainOpen(repoPath)
			if err != nil {
				return err
			}

			linter.ReportFunc = commitlinter.ZapReporter(zapctx.L(cmd.Context()).Named("lint-report"))

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return linter.Run(cmd.Context())
		},
	}

	defaultLinter := defaultlinter.New()
	linter.Linter = defaultLinter

	flagSet := command.Flags()
	flagSet.Var(revisionflag.New(plumbing.Revision(plumbing.HEAD)), "rev", "revision to start at")
	flagSet.Var(revisionflag.New("other"), "other", "revision for which the merge base of this and rev should stop the linting at")
	flagSet.BoolVar(&defaultLinter.AllowInitialCommit, "allow-initial-commit", false, "allow initial commit")

	cmd.RootCommand.AddCommand(command)
}
