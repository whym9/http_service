package main

import (
	"flag"

	server "http_service/internal/http_server"
)

func main() {
	addr := *flag.String("address", "localhost:8080", "http server address")
	saddr := *flag.String("sender_address", ":5005", "drpc serder address")

	server.HandleHTTP(addr, saddr)
}
