package sensors

// EthanolDivider is the resistor divider between the flex-fuel sensor's
// signal wire and the ADC input, sized so the sensor's worst-case 4.5V
// output can never exceed what the ADC (or the ESP32, if reading it
// directly) is rated for.
//
// R1=10k / R2=22k gives a ratio of 22/32 = 0.6875, so:
//   - 4.5V (sensor max) -> ~3.09V at the ADC pin
//   - 0.5V (sensor min) -> ~0.34V at the ADC pin
//
// Both standard E12 resistor values, and 3.09V leaves real headroom
// under a 3.3V rail (and well under the ~3.6V absolute-max most 3.3V
// parts, including the MCP3008 powered at 3.3V, can tolerate) --
// wiring the sensor's raw 4.5V straight into an ADC pin risks damaging
// it. See /docs/wiring.md for the physical layout.
var EthanolDivider = Divider{R1: 10_000, R2: 22_000}

// Sensor output range for a "0.5V-4.5V" linear ratiometric ethanol
// content sensor: 0.5V = 0% ethanol, 4.5V = 100% ethanol.
//
// CONFIRM THIS against your specific sensor's datasheet before
// trusting it: this is the common convention for this class of sensor,
// but the exact endpoints (and which end is 0% vs 100%) vary between
// manufacturers, and using the wrong curve will silently show a
// plausible-looking but wrong ethanol percentage.
const (
	EthanolSensorMinVolts float32 = 0.5 // 0% ethanol
	EthanolSensorMaxVolts float32 = 4.5 // 100% ethanol
)

// EthanolPercent converts the flex-fuel sensor's own output voltage
// (i.e. already undone from whatever divider scaled it down for the
// ADC -- see Divider.Undo) into an ethanol percentage in [0, 100].
func EthanolPercent(sensorVolts float32) float32 {
	span := EthanolSensorMaxVolts - EthanolSensorMinVolts
	pct := (sensorVolts - EthanolSensorMinVolts) / span * 100
	switch {
	case pct < 0:
		return 0
	case pct > 100:
		return 100
	default:
		return pct
	}
}
