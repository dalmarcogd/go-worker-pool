package worker

import (
	"fmt"
	"github.com/google/uuid"
	"strconv"
	"sync"
	"time"
)

//NewWorker
func NewWorker(name string, handle func() error, concurrency int, restartAlways bool) *Worker {
	id, _ := uuid.NewUUID()
	return &Worker{
		ID:            id.String(),
		Name:          name,
		Handle:        handle,
		Concurrency:   concurrency,
		RestartAlways: restartAlways,
		subWorkers:    make(map[string]*SubWorker),
		Restarts:      0,
	}
}

//Run
func (w *Worker) Run(errors chan WrapperHandleError) {
	var wg sync.WaitGroup
	w.StartAt = time.Now().UTC()
	for i := 1; i <= w.Concurrency; i++ {
		s := &SubWorker{ID: i, Status: STARTED, Worker: w}
		w.subWorkers[s.Name()] = s

		wg.Add(1)
		go func(handle func() error, sw *SubWorker) {
			defer wg.Done()
			if err := handle(); err != nil {
				errors <- WrapperHandleError{worker: w, err: err}
				sw.Status = ERROR
				sw.Error = err
			} else {
				sw.Status = FINISHED
			}
		}(w.Handle, s)
	}
	wg.Wait()
}

//Status
func (w *Worker) Status() map[string]string {
	status := map[string]string{}
	for _, subWorker := range w.subWorkers {
		status[subWorker.Name()] = subWorker.Status
	}
	return status
}

//IsUp
func (w *Worker) IsUp() bool {
	for _, v := range w.Status() {
		if v == STARTED {
			return true
		}
	}
	return false
}

//Name
func (s SubWorker) Name() string {
	return fmt.Sprintf("%s-%s", s.Worker.Name, strconv.Itoa(s.ID))
}
