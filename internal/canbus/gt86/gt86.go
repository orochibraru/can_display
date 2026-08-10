// Package gt86 decodes the factory CAN bus of the "Gen1" Toyota
// 86 / Subaru BRZ / Scion FR-S (2012-2020 model years, which covers a
// 2015 GT86).
//
// IMPORTANT: these IDs and formulas come from community
// reverse-engineering, not from any Toyota/Subaru factory
// documentation, and have not been verified against this specific car.
// Treat every value here as "probably right, please confirm" -- see
// /docs/can-reverse-engineering.md for how to check them with a cheap
// USB-CAN adapter before trusting the dash.
//
// Primary source: https://github.com/timurrrr/ft86/blob/main/can_bus/gen1.md
// (captured from the OBD-II port, pins 6/14 (CAN H/L), on a Gen1
// 86/BRZ/FR-S.)
package gt86

import (
	"time"

	"orochibraru/can_display/internal/canbus"
	"orochibraru/can_display/internal/signals"
)

// IDTemperatures (0x360, ~20Hz) carries oil and coolant temperature
// among other things (cruise control state). Both temperatures are
// encoded as (raw byte - 40) degrees Celsius.
const IDTemperatures uint32 = 0x360

// Not currently decoded, because they aren't on the factory bus at all
// (confirmed absent from every Gen1/Gen2 community reverse-engineering
// doc found so far):
//
//   - Battery voltage: needs further reverse-engineering work. If you
//     find it, add a decoder here.
//   - AFR / ethanol percentage: the factory ECU doesn't compute or
//     broadcast either. These only become available once an aftermarket
//     wideband O2 controller or flex-fuel sensor is added and configured
//     to broadcast onto the bus (or its own separate CAN channel).

// Register wires up every known decoder into reg, writing results into
// state.
func Register(reg *canbus.Registry, state *signals.State) {
	reg.Register(IDTemperatures, decodeTemperatures(state))
}

func decodeTemperatures(state *signals.State) canbus.Decoder {
	return func(f canbus.Frame, at time.Time) {
		if f.DLC < 4 {
			return
		}
		oilC := float32(f.Data[2]) - 40
		coolantC := float32(f.Data[3]) - 40
		state.SetOilTempC(oilC, at)
		state.SetCoolantTempC(coolantC, at)
	}
}
