package pool

import (
	"fmt"
	"testing"

	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/go_homepage/go/test"
)

func TestPool(t *testing.T) {
	testCases := []struct {
		tasksWithSuccess []bool
	}{
		{nil},
		{[]bool{}},
		{[]bool{true}},
		{[]bool{true, true}},
		{[]bool{true, true, true}},
		{[]bool{true, false, true}},
		{[]bool{false}},
		{[]bool{false, false, false}},
		{[]bool{false, true, true}},
	}

	log, _ := logTest.NewNullLogger()
	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":            testCaseIndex,
			"tasksWithSuccess": tc.tasksWithSuccess,
		})

		runCount := 0
		var errorTasks []*Task
		tasks := make([]*Task, len(tc.tasksWithSuccess))
		if tc.tasksWithSuccess == nil {
			tasks = nil
		} else {
			for i, success := range tc.tasksWithSuccess {
				ret := fmt.Errorf("error")
				if success {
					ret = nil
				}
				tasks[i] = NewTask(log, func() error {
					runCount++
					return ret
				})

				if !success {
					errorTasks = append(errorTasks, tasks[i])
				}
			}
		}
		p := NewPool(tasks, 10)
		p.Run()

		if runCount != len(tc.tasksWithSuccess) {
			t.Error(context.GotExpString("runCount", runCount, len(tc.tasksWithSuccess)))
		}

		errorCount := 0
		p.EachError(func(task *Task) {
			errorCount++
			if task.Error == nil {
				t.Error("EachError found task without error")
			}
			found := false
			for _, errorTask := range errorTasks {
				if errorTask == task {
					found = true
					break
				}
			}
			if !found {
				t.Error("EachError Error not found in errorTasks")
			}
		})
		if errorCount != len(errorTasks) {
			t.Error("errorCount does not match number of errorTasks")
		}
	}
}
