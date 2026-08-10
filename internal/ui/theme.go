package ui

import "image/color"

// Theme is a small, swappable palette so colors aren't hardcoded across
// every widget.
type Theme struct {
	Background color.RGBA
	Label      color.RGBA
	Good       color.RGBA
	Warning    color.RGBA
	Danger     color.RGBA
	Stale      color.RGBA
}

// Dark is the default theme: a dark background for readability in a
// sunlit car, with warning colors reserved for values actually out of
// range rather than used decoratively.
var Dark = Theme{
	Background: color.RGBA{R: 0x0b, G: 0x0e, B: 0x12, A: 0xff},
	Label:      color.RGBA{R: 0x8a, G: 0x94, B: 0xa3, A: 0xff},
	Good:       color.RGBA{R: 0x34, G: 0xd3, B: 0x99, A: 0xff},
	Warning:    color.RGBA{R: 0xf5, G: 0xa6, B: 0x23, A: 0xff},
	Danger:     color.RGBA{R: 0xe5, G: 0x4b, B: 0x4b, A: 0xff},
	Stale:      color.RGBA{R: 0x3a, G: 0x40, B: 0x4a, A: 0xff},
}
