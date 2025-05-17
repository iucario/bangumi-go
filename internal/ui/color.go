package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var Styles = tview.Theme{
	PrimitiveBackgroundColor:    tcell.NewHexColor(0x303446),
	ContrastBackgroundColor:     tcell.ColorBlue,
	MoreContrastBackgroundColor: tcell.ColorGreen,
	BorderColor:                 tcell.ColorWhite,
	TitleColor:                  tcell.ColorWhite,
	GraphicsColor:               tcell.ColorWhite,
	PrimaryTextColor:            tcell.ColorWhite,
	SecondaryTextColor:          tcell.ColorYellow,
	TertiaryTextColor:           tcell.ColorGreen,
	InverseTextColor:            tcell.ColorBlue,
	ContrastSecondaryTextColor:  tcell.ColorNavy,
}
