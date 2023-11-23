package zapatomicflag

import (
	"go.uber.org/zap"
)

type AtomicLevelFlag struct {
	level zap.AtomicLevel
}

func New(level zap.AtomicLevel) *AtomicLevelFlag {
	return &AtomicLevelFlag{
		level: level,
	}
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
