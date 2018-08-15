package test

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/s12chung/go_homepage/go/lib/utils"
)

const fixturePath = "./testdata"

func RandSeed() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func TimeDiff(callback func()) time.Duration {
	start := time.Now()
	callback()
	return time.Now().Sub(start)
}

func SandboxDir(t *testing.T, originalPath string) (string, func()) {
	dir, err := ioutil.TempDir("", "sandbox")
	if err != nil {
		t.Error(err)
	}
	return filepath.Join(dir, utils.CleanFilePath(originalPath)), func() { os.RemoveAll(dir) }
}

func ReadFixture(t *testing.T, filename string) []byte {
	bytes, err := ioutil.ReadFile(filepath.Join(fixturePath, filename))
	if err != nil {
		t.Error(err)
	}
	return bytes
}
