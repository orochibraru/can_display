# TARGET is the TinyGo board target for the firmware build. Override on
# the command line for a different board, e.g.:
#   make firmware TARGET=esp32-generic
# See `tinygo targets` for the full list.
TARGET ?= esp32-coreboard-v2

# PORT is the serial device to flash over. Override if `tinygo flash`
# can't find it automatically, e.g.:
#   make flash PORT=/dev/cu.usbserial-0001
PORT ?=

BIN_DIR := bin

.PHONY: sim build-sim firmware flash vet test tidy clean

## sim: run the desktop dashboard simulator (fake data, no hardware needed)
sim:
	go run ./cmd/simulator

## build-sim: build the simulator to bin/simulator instead of running it directly
build-sim:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/simulator ./cmd/simulator

## firmware: cross-compile the ESP32 firmware to bin/firmware.bin
firmware:
	mkdir -p $(BIN_DIR)
	tinygo build -target $(TARGET) -o $(BIN_DIR)/firmware.bin ./cmd/firmware

## flash: build and flash the firmware onto a connected ESP32
flash:
ifeq ($(strip $(PORT)),)
	tinygo flash -target $(TARGET) ./cmd/firmware
else
	tinygo flash -target $(TARGET) -port $(PORT) ./cmd/firmware
endif

## vet: static-check the shared packages and simulator (native Go) plus the firmware (TinyGo)
vet:
	go vet ./internal/... ./cmd/simulator/...
	tinygo build -target $(TARGET) -o /dev/null ./cmd/firmware

## test: run unit tests across the shared, hardware-agnostic packages
test:
	go test ./...

## tidy: sync go.mod/go.sum with actual imports
tidy:
	go mod tidy

## clean: remove build output
clean:
	rm -rf $(BIN_DIR)
