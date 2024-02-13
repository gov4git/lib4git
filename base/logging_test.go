package base

import "testing"

func TestLogger(t *testing.T) {
	LogVerbosely()
	Debugf("xxx")
}
