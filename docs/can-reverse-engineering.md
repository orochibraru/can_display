# Verifying and extending the GT86 CAN decoders

## Why this matters

The IDs and formulas in `internal/canbus/gt86` come from community
reverse-engineering (see the source link in that package's doc comment),
not from Toyota/Subaru factory documentation. They have **not** been
verified against this specific car. Treat every value as "probably
right, please confirm" before trusting anything the dash shows you —
especially before making driving decisions based on it.

## What's already wired up

| Signal | Status |
|---|---|
| Coolant temp | Decoded from `0x360` byte 3, `value - 40`°C. Unverified on this car. |
| Oil temp | Decoded from `0x360` byte 2, `value - 40`°C. Unverified on this car. |
| Battery voltage | **Not found** in any factory-bus reverse-engineering doc so far. Needs new reverse-engineering work (see below). |
| AFR | Not on the factory bus at all — the stock ECU doesn't compute or broadcast it. Needs an aftermarket wideband O2 controller broadcasting onto the bus. |
| Ethanol % | Not on the factory bus at all — needs an aftermarket flex-fuel sensor broadcasting onto the bus. |

## How to check (or find) an ID

You don't need this project's hardware to do this — any cheap USB-CAN
adapter (e.g. an $8 "CANable"-clone, or even another MCP2515 board on a
breadboard talking to a laptop) plus `candump`/`cansniffer`
(linux `can-utils`) or [SavvyCAN](https://www.savvycan.com/) (cross
platform, nicer UI) works.

1. Connect to the OBD-II port, pins 6 (CAN H) and 14 (CAN L).
2. With the engine idling and everything electrically "boring" (no
   accessories being toggled), log a minute or two of traffic.
3. To find a specific signal: change *one thing* at a time and diff the
   log against a baseline.
   - Coolant/oil temp: compare a cold start against 5 minutes of idling
     — the target bytes should climb steadily as the values in
     `0x360` are expected to (byte 2 and 3, `raw - 40` = °C). If they
     don't match on your car, `cansniffer` (which highlights bytes
     that change) makes this fast to spot.
   - Battery voltage: turn on/off a big electrical load (headlights,
     rear defroster) and watch for a byte that dips and recovers by
     roughly the right amount (idle ~14.2V, load ~13.5-13.8V).
4. Once you've found a candidate ID/byte/formula, add a decoder in
   `internal/canbus/gt86/gt86.go` following the pattern of
   `decodeTemperatures`, and register it in `Register()`.
5. Update the status table above and the comment citing your source
   (even if the source is "captured myself on <date>, see <log file>")
   so the next person (including future you) knows what's actually been
   confirmed versus copied from a forum post.

## Adding an aftermarket AFR/ethanol source

If you add a standalone wideband controller or flex-fuel sensor that
broadcasts onto CAN (many popular ones do, e.g. Haltech's CAN broadcast
protocol, AEM's, etc.), the pattern is the same: find its documented
frame layout (these are usually actually documented by the
manufacturer, unlike the factory bus), write a decoder function with
the same shape as `decodeTemperatures`, and `reg.Register(id, decoder)`
it. The rest of the pipeline — `internal/ui`, the simulator, staleness
handling — doesn't need to change at all; it already has tiles wired up
for `AFR` and `EthanolPercent` in `internal/signals`, just waiting for
something to call `state.SetAFR(...)` / `state.SetEthanolPercent(...)`.
