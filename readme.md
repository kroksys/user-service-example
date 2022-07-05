# User service example
Example micro service for users.

## Prepare buf to compile proto files
1. Install buf: https://docs.buf.build/installation
2. Visual Studio Code: Settings.json:
Including following code will remove proto file import errors.
```json
"protoc": {
    "options": [
        "-I=~/.cache/buf/v1/module/data/buf.build"
    ]
}
```


## Usage with docker compose
```bash
make up 

# Or just use the command from makefile:
docker-compose up --remove-orphans -d
```

## Usage with local server
``` bash
# Comment user-service from docker-compose.yaml

# Exectue this command to start mysql and redis services
docker-compose up --remove-orphans -d

# Start local server with:
go run ./cmd/server
```

## Documentation
Documentation is pretty empty and could be improved a lot.
``` bash
go get -u github.com/go-swagger/go-swagger
swagger serve pkg/pb/v1/user.swagger.json
```

## Info
* I assume email should be unique field.
* I used grpc-grateway for protobuf, http-proxy and openapi documentation. It makes proto file to be a single source of truth. 
* Mysql is database I have been working a lot, but the fact that I'm using gorm library gives option to switch database pretty easy.
* I decided to use redis to notify service acout user data changes.
* I used "buf" to compile proto files instead of protoc. Buf makes easier to import external resources and use plugins.

## Improvements
* Hashing password.
* Documentation, descriptions and explinations can be improved a lot.
* Instead of redis we could use Event-driven-arhitecture and use services like Kafka or RabbotMQ.
* Testing could be improved with more precise unit tests.
* A config file could be used instead of environment variables.

## Overall
This project was made over few evenings/nights and that made me create some shortcuts. With more time, effort and discussions it could become better. I dont think it is cleanest solution for the job, but it works as an example.