package pool

import (
	"sync"
)

//
// Task
//
type Task struct {
	Err error
	run func() error
}

func NewTask(run func() error) *Task {
	return &Task{run: run}
}

func (t *Task) Run(waitGroup *sync.WaitGroup) {
	t.Err = t.run()
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

func (p *Pool) EachError(callback func(*Task)) {
	for _, task := range p.Tasks {
		if task.Err != nil {
			callback(task)
		}
	}
}

func (p *Pool) Run() {
	for i := 0; i < p.concurrency; i++ {
		go p.work()
	}

	p.waitGroup.Add(len(p.Tasks))
	for _, task := range p.Tasks {
		p.tasksChan <- task
	}
	close(p.tasksChan)

	p.waitGroup.Wait()
}

func (p *Pool) work() {
	for task := range p.tasksChan {
		task.Run(&p.waitGroup)
	}
}
