build:
	go build -o bin/ddexport cmd/ddexport/main.go

fmt:
	go fmt ./...

test:
	go test -v ./...
