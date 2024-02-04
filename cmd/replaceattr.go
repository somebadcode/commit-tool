package cmd

import (
	"log/slog"

	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/somebadcode/commit-tool/commitslogvalue"
)

func replaceAttr(_ []string, attr slog.Attr) slog.Attr {
	switch attr.Value.Kind() {
	case slog.KindAny:
		switch attr.Value.Any().(type) {
		case *object.Commit:
			attr.Value = commitslogvalue.Value(attr.Value.Any().(*object.Commit))
		}
	}

	return attr
}
