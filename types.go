package main

import (
	"fmt"
	"sync"
	"time"
)

type callback func(kase *testKase, startTime, endTime, moTimeNow time.Time)

type testKase struct {
	sql  string
	dst  []any
	hook callback
}

type Result struct {
	localQueryStart time.Time
	localQueryEnd   time.Time
	moTimeNow       time.Time
	statement       string
	account         string
	status          string
	response_at     time.Time
}

func (r *Result) String() string {
	return fmt.Sprintf("startTime(local):%v execTime:%v moStartTime:%v statement: %s, account: %s, status: %s, response_at: %v\n",
		r.localQueryStart,
		r.localQueryEnd.Sub(r.localQueryStart),
		r.moTimeNow,
		r.statement,
		r.account,
		r.status,
		r.response_at)
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
