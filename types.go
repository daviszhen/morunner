package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

type callback func(kase *testKase, startTime, endTime, moTimeNow time.Time)

type testKase struct {
	sql         string
	sqlTemplate string
	dst         []any
	res         string
	inputParams []string
	prepare     callback
	hook        callback
	dropResult  bool
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
	outFile  string
	outF     *os.File
	csvWr    *csv.Writer
}

func (mr *MultiResult) Init() {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	var err error
	mr.outF, err = os.OpenFile(mr.outFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	mr.csvWr = csv.NewWriter(mr.outF)
}

func (mr *MultiResult) Close() {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.csvWr.Flush()
	_ = mr.outF.Close()
}

func (mr *MultiResult) Append(r *Result) {
	var err error
	mr.mu.Lock()
	defer mr.mu.Unlock()
	if mr.maxCount > 0 && len(mr.results) >= mr.maxCount {
		mr.results = mr.results[1:]
	}
	mr.results = append(mr.results, r)
	err = mr.csvWr.Write([]string{
		r.String(),
	})
	if err != nil {
		logger.Error("write %s into csv failed", zap.String("result", r.String()), zap.Error(err))
		return
	}
	mr.csvWr.Flush()
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
