lint:
	gofumpt -w . && golangci-lint run

run:
	BGM_ENV=dev go run . 2>app.log

ui:
	BGM_ENV=dev go run . ui 2>app.log

build:
	go build -o dist/bgm
