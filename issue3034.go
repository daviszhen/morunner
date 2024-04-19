package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
)

var kasesLoad = []*testKase{
	{
		sql: "drop database if exists morunner; create database morunner; use morunner;",
	},
	{
		sql: "create table t1(a int,b int);",
	},
	{
		sql:         "",
		sqlTemplate: "LOAD DATA INFILE '%s' INTO TABLE morunner.t1 FIELDS TERMINATED BY ',' ENCLOSED BY '\"' LINES TERMINATED BY '\\n' PARALLEL 'TRUE';\n",
		prepare: func(kase *testKase, startTime, endTime, moTimeNow time.Time) {
			csvPStr := path.Join(getCurrentAbPath(), "./test.csv")
			fmt.Println(csvPStr)
			kase.sql = fmt.Sprintf(kase.sqlTemplate, csvPStr)
		},
	},
	{
		sql:         "",
		sqlTemplate: "LOAD DATA INFILE '%s' INTO TABLE morunner.t1 FIELDS TERMINATED BY ',' ENCLOSED BY '\"' LINES TERMINATED BY '\\n' PARALLEL 'TRUE';\n",
		prepare: func(kase *testKase, startTime, endTime, moTimeNow time.Time) {
			csvPStr := path.Join(getCurrentAbPath(), "./test.csv")
			fmt.Println(csvPStr)
			kase.sql = fmt.Sprintf(kase.sqlTemplate, csvPStr)
		},
	},
	{
		sql:        "select * from t1;",
		dropResult: true,
	},
	{
		sql: "drop database if exists morunner;",
	},
}

func getCurrentAbPath() string {
	dir := getCurrentAbPathByExecutable()
	if strings.Contains(dir, getTmpDir()) {
		return getCurrentAbPathByCaller()
	}
	return dir
}

func getTmpDir() string {
	dir := os.Getenv("TEMP")
	if dir == "" {
		dir = os.Getenv("TMP")
	}
	res, _ := filepath.EvalSymlinks(dir)
	return res
}

func getCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

func getCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}

func load() {
	establishConn()
	var err error
	for i := 0; i < 1; i++ {
		for _, t := range kasesLoad {
			err = runCase(t)
			if err != nil {
				logger.Error("kase", zap.String("sql", t.sql), zap.Error(err))
				return
			}
		}
	}

}
