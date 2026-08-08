# 构建阶段
FROM golang:1.26-alpine AS builder

# 环境变量切换到国内源，否则docker pull容易卡死
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

# 将工作目录切换到该子系统的/app下
WORKDIR /app

# 复制依赖文件，并自动下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o word .

# 运行阶段
FROM alpine:latest

# 证书相关
RUN apk --no-cache add ca-certificates tzdata

# 将工作目录切换到该子系统的/app下
WORKDIR /app

# 复制构建阶段生成的可执行文件
COPY --from=builder /app/word .
COPY --from=builder /app/config ./config

# 暴露端口
EXPOSE 8080

# 运行可执行文件
CMD ["./word"]