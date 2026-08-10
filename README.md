# can_display

A DIY CAN bus gauge cluster for a 2015 Toyota GT86, built in TinyGo for
an ESP32 + SPI square display.

The project is split so the dashboard UI can be built and iterated on
without a car, an ESP32, or any wiring at all:

```
internal/
  signals/    vehicle signal state (current value + staleness), hardware-agnostic
  canbus/     CAN Frame/Bus/Decoder types, hardware-agnostic
  canbus/gt86/  GT86-specific decoders (which CAN IDs mean what)
  display/    the Display interface everything renders onto
  ui/         the actual dashboard: theme, tiles, layout
  sim/        fake data generator, for development without a car
cmd/
  simulator/  native desktop app (go run), fake data -> internal/ui
  firmware/   ESP32 firmware (tinygo build), real CAN -> internal/ui
```

`cmd/simulator` and `cmd/firmware` both render the *exact same*
`internal/ui.Dashboard` — the simulator is a true preview, not a
lookalike. Only pin wiring and how frames arrive (real MCP2515 vs. a
fake generator) differ between the two.

## Quick start

```sh
make sim        # opens a desktop window with the live dashboard, fake data
make vet         # go vet the shared/simulator code + tinygo-build-check the firmware
make test        # unit tests for the hardware-agnostic packages
make firmware    # cross-compile firmware to bin/firmware.bin (needs tinygo)
make flash       # build + flash onto a connected ESP32 (needs tinygo)
```

Requires [Go](https://go.dev) and, for the firmware targets,
[TinyGo](https://tinygo.org) (`brew tap tinygo-org/tools && brew install
tinygo` on macOS).

## Hardware this targets

- ESP32 dev board (classic ESP32, not C3/S3 — see `docs/wiring.md` if
  you're on a different chip)
- ST7789 2.0" 240x320 SPI TFT display
- MCP2515 SPI CAN controller module, wired to the OBD-II port's CAN H/L
  (pins 6/14)

Full pin mapping: [`docs/wiring.md`](docs/wiring.md).

## What actually works today

| Signal | Source | Status |
|---|---|---|
| Coolant temp | Factory CAN, `0x360` | Decoded, **unverified on this car** |
| Oil temp | Factory CAN, `0x360` | Decoded, **unverified on this car** |
| Battery voltage | — | Not found on the factory bus yet; tile is wired up and waiting |
| AFR | — | Not on the factory bus at all; needs an aftermarket wideband broadcasting on CAN |
| Ethanol % | — | Not on the factory bus at all; needs an aftermarket flex-fuel sensor broadcasting on CAN |

The CAN IDs/formulas in use come from community reverse-engineering,
not factory documentation — **do not trust them blindly**. See
[`docs/can-reverse-engineering.md`](docs/can-reverse-engineering.md)
for how to verify them on your own car (and how to add the missing
ones) with a cheap USB-CAN adapter.

The simulator defaults to previewing all five tiles, including the
three without a real data source yet, so the layout can be designed
before the hardware exists. Flip `Enable*` off in
`cmd/simulator/main.go`'s `main()` to see exactly what the firmware
shows today.

## Design notes

- **`internal/display.Display`** is a type alias for
  `tinygo.org/x/drivers.Displayer`. Any TinyGo display driver satisfies
  it for free, and the simulator's software framebuffer satisfies it
  too — so `internal/ui` never needs to know which one it's talking to.
- **Text rendering** uses `tinygo.org/x/tinyfont` (`freesans`) on both
  targets, on purpose — the simulator should look pixel-for-pixel like
  the real panel, not just "similar."
- **`internal/canbus.Registry`** dispatches frames to decoders by ID.
  The firmware feeds it real MCP2515 frames; a future `internal/sim`
  addition could replay a `candump` log through the same interface for
  regression testing against real captures.
- Every package under `internal/` builds with plain `go build`/`go
  test` — only `cmd/firmware` needs TinyGo (it's gated behind a
  `//go:build tinygo` tag so `go build ./...` from the repo root
  doesn't try and fail to compile it).
