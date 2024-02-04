package defaultlinter

type Option func(l *Linter)

func AllowInitialCommit() Option {
	return func(l *Linter) {
		l.AllowInitialCommit = true
	}
}

func WithRule(rules ...RuleFunc) Option {
	return func(l *Linter) {
		l.Rules = append(l.Rules, rules...)
	}
}
