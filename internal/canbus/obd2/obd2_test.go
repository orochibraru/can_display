package obd2

import (
	"testing"

	"orochibraru/can_display/internal/canbus"
)

func TestRequestFrame(t *testing.T) {
	f := RequestFrame(Request{Mode: ModeCurrentData, PID: PIDEquivalenceRatioO2S1})

	if f.ID != FunctionalRequestID {
		t.Errorf("ID = %#x, want %#x", f.ID, FunctionalRequestID)
	}
	if f.DLC != 8 {
		t.Errorf("DLC = %d, want 8", f.DLC)
	}
	want := [8]byte{0x02, 0x01, 0x24, 0, 0, 0, 0, 0}
	if f.Data != want {
		t.Errorf("Data = %v, want %v", f.Data, want)
	}
}

func TestParseSingleFrameResponse(t *testing.T) {
	req := Request{Mode: ModeCurrentData, PID: PIDEquivalenceRatioO2S1}

	// A well-formed positive response: PCI=0x06 (6 bytes follow), mode
	// echoed as 0x41 (0x01 | 0x40), PID 0x24, then 4 payload bytes.
	good := canbus.Frame{
		ID:   ResponseIDECU1,
		DLC:  8,
		Data: [8]byte{0x06, 0x41, 0x24, 0xAA, 0xBB, 0xCC, 0xDD, 0x00},
	}
	payload, ok := ParseSingleFrameResponse(good, req)
	if !ok {
		t.Fatal("expected ok=true for a well-formed response")
	}
	wantPayload := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	if len(payload) != len(wantPayload) {
		t.Fatalf("payload = %v, want %v", payload, wantPayload)
	}
	for i := range payload {
		if payload[i] != wantPayload[i] {
			t.Errorf("payload[%d] = %#x, want %#x", i, payload[i], wantPayload[i])
		}
	}

	cases := map[string]canbus.Frame{
		"wrong PID": {
			ID: ResponseIDECU1, DLC: 8,
			Data: [8]byte{0x06, 0x41, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0x00},
		},
		"negative response (mode not echoed as 0x41)": {
			ID: ResponseIDECU1, DLC: 8,
			Data: [8]byte{0x03, 0x7F, 0x01, 0x12, 0, 0, 0, 0},
		},
		"multi-frame first frame, not single frame": {
			ID: ResponseIDECU1, DLC: 8,
			Data: [8]byte{0x10, 0x14, 0x41, 0x24, 0xAA, 0xBB, 0xCC, 0xDD},
		},
		"too short": {
			ID: ResponseIDECU1, DLC: 1,
			Data: [8]byte{0x06},
		},
	}
	for name, f := range cases {
		if _, ok := ParseSingleFrameResponse(f, req); ok {
			t.Errorf("%s: expected ok=false", name)
		}
	}
}
