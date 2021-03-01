# REST + RabbitMQ

# Run
1. `docker run -d --hostname my-rabbit -p 5672:5672 --name some-rabbit rabbitmq:3`
2. `python -m flask run`

# Contents

### <a name="#login"></a> Login [POST /login]
+ Request have to contain JSON with this fields: username, password. Otherwise it will return status code **400** and JSON with {'success': false, 'message': 'Bad username or password'}
+ If pair of password and username doesn't exist in the system - it will return status code **401** and JSON with {'success': false, 'message': 'Bad username or password'}.

+ Example: 
```
Request: POST login
Response 200:
{ 
	"succsess: true",
	"token": "secret_token"
	"user" : "users/<id>"
}
```
