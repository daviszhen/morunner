package main

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

var execCase = []*testKase{
	{
		sql: "select connection_id();",
		dst: []any{
			new(int64),
		},
		prepare: func(kase *testKase, startTime, endTime, moTimeNow time.Time) {
			kase.dst = make([]any, 1)
			kase.dst[0] = new(int64)
		},
		hook: func(kase *testKase, startTime, endTime, _ time.Time) {
			fmt.Println("connection_id", *kase.dst[0].(*int64))
		},
	},
}

func execCmd(sigs chan os.Signal, sql string) error {
	establishConn()
	err := runCase(execCase[0])
	if err != nil {
		return err
	}
	result, err := conn.Query(sql)
	if err != nil {
		logger.Error("query failed", zap.String("sql", sql), zap.Error(err))
		return err
	}
	defer result.Close()
	return err
}
