package main

import (
	bs "../proto"

	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

const (
	port    = ":8080"
	simPort = ":50051"
)

var homeTemplate = template.Must(template.ParseFiles("home.html"))
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	simConn, err := grpc.Dial(simPort, grpc.WithInsecure())
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}

	defer simConn.Close()
	client := bs.NewBounceClient(simConn)
	stream, err := client.AddBall(context.Background(), &bs.BallRequest{Id: "ball 1"})
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

		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		fmt.Fprintf(w, "%d,%d", state.Xpos, state.Ypos)
	}

}

func serveWs2(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	simConn, err := grpc.Dial(simPort, grpc.WithInsecure())
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}

	defer simConn.Close()
	client := bs.NewBounceClient(simConn)
	stream, err := client.AddBallStream(context.Background())
	if err != nil {
		grpclog.Fatalf("%v.AddBallStream(_) = _, %v", client, err)
	}

	// read from Simulator, write to client
	go func() {
		for {
			state, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				grpclog.Fatalf("%v.AddBallStream(_) = _, %v", client, err)
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			j, _ := json.Marshal(state)
			fmt.Fprintf(w, "%s\n", string(j))
		}
	}()

	stream.Send(&bs.BallRequest{Id: "ball 1"})

	// read from client, write to sim
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		log.Printf("message received: %s\n", message)
		stream.Send(&bs.BallRequest{Id: "ball 1"})
	}

}

func registerWithSim(id string) (boardSize, ballRadius int32) {
	conn, err := grpc.Dial(simPort, grpc.WithInsecure())
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}

	defer conn.Close()
	client := bs.NewBounceClient(conn)

	reg, err := client.RegisterClient(context.Background(), &bs.RegisterRequest{ClientId: id})
	if err != nil {
		grpclog.Fatalf("%v.RegisterClient(_) = _, %v", client, err)
	}

	return reg.BoardSize, reg.BallRadius
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	boardsize, ballradius := registerWithSim("whatevs")

	worldinfo := struct {
		BoardSize  int32
		BallRadius int32
		Host       string
	}{
		boardsize,
		ballradius,
		r.Host,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTemplate.Execute(w, worldinfo)
}

func main() {
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	http.HandleFunc("/ws2", serveWs2)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
