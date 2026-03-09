# binary-scanner

A simple binary scanner composed of a client and server, written in Go. 

## Server
1. Accepts file metadata via:
   POST /files
2. Stores the data persistently using sqlite
3. Endpoint: GET /files?limit=20

### Setup and usage
From project directory, either: 
- Run the server directly:
  `go run server/main.go`
- Or you can run it directly from the go file
- Data can be viewed by accessing:
  `curl http://localhost:8080/files?limit=20`

## Client
1. Accepts a directory path as input, for example:
   binary-scan ./my-folder
2. Recursively scans the directory.
3. For each file, sends metadata to the server:
   • file path
   • file size
   • last modified time only if the file is a binary, executable file
4. Sends the data over HTTP.
5. Handles temporary network failures 
6. Feature simple retry logic

## Setup and Usage
From project directory, either:
- Run the client directly:
  `go run client/main.go /path/to/scan`
- Build the client binary and run it:
  `go build -o binary-scan client/main.go`
- Then execute the binary with the directory path:
  `./binary-scan /path/to/scan`
   

