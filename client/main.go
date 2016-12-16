package main

import (
	bs "../proto"
	"io"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

const (
	port = ":50051"
)

func main() {
	conn, err := grpc.Dial(port, grpc.WithInsecure())
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}

	defer conn.Close()
	client := bs.NewBounceClient(conn)

	stream, err := client.AddBall(context.Background(), &bs.BallRequest{Id: "whatevs"})
	if err != nil {
		grpclog.Fatalf("%v.AddBall(_) = _, %v", client, err)
	}

	for {
		state, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			grpclog.Fatalf("%v.AddBall(_) = _, %v", client, err)
		}
		grpclog.Println(state)
	}
}