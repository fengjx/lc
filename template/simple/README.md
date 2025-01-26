# simple

## 安装依赖

```bash
go mod tidy
```

## 启动服务

```bash
go run main.go
```

## 访问接口

```bash
curl -i 'http://localhost:8080/hello/say-hello' \
--header 'Content-Type: application/json' \
--data '{
    "name": "error"
}'
```
