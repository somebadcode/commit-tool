package commitslogvalue

import (
	"log/slog"

	"github.com/go-git/go-git/v5/plumbing/object"
)

func Value(commit *object.Commit) slog.Value {
	attrs := make([]slog.Attr, 2, 3)
	attrs[0] = slog.String("hash", commit.Hash.String())
	attrs[1] = slog.Group("author",
		slog.String("name", commit.Author.Name),
		slog.String("email", commit.Author.Email),
		slog.Time("when", commit.Author.When),
	)

	if commit.Committer.String() != commit.Author.String() {
		attrs = append(attrs,
			slog.Group("committer",
				slog.String("name", commit.Committer.Name),
				slog.String("email", commit.Committer.Name),
				slog.Time("when", commit.Committer.When),
			),
		)
	}

	return slog.GroupValue(attrs...)
}
