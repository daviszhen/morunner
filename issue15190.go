package main

import (
	"fmt"
	"go.uber.org/zap"
	"time"
)

var kases15190 = []*testKase{
	{
		sql: "show session variables like '%save_query_result%';",
		dst: []any{
			new(string),
			new(string),
		},
		prepare: func(kase *testKase, startTime, endTime, moTimeNow time.Time) {
			kase.dst = make([]any, 2)
			kase.dst[0] = new(string)
			kase.dst[1] = new(string)
		},
		hook: func(kase *testKase, startTime, endTime, _ time.Time) {
			fmt.Println(*kase.dst[0].(*string), *kase.dst[1].(*string))
		},
	},
	{
		sql: "explain select * from mo_catalog.mo_user;",
		dst: []any{
			new(string),
		},
		prepare: func(kase *testKase, startTime, endTime, moTimeNow time.Time) {
			kase.dst = make([]any, 1)
			kase.dst[0] = new(string)
		},
		hook: func(kase *testKase, startTime, endTime, _ time.Time) {
			fmt.Println(*kase.dst[0].(*string))
		},
	},
	{
		sql: "select database() as db,last_query_id() as query_id;",
		dst: []any{
			new(string),
			new(string),
		},
		prepare: func(kase *testKase, startTime, endTime, moTimeNow time.Time) {
			kase.dst = make([]any, 2)
			kase.dst[0] = new(string)
			kase.dst[1] = new(string)
		},
		hook: func(kase *testKase, startTime, endTime, _ time.Time) {
			fmt.Println(*kase.dst[0].(*string), *kase.dst[1].(*string))
		},
	},
	{
		sql:         "",
		sqlTemplate: "select * from result_scan('%s') as t limit 0,1000;",
		dst: []any{
			new(string),
		},
		inputParams: make([]string, 1),
		prepare: func(kase *testKase, startTime, endTime, moTimeNow time.Time) {
			kase.dst = make([]any, 1)
			kase.dst[0] = new(string)
			kase.sql = fmt.Sprintf(kase.sqlTemplate, kase.inputParams[0])
			fmt.Println(kase.sql)
		},
		hook: func(kase *testKase, startTime, endTime, _ time.Time) {

		},
	},
}

func issue15190() {
	establishConn()
	err := runCase(kases15190[0])
	if err != nil {
		logger.Error("kase1", zap.Error(err))
		return
	}
	err = runCase(kases15190[1])
	if err != nil {
		logger.Error("kase2", zap.Error(err))
		return
	}
	err = runCase(kases15190[2])
	if err != nil {
		logger.Error("kase3", zap.Error(err))
		return
	}
	kases15190[3].inputParams[0] = *(kases15190[2].dst[1].(*string))
	err = runCase(kases15190[3])
	if err != nil {
		logger.Error("kase4", zap.Error(err))
		return
	}
	want := *(kases15190[1].dst[0].(*string))
	realRes := *(kases15190[3].dst[0].(*string))
	if want != realRes {
		logger.Error("not equal", zap.String("want", want), zap.String("real", realRes))
	}
}
