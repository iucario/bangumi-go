lint:
	gofumpt -w . && golangci-lint run

run:
	BGM_ENV=dev go run .

ui:
	BGM_ENV=dev go run . ui


build:
	go build -o dist/bgm
