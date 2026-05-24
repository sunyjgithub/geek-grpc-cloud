# ==========================================
# 第一阶段：编译阶段 (Stage 1: Build)
# ==========================================
# 选用带有完整 Go 工具链的轻量级 alpine 镜像作为编译温床
FROM golang:1.22-alpine AS builder

# 1. 为 Alpine 配置国内加速源，并安装基础构建工具（如 make）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && apk add --no-cache git make

# 2. 设置容器内部的安全工作目录
WORKDIR /app

# 3. 优先拷贝依赖描述文件，完美利用 Docker 的 Layer 缓存机制机制
COPY go.mod go.sum ./

# 4. 开启 Go 模块代理，一键加速依赖下载
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

# 5. 拷贝本地全量源码进入容器
COPY . .

# 6. 【硬核静态编译】
# CGO_ENABLED=0：关闭动态链接，确保编译出的二进制不依赖任何物理机 Linux 动态链接库
# GOOS=linux：强行指定编译为纯正的 Linux 64位可执行文件（即使你在 Win10 上跑）
# -ldflags="-s -w"：彻底擦除符号表与调试信息，极大压缩二进制体积
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o order-server ./cmd/server/main.go


# ==========================================
# 第二阶段：运行阶段 (Stage 2: Run)
# ==========================================
# “过河拆桥”：丢弃第一阶段几百兆的 Go 编译环境，只选用极简、安全的纯净 alpine 运行底座（仅约 5MB）
FROM alpine:3.19

WORKDIR /app

# 7. 从第一阶段名叫 builder 的虚拟镜像中，精准地把编译好的静态二进制文件“偷”过来
COPY --from=builder /app/order-server .

# 8. 声明容器内部需要暴露的 gRPC 核心端口
EXPOSE 50051

# 9. 容器启动的终极终点
CMD ["./order-server"]