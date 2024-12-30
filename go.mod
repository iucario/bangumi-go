module github.com/iucario/bangumi-go

go 1.22

replace github.com/iucario/bangumi-go/cmd => ./cmd

replace github.com/iucario/bangumi-go/cmd/auth => ./cmd/auth

replace github.com/iucario/bangumi-go/cmd/list => ./cmd/list

replace github.com/iucario/bangumi-go/cmd/subject => ./cmd/subject

replace github.com/iucario/bangumi-go/api => ./api

require (
	github.com/gdamore/tcell/v2 v2.7.4
	github.com/spf13/cobra v1.8.1
)

require (
	github.com/gdamore/encoding v1.0.1 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/term v0.27.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/rivo/tview v0.0.0-20241227133733-17b7edb88c57
	github.com/spf13/pflag v1.0.5 // indirect
)
