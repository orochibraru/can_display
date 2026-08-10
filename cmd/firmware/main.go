//go:build tinygo

// Command firmware is the ESP32 program: it reads the GT86's CAN bus
// through an MCP2515 SPI controller and renders internal/ui's dashboard
// to an ST7789 SPI display. It shares every package under internal/
// with cmd/simulator -- only this file and the pin wiring are specific
// to real hardware.
//
// Build with TinyGo, not `go build` (this file needs the `machine`
// package, which only exists under the TinyGo compiler). See the
// Makefile for the exact command.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/mcp2515"
	"tinygo.org/x/drivers/st7789"

	"orochibraru/can_display/internal/canbus"
	"orochibraru/can_display/internal/canbus/gt86"
	"orochibraru/can_display/internal/signals"
	"orochibraru/can_display/internal/ui"
)

// Pin wiring -- adjust these to match your actual breadboard/PCB layout;
// see /docs/wiring.md for the reference wiring this project assumes.
//
// Two separate SPI buses are used (SPI2 for the display, SPI3 for the
// CAN controller) so a display refresh never blocks a CAN read, and
// vice versa.
const (
	displaySCK   = machine.GPIO18
	displaySDO   = machine.GPIO23 // MOSI; display is write-only, no MISO needed
	displayReset = machine.GPIO4
	displayDC    = machine.GPIO2
	displayCS    = machine.GPIO5
	displayBL    = machine.GPIO15

	canSCK = machine.GPIO14
	canSDO = machine.GPIO13 // MOSI
	canSDI = machine.GPIO12 // MISO
	canCS  = machine.GPIO27
)

// panelWidth/panelHeight must match cmd/simulator's constants -- both
// feed the same internal/ui.Dashboard, which lays itself out to
// whatever size it's given.
const (
	panelWidth  = 240
	panelHeight = 320
)

// renderInterval caps how often the dashboard redraws. The CAN bus
// itself updates much faster than this; there's no point redrawing the
// panel faster than a human can read it, and every full-screen redraw
// costs real SPI bus time.
const renderInterval = 100 * time.Millisecond

func main() {
	display := setupDisplay()
	can := setupCAN()

	state := &signals.State{}
	reg := canbus.NewRegistry()
	gt86.Register(reg, state)

	// The CAN read loop runs forever in its own goroutine; the main
	// goroutine just redraws whatever the latest state is, on a timer.
	go func() {
		err := reg.Run(mcp2515Bus{dev: can})
		// Run only returns on a read error, which means the MCP2515 (or
		// its SPI link) is in a bad state -- crash loudly rather than
		// silently freezing the dash with stale numbers.
		panic("CAN read loop exited: " + err.Error())
	}()

	dash := ui.Dashboard{Theme: ui.Dark, Tiles: ui.DefaultTiles()}
	for {
		snap := state.Snapshot()
		if err := dash.Render(&display, snap); err != nil {
			println("render error:", err.Error())
		}
		time.Sleep(renderInterval)
	}
}

func setupDisplay() st7789.Device {
	spi := machine.SPI2
	err := spi.Configure(machine.SPIConfig{
		Frequency: 40e6,
		SCK:       displaySCK,
		SDO:       displaySDO,
		SDI:       machine.NoPin,
	})
	if err != nil {
		panic("display SPI configure: " + err.Error())
	}

	dev := st7789.New(spi, displayReset, displayDC, displayCS, displayBL)
	dev.Configure(st7789.Config{
		Width:  panelWidth,
		Height: panelHeight,
	})
	return dev
}

func setupCAN() *mcp2515.Device {
	spi := machine.SPI3
	err := spi.Configure(machine.SPIConfig{
		Frequency: 10e6,
		SCK:       canSCK,
		SDO:       canSDO,
		SDI:       canSDI,
	})
	if err != nil {
		panic("CAN SPI configure: " + err.Error())
	}

	dev := mcp2515.New(spi, canCS)
	dev.Configure(mcp2515.Configuration{Extended: false})

	// Most cheap MCP2515 breakout boards ship with a 16MHz crystal; if
	// yours uses an 8MHz crystal, change Clock16MHz to Clock8MHz.
	// 500kbps matches the GT86's factory high-speed CAN bus.
	if err := dev.Begin(mcp2515.CAN500kBps, mcp2515.Clock16MHz); err != nil {
		panic("CAN begin: " + err.Error())
	}
	return dev
}
