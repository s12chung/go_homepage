package cli

import (
	"github.com/golang/mock/gomock"
	"github.com/s12chung/go_homepage/go/test"
	"github.com/s12chung/go_homepage/go/test/mocks"
	"testing"
)

//go:generate mockgen -destination=../test/mocks/cli_app.go -package=mocks github.com/s12chung/go_homepage/go/cli App

func defaultCli(app App) *Cli {
	return NewCli("random name", app)
}

func TestCli_Run(t *testing.T) {
	testCases := []struct {
		args         []string
		functionName string
	}{
		{nil, "Generate"},
		{[]string{}, "Generate"},
		{[]string{"-file-server"}, "RunFileServer"},
		{[]string{"-server"}, "Host"},
		{[]string{"-file-server", "-server"}, "RunFileServer"},
		{[]string{"-blah"}, ""},
		{[]string{"-file-server", "-blah"}, ""},
	}
	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":        testCaseIndex,
			"args":         tc.args,
			"functionName": tc.functionName,
		})

		controller := gomock.NewController(t)
		app := mocks.NewMockApp(controller)
		expect := app.EXPECT()

		expect.GeneratedPath().Return("the_generated")
		expect.FileServerPort().Return(999)
		expect.ServerPort().Return(100)

		map[string]func() *gomock.Call{
			"":              func() *gomock.Call { return nil },
			"Generate":      expect.Generate,
			"RunFileServer": expect.RunFileServer,
			"Host":          expect.Host,
		}[tc.functionName]()

		defaultCli(app).Run(tc.args)

		t.Log(context.FieldsString())
		controller.Finish()
	}
}
