//go:build tinygo

package main

import (
	"time"

	"tinygo.org/x/drivers/mcp2515"

	"orochibraru/can_display/internal/canbus"
)

// mcp2515Bus adapts an *mcp2515.Device -- which exposes polling methods,
// Received() and Rx() -- to canbus.Bus's blocking ReadFrame(), so the
// same canbus.Registry.Run loop works regardless of what kind of bus
// it's reading from.
type mcp2515Bus struct {
	dev *mcp2515.Device
}

// pollInterval is how often we check the controller for a new message
// while waiting. The MCP2515's INT pin could be wired up for a
// zero-latency interrupt-driven version of this instead; polling is
// simpler and fast enough for dashboard-refresh-rate data.
const pollInterval = time.Millisecond

func (b mcp2515Bus) ReadFrame() (canbus.Frame, error) {
	for !b.dev.Received() {
		time.Sleep(pollInterval)
	}
	msg, err := b.dev.Rx()
	if err != nil {
		return canbus.Frame{}, err
	}
	f := canbus.Frame{ID: msg.ID, DLC: msg.Dlc}
	copy(f.Data[:], msg.Data)
	return f, nil
}
