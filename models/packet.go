package models

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	MaxPacketSize = 64
)

const (
	REPORT_ID_INPUT_BUTTON = iota + 1
	REPORT_ID_INPUT_ENCODER
)

const (
	ACTION_LAUNCH = 0x03
)

type Button struct {
	Pin     uint8
	ID      uint8
	Pressed uint8
}

type Encoder struct {
	PinA          uint8
	PinB          uint8
	LastBPosition uint8
	ID            uint8
	Percentage    uint8
}

type Action struct {
	Type  uint8
	Flags uint8
	Id    uint8
	Data  [13]byte
}

func (a Action) String() string {
	return fmt.Sprintf("Action(Type=0x%02X, Flags=0x%02X, Id=%d, Data=[% X])",
		a.Type, a.Flags, a.Id, a.Data[:])
}

type DeviceSlot struct {
	Row       uint8
	Col       uint8
	Changed   uint8
	Type      uint8
	Action    Action
	UnionData [5]byte
}

func (d DeviceSlot) String() string {
	return fmt.Sprintf("DeviceSlot[Row: %d, Col: %d] Type: %d, Action: %v",
		d.Row, d.Col, d.Type, d.Action)
}

func (d *DeviceSlot) GetEncoder() (*Encoder, error) {
	var enc Encoder
	buf := bytes.NewReader(d.UnionData[:5])
	err := binary.Read(buf, binary.LittleEndian, &enc)
	return &enc, err
}

func (d *DeviceSlot) GetButton() (*Button, error) {
	var btn Button
	buf := bytes.NewReader(d.UnionData[:5])
	err := binary.Read(buf, binary.LittleEndian, &btn)
	return &btn, err
}

func (s *DeviceSlot) Parse(d []byte) (uint8, error) {
	buf := bytes.NewReader(d[1:])
	err := binary.Read(buf, binary.LittleEndian, s)
	return d[0], err
}
