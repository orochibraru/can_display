# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

DIY CAN bus gauge cluster for a 2015 Toyota GT86: TinyGo firmware for a Raspberry Pi Pico W(H) + ST7789 V2.2 2" 240x320 SPI display + MCP2515 CAN controller, plus a desktop simulator that renders the same UI.

## Commands

```sh
make sim         # desktop simulator (ebiten window, fake data), no hardware needed
make test        # go test ./... (all tests live in tests/)
make vet         # go vet internal + simulator, then tinygo build-check of the firmware
make firmware    # tinygo build -> bin/firmware.uf2 (TARGET defaults to pico-w)
make flash       # tinygo flash; Pico must be in BOOTSEL mode
go test ./tests/ -run TestDecodeRPM   # single test
```

The firmware needs TinyGo (`brew tap tinygo-org/tools && brew install tinygo`). Everything else is plain Go.

## Architecture

Data flows one way: **source -> `signals.State` -> `ui.Dashboard.Render(display, snapshot)`**.

- `cmd/firmware` (build tag `//go:build tinygo`, so `go build ./...` skips it) is the only hardware-specific code: pin constants, SPI setup, the MCP2515 adapter (`mcp2515_bus.go` turns the driver's polling API into the blocking `canbus.Bus`/`canbus.Writer`), and the RP2040 ADC read for ethanol. It runs the CAN read loop and the AFR OBD-II poller as goroutines and redraws every 100ms.
- `cmd/simulator` feeds `internal/sim.Generator` fake values into the same `signals.State` and renders the same `ui.Dashboard` onto a software framebuffer. It is meant to be a pixel-exact preview (both targets use `tinyfont`), so UI changes are developed there.
- `internal/display.Display` is a type alias for `tinygo.org/x/drivers.Displayer`, which is why any TinyGo display driver and the simulator framebuffer both plug in without glue.
- `internal/canbus`: `Frame`, read-only `Bus`, separate `Writer` (only needed for request/response), and `Registry`, which dispatches frames by CAN ID to `Decoder` funcs.
- `internal/canbus/gt86`: car-specific decoders registered via `gt86.Register(reg, state)`. That covers the passively broadcast IDs (`0x140` RPM, `0x360` temps) and also registers the AFR response decoder. `AFRPoller` sends the matching requests.
- `internal/canbus/obd2`: generic ISO 15765-4 Mode 01 request framing, single-frame response parsing, and `Poller`.
- `internal/sensors`: pure conversion math for locally wired analog sensors (ADC raw -> volts, resistor divider undo, ethanol sensor curve). No hardware access.
- `internal/signals`: thread-safe `State` with one `Reading` (value, valid, timestamp) per signal. `Snapshot()` is what the UI consumes, and stale or never-set readings render as "--" instead of a frozen or zero value.
- `internal/ui`: `config.go` holds all normal/warning/danger thresholds; retune gauges there, not in `dashboard.go`, which only wires those values into tiles.

Adding a signal usually touches: a setter/field in `signals`, a decoder in `gt86` (or a sensor function plus a firmware read), a tile in `ui/dashboard.go` + range in `ui/config.go`, and optionally `sim.Generator`.

## Conventions and gotchas

- Tests live in `tests/` as a single external package (`package tests`), so they can only reach exported identifiers. Anything a test needs must be exported.
- CAN IDs and formulas come from community reverse-engineering and are **unverified on this car**. AFR via PID 0x24 is speculative. Don't present them as confirmed. `docs/can-reverse-engineering.md` covers how to verify them.
- Hardware facts that constrain the code (see `docs/wiring.md`): the RP2040 is not 5V tolerant (MCP2515 SO line needs level shifting). GPIO23/24/25/29 are reserved by the Pico W's wireless chip. The flex-fuel sensor outputs 0.5-4.5V, so it goes through the `sensors.EthanolDivider` (a test asserts the divided max stays under 3.3V). `adcVref` is a calibration knob.
- Pin constants in `cmd/firmware/main.go` and the tables in `docs/wiring.md` must stay in sync. Same for the panel size constants in `cmd/firmware` and `cmd/simulator`.
