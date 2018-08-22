package webpack

import (
	"testing"

	"github.com/s12chung/go_homepage/go/test"
)

func TestDefaultSettings(t *testing.T) {
	test.TestEnvSetting(t, "ASSETS_PATH", "assets", func() string {
		return DefaultSettings().AssetsPath
	})
}
