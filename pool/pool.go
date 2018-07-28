package pool

import (
	"sync"

	"github.com/Sirupsen/logrus"
)

//
// Task
//
type Task struct {
	run   func() error
	Log   logrus.FieldLogger
	Error error
}

func NewTask(log logrus.FieldLogger, run func() error) *Task {
	return &Task{run, log, nil}
}

func (task *Task) Run(waitGroup *sync.WaitGroup) {
	task.Error = task.run()
	waitGroup.Done()
}

//
// Pool
//
type Pool struct {
	Tasks []*Task

	concurrency int
	tasksChan   chan *Task
	waitGroup   sync.WaitGroup
}

func NewPool(tasks []*Task, concurrency int) *Pool {
	return &Pool{
		Tasks:       tasks,
		concurrency: concurrency,
		tasksChan:   make(chan *Task),
	}
}

func (pool *Pool) EachError(callback func(*Task)) {
	for _, task := range pool.Tasks {
		if task.Error != nil {
			callback(task)
		}
	}
}

func (pool *Pool) Run() {
	for i := 0; i < pool.concurrency; i++ {
		go pool.work()
	}

	pool.waitGroup.Add(len(pool.Tasks))
	for _, task := range pool.Tasks {
		pool.tasksChan <- task
	}
	close(pool.tasksChan)

	pool.waitGroup.Wait()
}

func (pool *Pool) work() {
	for task := range pool.tasksChan {
		task.Run(&pool.waitGroup)
	}
}
