build:
	go build -v ./...

tests:
	go test -v -cover .
format:
	go fmt .