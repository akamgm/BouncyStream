package main

import (
	bs "../proto"
	"log"
	"math/rand"
	"net"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	port        = ":50051"
	TICK_MS     = 100
	BALL_RADIUS = 5
	BOARD_SIZE  = 500
)

type Ball struct {
	Id     string
	Xpos   int
	Ypos   int
	Xspeed int
	Yspeed int
}

func NewBall(id string) *Ball {
	// just initialize with some random values
	var b Ball
	b.Id = id
	b.Xpos = rand.Intn(BOARD_SIZE - 2*BALL_RADIUS)
	b.Ypos = rand.Intn(BOARD_SIZE - 2*BALL_RADIUS)
	b.Xspeed = rand.Intn(2*BALL_RADIUS) - BALL_RADIUS
	b.Yspeed = rand.Intn(2*BALL_RADIUS) - BALL_RADIUS
	return &b
}

func (b *Ball) UpdatePosition() {
	// only worry about collisions with the walls, not other balls
	b.Xpos += b.Xspeed
	if b.Xpos+BALL_RADIUS > BOARD_SIZE {
		b.Xpos = BOARD_SIZE - BALL_RADIUS
		b.Xspeed *= -1
	} else if b.Xpos < BALL_RADIUS {
		b.Xpos = BALL_RADIUS
		b.Xspeed *= -1
	}

	b.Ypos += b.Yspeed
	if b.Ypos+BALL_RADIUS > BOARD_SIZE {
		b.Ypos = BOARD_SIZE - BALL_RADIUS
		b.Yspeed *= -1
	} else if b.Ypos < BALL_RADIUS {
		b.Ypos = BALL_RADIUS
		b.Yspeed *= -1
	}
}

func (b *Ball) ToProto() *bs.SceneState {
	return &bs.SceneState{Xpos: int32(b.Xpos), Ypos: int32(b.Ypos), Id: b.Id}
}

type server struct{}

func (s *server) AddBall(in *bs.BallRequest, stream bs.Bounce_AddBallServer) error {
	b := NewBall(in.Id)
	for {
		if err := stream.Send(b.ToProto()); err != nil {
			log.Println(err)
			return err
		}
		time.Sleep(TICK_MS * time.Millisecond)
		b.UpdatePosition()
	}
	return nil
}

func (s *server) RegisterClient(ctx context.Context, req *bs.RegisterRequest) (*bs.RegisterResponse, error) {
	log.Printf("Hello %s\n", req.ClientId)
	return &bs.RegisterResponse{BoardSize: BOARD_SIZE, BallRadius: BALL_RADIUS}, nil
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
