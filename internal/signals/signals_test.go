package signals

import (
	"testing"
	"time"
)

func TestReadingStale(t *testing.T) {
	now := time.Now()

	unset := Reading{}
	if !unset.Stale(now, time.Second) {
		t.Error("a never-set Reading should be stale")
	}

	fresh := Reading{Value: 42, Valid: true, UpdatedAt: now}
	if fresh.Stale(now, time.Second) {
		t.Error("a Reading updated right now should not be stale")
	}

	old := Reading{Value: 42, Valid: true, UpdatedAt: now.Add(-10 * time.Second)}
	if !old.Stale(now, time.Second) {
		t.Error("a Reading older than maxAge should be stale")
	}
}

func TestStateSetAndSnapshot(t *testing.T) {
	s := &State{}
	now := time.Now()

	s.SetCoolantTempC(88, now)
	s.SetBatteryVoltage(13.8, now)

	snap := s.Snapshot()
	if !snap.CoolantTempC.Valid || snap.CoolantTempC.Value != 88 {
		t.Errorf("CoolantTempC = %+v, want valid 88", snap.CoolantTempC)
	}
	if !snap.BatteryVoltage.Valid || snap.BatteryVoltage.Value != 13.8 {
		t.Errorf("BatteryVoltage = %+v, want valid 13.8", snap.BatteryVoltage)
	}
	// Never set -- should stay invalid rather than reading as zero.
	if snap.AFR.Valid {
		t.Error("AFR should still be invalid: never set")
	}
}
