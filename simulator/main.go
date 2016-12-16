package main

import (
	bs "../proto"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
)

const (
	port = ":50051"
)


type server struct{}

func (s *server) AddBall(in *bs.BallRequest, stream bs.Bounce_AddBallServer) error {
	for i:=0; i < 100; i++ {
		if err := stream.Send(&bs.SceneState{Xpos: int32(3+i), Ypos: int32(4+i)}); err != nil {
			log.Println(err)
			return err
		}
		time.Sleep(500*time.Millisecond)
	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	bs.RegisterBounceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
