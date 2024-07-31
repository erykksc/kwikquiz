# kwikquiz
Quiz-platform written in Golang and HTMX


## Run using Docker
Requirements: Docker installed

```bash
docker build -t kwikquiz .
docker run -t kwikquiz
```
_The app will be available at `http://localhost:3000`_

## How to run locally
Requirements: Go 1.22.5 installed

```bash
go run kwikquiz.go -prod
```
_The app will be available at `http://localhost:3000`_

## Running during development
Requirements: Go 1.22.5, air installed

Air is a tool that auto rebuilds and runs the application when changes are detected.
```bash
air
```

## CLI interface
For available flags and arguments run the command
```bash
go run kwikquiz.go -help
```

## Contributing
Please read the [CONTRIBUTING.md](CONTRIBUTING.md) file for more information on how to contribute to this project.
