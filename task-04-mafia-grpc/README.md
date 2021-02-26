# Simple Mafia gRPC
Simple Mafia game (with 2 Citizens, 1 Mafia, 1 Sheriff) written in GoLang

Communication between clients and the server based on gRPC technology.

When 4 players connect to the server, the game starts automatically.

Note: there is now "endDay" call because I use vote and skip instead of it

Run server with Docker:

https://hub.docker.com/repository/docker/larorr/mafia-grpc

=> `docker pull larorr/mafia-grpc`

=> `docker run -p 8080:8080 larorr/mafia-grpc`

Run client:

=> `go mod download`

=> `go build client.go -n=<your_name> -a=<server_address>`

Note: "127.0.0.1:8080" is default address

### Commands from the client

All roles
* list - get list of players (Any time)
* vote <player_id> - vote for player execution (DAY)
* skip - skip voting (DAY)
* newGame - start new game (AFTER LAST GAME IS OVER)

Mafia:
* kill <player_id> - kill player (NIGHT), note: can't kill himself

Sheriff:
* check <player_id> - check if player is the mafia (NIGHT)
* expose - publish data on the mafia (DAY)