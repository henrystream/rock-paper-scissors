run:
	go run ./cmd/server .
tests:
	export CGO_ENABLED=1

	go test -v -cover ./internal/game
	go test -v -cover ./internal/server
	#go test -v -cover -race ./internal/game
	#go test -v -cover -race ./internal/server