package main

import (
	"fmt"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"mafia_grpc/proto"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

const PLAYERS_NUM = 4

// Roles
const (
	ROLE_NOT_ASSIGNED = iota
	ROLE_CITIZEN
	ROLE_MAFIA
	ROLE_SHERIFF
)

// Current state
const (
	NOT_STARTED = iota
	DAY
	NIGHT
)

// TODO MOVE this to the client part!
func getRoleDescription(role int) string {
	switch role {
	case ROLE_CITIZEN:
		return `You are a CITIZEN!
			Your commands are "vote <player_id>", "skip"`

	case ROLE_MAFIA:
		return `You are a MAFIA!
			Your commands are "kill <player_id>", "vote <player_id>", "skip"`

	case ROLE_SHERIFF:
		return `You are a SHERIFF!
			Your commands are "check <player_id>", expose "<player_id>", "vote <player_id>", "skip"`
	default:
		panic("No such role!")
	}
}

const (
	CITIZEN_NUM = 2
	MAFIA_NUM   = 1
	SHERIFF_NUM = 1
)

var glog grpclog.LoggerV2

func init() {
	glog = grpclog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

type Player struct {
	Id   int
	Name string
	Role int
	// Use Die instead of IsAlive = false!
	IsAlive        bool
	VotedThisRound bool
	// If Mafia or Sheriff did their special actions this round
	DidSpecialAction bool
	Conn             *Connection
}

func (p *Player) Die() {
	p.IsAlive = false
	sendMessageStr("You are DEAD now! ;(", p.Conn)
}

func (p *Player) String() string {
	return fmt.Sprintf("Name: %v - Id: %v - Is Alive: %v", p.Name, p.Id, p.IsAlive)
}

type Connection struct {
	stream proto.Mafia_CreateStreamServer
	//id     string
	active bool
	error  chan error
}

type Server struct {
	//Connection []*Connection
	Players          []*Player
	idCounter        int
	currentState     int
	chanSheriffCheck chan interface{}
	chanMafiaKill    chan *Player
	authKeyToPlayer  map[string]*Player
	idToPlayer       map[int]*Player
	votes            map[int]int
	waitVote         *sync.WaitGroup
}

func (s *Server) StartNewGame(ctx context.Context, empty *proto.Empty) (*proto.Empty, error) {
	if len(s.Players) == PLAYERS_NUM && s.currentState == NOT_STARTED {
		go s.StartGame()
	}
	return &proto.Empty{}, nil
}

func (s *Server) getPlayers(action *proto.PlayerAction) (*Player, *Player, *proto.ActionRespond) {
	fromPlayer := s.authKeyToPlayer[action.SenderAuthKey]
	if fromPlayer == nil {
		return nil, nil, &proto.ActionRespond{Message: "Error: No such player!"} //errors.New("no such player")
	}
	targetPlayer := s.idToPlayer[int(action.TargetId)]
	if targetPlayer == nil {
		return nil, nil, &proto.ActionRespond{Message: "Error: No such target!"} //errors.New("no such target")
	}
	return fromPlayer, targetPlayer, nil
}

func (s *Server) Kill(ctx context.Context, action *proto.PlayerAction) (*proto.ActionRespond, error) {
	// TODO fix possible race condition problem
	fromPlayer, targetPlayer, err := s.getPlayers(action)
	if err != nil {
		return err, nil
	}
	if !fromPlayer.IsAlive ||
			fromPlayer.DidSpecialAction ||
			fromPlayer.Role != ROLE_MAFIA ||
			s.currentState != NIGHT ||
			targetPlayer.Id == fromPlayer.Id ||
			!targetPlayer.IsAlive {
		return &proto.ActionRespond{Message: "Error: You can't do such action right now!"}, nil
	}

	targetPlayer.Die()
	defer func() {
		s.chanMafiaKill <- targetPlayer
	}()
	fromPlayer.DidSpecialAction = true
	return &proto.ActionRespond{Message: fmt.Sprintf("You killed %v", targetPlayer.Name)}, nil
}

func (s *Server) CheckIfMafia(ctx context.Context, action *proto.PlayerAction) (*proto.ActionRespond, error) {
	fromPlayer, targetPlayer, err := s.getPlayers(action)
	if err != nil {
		return err, nil
	}
	if !fromPlayer.IsAlive ||
			fromPlayer.Role != ROLE_SHERIFF ||
			fromPlayer.DidSpecialAction ||
			// I decided that sheriff can check dead bodies
			//!targetPlayer.IsAlive ||
			targetPlayer.Id == fromPlayer.Id ||
			s.currentState != NIGHT {
		return &proto.ActionRespond{Message: "Error: You can't do such action right now!"}, nil
	}

	var result string
	if targetPlayer.Role == ROLE_MAFIA {
		result = fmt.Sprintf("Player %v is mafia!", targetPlayer.Name)
	} else {
		result = fmt.Sprintf("Player %v is NOT mafia", targetPlayer.Name)
	}
	fromPlayer.DidSpecialAction = true
	go func() {
		s.chanSheriffCheck <- nil
	}()
	return &proto.ActionRespond{Message: result}, nil
}

func (s *Server) ExposeMafia(ctx context.Context, action *proto.PlayerAction) (*proto.ActionRespond, error) {
	fromPlayer, targetPlayer, err := s.getPlayers(action)
	if err != nil {
		return err, nil
	}
	if !fromPlayer.IsAlive ||
			fromPlayer.Role != ROLE_SHERIFF ||
			s.currentState != DAY ||
			!targetPlayer.IsAlive {
		return &proto.ActionRespond{Message: "Error: You can't do such action right now!"}, nil
	}

	s.BroadcastMessageStr(fmt.Sprintf("Sheriff says %v is a mafia!", targetPlayer.Name))
	return &proto.ActionRespond{Message: "You exposed the data!"}, nil
}

func (s *Server) VoteForExecution(ctx context.Context, action *proto.PlayerAction) (*proto.ActionRespond, error) {
	fromPlayer, targetPlayer, err := s.getPlayers(action)
	if err != nil {
		return err, nil
	}
	if !fromPlayer.IsAlive ||
			s.currentState != DAY ||
			!targetPlayer.IsAlive ||
			targetPlayer.Id == fromPlayer.Id ||
			fromPlayer.VotedThisRound {
		return &proto.ActionRespond{Message: "Error: You can't do such action right now!"}, nil
	}
	fromPlayer.VotedThisRound = true
	s.votes[targetPlayer.Id]++
	s.waitVote.Done()
	return &proto.ActionRespond{Message: fmt.Sprintf("You voted to execute %v", targetPlayer.Name)}, nil
}

func (s *Server) SkipVote(ctx context.Context, action *proto.PlayerAction) (*proto.ActionRespond, error) {
	fromPlayer := s.authKeyToPlayer[action.SenderAuthKey]
	if fromPlayer == nil {
		return &proto.ActionRespond{Message: "Error: No such player!"}, nil //errors.New("no such player")
	}
	if !fromPlayer.IsAlive ||
			s.currentState != DAY ||
			fromPlayer.VotedThisRound {
		return &proto.ActionRespond{Message: "Error: You can't do such action right now!"}, nil
	}
	fromPlayer.VotedThisRound = true
	s.waitVote.Done()
	return &proto.ActionRespond{Message: "You skipped vote"}, nil
}

func (s *Server) GetPlayers(ctx context.Context, empty *proto.Empty) (*proto.Message, error) {
	return &proto.Message{Text: fmt.Sprintf("List of players:\n%v", s.getPlayersList())}, nil
}

func (s *Server) getPlayersList() string {
	result := ""
	for _, p := range s.Players {
		result += p.String() + "\n"
	}
	return result
}

func (s *Server) CreateStream(pconn *proto.Connect, stream proto.Mafia_CreateStreamServer) error {

	conn := &Connection{
		stream: stream,
		active: true,
		error:  make(chan error),
	}

	player := &Player{
		Id:      s.idCounter,
		Name:    pconn.Player.Name,
		Role:    ROLE_NOT_ASSIGNED,
		IsAlive: false,
		Conn:    conn,
	}

	if s.idCounter > PLAYERS_NUM {
		sendMessageStr("Sorry! The game is full!", conn)
	}

	s.Players = append(s.Players, player)
	s.authKeyToPlayer[pconn.Player.AuthKey] = player
	s.idToPlayer[player.Id] = player
	sendMessageStr(fmt.Sprintf("You have been connected!\nPlayers:\n%v", s.getPlayersList()), conn)

	s.BroadcastMessageStr(fmt.Sprintf("Player %v (id: %v) has connected!", player.Name, player.Id))
	glog.Infof("Player %v (id: %v) has connected!", player.Name, player.Id)

	if s.idCounter == PLAYERS_NUM {
		go s.StartGame()
	}

	s.idCounter++
	return <-conn.error
}

func (s *Server) BroadcastMessage(msg *proto.Message) error {
	wait := sync.WaitGroup{}
	done := make(chan int)

	for _, player := range s.Players {
		conn := player.Conn
		wait.Add(1)

		go func() {
			defer wait.Done()
			sendMessage(msg, conn)
		}()
	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done

	//&proto.Close{},
	return nil
}

func (s *Server) BroadcastMessageStr(msg string) {
	s.BroadcastMessage(&proto.Message{Text: msg})
}

func sendMessage(msg *proto.Message, conn *Connection) {
	if conn.active {
		err := conn.stream.Send(msg)
		//glog.Info("Sending message to: ", conn.stream)

		if err != nil {
			glog.Errorf("Error with Stream: %v - Error: %v", conn.stream, err)
			conn.active = false
			conn.error <- err
		}
	}
}

func sendMessageStr(msg string, conn *Connection) {
	sendMessage(&proto.Message{Text: msg}, conn)
}

func AssignRoles(s *Server) {

	for i := 0; i < MAFIA_NUM; i++ {
		s.Players[rand.Intn(PLAYERS_NUM)].Role = ROLE_MAFIA
	}
	for i := 0; i < SHERIFF_NUM; i++ {
		p := s.Players[rand.Intn(PLAYERS_NUM)]
		if p.Role == 0 {
			p.Role = ROLE_SHERIFF
		} else {
			i--
		}
	}

	for _, player := range s.Players {
		if player.Role == 0 {
			player.Role = ROLE_CITIZEN
		}
	}
}

func (s *Server) FindByRole(role int, onlyAlive bool) []*Player {
	players := make([]*Player, 0)
	for _, player := range s.Players {
		if player.Role == role && (!onlyAlive || player.IsAlive) {
			players = append(players, player)
		}
	}
	return players
}

func (s *Server) FindAlive() []*Player {
	players := make([]*Player, 0)
	for _, player := range s.Players {
		if player.IsAlive {
			players = append(players, player)
		}
	}
	return players
}

func (s *Server) StartGame() {
	glog.Info("Game started!")
	s.BroadcastMessageStr("Game started!")
	for _, player := range s.Players {
		player.Role = ROLE_NOT_ASSIGNED
	}
	AssignRoles(s)
	for _, player := range s.Players {
		player.IsAlive = true
		msg := getRoleDescription(player.Role)
		sendMessageStr(msg, player.Conn)
	}
	// Main game loop
	for {
		for _, player := range s.Players {
			player.DidSpecialAction = false
		}
		// Night
		s.currentState = NIGHT
		s.BroadcastMessageStr("It is NIGHT now")
		s.BroadcastMessageStr("Now Mafia should kill and Sheriff should check")
		// Wait for check and kill
		killed := <-s.chanMafiaKill
		if len(s.FindByRole(ROLE_SHERIFF, true)) > 0 {
			<-s.chanSheriffCheck
		}

		// Day
		s.currentState = DAY
		s.BroadcastMessageStr("It is DAY now")
		// Check victory
		s.BroadcastMessageStr(fmt.Sprintf("Player %v was killed!", killed.Name))
		if s.checkVictory() {
			break
		}
		// Vote
		// Set votes to zero
		for _, p := range s.Players {
			p.VotedThisRound = false
		}
		s.votes = make(map[int]int)

		s.waitVote.Add(len(s.FindAlive()))
		s.BroadcastMessageStr("Vote for execution now! \n Send \"vote <player_id>\" \n If you want to skip vote send \"skip\"")
		s.waitVote.Wait()

		execId, allEqual :=  max(s.votes)
		if !allEqual {
			execPlayer := s.idToPlayer[execId]
			execPlayer.Die()
			s.BroadcastMessageStr(fmt.Sprintf("%v is executed!", execPlayer.Name))
			if s.checkVictory() {
				break
			}
		}  else {
			s.BroadcastMessageStr("Votes are equal!")
		}
	}
	s.BroadcastMessageStr("You can now start a new game with \"newGame\" command")
	s.currentState = NOT_STARTED
}

func (s *Server) checkVictory() bool {
	maf := s.FindByRole(ROLE_MAFIA, true)
	mafCount := len(maf)
	otherCount := len(s.FindByRole(ROLE_CITIZEN, true)) + len(s.FindByRole(ROLE_SHERIFF, true))
	if mafCount == 0 {
		// We have only one mafia so it's ok
		s.BroadcastMessageStr(fmt.Sprintf("Citizens have won! %v was mafia!", s.FindByRole(ROLE_MAFIA, false)[0].Name))
		return true
	} else if mafCount >= otherCount {
		s.BroadcastMessageStr(fmt.Sprintf("Mafia have won! %v was mafia!", maf[0].Name))
		return true
	}
	return false
}

func max(numbers map[int]int) (key int, allMaxEqual bool) {
	var maxNumber int
	for key, maxNumber = range numbers {
		break
	}

	for k, v := range numbers {
		if v > maxNumber {
			maxNumber = v
			key = k
		}
	}
	count := 0
	for _, v := range numbers {
		if v == maxNumber {
			count++
		}
	}
	// If there is several max numbers
	allMaxEqual = count > 1

	return key, allMaxEqual
}

func main() {
	//var connections []*Connection
	var players []*Player

	server := &Server{players,
		1,
		NOT_STARTED,
		make(chan interface{}),
		make(chan *Player),
		make(map[string]*Player),
		make(map[int]*Player),
		make(map[int]int),
		&sync.WaitGroup{}}

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		glog.Fatalf("error creating the server %v", err)
	}

	glog.Info("Starting server at port :8080")

	// Seed for random actions
	rand.Seed(time.Now().UnixNano())
	proto.RegisterMafiaServer(grpcServer, server)
	grpcServer.Serve(listener)
}
