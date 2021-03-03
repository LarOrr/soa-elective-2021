# Python GraphQL server
Created with [strawberry](https://github.com/strawberry-graphql/strawberry)

Run locally:

`pip install strawberry-graphql[debug-server]` 

or `pip install --no-cache-dir -r requirements.txt`

`strawberry server graphql_server`

Run with docker:
 
 `docker run -p 8000:8000 larorr/soa-graphql-py`
 (Docker hub: https://hub.docker.com/repository/docker/larorr/soa-graphql-py)

## Example requests:
+ Get all unfinished games:
```graphql
{
  games(finished:true) {
    id
    finished
    scoreboard {
      mafiaScore
      citizenScore
    }
    comments {
      author
      content
    }
  }
}
```

Response:
```graphql
{
  "data": {
    "games": [
      {
        "id": "0",
        "finished": false,
        "scoreboard": {
          "mafiaScore": 1,
          "citizenScore": 2
        },
        "comments": [
          {
            "author": "user1",
            "content": "I hate this game!!!"
          }
        ]
      },
      {
        "id": "2",
        "finished": false,
        "scoreboard": {
          "mafiaScore": 0,
          "citizenScore": 1
        },
        "comments": []
      }
    ]
  }
}
```

+ Get game by id only with it's scoreboard:
```graphql
{
  game(id:1) {
    scoreboard {
      mafiaScore
      citizenScore
    }
  }
}
```
Response:
```graphql
{
  "data": {
    "game": {
      "scoreboard": {
        "mafiaScore": 10,
        "citizenScore": 0
      }
    }
  }
}
```

+ Add comment to the game with id:
```graphql
mutation {
   addComment(gameId:0,comment:{
    author:"newUser",
    content:"I'm much better than all of you! AND this comment is the best!",
  }) {
    id
    author
    content
  }
}
```
Response:
```graphql
{
  "data": {
    "addComment": {
      "id": "3",
      "author": "newUser",
      "content": "I'm much better than all of you! AND this comment is the best!"
    }
  }
}
```