default: test build

build:
	go build ./cmd/timerec
	go build ./cmd/timerec-server

test:
	go test ./... -cover
