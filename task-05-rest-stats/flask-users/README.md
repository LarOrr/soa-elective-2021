# REST + RabbitMQ
REST api for user resource with JWT auth and RabbitMQ.

Application uses RabbitMQ to generate PDF files. 
1) First application gets HTTP request for PDF generation from POST /users/{user_id}/pdf (which returns future URI of the generated pdf).
2) Server sends request for generation to RabbitMQ
3) Process pdf_generator.pdf listens to RabbitMQ and generates file from request.
4) File can be accessed from URI


Application uses JWT authorisation for all actions except [GET /users] , [GET /users/{user_id}] and [POST /users].

Tokens are given in [POST /users] an [POST /login]

SQLLite is used as database, it's accessed with SQLAlchemy.

# Run
1) `docker run -d --hostname my-rabbit -p 5672:5672 --name some-rabbit rabbitmq:3`

2) Local:
 `python -m flask run`
 
3) OR with docker (https://hub.docker.com/r/larorr/soa-rest): 
`docker run -p 5000:5000 larorr/soa-rest`

Note: When building image change config first (address + mq_address)

# Resources

###  Users [/users]
Possible operations:
+ GET /users - get all users
+ GET /users/{user_id} - get user by id
+ POST /users - create new user => returns created user and auth JWT token
+ PUT or PATCH /users/{user_id} - update user (updates only fields from the request)
+ DELETE users/{user_id}

Examples: 
```
Request: GET users
Response 200:
[
    {
        "email": "user1@ggmail.com",
        "gender": "male",
        "profile_pic": null,
        "self": "/users/1",
        "stats": "/users/1/stats",
        "username": "user1"
    },
    {
        "email": "user2@yahooo.org",
        "gender": null,
        "profile_pic": null,
        "self": "/users/2",
        "stats": "/users/2/stats",
        "username": "user2"
    }
]
```

###  Users stats [/users/{user_id}/stats]

+ GET /users/{user_id}/stats 
+ PATCH or PUT /users/{user_id}/stats - update stats

Examples: 
```
Request: GET /users/2/stats
Response 200:
{
    "loss_count": 4,
    "session_count": 12,
    "total_time": 5000,
    "victory_count": 8
}
```

###  User PDF [/users/{user_id}/pdf]

+ GET /users/{user_id}/pdf - get previously generated pdf
+ POST /users/{user_id}/pdf - request PDF generation. Returns URI where pdf will be located.

Examples: 
```
Request: POST /users/2/pdf
Response 200:
{
    "url":"/users/2/pdf"
}
```

### Login [POST /login]
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
