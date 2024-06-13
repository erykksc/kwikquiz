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
DEBUG=1 flag is used to enable debug mode. Application will add:
* Set log level to debug
* Add example quizzes to the database
* Add example lobby with pin: 1234

```bash
DEBUG=1 air
```

## Contributing
Please read the [CONTRIBUTING.md](CONTRIBUTING.md) file for more information on how to contribute to this project.
