lint:
	gofumpt -w . && golangci-lint run

run:
	go run .