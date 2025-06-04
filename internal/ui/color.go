package ui

import (
	"fmt"

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
	ContrastBackgroundColor:     tcell.NewHexColor(color["blue"]),
	MoreContrastBackgroundColor: tcell.NewHexColor(color["purple"]),
	BorderColor:                 tcell.NewHexColor(color["white"]),
	TitleColor:                  tcell.NewHexColor(color["white"]),
	GraphicsColor:               tcell.NewHexColor(color["red"]),
	PrimaryTextColor:            tcell.NewHexColor(color["white"]),
	SecondaryTextColor:          tcell.NewHexColor(color["yellow"]),
	TertiaryTextColor:           tcell.NewHexColor(color["green"]),
	InverseTextColor:            tcell.NewHexColor(color["blue"]),
	ContrastSecondaryTextColor:  tcell.NewHexColor(color["white"]),
}

func Yellow(text string) string {
	return fmt.Sprintf("[#%06x]%s[-]", color["yellow"], text)
}

func Blue(text string) string {
	return fmt.Sprintf("[#%06x]%s[-]", color["blue"], text)
}

func Green(text string) string {
	return fmt.Sprintf("[#%06x]%s[-]", color["green"], text)
}

func Cyan(text string) string {
	return fmt.Sprintf("[#%06x]%s[-]", color["cyan"], text)
}

func Red(text string) string {
	return fmt.Sprintf("[#%06x]%s[-]", color["red"], text)
}

func Black(text string) string {
	return fmt.Sprintf("[#%06x]%s[-]", color["black"], text)
}

func Purple(text string) string {
	return fmt.Sprintf("[#%06x]%s[-]", color["purple"], text)
}

func White(text string) string {
	return fmt.Sprintf("[#%06x]%s[-]", color["white"], text)
}

func ColorToHex(color tcell.Color) string {
	r, g, b := color.RGB()
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

func TertiaryText(text string) string {
	return fmt.Sprintf("[%s]%s[-]", ColorToHex(Styles.TertiaryTextColor), text)
}

func SecondaryText(text string) string {
	return fmt.Sprintf("[%s]%s[-]", ColorToHex(Styles.SecondaryTextColor), text)
}

func GraphicsColor(text string) string {
	return fmt.Sprintf("[%s]%s[-]", ColorToHex(Styles.GraphicsColor), text)
}
