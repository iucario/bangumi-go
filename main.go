package main

import (
	"github.com/iucario/bangumi-go/cmd"
	_ "github.com/iucario/bangumi-go/cmd/auth"
)

func main() {
	cmd.Execute()
}
