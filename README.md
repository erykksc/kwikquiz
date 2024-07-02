# kwikquiz
Quiz-platform written in Golang and HTMX


## Run using Docker Compose
Requirements: Docker and Docker Compose installed

```bash
docker-compose up
```
_The app will be available at `http://localhost:3000`_

## How to run locally
Requirements: Go 1.22 installed

```bash
go run cmd/main.go
```
_The app will be available at `http://localhost:3000`_

## Running during development
Requirements: Go 1.22, air installed

Air is a tool that auto rebuilds and runs the application when changes are detected.
DEBUG=1 flag is used to set log level to debug.
```bash
DEBUG=1 air
```

## Running in production mode
To run in production mode, set the environment variable `PROD` to `true` or `1`.
```bash
PROD=1 go run cmd/main.go
```

### SECURITY WARNING
__If you plan to host the app on the web, you should change .env file so that the password to your database is not known to the public.__

## Contributing
Please read the [CONTRIBUTING.md](CONTRIBUTING.md) file for more information on how to contribute to this project.
