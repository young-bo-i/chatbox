# backend/Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY go.mod ./

# 复制源代码后再下载依赖（让 go mod tidy 生成正确的 go.sum）
COPY . .
RUN go mod tidy

# 构建主服务
RUN CGO_ENABLED=0 go build -o server .

# 构建重置密码工具
RUN CGO_ENABLED=0 go build -o reset-password ./cmd/reset-password

# 运行阶段
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# 复制构建产物
COPY --from=builder /app/server .
COPY --from=builder /app/reset-password .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./server"]
