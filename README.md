# 对于 Go 编译器来说，它有两条死命令

1.同一个文件夹（目录）下的所有 .go 文件，第一行的 package 名字必须完全相同
2.文件夹的名字，决定了别人怎么 import 它；而文件内 package 的名字，决定了别人在代码里怎么 use（调用）它
也就是说import导的是物理路径，而代码调用走的是命名空间 package


 
gRPC 的底层是基于 HTTP/2  网络上真正发出的，是一个 HTTP/2 的 POST 请求

这个 HTTP/2 请求的 Path（路径） 到底长什么样？它的计算公式是硬编码在 gRPC 规范里的：

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/order/v1/order.proto



# dockerfile
切到项目根目录下，执行以下构建命令：
docker build -t geek-grpc-cloud/order-server:v1.0.0 .


末尾有一个英文句号 .，代表以当前目录作为上下文进行构建