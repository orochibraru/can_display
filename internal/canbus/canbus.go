// Package canbus provides the hardware-agnostic pieces of talking CAN:
// a Frame type, a Bus interface any source of frames can implement (a
// real MCP2515 controller, a candump log replayed for testing, a fake
// generator), and a Registry that dispatches frames to decoders by ID.
package canbus

import "time"

// Frame is a single classic CAN 2.0 frame. This project only needs
// classic CAN (no FD), which is what the GT86's factory bus runs.
type Frame struct {
	ID   uint32
	DLC  uint8
	Data [8]byte
}

// Bus is anything that can hand us frames as they arrive.
type Bus interface {
	// ReadFrame blocks until a frame is available and returns it, or
	// returns an error (including io.EOF for a finite source like a
	// replayed log).
	ReadFrame() (Frame, error)
}

// Writer is implemented by a Bus that can also transmit frames. Kept
// separate from Bus (rather than folded into it) because most of this
// project only ever listens -- a replayed candump log, for instance,
// has no meaningful way to "send" anything. Only protocols that need
// to ask for data, like OBD-II request/response (see
// internal/canbus/obd2), need this.
type Writer interface {
	WriteFrame(f Frame) error
}

// Decoder updates application state from one CAN frame. Decoders should
// return quickly (no blocking I/O) since they run inline with the read
// loop.
type Decoder func(f Frame, at time.Time)

// Registry dispatches incoming frames to whichever decoder(s) are
// registered for that frame's ID. Frames with no registered decoder are
// silently ignored -- the bus carries plenty of IDs we don't care about.
type Registry struct {
	decoders map[uint32][]Decoder
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{decoders: make(map[uint32][]Decoder)}
}

// Register adds d as a decoder for frames with the given ID. Multiple
// decoders may be registered for the same ID.
func (r *Registry) Register(id uint32, d Decoder) {
	r.decoders[id] = append(r.decoders[id], d)
}

// Dispatch runs every decoder registered for f.ID.
func (r *Registry) Dispatch(f Frame, at time.Time) {
	for _, d := range r.decoders[f.ID] {
		d(f, at)
	}
}

// Run reads frames from bus and dispatches each one, until ReadFrame
// returns an error. It's meant to run in its own goroutine (or, on the
// firmware, as the body of the main loop).
func (r *Registry) Run(bus Bus) error {
	for {
		f, err := bus.ReadFrame()
		if err != nil {
			return err
		}
		r.Dispatch(f, time.Now())
	}
}
