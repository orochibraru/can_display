// Package sim fabricates plausible-looking signal data so the dashboard
// UI can be developed and demoed without a car.
package sim

import (
	"math"
	"math/rand"
	"time"

	"orochibraru/can_display/internal/signals"
)

// Generator writes fake readings straight into a signals.State on every
// Tick, standing in for what a real canbus.Registry of decoders would
// otherwise do. It is not a CAN bus emulator -- it never produces
// frames, only final signal values.
type Generator struct {
	start time.Time
	rng   *rand.Rand

	// EnableAFR/EnableEthanol/EnableBattery let you preview what the
	// dashboard looks like once those signals are actually wired up to
	// a data source, even though the real firmware doesn't have one yet
	// (see internal/canbus/gt86). Off by default so the simulator's
	// default view matches what the firmware can actually show today.
	EnableAFR, EnableEthanol, EnableBattery bool
}

// NewGenerator returns a Generator whose clock starts now.
func NewGenerator() *Generator {
	return &Generator{start: time.Now(), rng: rand.New(rand.NewSource(1))}
}

// Tick writes one fresh set of fake readings into state, timestamped at
// now.
func (g *Generator) Tick(state *signals.State, now time.Time) {
	t := now.Sub(g.start).Seconds()

	// Coolant/oil temp: ramp up from a cold start over ~90s, then hover
	// with small noise, like a warming engine settling at temperature.
	warmup := clamp(float32(t/90), 0, 1)
	coolant := 20 + warmup*70 + noise(g.rng, 1.5)
	oil := 20 + warmup*95 + noise(g.rng, 2)
	state.SetCoolantTempC(coolant, now)
	state.SetOilTempC(oil, now)

	if g.EnableBattery {
		vibration := float32(math.Sin(t / 5))
		state.SetBatteryVoltage(14.2+vibration*0.3+noise(g.rng, 0.05), now)
	}
	if g.EnableAFR {
		swing := float32(math.Sin(t / 2))
		state.SetAFR(14.7+swing*1.5+noise(g.rng, 0.2), now)
	}
	if g.EnableEthanol {
		state.SetEthanolPercent(10+noise(g.rng, 1), now)
	}
}

func noise(rng *rand.Rand, amplitude float32) float32 {
	return (rng.Float32()*2 - 1) * amplitude
}

func clamp(v, min, max float32) float32 {
	switch {
	case v < min:
		return min
	case v > max:
		return max
	default:
		return v
	}
}
