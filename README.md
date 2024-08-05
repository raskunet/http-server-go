### Description 
The aim of this project was to learn about the HTTP Protocol and how http server handle requests, send responses, respond to different routes, how http header works.
This play server supports concurrency, handling request to different routes/paths and serving files from a directory 

---
## Installation

1. clone the repo and cd into it
```sh
   git clone 
   cd http-server-go
```
2. Install the executable & run
```go
   go build app/server.go
   ./server
```
3. Running the server with `--directory` flag with path to directory allows the server to serve the files from that directory
```go
   ./server --directory /tmp/
```
## Working
To send requests to the server you can use curl. By default the server runs on port 4221
1. Sending requests to base url `/`
```sh
   curl -v http://localhost:4221/
```
2. Sending requests to `/echo/[str]` returns the `str` back
```sh
   curl -v http://localhost:4221/echo/
```
3. Sending request to `/user-agent` responds with the client `user-agent`
```sh
   curl -v https://localhost:4221/user-agent
```
4. *Sending `POST` request to `/files/[name of file]` with `data` creates a files containing that data
```sh
   curl -v -X POST \
   --data "request form curl: lorem ipsum"
   https://localhost/files/hello.txt
```
5. *Sending `GET` request to `files/[name of file]` responds with data in that file if it exists
```sh
   curl -v http://localhost:4221/files/hello.txt
```
*Only if the server was run with the `--directory` flag