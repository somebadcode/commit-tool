package logging

import (
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

type AtomicLevelFlag struct {
	level zap.AtomicLevel
}

func (lvl AtomicLevelFlag) String() string {
	return lvl.level.String()
}

func (lvl AtomicLevelFlag) Set(s string) error {
	return lvl.level.UnmarshalText([]byte(s))
}

func (lvl AtomicLevelFlag) Type() string {
	return "level"
}

func NewAtomicLevelValue(level zap.AtomicLevel) pflag.Value {
	return &AtomicLevelFlag{
		level: level,
	}
}
