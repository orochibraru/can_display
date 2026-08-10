//go:build tinygo

// Command firmware is the ESP32 program: it reads the GT86's CAN bus
// through an MCP2515 SPI controller, reads the flex-fuel sensor through
// an MCP3008 SPI ADC, and renders internal/ui's dashboard to an ST7789
// SPI display. It shares every package under internal/ with
// cmd/simulator -- only this file and the pin wiring are specific to
// real hardware.
//
// Build with TinyGo, not `go build` (this file needs the `machine`
// package, which only exists under the TinyGo compiler). See the
// Makefile for the exact command.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/mcp2515"
	"tinygo.org/x/drivers/mcp3008"
	"tinygo.org/x/drivers/st7789"

	"orochibraru/can_display/internal/canbus"
	"orochibraru/can_display/internal/canbus/gt86"
	"orochibraru/can_display/internal/sensors"
	"orochibraru/can_display/internal/signals"
	"orochibraru/can_display/internal/ui"
)

// Pin wiring -- adjust these to match your actual breadboard/PCB layout;
// see /docs/wiring.md for the reference wiring this project assumes.
//
// The classic ESP32 only has two general-purpose SPI peripherals
// (SPI2/SPI3 -- SPI0/SPI1 are reserved for flash). The display gets its
// own bus (SPI2); the MCP2515 CAN controller and MCP3008 ADC share the
// other (SPI3) with separate chip-selects, since neither needs the bus
// often enough for that to matter.
const (
	displaySCK   = machine.GPIO18
	displaySDO   = machine.GPIO23 // MOSI; display is write-only, no MISO needed
	displayReset = machine.GPIO4
	displayDC    = machine.GPIO2
	displayCS    = machine.GPIO5
	displayBL    = machine.GPIO15

	sharedSCK = machine.GPIO14
	sharedSDO = machine.GPIO13 // MOSI
	sharedSDI = machine.GPIO12 // MISO
	canCS     = machine.GPIO27
	ethanolCS = machine.GPIO26

	// ethanolADCChannel is which MCP3008 input the flex-fuel sensor's
	// (divided-down) signal is wired to. See /docs/wiring.md.
	ethanolADCChannel = 0
)

// adcVref is the MCP3008's reference voltage. The MCP3008 is powered
// from the ESP32's 3.3V rail here (so its SPI logic levels match the
// ESP32 without a level shifter), which also sets its analog reference.
const adcVref float32 = 3.3

// panelWidth/panelHeight must match cmd/simulator's constants -- both
// feed the same internal/ui.Dashboard, which lays itself out to
// whatever size it's given.
const (
	panelWidth  = 240
	panelHeight = 320
)

// tickInterval caps how often the dashboard redraws and the ethanol
// sensor is re-read. The CAN bus itself updates much faster than this;
// there's no point redrawing the panel faster than a human can read it,
// and every full-screen redraw costs real SPI bus time.
const tickInterval = 100 * time.Millisecond

func main() {
	display := setupDisplay()
	sharedSPI := setupSharedSPI()
	can := setupCAN(sharedSPI)
	ethanolADC := setupEthanolADC(sharedSPI)

	state := &signals.State{}
	reg := canbus.NewRegistry()
	gt86.Register(reg, state)
	bus := mcp2515Bus{dev: can}

	// The CAN read loop runs forever in its own goroutine; the main
	// goroutine polls the ADC and redraws on a timer.
	go func() {
		err := reg.Run(bus)
		// Run only returns on a read error, which means the MCP2515 (or
		// its SPI link) is in a bad state -- crash loudly rather than
		// silently freezing the dash with stale numbers.
		panic("CAN read loop exited: " + err.Error())
	}()

	// AFR needs an active request/response, not passive listening --
	// see internal/canbus/gt86's afr.go for why, and for how
	// speculative this one is. Its response comes back through reg
	// above like any other frame.
	go func() {
		if err := gt86.AFRPoller(bus).Run(); err != nil {
			panic("AFR poll loop exited: " + err.Error())
		}
	}()

	dash := ui.Dashboard{Theme: ui.Dark, Tiles: ui.DefaultTiles()}
	for {
		updateEthanolPercent(ethanolADC, state)

		snap := state.Snapshot()
		if err := dash.Render(&display, snap); err != nil {
			println("render error:", err.Error())
		}
		time.Sleep(tickInterval)
	}
}

// updateEthanolPercent reads the flex-fuel sensor through the MCP3008,
// undoes the resistor divider, and converts the result to a percentage.
// See internal/sensors for the conversion math and its caveats.
func updateEthanolPercent(adc *mcp3008.Device, state *signals.State) {
	raw, err := adc.Read(ethanolADCChannel)
	if err != nil {
		// Leave the last known reading in place; if the ADC is actually
		// dead this signal will go stale on its own and the dashboard
		// will show "--" rather than a frozen number.
		return
	}
	adcVolts := sensors.RawToVolts(raw, adcVref)
	sensorVolts := sensors.EthanolDivider.Undo(adcVolts)
	state.SetEthanolPercent(sensors.EthanolPercent(sensorVolts), time.Now())
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

// setupSharedSPI configures the bus the MCP2515 and MCP3008 share.
//
// The clock is deliberately conservative (1MHz): the MCP3008 is only
// rated for a couple MHz when powered at 3.3V (its max SPI clock scales
// with VDD), while the MCP2515 is happy anywhere up to several MHz. One
// shared frequency has to satisfy both, and neither device is read
// often enough for 1MHz to be a bottleneck.
func setupSharedSPI() drivers.SPI {
	spi := machine.SPI3
	err := spi.Configure(machine.SPIConfig{
		Frequency: 1e6,
		SCK:       sharedSCK,
		SDO:       sharedSDO,
		SDI:       sharedSDI,
	})
	if err != nil {
		panic("shared SPI configure: " + err.Error())
	}
	return spi
}

func setupCAN(spi drivers.SPI) *mcp2515.Device {
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

func setupEthanolADC(spi drivers.SPI) *mcp3008.Device {
	dev := mcp3008.New(spi, ethanolCS)
	dev.Configure()
	return dev
}
