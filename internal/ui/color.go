package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// "name": "synthwave-everything",
var color = map[string]int32{
	"black":               0xfefefe,
	"red":                 0xf97e72,
	"green":               0x72f1b8,
	"yellow":              0xfede5d,
	"blue":                0x6d77b3,
	"purple":              0xc792ea,
	"cyan":                0xf772e0,
	"white":               0xfefefe,
	"brightBlack":         0xfefefe,
	"brightRed":           0xf88414,
	"brightGreen":         0x72f1b8,
	"brightYellow":        0xfff951,
	"brightBlue":          0x36f9f6,
	"brightPurple":        0xe1acff,
	"brightCyan":          0xf92aad,
	"brightWhite":         0xfefefe,
	"background":          0x2a2139,
	"foreground":          0xf0eff1,
	"cursorColor":         0x72f1b8,
	"selectionBackground": 0x181521,
}

var Styles = tview.Theme{
	PrimitiveBackgroundColor:    tcell.NewHexColor(color["background"]),
	ContrastBackgroundColor:     tcell.NewHexColor(color["cyan"]),
	MoreContrastBackgroundColor: tcell.NewHexColor(color["background"]),
	BorderColor:                 tcell.NewHexColor(color["white"]),
	TitleColor:                  tcell.NewHexColor(color["white"]),
	GraphicsColor:               tcell.NewHexColor(color["red"]),
	PrimaryTextColor:            tcell.NewHexColor(color["white"]),
	SecondaryTextColor:          tcell.NewHexColor(color["yellow"]),
	TertiaryTextColor:           tcell.NewHexColor(color["green"]),
	InverseTextColor:            tcell.NewHexColor(color["blue"]),
	ContrastSecondaryTextColor:  tcell.NewHexColor(color["brightGreen"]),
}
