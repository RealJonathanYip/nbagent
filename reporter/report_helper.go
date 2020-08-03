package reporter

import (
	"time"
)

type AutoReportHelper struct {
	key   string
	start time.Time
}

func NewAutoReportHelper(key string) *AutoReportHelper {
	return &AutoReportHelper{
		key:   key,
		start: time.Now(),
	}
}

func (m *AutoReportHelper) Report(szResult string) {
	Result(m.key, szResult, int64(time.Since(m.start)/time.Microsecond), 1)
}
