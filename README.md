# morunner

## build
go build -o main main.go types.go

## 测试OOM

每秒500个短链接。2ms 1个链接。
./main -reqcount 500

## 测试查询一致性
./main --loop --url freetier-01.cn-hangzhou.cluster.aliyun-dev.matrixone.tech --user dump --password 

http_proxy="" all_proxy="" curl http 127.0.0.1:8080/status
