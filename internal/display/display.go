// Package display defines the drawing surface the UI renders onto.
package display

import "tinygo.org/x/drivers"

// Display is the drawing surface every screen renders onto: a real SPI
// panel on the Pico W, or a software framebuffer in the simulator.
//
// It's a direct alias for drivers.Displayer, so any TinyGo display
// driver (ST7789, ILI9341, ...) satisfies it with zero glue code, and
// tinyfont's drawing functions accept it as-is.
type Display = drivers.Displayer
