# Simple console client for GraphQL 
Using https://github.com/graphql-python/gql

The GraphQL server is supposed to run on localhost:8000

Run:

`pip install --pre gql[all]`

or `pip install --no-cache-dir -r requirements.txt`

`python ./client.py -n <your name>`

### Commands
+ Enter 'exit' to exit the program
+ Enter 'games <optional:finished (bool)>' to get list of games. Note finished must be 'true' or 'false'.
+ Enter 'game <game_id (int)>' to get game
+ Enter 'game <game_id (int)> score' to get score of the game
+ Enter 'addComment <game_id (int)> <comment (str)>' to comment the game