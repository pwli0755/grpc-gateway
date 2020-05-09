package server

import (
	"context"
	pb "grpc_gateway/proto"
)

type helloService struct{}

func NewHelloService() *helloService {
	return &helloService{}
}

func (h *helloService) SayHelloWorld(ctx context.Context, req *pb.HelloWorldRequest) (*pb.HelloWorldResponese, error) {
	return &pb.HelloWorldResponese{
		Message: req.Referer + " world",
	}, nil
}
