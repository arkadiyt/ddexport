build:
	go build -o bin/ddexport cmd/main.go

fmt:
	go fmt ./...

test:
	go test -v ./...
