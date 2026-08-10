package obd2

// PIDEquivalenceRatioO2S1 is Mode 01 PID 0x24: "O2 Sensor 1, Wide Range
// Equivalence Ratio and Voltage" -- the standard SAE J1979 PID for
// reading a wideband/wide-range A/F sensor's output. PIDs 0x25-0x2B
// cover O2 sensors 2-8 with the same payload layout.
const PIDEquivalenceRatioO2S1 byte = 0x24

// EquivalenceRatio decodes the lambda (equivalence ratio) portion of a
// PID 0x24-0x2B response payload: the first two of its four bytes (A,
// B). The remaining two bytes (C, D) carry the sensor's raw voltage,
// which this project doesn't currently use.
//
// Formula per SAE J1979: lambda = 2/65536 * (256*A + B), range [0, 2).
// lambda of 1.0 is stoichiometric; multiply by a fuel's stoichiometric
// AFR (14.7 for gasoline) to get a conventional AFR reading.
func EquivalenceRatio(payload []byte) (lambda float32, ok bool) {
	if len(payload) < 2 {
		return 0, false
	}
	raw := uint16(payload[0])<<8 | uint16(payload[1])
	return float32(raw) * 2 / 65536, true
}
