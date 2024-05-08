# morunner



## FE回归

### build
go build -o main *.go

### 测试OOM

每秒500个短链接。2ms 1个链接。
./main -reqcount 500

### 测试查询一致性
./main --loop --url freetier-01.cn-hangzhou.cluster.aliyun-dev.matrixone.tech --user dump --password 

http_proxy="" all_proxy="" curl http 127.0.0.1:8080/status

### 测试ISSUE15190 query_result
./main -testcase 1

### 测试在FE拼接结果的sql

./main -testcase 2 

### 测试load

./main -testcase 3


## PR checklist

- [ ]  bvt
- [ ]  ut
  - [ ] FE UT
- [ ]  sca
- [ ]  事务
    - [ ]  乐观 bvt
    - [ ]  悲观 bvt
    - [ ]  泄露
- [ ]  回归
    - [ ]  ci regression
    - [ ]  morunner
        - [ ]  短链
        - [ ]  FE 处理的语句回归
- [ ]  proxy和session迁移


      //启动集群
    
      ./mo-service -debug-http 127.0.0.1:6060 -cfg etc/launch-with-proxy/log.toml >log.log
      ./mo-service -debug-http 127.0.0.1:6061 -cfg etc/launch-with-proxy/tn.toml >tn.log
      ./mo-service -debug-http 127.0.0.1:6062 -cfg etc/launch-with-proxy/cn1.toml >cn1.log
      ./mo-service -debug-http 127.0.0.1:6063 -cfg etc/launch-with-proxy/proxy.toml >proxy.log
    
      //跑tpcc
    
      create database tpcc;
    
      cd mo-tpcc
      ./runSQL.sh [props.mo](http://props.mo/) tableCreates
      mysql -h127.0.0.1 -udump -P6009 -p111 tpcc <load-tpcc-w1.sql
      ./runBenchmark.sh [props.mo](http://props.mo/)
    
      然后再启动cn2:
    
      ./mo-service -cfg etc/launch-with-proxy/cn2.toml >cn2.log
    
      迁移前后用sql命令观察：
      select node_id, conn_id, user, account from processlist() a where user='dump';
    
      有些session的node_id 会变化
    
      //用这个命令把一个cn设置成draining状态，就会迁移到另外一个cn上。uuid换成具体的cn node-id
    
      select mo_ctl('cn', 'workstate', 'uuid:2');
    
      //这个是设置会working状态，就会再次均衡连接
    
      select mo_ctl('cn', 'workstate', 'uuid:1');

- [ ] Kill问题
- [ ] sysbench 点查

    main: 6b1a10d62ec53a54394120d8c6327c7886c1ce15
    
    sysbench:
    内存 1.3g
    19000

- [ ] tpcc

    main: 6b1a10d62ec53a54394120d8c6327c7886c1ce15
    
    10 仓 10 并发
    
    内存4.6g
    
    **avg** tpm total 7100
    
    **current** tpm total 640000
    
    memory 420Mb/1000MB