package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	//params
	loop                      bool
	sleep                     int
	reconnectInterval         int
	url, port, user, password string
	httpPort                  string

	runCount int
	runStart time.Time
)

var conn *sql.DB
var logger *zap.Logger

var moStartTime time.Time

const maxLen = 100

var failedResults = &MultiResult{
	maxCount: maxLen,
}

var lastResults = &MultiResult{
	maxCount: 10,
}

var kases = []*testKase{
	{
		sql: "select now()",
		hook: func(kase *testKase, startTime, endTime, _ time.Time) {
			if ptr, ok := kase.dst[0].(*time.Time); ok {
				moStartTime = *ptr
				//fmt.Fprintf(os.Stderr, "now: %v\n", moTimeNow)
			}
		},
		dst: []any{
			new(time.Time),
		},
	},
	{
		sql: "select statement,account,status,response_at from system.statement_info order by response_at desc limit 5",
		hook: func(kase *testKase, startTime, endTime, moTimeNow time.Time) {
			if ptr, ok := kase.dst[3].(*time.Time); ok {
				r := &Result{
					localQueryStart: startTime,
					localQueryEnd:   endTime,
					moTimeNow:       moTimeNow,
					statement:       *kase.dst[0].(*string),
					account:         *kase.dst[1].(*string),
					status:          *kase.dst[2].(*string),
					response_at:     *kase.dst[3].(*time.Time),
				}
				if ptr.Before(moTimeNow.Add(-time.Hour)) {
					//fmt.Fprintf(os.Stderr,"invalid time. statement: %s, account: %s, status: %s, response_at: %v\n",
					//   *kase.dst[0].(*string),
					//   *kase.dst[1].(*string),
					//   *kase.dst[2].(*string),
					//   *kase.dst[3].(*time.Time))
					failedResults.Append(r)
				}
				lastResults.Append(r)
			}
		},
		dst: []any{
			new(string),
			new(string),
			new(string),
			new(time.Time),
		},
	},
}

func main() {
	flag.BoolVar(&loop, "loop", false, "loop")
	flag.IntVar(&sleep, "sleep", 60, "sleep timeout seconds")
	flag.IntVar(&reconnectInterval, "reconnect-gap", 30, "reconnect interval seconds")
	flag.StringVar(&url, "url", "127.0.0.1", "url")
	flag.StringVar(&port, "port", "6001", "port")
	flag.StringVar(&user, "user", "dump", "user")
	flag.StringVar(&password, "password", "111", "password")
	flag.StringVar(&httpPort, "http-port", "8080", "http port")
	flag.Parse()

	logger, _ = zap.NewProduction()
	defer logger.Sync()
	defer func() {
		logger.Info("exit")
	}()

	go httpServer()

	runStart = time.Now()
	runCases()
	runCount++
	for loop {
		runCases()
		if loop {
			time.Sleep(time.Duration(sleep) * time.Second)
		}
		runCount++
	}
}

func httpServer() {
	if !loop {
		return
	}
	http.HandleFunc("/status", func(writer http.ResponseWriter, request *http.Request) {
		//params
		_, _ = writer.Write([]byte(fmt.Sprintf("loop %v\n", loop)))
		_, _ = writer.Write([]byte(fmt.Sprintf("sleep %d\n", sleep)))
		_, _ = writer.Write([]byte(fmt.Sprintf("reconnectInterval %d\n", reconnectInterval)))
		_, _ = writer.Write([]byte(fmt.Sprintf("url %s\n", url)))
		_, _ = writer.Write([]byte(fmt.Sprintf("port %s\n", port)))
		_, _ = writer.Write([]byte(fmt.Sprintf("user %s\n", user)))
		_, _ = writer.Write([]byte(fmt.Sprintf("password %s\n", password)))
		_, _ = writer.Write([]byte(fmt.Sprintf("httpPort %s\n", httpPort)))

		//status
		_, _ = writer.Write([]byte(fmt.Sprintf("runStart %v last %v\n", runStart, time.Since(runStart))))
		_, _ = writer.Write([]byte(fmt.Sprintf("runCount %d\n", runCount)))
		_, _ = writer.Write([]byte(fmt.Sprintf("queryStartTime(matrixone) %v\n", moStartTime)))

		_, _ = writer.Write([]byte(fmt.Sprintf("\n\n")))

		_, _ = writer.Write([]byte(fmt.Sprintf("test kases:\n")))
		//test sql
		for i, kase := range kases {
			_, _ = writer.Write([]byte(fmt.Sprintf("kase %d: sql: %s\n", i, kase.sql)))
		}
		_, _ = writer.Write([]byte(fmt.Sprintf("\n\n")))

		printResult := func(i int, result *Result) {
			_, _ = writer.Write([]byte(fmt.Sprintf("result %d: %v",
				i, result.String())))
		}

		//last results
		_, _ = writer.Write([]byte(fmt.Sprintf("last succeeded kases:\n")))
		lastResults.List(printResult)

		//failed results
		_, _ = writer.Write([]byte(fmt.Sprintf("\n\nlast failed kases:\n")))
		cnt := failedResults.List(printResult)

		if cnt > 0 {
			_, _ = writer.Write([]byte(fmt.Sprintf("results count: %d\n", cnt)))
		} else {
			_, _ = writer.Write([]byte(fmt.Sprintf("no failed result\n")))
		}
	})
	_ = http.ListenAndServe(fmt.Sprintf("127.0.0.1:%s", httpPort), nil)
}

func connectDb(url, port, user, password string) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true", user, password, url, port))
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(-1)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, err
}

func establishConn() {
	var err error
	for conn == nil {
		conn, err = connectDb(url, port, user, password)
		if err != nil {
			logger.Error("connect db failed", zap.String("url", url), zap.String("port", port), zap.String("user", user), zap.Error(err))
			if conn != nil {
				_ = conn.Close()
				conn = nil
			}

			time.Sleep(time.Duration(reconnectInterval) * time.Second)
		}
	}
}

func runCases() {
	establishConn()
	for _, kase := range kases {
		err := runCase(kase)
		if err != nil {
			logger.Error("run case failed", zap.String("sql", kase.sql), zap.Error(err))
			break
		}
	}
}

func runCase(kase *testKase) error {
	start := time.Now()
	result, err := conn.Query(kase.sql)
	if err != nil {
		return err
	}
	defer result.Close()
	end := time.Now()

	for result.Next() {
		err = result.Scan(kase.dst...)
		if err != nil {
			return errors.Join(err, result.Err())
		}
		if kase.hook != nil {
			kase.hook(kase, start, end, moStartTime)
		}
	}
	return err
}
