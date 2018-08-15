package test

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"
)

func PrintLogEntries(t *testing.T, hook *logTest.Hook) {
	for _, entry := range hook.AllEntries() {
		s, err := entry.String()
		if err != nil {
			t.Error(err)
		}
		t.Log(strings.TrimSpace(s))
	}
}

func SafeLogEntries(hook *logTest.Hook) bool {
	for _, entry := range hook.AllEntries() {
		if entry.Level <= logrus.WarnLevel {
			return false
		}
	}
	return true
}
