# 强制指定在 Windows CMD 模式下执行命令
SHELL := cmd.exe

.PHONY: proto
# 一键生成 gRPC 代码
proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/order/v1/order.proto

.PHONY: clean
# 适配 Windows 的清理命令（/Q 代表安静模式不提示，/F 代表强制删除）
clean:
	if exist api\order\v1\*.pb.go del /Q /F api\order\v1\*.pb.go