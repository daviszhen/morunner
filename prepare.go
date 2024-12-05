package main

import (
	"time"
)

var prepareInitCases = []*testKase{
	{
		sql: "drop database if exists morunner_prepare; create database morunner_prepare; use morunner_prepare;",
	},
	{
		sql: "create table t1(a int);",
	},
	{
		sql: "insert into t1 values(1),(2),(3),(4),(5);",
	},
}

var prepareCase = []*testKase{
	{
		sql: "select * from t1 where a = ?",
	},
	{
		initPrepare: func(kase *testKase, startTime, endTime, moTimeNow time.Time) {
			kase.prepareParams = make([]any, 5)
			for i := 0; i < 5; i++ {
				kase.prepareParams[i] = i + 1
			}
		},
		prepare: func(kase *testKase, startTime, endTime, moTimeNow time.Time) {
			kase.dst = make([]any, 1)
			kase.dst[0] = new(int)
		},
	},
}

func prepare() error {
	establishConn()
	for _, initCase := range prepareInitCases {
		err := runCase(initCase)
		if err != nil {
			return err
		}
	}
	stmt, err := conn.Prepare(prepareCase[0].sql)
	if err != nil {
		return err
	}
	defer stmt.Close()
	prepareCase[1].initPrepare(prepareCase[1], time.Now(), time.Now(), time.Now())
	for i := 1; i < len(prepareCase); i++ {
		for j := 0; j < len(prepareCase[i].prepareParams); j++ {
			param := prepareCase[i].prepareParams[j]
			err = runPrepareCase(prepareCase[i], stmt, []any{param})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
