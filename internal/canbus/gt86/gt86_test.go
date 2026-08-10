package gt86

import (
	"testing"
	"time"

	"orochibraru/can_display/internal/canbus"
	"orochibraru/can_display/internal/signals"
)

func TestDecodeTemperatures(t *testing.T) {
	state := &signals.State{}
	reg := canbus.NewRegistry()
	Register(reg, state)

	// Byte 2 = oil temp, byte 3 = coolant temp, both offset by 40 per
	// gen1.md (see package doc for source/caveats).
	frame := canbus.Frame{
		ID:   IDTemperatures,
		DLC:  8,
		Data: [8]byte{0, 0, 100, 85, 0, 0, 0, 0}, // oil=60C, coolant=45C
	}
	reg.Dispatch(frame, time.Now())

	snap := state.Snapshot()
	if !snap.OilTempC.Valid || snap.OilTempC.Value != 60 {
		t.Errorf("OilTempC = %+v, want valid 60", snap.OilTempC)
	}
	if !snap.CoolantTempC.Valid || snap.CoolantTempC.Value != 45 {
		t.Errorf("CoolantTempC = %+v, want valid 45", snap.CoolantTempC)
	}
}

func TestDecodeTemperaturesIgnoresShortFrames(t *testing.T) {
	state := &signals.State{}
	reg := canbus.NewRegistry()
	Register(reg, state)

	// DLC < 4 means bytes 2/3 aren't guaranteed present; the decoder
	// should skip it rather than read garbage.
	reg.Dispatch(canbus.Frame{ID: IDTemperatures, DLC: 2}, time.Now())

	if state.Snapshot().OilTempC.Valid {
		t.Error("OilTempC should still be invalid after a too-short frame")
	}
}

func TestUnregisteredIDIsIgnored(t *testing.T) {
	state := &signals.State{}
	reg := canbus.NewRegistry()
	Register(reg, state)

	reg.Dispatch(canbus.Frame{ID: 0xDEAD, DLC: 8}, time.Now())

	if state.Snapshot().OilTempC.Valid {
		t.Error("an unrelated CAN ID should not produce a reading")
	}
}
