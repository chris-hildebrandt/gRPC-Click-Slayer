package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"sync"
	"time"

	pb "grpc-click-slayer/server/proto"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type gameServer struct {
	pb.UnimplementedMonsterSlayerServer
	mu          sync.RWMutex
	players     map[string]*pb.Player
	monster     *pb.Monster
	subscribers []chan *pb.GameUpdate
}

func newGameServer() *gameServer {
	gs := &gameServer{
		players:     make(map[string]*pb.Player),
		subscribers: make([]chan *pb.GameUpdate, 0),
	}
	gs.loadPlayerData()
	gs.createNewMonster()
	return gs
}

func (gs *gameServer) loadPlayerData() {
	data, err := os.ReadFile("players.json")
	if err != nil {
		log.Println("No existing player data, starting fresh")
		return
	}

	var players map[string]*pb.Player
	err = json.Unmarshal(data, &players)
	if err != nil {
		log.Printf("Error loading player data: %v", err)
		return
	}

	gs.mu.Lock()
	gs.players = players
	gs.mu.Unlock()

	log.Printf("Loaded %d players", len(players))
}

func (gs *gameServer) savePlayerData() {
	gs.mu.RLock()
	players := make(map[string]*pb.Player)
	for name, player := range gs.players {
		players[name] = player
	}
	gs.mu.RUnlock()

	data, err := json.MarshalIndent(players, "", "  ")
	if err != nil {
		log.Printf("Error saving player data: %v", err)
		return
	}

	err = os.WriteFile("players.json", data, 0644)
	if err != nil {
		log.Printf("Error writing player data: %v", err)
	}
}

func (gs *gameServer) createNewMonster() {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.monster = &pb.Monster{
		Name:          "Dragon",
		MaxHealth:     10000,
		CurrentHealth: 10000,
		Level:         1,
		LootTable:     []string{"sword", "shield", "potion", "bow", "armor"},
		LastUpdated:   time.Now().Unix(),
	}

	log.Println("Created new monster:", gs.monster.Name)
}

func (gs *gameServer) JoinGame(ctx context.Context, req *pb.JoinGameRequest) (*pb.JoinGameResponse, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	player, exists := gs.players[req.PlayerName]
	if !exists {
		player = &pb.Player{
			Name:          req.PlayerName,
			TotalClicks:   0,
			LootCollected: []string{},
			Level:         1,
			Experience:    0,
			LastPlayed:    time.Now().Unix(),
		}
		gs.players[req.PlayerName] = player
		gs.savePlayerData()
	}

	return &pb.JoinGameResponse{
		Player:  player,
		Monster: gs.monster,
	}, nil
}

func (gs *gameServer) AttackMonster(ctx context.Context, req *pb.AttackRequest) (*pb.AttackResponse, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	player, exists := gs.players[req.PlayerName]
	if !exists {
		return nil, grpc.Errorf(grpc.Code(grpc.NotFound), "Player not found")
	}

	// Update player stats
	player.TotalClicks++
	player.Experience += req.Damage
	player.LastPlayed = time.Now().Unix()

	// Damage monster
	gs.monster.CurrentHealth -= req.Damage
	gs.monster.LastUpdated = time.Now().Unix()

	// Check if monster is defeated
	monsterDefeated := gs.monster.CurrentHealth <= 0
	lootGained := ""

	if monsterDefeated {
		// Give loot
		if len(gs.monster.LootTable) > 0 {
			lootGained = gs.monster.LootTable[0] // Simple loot system
			player.LootCollected = append(player.LootCollected, lootGained)
			player.Experience += 100 // Bonus XP for killing monster
		}
		// Create new monster
		gs.createNewMonster()
	}

	// Save player data
	gs.savePlayerData()

	// Broadcast update
	gs.broadcastUpdate()

	return &pb.AttackResponse{
		Player:          player,
		Monster:         gs.monster,
		MonsterDefeated: monsterDefeated,
		LootGained:      lootGained,
	}, nil
}

func (gs *gameServer) StreamGameUpdates(empty *emptypb.Empty, stream pb.MonsterSlayer_StreamGameUpdatesServer) error {
	ch := make(chan *pb.GameUpdate, 1)
	gs.mu.Lock()
	gs.subscribers = append(gs.subscribers, ch)
	gs.mu.Unlock()

	// Send initial state
	gs.mu.RLock()
	gameUpdate := &pb.GameUpdate{
		Monster:   gs.monster,
		Players:   gs.getPlayersList(),
		Timestamp: time.Now().Unix(),
	}
	gs.mu.RUnlock()

	if err := stream.Send(gameUpdate); err != nil {
		return err
	}

	// Stream updates
	for update := range ch {
		if err := stream.Send(update); err != nil {
			gs.mu.Lock()
			// Remove this subscriber
			for i, sub := range gs.subscribers {
				if sub == ch {
					gs.subscribers = append(gs.subscribers[:i], gs.subscribers[i+1:]...)
					break
				}
			}
			gs.mu.Unlock()
			return err
		}
	}
	return nil
}

func (gs *gameServer) getPlayersList() []*pb.Player {
	players := make([]*pb.Player, 0, len(gs.players))
	for _, player := range gs.players {
		players = append(players, player)
	}
	return players
}

func (gs *gameServer) broadcastUpdate() {
	gameUpdate := &pb.GameUpdate{
		Monster:   gs.monster,
		Players:   gs.getPlayersList(),
		Timestamp: time.Now().Unix(),
	}

	for _, sub := range gs.subscribers {
		select {
		case sub <- gameUpdate:
		default:
			// Channel is full, skip this subscriber
		}
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMonsterSlayerServer(grpcServer, newGameServer())

	log.Println("🐉 Monster Slayer gRPC server listening on :50051")
	log.Println("📊 Join game: JoinGame")
	log.Println("⚔️  Attack monster: AttackMonster")
	log.Println("🔌 Stream updates: StreamGameUpdates")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
