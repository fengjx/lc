# standard

## 快速开始

根据 proto 生成接口代码
```
lc pbgen -f proto/hello/hello.proto
```

启动服务
```bash
go mod tidy
go run main.go
```

请求接口
```bash
curl -i -X POST 'http://localhost:8080/hello/say-hello' \
-H 'Content-Type: application/json' \
-d '{
"name": "fengjx"
}'
```