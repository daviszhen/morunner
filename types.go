package main

import (
	"sync"
	"time"
)

type callback func(kase *testKase)

type testKase struct {
	sql  string
	dst  []any
	hook callback
}

type Result struct {
	statement   string
	account     string
	status      string
	response_at time.Time
}

type MultiResult struct {
	mu       sync.Mutex
	results  []*Result
	maxCount int
}

func (mr *MultiResult) Append(r *Result) {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	if mr.maxCount > 0 && len(mr.results) >= mr.maxCount {
		mr.results = mr.results[1:]
	}
	mr.results = append(mr.results, r)
}

func (mr *MultiResult) List(fn func(int, *Result)) int {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	for i, result := range mr.results {
		if fn != nil {
			fn(i, result)
		}
	}
	return len(mr.results)
}
