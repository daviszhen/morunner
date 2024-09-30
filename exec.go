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

func execCmd(sigs chan os.Signal, sql string, isQuery bool, d time.Duration) error {
	establishConn()
	err := runCase(execCase[0])
	if err != nil {
		return err
	}

	fmt.Println("isquery", isQuery, "interval", d)

	if d > 0 {
		tk := time.NewTicker(d)
		for {
			select {
			case <-tk.C:
			case <-sigs:
				return nil
			}

			_ = execSql(sql, isQuery)
		}
	} else {
		err = execSql(sql, isQuery)
		if err != nil {
			return err
		}
	}

	return err
}

func execSql(sql string, isQuery bool) error {
	if isQuery {
		result, err := conn.Query(sql)
		if err != nil {
			logger.Error("query failed", zap.String("sql", sql), zap.Error(err))
			return err
		}
		defer result.Close()
	} else {
		_, err := conn.Exec(sql)
		if err != nil {
			logger.Error("execute failed", zap.String("sql", sql), zap.Error(err))
			return err
		}
	}
	return nil
}
