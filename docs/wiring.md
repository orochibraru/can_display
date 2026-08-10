# Reference wiring

This matches the pin constants at the top of `cmd/firmware/main.go`.
Change either the wiring or the constants if yours differs — they just
need to agree.

Board target: `esp32-coreboard-v2` (a generic ESP32 dev board). If you're
using a different board, `tinygo targets` lists the alternatives, and
you'll likely need different GPIO numbers (check your board's pinout
diagram).

## Display: ST7789, 2.0" 240x320 SPI TFT

| Display pin | ESP32 pin | Notes |
|---|---|---|
| SCK  | GPIO18 | SPI2 clock |
| SDA/MOSI | GPIO23 | SPI2 data out |
| RES/RST | GPIO4 | |
| DC   | GPIO2  | |
| CS   | GPIO5  | |
| BLK  | GPIO15 | backlight; driven high in software |
| VCC  | 3V3 | most of these panels are 3.3V logic *and* power — double check yours |
| GND  | GND | |

No MISO connection is used — the display is write-only.

## CAN: MCP2515 SPI CAN controller module

| MCP2515 pin | ESP32 pin | Notes |
|---|---|---|
| SCK  | GPIO14 | SPI3 clock |
| SI (MOSI) | GPIO13 | |
| SO (MISO) | GPIO12 | |
| CS   | GPIO27 | |
| INT  | *(unused)* | firmware polls instead of using the interrupt pin; wire it up later if you want an interrupt-driven read loop |
| VCC  | 5V | most MCP2515 boards want 5V, with an onboard regulator/level-shifter for the SPI lines — check your specific board |
| GND  | GND | |
| CAN-H / CAN-L | car's OBD-II port, pins 6 and 14 | see below |

The firmware assumes a **16MHz crystal** on the MCP2515 board (the
common default on cheap breakout boards) and **500kbps**, which matches
the GT86's factory high-speed CAN bus. If your board uses an 8MHz
crystal, change `mcp2515.Clock16MHz` to `mcp2515.Clock8MHz` in
`setupCAN()`.

## Ethanol content: analog flex-fuel sensor, via MCP3008

TinyGo has no ADC support at all for the classic ESP32 (`machine.ADC`
only exists for ESP32-S3/C3 as of TinyGo 0.41) — confirmed by trying to
compile against it, not assumed. Rather than switch board or write an
unproven bare-metal ADC driver, the sensor is read through a cheap
external SPI ADC instead, sharing the CAN controller's bus.

**The sensor outputs 0.5V-4.5V. That exceeds what a 3.3V-rail chip can
safely see on an analog input — do not wire it in directly.** A
resistor divider brings it down first:

```
sensor signal ──[ R1: 10kΩ ]──┬── MCP3008 CH0
                                │
                          [ R2: 22kΩ ]
                                │
                               GND
```

This divider (ratio 22/32 ≈ 0.6875) brings the sensor's worst-case 4.5V
down to ~3.09V at the ADC pin — safely under the MCP3008's 3.3V rail,
with the math in `internal/sensors` (and a test asserting it stays
under the rail voltage, so a future edit can't silently make this
unsafe).

| MCP3008 pin | ESP32 pin | Notes |
|---|---|---|
| CLK  | GPIO14 | shared with MCP2515 (SPI3) |
| DOUT | GPIO12 | shared (MISO) |
| DIN  | GPIO13 | shared (MOSI) |
| CS   | GPIO26 | own chip-select |
| CH0  | divider output (above) | flex-fuel sensor, through the divider |
| VDD, VREF | 3.3V | **must match the ESP32's logic rail** — this is what makes CH0's usable range 0-3.3V |
| AGND, DGND | GND | |

If you're using a different sensor with a different voltage range (or
find out this one's endpoints/direction are different from the
assumed 0.5V=0%/4.5V=100%), the two things to change are the divider
resistor values (`sensors.EthanolDivider`) and the sensor curve
(`sensors.EthanolPercent`) — both isolated in `internal/sensors` with
their own unit tests.

## Connecting to the car

The GT86/BRZ/FR-S OBD-II port carries CAN H/L on the standard pins:

- Pin 6: CAN H
- Pin 14: CAN L

Do **not** connect the MCP2515's ground to anything other than a real
chassis/OBD ground, and don't power the ESP32 from the car's 12V rail
without a proper automotive-grade buck converter (a car's electrical
system has voltage spikes well outside what a USB power path is rated
for). A cheap USB car charger or a dedicated 12V→5V buck module is the
easy, safe option for prototyping on the bench before it goes anywhere
near the dash.

## Two SPI buses, three devices

The classic ESP32 only exposes two general-purpose hardware SPI
peripherals through TinyGo (`machine.SPI2` and `machine.SPI3` — SPI0/1
are reserved for flash), so the display gets its own bus (SPI2) and the
MCP2515 + MCP3008 share the other (SPI3) with separate chip-selects.

A shared bus runs at one clock frequency for every device on it. The
firmware clocks SPI3 at a conservative 1MHz, because the MCP3008's
maximum SPI clock scales down with its supply voltage and it's only
rated for a couple MHz at 3.3V — the MCP2515 is comfortable at 1MHz
too, just doing slightly less work per second, which doesn't matter at
the polling rate this project reads it at. The display keeps its own
faster 40MHz bus since full-screen redraws are the one place SPI speed
actually matters here.

If you're pin-constrained and want to put the display on the shared bus
too, that works, but a slow full-screen redraw would then briefly delay
CAN/ADC reads (and vice versa) — worth checking the CAN read loop still
keeps up before relying on it.
