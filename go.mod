module github.com/iucario/bangumi-go

go 1.22

replace github.com/iucario/bangumi-go/cmd => ./cmd

replace github.com/iucario/bangumi-go/cmd/auth => ./cmd/auth

require github.com/spf13/cobra v1.8.0

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)
