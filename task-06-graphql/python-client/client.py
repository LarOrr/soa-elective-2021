from gql import gql, Client
from gql.transport.aiohttp import AIOHTTPTransport
import argparse

# Select your transport with a defined url endpoint
from gql.transport.exceptions import TransportQueryError

transport = AIOHTTPTransport(url="http://localhost:8000")

# Create a GraphQL client using the defined transport
client = Client(transport=transport, fetch_schema_from_transport=True)


def games_request(finished: str) -> str:
    inp = ""
    if finished:
        inp = f"(finished:{finished})"
    return f"""
                   {{
                     games{inp} {{
                       id
                       finished
                       scoreboard {{
                         mafiaScore
                         citizenScore
                       }}
                       comments {{
                         author
                         content
                       }}
                     }}
                   }}
                   """


def game_request(id: str) -> str:
    return """
                     {{
                      game(id:{}) {{
                        comments {{
                         author
                         content
                       }}
                        
                        scoreboard {{
                          mafiaScore
                          citizenScore
                        }}
                        }}
                      }}
                   """.format(id)


def game_scoreboard(id: str) -> str:
    return """
                     {{
                      game(id:{}) {{
                        scoreboard {{
                          mafiaScore
                          citizenScore
                        }}
                        }}
                      }}
                   """.format(id)


def add_comment(game_id: str, author: str, content: str) -> str:
    return f"""
            mutation {{
                       addComment(gameId:{game_id},comment:{{
                        author:"{author}",
                        content:"{content}",
                      }}) {{
                        id
                        author
                        content
                      }}
                    }}
                   """


parser = argparse.ArgumentParser(description='Process some integers.')
parser.add_argument('-name', action='store', type=str, required=False,
                    help='Your name', default='Anon')
args = parser.parse_args()
name = args.name
print(f"Hello, {name}!")

print("---------------------------------")
print("Enter 'exit' to exit the program")
print("Enter 'games <optional:finished (bool)>' to get list of games")
print("Enter 'game <game_id (int)>' to get game")
print("Enter 'game <game_id (int)> score' to get score of the game")
print("Enter 'addComment <game_id (int)> <comment (str)>' to comment the game")
print("---------------------------------")
while 1:
    # pass
    command = input()
    try:
        parts = command.split(' ')
        if parts[0] == 'exit':
            print("Bye!")
            break
        elif parts[0] == "games":
            if len(parts) > 1:
                finished = parts[1]
            else:
                finished = None
            query = gql(games_request(finished))
        elif parts[0] == "game":
            id = parts[1]
            if len(parts) < 3:
                query = gql(game_request(id))
            else:
                assert parts[2] == 'score'
                query = gql(game_scoreboard(id))
        elif parts[0] == "addComment":
            id = parts[1]
            comment = ' '.join(parts[2:])
            query = gql(add_comment(id, name, comment))
        else:
            assert False

        result = client.execute(query)
        print(result)
    except IndexError or AssertionError:
        print("Command '{}' is incorrect!".format(command))
    except TransportQueryError:
        print(f"Something wrong with your request or such object doesn't exist")

# Execute the query on the transport
# result = client.execute(query)
# print(result)
