package main

import (
	"context"
	"wb-level0/connections"
)

func main() {
	conn := *connections.DataBase()
	cache := connections.CreateCache(&conn)
	sc := *connections.NatsStreaming(&conn, cache)
	connections.RunHttpServer(cache)

	defer sc.Close()
	defer conn.Close(context.Background())
}
