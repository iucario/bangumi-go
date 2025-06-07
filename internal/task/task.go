package task

import (
	"sync"
)

type Task struct {
	ID string
	Do func() (any, error)
}

type Result struct {
	Data  any
	Error error
}

func Run(tasks []Task) map[string]Result {
	var results sync.Map
	var wg sync.WaitGroup

	for _, task := range tasks {
		t := task
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := t.Do()
			results.Store(t.ID, Result{
				Data:  res,
				Error: err,
			})
		}()
	}
	wg.Wait()

	ret := make(map[string]Result)
	results.Range(func(key, value any) bool {
		id, _ := key.(string)
		ret[id] = value.(Result)
		return true
	})
	return ret
}
