package commitslogvalue

import (
	"log/slog"

	"github.com/go-git/go-git/v5/plumbing/object"
)

func ReplaceAttr(_ []string, attr slog.Attr) slog.Attr {
	switch attr.Value.Kind() {
	case slog.KindAny:
		switch attr.Value.Any().(type) {
		case *object.Commit:
			attr.Value = Commit(attr.Value.Any().(*object.Commit))
		case object.Signature:
			attr.Value = Signature(attr.Value.Any().(object.Signature))
		}
	default:
		return attr
	}

	return attr
}

func Commit(commit *object.Commit) slog.Value {
	attrs := make([]slog.Attr, 2, 3)
	attrs[0] = slog.String("hash", commit.Hash.String())
	attrs[1] = slog.Any("author", commit.Author)

	if commit.Committer.String() != commit.Author.String() {
		attrs = append(attrs,
			slog.Any("committer", commit.Committer),
		)
	}

	return slog.GroupValue(attrs...)
}

func Signature(signature object.Signature) slog.Value {
	return slog.GroupValue(
		slog.String("name", signature.Name),
		slog.String("email", signature.Email),
		slog.Time("when", signature.When),
	)
}
