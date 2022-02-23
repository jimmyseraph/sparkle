package grpc

import (
	"fmt"

	"github.com/jimmyseraph/sparkle/easy_grpc"
	"github.com/jimmyseraph/sparkle/example/grpc/proto"
)

func CallGrpc() {
	handler := easy_grpc.NewGRPCHandler("localhost:8082")
	client := proto.NewHelloClient(handler.Conn)
	defer handler.Close()
	reply, err := client.SayHi(handler.Ctx, &proto.SayHiRequest{
		Name: "liudao",
		Age:  18,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", reply.Message)
}

func CallGrpcWithDynamicpb() {
	handler := easy_grpc.NewGRPCHandler("localhost:8082")
	defer handler.Close()
	dyGRPC, err := easy_grpc.GenerateDynamicGRPC("/tmp")
	if err != nil {
		panic(err)
	}
	fd, err := dyGRPC.GetDescriptorFileByName("hello.proto")
	if err != nil {
		panic(err)
	}
	dyAPI := easy_grpc.NewDynamicAPI(fd, "Hello", "SayHi")
	reply, err := dyAPI.Invoke(handler, `{
		"name": "liudao",
		"age": 18
	}`)
	if err != nil {
		panic(err)
	}
	fmt.Println(reply)
}
