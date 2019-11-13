package glowglasses

import (
	"github.com/muka/go-bluetooth/bluez/profile/device"
	"github.com/muka/go-bluetooth/bluez/profile/gatt"
)

type GlowGlassesX struct {
	Device *device.Device1
	Char   *gatt.GattCharacteristic1
}

func New(dev *device.Device1, ch *gatt.GattCharacteristic1) *GlowGlassesX {
	return &GlowGlassesX{Device: dev, Char: ch}
}

func (gg *GlowGlassesX) Call(cmd []byte) error {
	return gg.Char.WriteValue(cmd, nil)
}

func (gg *GlowGlassesX) On() error {
	return gg.Call([]byte{0x01, 0x00, 0x02, 0x06, 0x09, 0x02, 0x05, 0x03})
}

func (gg *GlowGlassesX) Off() error {
	return gg.Call([]byte{0x01, 0x00, 0x02, 0x06, 0x09, 0x00, 0x03})
}
