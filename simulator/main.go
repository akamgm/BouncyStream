package main

import (
	bs "../proto"
	"log"

	"net"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	port        = ":50051"
	TICK_MS     = 10
	BALL_RADIUS = 10
	BOARD_SIZE  = 500
)

type SimServer struct {
	population []*Ball
	snapshots  chan *bs.WorldState
}

func NewSimServer() *SimServer {
	var ss SimServer
	ss.snapshots = make(chan *bs.WorldState, 64)
	return &ss
}

func (s *SimServer) RunWorld() {
	for {
		s.Tick()
		time.Sleep(TICK_MS * time.Millisecond)
	}
}

func (s *SimServer) Tick() {
	// since population is empty if there are no clients, use
	// this as a proxy for whether to emit state
	if len(s.population) == 0 {
		return
	}

	s.UpdatePositions()

	// emit scene here
	var world bs.WorldState
	for _, b := range s.population {
		world.Balls = append(world.Balls, b.ToProto())
	}
	s.snapshots <- &world
}

func (s *SimServer) UpdatePositions() {
	for _, b := range s.population {
		b.UpdatePosition()
	}
}

func (s *SimServer) AddBall(in *bs.BallRequest, stream bs.Bounce_AddBallServer) error {
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

func (s *SimServer) AddBallStream(stream bs.Bounce_AddBallStreamServer) error {
	reqs := make(chan *bs.BallRequest, 32)
	go func() {
		for {
			in, err := stream.Recv()
			if err != nil {
				log.Println(err)
				return
			}
			reqs <- in
		}
	}()

	for {
		select {
		case r := <-reqs:
			log.Printf("r: %v\n", r)
			s.population = append(s.population, NewBall(r.Id))
		case e := <-s.snapshots:
			if err := stream.Send(e); err != nil {
				log.Println(err)
				return err
			}
		}
	}
}

func (s *SimServer) RegisterClient(ctx context.Context, req *bs.RegisterRequest) (*bs.RegisterResponse, error) {
	log.Printf("Hello %s\n", req.ClientId)
	return &bs.RegisterResponse{BoardSize: BOARD_SIZE, BallRadius: BALL_RADIUS}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	ss := NewSimServer()
	go ss.RunWorld()

	s := grpc.NewServer()
	bs.RegisterBounceServer(s, ss)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
