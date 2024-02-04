package lint

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/spf13/cobra"

	"github.com/somebadcode/commit-tool/cmd"
	"github.com/somebadcode/commit-tool/commitlinter"
	"github.com/somebadcode/commit-tool/commitlinter/defaultlinter"
	"github.com/somebadcode/commit-tool/revisionflag"
	"github.com/somebadcode/commit-tool/slogctx"
)

const (
	FlagRev                = "rev"
	FlagOtherRev           = "other"
	FlagAllowInitialCommit = "allow-initial-commit"
)

var linter commitlinter.CommitLinter
var command = &cobra.Command{
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

		linter.Rev = cmd.Flags().Lookup(FlagRev).Value.(*revisionflag.Revision).Rev
		linter.OtherRev = cmd.Flags().Lookup(FlagOtherRev).Value.(*revisionflag.Revision).Rev
		linter.ReportFunc = commitlinter.SlogReporter(slogctx.L(cmd.Context()))

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return linter.Run(cmd.Context())
	},
}

func init() {
	defaultLinter := defaultlinter.New()
	linter.Linter = defaultLinter

	flagSet := command.Flags()
	flagSet.Var(revisionflag.New(plumbing.Revision(plumbing.HEAD)), FlagRev, "revision to start at")
	flagSet.Var(revisionflag.New(""), FlagOtherRev, "revision for which the merge base of this and rev should stop the linting at")
	flagSet.BoolVar(&defaultLinter.AllowInitialCommit, FlagAllowInitialCommit, false, "allow initial commit")

	cmd.AddCommand(command)
}
