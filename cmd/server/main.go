package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	// 🛠️ 严格对齐你的 go.mod 模块名与 api 路径
	orderv1 "geek-grpc-cloud/api/order/v1"
)

// OrderServer 1. 定义订单服务结构体
type OrderServer struct {
	orderv1.UnimplementedOrderServiceServer
}

// CreateOrder 2. 完美实现契约定义的 CreateOrder 接口
func (s *OrderServer) CreateOrder(_ context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	// 🔥 严格对齐契约入参：req.UserId, req.ProductId, req.Quantity, req.Price
	log.Printf("【gRPC 收到订单请求】用户ID: %s, 商品ID: %d, 数量: %d, 单价: %.2f",
		req.UserId, req.ProductId, req.Quantity, req.Price)

	// 故意埋一个雷：如果用户ID是 "panic"，直接触发崩溃，测试我们的 Recovery 拦截器
	if req.UserId == "panic" {
		panic("🔥 模拟严重的运行时数据库连接断开异常！")
	}

	// 模拟工业级分布式生成全局唯一 ID（如雪花算法 Snowflake）
	mockOrderID := fmt.Sprintf("ORD-%d-%d", time.Now().UnixNano(), req.ProductId)

	// 🔥 严格对齐契约出参：OrderId, Status, CreatedAt
	return &orderv1.CreateOrderResponse{
		OrderId:   mockOrderID,
		Status:    "PAID_SUCCESS",    // 工业实践：使用大写字符串状态机，便于可读与排查
		CreatedAt: time.Now().Unix(), // 彻底避开时区地狱，落盘秒级 Unix 时间戳
	}, nil
}

// 🛠️ 大厂级核心：手写一元服务端拦截器
func myServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		startTime := time.Now()

		// 1. 模拟分布式全链路追踪：生成一个伪 TraceID
		mockTraceID := fmt.Sprintf("trace-id-%d", startTime.UnixNano())

		// 2. 统一防御性设计：异常捕获（Recovery）
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[🚨 CRITICAL - TraceID: %s] 捕获到服务崩溃! 错误原因: %v \n 堆栈信息:\n%s",
					mockTraceID, r, string(debug.Stack()))
				// 将运行时的 panic 包装为体面的 gRPC 状态码返回给调用方，防止客户端挂起
				err = status.Errorf(codes.Internal, "Internal server panic caught by interceptor")
			}
		}()

		log.Printf("[📥 入向请求 - TraceID: %s] 调用方法: %s", mockTraceID, info.FullMethod)

		// 3. 执行真正的业务逻辑（进入 CreateOrder 方法）
		resp, err = handler(ctx, req)

		// 4. 统一后置处理：耗时统计与日志审计
		duration := time.Since(startTime)
		log.Printf("[📤 出向响应 - TraceID: %s] 方法: %s, 耗时: %v, 是否报错: %v",
			mockTraceID, info.FullMethod, duration, err != nil)

		return resp, err
	}
}

func main() {
	port := ":50051"

	// 3. 建立物理层面的四层 TCP 监听
	listen, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen TCP port %s: %v", port, err)
	}

	// 4. 引入 Keepalive 机制，防御云原生多层网络下的僵尸连接
	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Minute,
			MaxConnectionAge:      30 * time.Minute, // 强制物理长连接轮转，打破大厂 L4 负载均衡（如 LVS）的连接粘性
			MaxConnectionAgeGrace: 5 * time.Minute,
			Time:                  2 * time.Hour,
			Timeout:               20 * time.Second,
		}),
		grpc.UnaryInterceptor(myServerInterceptor()),
	}

	// 5. 实例化 gRPC 服务端
	grpcServer := grpc.NewServer(opts...)

	// 6. 将严格契约化的业务逻辑注册到 gRPC 运行时中
	orderv1.RegisterOrderServiceServer(grpcServer, &OrderServer{})

	// 7. 优雅启停（Graceful Shutdown）—— 避免大厂滚动发布时订单断流
	go func() {
		log.Printf("🚀 工业级 gRPC 订单微服务已启动，正在监听端口 %s ...", port)
		if err := grpcServer.Serve(listen); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("gRPC server run failed: %v", err)
		}
	}()

	// 监听系统退出信号（Ctrl+C, K8s Pod 销毁信号 SIGTERM）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⚠️ 收到停机信号，正在触发优雅停机...")

	// 发送 HTTP/2 GOAWAY 帧，拒绝新请求，安全消化积压请求，体面下班
	grpcServer.GracefulStop()

	log.Println("🛑 订单服务已安全关闭！")
}
