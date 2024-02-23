package vfs

import (
	"time"

	"github.com/souvikdeyrit/spinel/pkg/meta"
)

type LogContext interface {
	meta.Context
	Duration() time.Duration
}
