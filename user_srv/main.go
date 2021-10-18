package main

import (
	"flag"
	"fmt"
	"mytest/user_srv/handler"
	"mytest/user_srv/proto"
	"net"

	"google.golang.org/grpc"
)

func main() {
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int("port", 50051, "端口号")

	flag.Parse()
	fmt.Println("ip: ", *IP)
	fmt.Println("port: ", *Port)

	server := grpc.NewServer()
	proto.RegisterUserServer(server, &handler.UserServer{})
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		panic("Failed to listen:" + err.Error())
	}
	err = server.Serve(lis)
	if err != nil {
		panic("Failed to start grpc:" + err.Error())
	}
}
