// cmd/server/main.go
package main

import "rock-paper-scissors/internal/server"

func main() {
	s := server.NewServer()
	s.Run()
}
