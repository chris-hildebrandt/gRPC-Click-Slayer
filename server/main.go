package main

import (
    pb "grpc-click-slayer/server/proto"
    "google.golang.org/grpc"
    "google.golang.org/protobuf/types/known/emptypb"
    "net"
    "log"
    "sync"
)

type server struct {
    pb.UnimplementedClickRaceServer
    mu     sync.Mutex
    scores map[string]int32
    subs   []chan *pb.Leaderboard
}

func newServer() *server {
    return &server{scores: make(map[string]int32)}
}

func (s *server) SendClicks(stream pb.ClickRace_SendClicksServer) error {
    for {
        click, err := stream.Recv()
        if err != nil {
            return err
        }
        s.mu.Lock()
        s.scores[click.Player] += click.Amount
        s.broadcast()
        score := &pb.Score{Player: click.Player, Score: s.scores[click.Player]}
        s.mu.Unlock()
        if err := stream.SendAndClose(score); err != nil {
            return err
        }
    }
}

func (s *server) GetLeaderboard(_ *emptypb.Empty, stream pb.ClickRace_GetLeaderboardServer) error {
    ch := make(chan *pb.Leaderboard, 1)
    s.mu.Lock()
    s.subs = append(s.subs, ch)
    s.mu.Unlock()

    for lb := range ch {
        if err := stream.Send(lb); err != nil {
            return err
        }
    }
    return nil
}

func (s *server) broadcast() {
    lb := &pb.Leaderboard{}
    for p, sc := range s.scores {
        lb.Scores = append(lb.Scores, &pb.Score{Player: p, Score: sc})
    }
    for _, sub := range s.subs {
        sub <- lb
    }
}

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    grpcServer := grpc.NewServer()
    pb.RegisterClickRaceServer(grpcServer, newServer())
    log.Println("Server listening on :50051")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}