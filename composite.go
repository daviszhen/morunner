package main

import (
	"go.uber.org/zap"
)

var kasesComposite = []*testKase{
	{
		sql:        "show errors",
		dropResult: true,
	},
	{
		sql:        "show subscriptions",
		dropResult: true,
	},
	{
		sql:        "show accounts",
		dropResult: true,
	},
	{
		sql:        "show warnings;",
		dropResult: true,
	},
	{
		sql:        "show variables",
		dropResult: true,
	},
	{
		sql:        "show backend servers",
		dropResult: true,
	},
	{
		sql:        "show connectors",
		dropResult: true,
	},
	{
		sql:        "explain select * from mo_catalog.mo_account ma ;",
		dropResult: true,
	},
	{
		sql:        "show collation",
		dropResult: true,
	},
	{
		sql:        "explain verbose select * from mo_catalog.mo_account ma ;",
		dropResult: true,
	},
	{
		sql:        "explain analyze select * from mo_catalog.mo_account ma ;",
		dropResult: true,
	},
	{
		sql:        "explain analyze verbose select * from mo_catalog.mo_account ma ;",
		dropResult: true,
	},
	{
		sql:        "prepare st1 from select * from  mo_catalog.mo_account ma ;",
		dropResult: true,
	},
	{
		sql:        "execute st1 ",
		dropResult: true,
	},
	{
		sql:        "explain force execute st1 ",
		dropResult: true,
	},
	{
		sql:        "explain verbose force execute st1 ",
		dropResult: true,
	},
	{
		sql:        "explain analyze force execute st1 ",
		dropResult: true,
	},
	{
		sql:        "explain analyze verbose force execute st1 ",
		dropResult: true,
	},
	{
		sql:        "deallocate prepare st1 ",
		dropResult: true,
	},
}

func composite() {
	establishConn()
	var err error
	for i := 0; i < 5; i++ {
		for _, t := range kasesComposite {
			err = runCase(t)
			if err != nil {
				logger.Error("kase", zap.String("sql", t.sql), zap.Error(err))
				return
			}
		}
	}

}
