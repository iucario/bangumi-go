package main

import (
	"github.com/iucario/bangumi-go/cmd"
	_ "github.com/iucario/bangumi-go/cmd/auth"
	_ "github.com/iucario/bangumi-go/cmd/list"
	_ "github.com/iucario/bangumi-go/cmd/ui"
)

func main() {
	cmd.Execute()
}
