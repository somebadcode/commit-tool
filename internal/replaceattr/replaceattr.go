package replaceattr

import (
	"log/slog"

	"github.com/go-git/go-git/v5/plumbing/object"
)

type MultiError interface {
	Unwrap() []error
}

func ReplaceAttr(_ []string, attr slog.Attr) slog.Attr {
	switch attr.Value.Kind() {
	case slog.KindAny:
		switch attr.Value.Any().(type) {
		case *object.Commit:
			attr.Value = Commit(attr.Value.Any().(*object.Commit))
		case object.Signature:
			attr.Value = Signature(attr.Value.Any().(object.Signature))
		case MultiError:
			attr.Value = Errors(attr.Value.Any().(MultiError).Unwrap())
		case error:
			attr.Value = slog.StringValue(attr.Value.Any().(error).Error())
		}
	default:
		return attr
	}

	return attr
}

func Errors(errs []error) slog.Value {
	values := make([]string, len(errs))
	for i, err := range errs {
		values[i] = err.Error()
	}

	return slog.AnyValue(values)
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
