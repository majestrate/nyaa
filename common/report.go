package common

import (
	"time"
)

type ReportParam struct {
	Limit   uint32
	Offset  uint32
	AllTime bool
	Before  time.Time
	After   time.Time
}
