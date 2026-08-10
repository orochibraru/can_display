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

## Two SPI buses, not one

The firmware puts the display on `machine.SPI2` and the CAN controller
on `machine.SPI3` — two independent hardware SPI peripherals on the
classic ESP32 — rather than sharing one bus with two chip-selects. This
means a slow full-screen display redraw can never stall CAN reception
(or vice versa). If you're pin-constrained, sharing a single bus with
separate CS lines does work, but you'll want to make sure the CAN read
loop keeps up even during a redraw.
