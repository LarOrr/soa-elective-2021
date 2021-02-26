package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"strconv"
	"strings"
	"time"

	//"flag"
	"fmt"
	"mafia_grpc/proto"
	"os"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"sync"
)

var authKey string
var client proto.MafiaClient
var wait *sync.WaitGroup

func init() {
	wait = &sync.WaitGroup{}
}

func connect(user *proto.ConnectInfo) error {
	var streamerror error

	stream, err := client.CreateStream(context.Background(), &proto.Connect{
		Player: user,
		Active: true,
	})

	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}

	wait.Add(1)
	go func(str proto.Mafia_CreateStreamClient) {
		defer wait.Done()

		for {
			msg, err := str.Recv()
			if err != nil {
				streamerror = fmt.Errorf("Error reading message: %v", err)
				break
			}

			printInfo(msg.Text)

		}
	}(stream)

	return streamerror
}

func printInfo(text string) {
	fmt.Printf("%s > %s\n", time.Now().Format("15:04:05"), text)
}

func main() {
	//timestamp := time.Now()
	done := make(chan int)

	// Flags for connection
	addr_p := flag.String("a", "127.0.0.1:8080", "The address of the server")
	name_p := flag.String("n", "Anon", "The name of the user")
	flag.Parse()
	addr := *addr_p
	name := *name_p
	//id := sha256.Sum256([]byte(timestamp.String() + name))

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Couldnt connect to service: %v", err)
	}

	timestamp := time.Now()
	auth := sha256.Sum256([]byte(timestamp.String() + name))
	authKey = hex.EncodeToString(auth[:])
	client = proto.NewMafiaClient(conn)
	user := &proto.ConnectInfo{
		Name:    name,
		AuthKey: authKey,
	}

	connect(user)

	wait.Add(1)
	go func() {
		defer wait.Done()

		mapActions := make(map[string]func(ctx context.Context, in *proto.PlayerAction, opts ...grpc.CallOption) (*proto.ActionRespond, error))
		mapActions["kill"] = client.Kill
		mapActions["check"] = client.CheckIfMafia
		mapActions["expose"] = client.ExposeMafia
		mapActions["vote"] = client.VoteForExecution

		fmt.Printf("Use 'list' to get list of players\n")
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input := strings.Split(scanner.Text(), " ")

			//_, err := client.BroadcastMessage(context.Background(), msg)
			//rpc CreateStream(Connect) returns (stream Message);
			////  rpc BroadcastMessage(Message) returns (Close);
			//rpc Kill(PlayerAction) returns (ActionRespond);
			//rpc CheckIfMafia(PlayerAction) returns (ActionRespond);
			//rpc ExposeMafia(PlayerAction) returns (ActionRespond);
			//rpc VoteForExecution(PlayerAction) returns (ActionRespond);
			//rpc SkipVote(PlayerAction) returns (ActionRespond);
			//rpc GetPlayers(Empty) returns (Message);

			if input[0] == "list" {
				msg, err := client.GetPlayers(context.Background(), &proto.Empty{})
				if err != nil {
					fmt.Printf("Error Sending Message: %v", err)
					break
				}
				printInfo(msg.Text)

			} else if input[0] == "kill" || input[0] == "check" || input[0] == "expose" || input[0] == "vote" {
				player_id, err := strconv.Atoi(input[1])
				if err != nil {
					printInfo("You should add <player_id>!")
				} else {
					resp, err := mapActions[input[0]](context.Background(), &proto.PlayerAction{SenderAuthKey: authKey,
						TargetId: int32(player_id)})
					if err != nil {
						printInfo(fmt.Sprintf("Error Sending Message: %v", err))
						break
					}
					printInfo(resp.Message)
				}
			} else if input[0] == "skip" {
				resp, err := client.SkipVote(context.Background(),
					&proto.PlayerAction{SenderAuthKey: authKey,
						TargetId: 0})
				if err != nil {
					fmt.Printf("Error Sending Message: %v", err)
					break
				}
				fmt.Printf(resp.Message + "\n")
			} else if input[0] == "newGame" {
				_, err := client.StartNewGame(context.Background(), &proto.Empty{})
				if err != nil {
					fmt.Printf("Error Sending Message: %v", err)
					break
				}
			} else {
				fmt.Printf("Wrong command!")
			}
		}
	}()

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
}
