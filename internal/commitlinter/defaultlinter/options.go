package defaultlinter

type Option func(l *Linter)

func AllowInitialCommit() Option {
	return func(l *Linter) {
		l.AllowInitialCommit = true
	}
}

func WithCommitTypes(types ...string) Option {
	return func(l *Linter) {
		if l.AllowedTypes == nil {
			l.AllowedTypes = make(map[string]struct{})
		}

		for _, typ := range types {
			l.AllowedTypes[typ] = struct{}{}
		}
	}
}
