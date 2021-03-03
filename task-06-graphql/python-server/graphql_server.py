import typing
from typing import List

import strawberry
from strawberry.scalars import ID


# ID (str) in types: https://stackoverflow.com/questions/47874344/should-i-handle-a-graphql-id-as-a-string-on-the-client

@strawberry.type
class Scoreboard:
    id: ID
    mafiaScore: int
    citizenScore: int


@strawberry.type
class Comment:
    id: ID
    author: str
    content: str


@strawberry.input
class CommentInput:
    author: str
    content: str


@strawberry.type
class Game:
    id: ID
    finished: bool
    scoreboard: Scoreboard
    comments: List[Comment]


gameDict: typing.Dict[int or ID, Game] = {
    0: Game(id=0, finished=False, scoreboard=Scoreboard(id=2, mafiaScore=1, citizenScore=2),
            comments=[Comment(id=0, author='user1', content='I hate this game!!!')]),
    1: Game(id=1, finished=True, scoreboard=Scoreboard(id=5, mafiaScore=10, citizenScore=0),
            comments=[Comment(id=1, author='proPlayer', content='gg ez'),
                      Comment(id=2, author='[TEAM] team_player', content='>:(')]),
    2: Game(id=2, finished=False, scoreboard=Scoreboard(id=6, mafiaScore=0, citizenScore=1),
            comments=[]),
}


@strawberry.type
class Query:
    @strawberry.field
    def games(self, finished: typing.Optional[bool] = None) -> List[Game]:
        if finished is not None:
            return list(filter(lambda game: game.finished == finished, gameDict.values()))
        else:
            return list(gameDict.values())

    @strawberry.field
    def game(self, id: int) -> Game:
        # id = int(ID)
        return gameDict[id]


@strawberry.type
class Mutation:
    @strawberry.mutation
    def addComment(self, gameId: int, comment: CommentInput) -> Comment:
        # Max existing id
        max_id = max([comment.id for g in gameDict.values() for comment in g.comments])
        newComment = Comment(id=int(max_id) + 1, author=comment.author, content=comment.content)
        gameDict[gameId].comments.append(newComment)
        return newComment


schema = strawberry.Schema(query=Query, mutation=Mutation)
f = open("schema.graphql", "w+")
f.write(schema.as_str())
f.close()
print(schema)
