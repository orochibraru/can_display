// Package sensors converts raw ADC readings from locally-wired analog
// sensors (things that aren't on the CAN bus at all) into real-world
// units. It's the analog-sensor counterpart to internal/canbus/gt86:
// pure conversion math, no hardware access, so it's testable without
// a board and reusable between the firmware and any future replay/sim
// tooling.
package sensors

// RawToVolts converts a TinyGo-style scaled ADC reading (0..0xffff, see
// machine.ADC) to the voltage actually present at the ADC's input pin,
// given the ADC's reference voltage.
func RawToVolts(raw uint16, vref float32) float32 {
	return float32(raw) / 65535 * vref
}

// Divider describes a resistor voltage divider placed between a sensor
// and an ADC input, used to bring a sensor's output range down to
// something the ADC can safely read. R1 is the resistor between the
// sensor signal and the ADC pin; R2 is the resistor between the ADC
// pin and ground.
type Divider struct {
	R1, R2 float32 // ohms
}

// Ratio is Vout/Vin for this divider: R2/(R1+R2).
func (d Divider) Ratio() float32 {
	return d.R2 / (d.R1 + d.R2)
}

// Undo recovers the original (pre-divider) sensor voltage from the
// voltage actually measured at the ADC pin.
func (d Divider) Undo(measuredVolts float32) float32 {
	return measuredVolts / d.Ratio()
}
