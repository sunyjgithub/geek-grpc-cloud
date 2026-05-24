package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	// 🛠️ 严格引入你的 API 契约层
	orderv1 "geek-grpc-cloud/api/order/v1"
)

func main() {
	target := "127.0.0.1:50051"

	// 1. 配置工业级客户端保活策略（防御网络隐形断连）
	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second, // 每 10 秒空闲发送一次 HTTP/2 PING 帧
		Timeout:             time.Second,      // PING 帧 1 秒无响应视为断连
		PermitWithoutStream: true,             // 即使没有活跃的 RPC 流量，也允许发送 PING
	}

	// 2. 声明式配置连接参数（大厂标配）
	// 1. 声明式配置连接参数（去掉过时的 WithBlock）
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(kacp),
	}

	log.Printf("🔌 正在尝试建立与 gRPC 服务端 [%s] 的骨干长连接...", target)

	// 3. 建立基于 HTTP/2 的物理长连接（在整个应用生命周期中，复用这一个 conn！）
	// 2. 🔥 替换为官方推荐的最新 API：grpc.NewClient（该方法天然异步非阻塞，无需 Context 限制）
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		log.Fatalf("❌ 无法创建 gRPC 客户端连接: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	// 4. 实例化由契约生成的客户端 Stub
	client := orderv1.NewOrderServiceClient(conn)

	// 5. 构造契约规定的请求体
	req := &orderv1.CreateOrderRequest{
		UserId:    "panic",
		ProductId: 88888888,
		Quantity:  2,
		Price:     199.99,
	}

	log.Println("📤 正在发起 CreateOrder (Unary RPC) 单向请求...")

	// 6. 严密附加防御性超时控制（防御长尾延迟，大厂核心天条！）
	rpcCtx, rpcCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer rpcCancel()

	// 7. 触发真正的网络通信
	resp, err := client.CreateOrder(rpcCtx, req)
	if err != nil {
		log.Fatalf("❌ 订单创建失败，发生 RPC 错误: %v", err)
	}

	// 8. 打印响应结果，并演示生产环境如何将其转为 JSON
	log.Printf("🎉 【gRPC 响应成功】订单ID: %s, 状态: %s", resp.OrderId, resp.Status)

	// 炫技：将 Protobuf 结构体转为标准 JSON 字符串输出
	jsonBytes, _ := json.MarshalIndent(resp, "", "  ")
	log.Printf("📜 格式化标准业务 JSON 出参:\n%s", string(jsonBytes))
}
