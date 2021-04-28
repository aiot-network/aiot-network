package gorutinue

import (
	"errors"
	log "github.com/aiot-network/aiotchain/tools/log/log15"
	"sync"
)

const (
	maxWorkCount  = 10000
	maxReadyCount = 10000
	module        = "goroutine"
)

type Task struct {
	f func() error
}

func NewTask(f func() error) *Task {
	return &Task{f: f}
}

func (t *Task) Run(flagCh chan bool) {
	flagCh <- true
	go func() {
		<-flagCh
		t.f()
	}()
}

type Pool struct {
	mutex         sync.RWMutex
	maxReadyCount uint32
	worksCh       chan *Task
	readyCh       chan *Task
	flagCh        chan bool
	wg            sync.WaitGroup
}

func NewPool() *Pool {
	return &Pool{
		maxReadyCount: maxReadyCount,
		worksCh:       make(chan *Task, maxWorkCount),
		readyCh:       make(chan *Task, maxReadyCount),
		flagCh:        make(chan bool, maxWorkCount),
	}
}

func (p *Pool) Name() string {
	return module
}

func (p *Pool) Start() error {
	go p.worksRun()
	go p.readyRun()
	return nil
}

func (p *Pool) Stop() error {
	close(p.worksCh)
	close(p.readyCh)
	close(p.flagCh)
	p.wg.Wait()
	log.Info("Goroutine pool was stopped", "module", module)
	return nil
}

func (p *Pool) Info() map[string]interface{} {
	return make(map[string]interface{}, 0)
}

func (p *Pool) worksRun() {
	p.wg.Add(1)
	defer p.wg.Done()

	for {
		select {
		case task, ok := <-p.worksCh:
			if !ok {
				return
			}
			task.Run(p.flagCh)
		}
	}
}

func (p *Pool) readyRun() {
	p.wg.Add(1)
	defer p.wg.Done()

	for {
		select {
		case task, ok := <-p.readyCh:
			if !ok {
				return
			}
			p.worksCh <- task
		}
	}
}

func (p *Pool) AddTask(task *Task) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if uint32(len(p.readyCh)) == p.maxReadyCount {
		return errors.New("the pool is full, please wait")
	}
	p.readyCh <- task
	return nil
}
