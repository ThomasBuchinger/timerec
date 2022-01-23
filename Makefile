default: test build

build:
	go build ./cmd/timerec
	go build ./cmd/timerec-server

test:
	go test ./... -cover

dev: test build
	go run ./cmd/timerec-server