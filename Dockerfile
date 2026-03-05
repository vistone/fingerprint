# Fingerprint Go v3.0 - Multi-stage Docker Build
FROM golang:1.25-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /build

# 复制 go mod 文件
COPY go.mod go.sum go.work ./
COPY modules/core/go.mod modules/core/
COPY modules/profiles/go.mod modules/profiles/
COPY modules/tls/go.mod modules/tls/
COPY modules/http/go.mod modules/http/
COPY modules/ml/go.mod modules/ml/
COPY modules/defense/go.mod modules/defense/
COPY modules/frontend/go.mod modules/frontend/
COPY modules/gateway/go.mod modules/gateway/
COPY modules/metrics/go.mod modules/metrics/

# 下载依赖
RUN go work sync && go mod download

# 复制源代码
COPY . .

# 构建二进制文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o fingerprint-gateway \
    ./cmd/gateway

# 构建示例程序
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s' \
    -o fingerprint-demo \
    ./examples/v3

# 运行时镜像
FROM scratch

# 从 builder 复制证书
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# 复制二进制文件
COPY --from=builder /build/fingerprint-gateway /usr/local/bin/
COPY --from=builder /build/fingerprint-demo /usr/local/bin/

# 非 root 用户运行
USER 65534:65534

# 暴露端口
EXPOSE 8080 9090

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/fingerprint-gateway", "-health-check"] || exit 1

# 入口
ENTRYPOINT ["/usr/local/bin/fingerprint-gateway"]
CMD ["-http-port=8080", "-grpc-port=9090"]
