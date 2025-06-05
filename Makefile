lint:
	gofumpt -w . && golangci-lint run

run:
	go run .

ui:
	go run . ui


build:
	go build -o dist/bgm
