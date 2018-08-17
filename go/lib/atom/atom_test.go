package atom

import (
	"testing"

	"github.com/s12chung/go_homepage/go/test"
)

func defaultRenderer() *Renderer {
	return NewRenderer(DefaultSettings())
}

func TestRenderer_NewFeed(t *testing.T) {
	renderer := defaultRenderer()

	bytes, err := renderer.NewFeed("posts", test.Time(1), "the_posts", "something.png").Marhshall()
	if err != nil {
		t.Error(err)
	}

	fixtureFilename := "new_feed.xml"
	if *updateFixturesPtr {
		test.WriteFixture(t, fixtureFilename, bytes)
		return
	}
	got := string(bytes)
	exp := string(test.ReadFixture(t, fixtureFilename))
	test.AssertLabel(t, "Result", got, exp)
}
