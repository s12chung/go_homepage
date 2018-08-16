package test

import (
	"flag"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const FixturePath = "./testdata"

func AssertBasic(t *testing.T, input, got, exp interface{}) {
	context := NewContext().SetFields(ContextFields{
		"input": input,
	})
	if got != exp {
		t.Error(context.GotExpString("Result", got, exp))
	}
}

func RandSeed() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func TimeDiff(callback func()) time.Duration {
	start := time.Now()
	callback()
	return time.Now().Sub(start)
}

func Time(i int) time.Time {
	return time.Date(2018, 1, i, i, i, i, i, time.UTC)
}

func cleanFilePath(filePath string) string {
	filePath = strings.TrimLeft(filePath, ".")
	return strings.Trim(filePath, "/")
}

func SandboxDir(t *testing.T, originalPath string) (string, func()) {
	dir, err := ioutil.TempDir("", "sandbox")
	if err != nil {
		t.Error(err)
	}
	return filepath.Join(dir, cleanFilePath(originalPath)), func() { os.RemoveAll(dir) }
}

func UpdateFixtureFlag() *bool {
	return flag.Bool("update", false, "Update fixtures")
}

func ReadFixture(t *testing.T, filename string) []byte {
	bytes, err := ioutil.ReadFile(filepath.Join(FixturePath, filename))
	if err != nil {
		t.Error(err)
	}
	return bytes
}

func WriteFixture(t *testing.T, filename string, data []byte) {
	err := ioutil.WriteFile(filepath.Join(FixturePath, filename), data, 0755)
	if err != nil {
		t.Error(err)
	}
}
