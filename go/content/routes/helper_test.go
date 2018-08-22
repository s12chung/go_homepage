package routes

import (
	"testing"

	"github.com/golang/mock/gomock"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/gostatic/go/test"
	"github.com/s12chung/gostatic/go/test/mocks"
)

//go:generate mockgen -destination=../../test/mocks/router_context.go -package=mocks github.com/s12chung/gostatic/go/lib/router Context

func TestTemplateName(t *testing.T) {
	testCases := []struct {
		templateName string
		urlParts     []string
		exp          string
		panic        bool
	}{
		{"", []string{"testy"}, "testy", false},
		{"", []string{"one", "two"}, "one", false},
		{"the_in", []string{"testy"}, "the_in", false},
		{"the_in", []string{}, "the_in", false},
		{"the_in", []string{"one", "two"}, "the_in", false},
		{"", []string{}, "", true},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":        testCaseIndex,
			"templateName": tc.templateName,
			"urlParts":     tc.urlParts,
		})

		controller := gomock.NewController(t)
		ctx := mocks.NewMockContext(controller)
		if tc.templateName == "" {
			ctx.EXPECT().UrlParts().AnyTimes().Return(tc.urlParts)
		}
		if tc.panic {
			log, _ := logTest.NewNullLogger()
			ctx.EXPECT().Log().Return(log)
		}

		func() {
			defer func() {
				t.Log(context.FieldsString())
				controller.Finish()
				if tc.panic {
					if r := recover(); r == nil {
						t.Errorf("Did not panic for duplicate route setup.")
					}
				}
			}()

			got := templateName(ctx, tc.templateName)
			if got != tc.exp {
				t.Error(context.GotExpString("Result", got, tc.exp))
			}
		}()

	}
}

func TestDefaultTitle(t *testing.T) {
	testCases := []struct {
		templateName string
		urlParts     []string
		exp          string
	}{
		{"temp_name", []string{"testy"}, "temp_name"},
		{"", []string{}, ""},
		{"", []string{"testy"}, ""},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":        testCaseIndex,
			"templateName": tc.templateName,
			"urlParts":     tc.urlParts,
		})

		controller := gomock.NewController(t)
		ctx := mocks.NewMockContext(controller)
		ctx.EXPECT().UrlParts().Return(tc.urlParts)

		got := defaultTitle(ctx, tc.templateName)
		if got != tc.exp {
			t.Error(context.GotExpString("Result", got, tc.exp))
		}
	}
}
