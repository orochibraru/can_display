package ui

import (
	"time"

	"orochibraru/can_display/internal/display"
	"orochibraru/can_display/internal/signals"
)

// DefaultStaleAfter is how long a signal can go unrefreshed before the
// dashboard shows "--" instead of a frozen last value -- long enough to
// ride out normal bus jitter, short enough to notice a disconnected
// sensor or a car that's been switched off within a couple of seconds.
const DefaultStaleAfter = 2 * time.Second

// DashTile pairs a Tile's look with the accessor that pulls its current
// reading out of a Snapshot.
type DashTile struct {
	Tile
	Value func(signals.Snapshot) signals.Reading
}

// DefaultTiles is the stock GT86 layout. Coolant and oil temp have real
// decoders today (internal/canbus/gt86); battery/AFR/ethanol are wired
// up and waiting on a data source -- see that package's doc comment.
func DefaultTiles() []DashTile {
	return []DashTile{
		{
			Tile: Tile{
				Label: "COOLANT", Unit: "°C", Format: "%.0f", StaleAfter: DefaultStaleAfter,
				Range: Range{Min: 40, Max: 130, WarnLow: 60, WarnHigh: 105, DangerLow: 50, DangerHigh: 115},
			},
			Value: func(s signals.Snapshot) signals.Reading { return s.CoolantTempC },
		},
		{
			Tile: Tile{
				Label: "OIL TEMP", Unit: "°C", Format: "%.0f", StaleAfter: DefaultStaleAfter,
				Range: Range{Min: 40, Max: 150, WarnLow: 60, WarnHigh: 120, DangerLow: 50, DangerHigh: 135},
			},
			Value: func(s signals.Snapshot) signals.Reading { return s.OilTempC },
		},
		{
			Tile: Tile{
				Label: "BATTERY", Unit: "V", Format: "%.1f", StaleAfter: DefaultStaleAfter,
				Range: Range{Min: 10, Max: 16, WarnLow: 12, WarnHigh: 15, DangerLow: 11.5, DangerHigh: 15.5},
			},
			Value: func(s signals.Snapshot) signals.Reading { return s.BatteryVoltage },
		},
		{
			Tile: Tile{
				Label: "AFR", Unit: "", Format: "%.1f", StaleAfter: DefaultStaleAfter,
				Range: Range{Min: 10, Max: 18, WarnLow: 12, WarnHigh: 16, DangerLow: 11, DangerHigh: 17},
			},
			Value: func(s signals.Snapshot) signals.Reading { return s.AFR },
		},
		{
			Tile: Tile{
				Label: "ETHANOL", Unit: "%", Format: "%.0f", StaleAfter: DefaultStaleAfter,
				Range: Range{Min: 0, Max: 100, WarnLow: -1, WarnHigh: 101, DangerLow: -1, DangerHigh: 101},
			},
			Value: func(s signals.Snapshot) signals.Reading { return s.EthanolPercent },
		},
	}
}

// Dashboard is the top-level screen: a grid of tiles, one per signal,
// laid out to fill whatever panel it's given.
type Dashboard struct {
	Theme Theme
	Tiles []DashTile
}

// gridGap is the pixel gap left between tiles (and around the screen
// edge) so adjacent tiles don't visually merge into one another.
const gridGap = 3

// Render lays the tiles out in a grid sized to d and draws every one
// against snap, then flushes the frame.
func (dash Dashboard) Render(d display.Display, snap signals.Snapshot) error {
	w, h := d.Size()
	fillRect(d, Rect{X: 0, Y: 0, W: w, H: h}, dash.Theme.Background)

	const cols = 2
	rows := (int16(len(dash.Tiles)) + cols - 1) / cols

	cellW := w / cols
	cellH := h / rows

	for i, t := range dash.Tiles {
		col := int16(i) % cols
		row := int16(i) / cols
		rect := Rect{
			X: col*cellW + gridGap,
			Y: row*cellH + gridGap,
			W: cellW - 2*gridGap,
			H: cellH - 2*gridGap,
		}
		t.Draw(d, dash.Theme, rect, t.Value(snap), snap.At)
	}
	return d.Display()
}
