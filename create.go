package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"time"
)

func createConn(d time.Duration) {
	db, err := sql.Open("mysql", "dump:111@tcp(127.0.0.1:6001)/")
	if err != nil {
		panic(err)
	}
	// See "Important settings" section.
	db.SetConnMaxLifetime(d)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	defer db.Close()

	res, err := db.Query("select account_name from mo_catalog.mo_account")
	if err != nil {
		panic(err)
	}
	defer res.Close()
	//var name string
	//for res.Next() {
	//	res.Scan(&name)
	//	fmt.Println(name)
	//}
}

func startTicker(sigs chan os.Signal, reqCount int) {
	reqCount = max(1, reqCount)
	d := time.Second / time.Duration(reqCount)
	tk := time.NewTicker(d)
	quit := false
	fmt.Println(d)
	for {
		select {
		case <-tk.C:
			createConn(d)
			//fmt.Println(time.Now())
		case <-sigs:
			quit = true
		}
		if quit {
			break
		}
	}
}
