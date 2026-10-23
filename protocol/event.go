// Package protocol implements the transport-independent Nodis event wire format.
package protocol

import (
	"encoding/binary"
	"errors"
	"hash/crc32"

	"github.com/diiyw/nodis"
)

const (
	HeaderSize   = 53
	MaxFrameSize = 64 << 20
)

var (
	ErrInvalidFrame = errors.New("nodis protocol: invalid frame")
	magic           = [4]byte{'N', 'D', 'E', '1'}
	table           = crc32.MakeTable(crc32.Castagnoli)
)

func AppendEvent(dst []byte, event nodis.Event) []byte {
	start := len(dst)
	dst = append(dst, magic[:]...)
	dst = append(dst, 0, 0, 0, 0)
	dst = append(dst, byte(event.Type))
	dst = binary.LittleEndian.AppendUint64(dst, event.Sequence)
	dst = binary.LittleEndian.AppendUint64(dst, event.PermissionRevision)
	dst = binary.LittleEndian.AppendUint64(dst, uint64(event.Document))
	dst = binary.LittleEndian.AppendUint64(dst, uint64(event.Object))
	dst = binary.LittleEndian.AppendUint64(dst, event.Revision)
	dst = binary.LittleEndian.AppendUint32(dst, uint32(len(event.Data)))
	dst = append(dst, event.Data...)
	binary.LittleEndian.PutUint32(dst[start+4:start+8], uint32(len(dst)-start+4))
	checksum := crc32.Checksum(dst[start:], table)
	dst = binary.LittleEndian.AppendUint32(dst, checksum)
	return dst
}

func DecodeEvent(frame []byte) (nodis.Event, error) {
	if len(frame) < HeaderSize+4 || len(frame) > MaxFrameSize || string(frame[:4]) != string(magic[:]) {
		return nodis.Event{}, ErrInvalidFrame
	}
	if binary.LittleEndian.Uint32(frame[4:8]) != uint32(len(frame)) {
		return nodis.Event{}, ErrInvalidFrame
	}
	if crc32.Checksum(frame[:len(frame)-4], table) != binary.LittleEndian.Uint32(frame[len(frame)-4:]) {
		return nodis.Event{}, ErrInvalidFrame
	}
	dataSize := int(binary.LittleEndian.Uint32(frame[49:53]))
	if dataSize != len(frame)-HeaderSize-4 {
		return nodis.Event{}, ErrInvalidFrame
	}
	event := nodis.Event{
		Type:               nodis.EventType(frame[8]),
		Sequence:           binary.LittleEndian.Uint64(frame[9:17]),
		PermissionRevision: binary.LittleEndian.Uint64(frame[17:25]),
		Document:           nodis.DocumentID(binary.LittleEndian.Uint64(frame[25:33])),
		Object:             nodis.ObjectID(binary.LittleEndian.Uint64(frame[33:41])),
		Revision:           binary.LittleEndian.Uint64(frame[41:49]),
		Data:               frame[53 : len(frame)-4],
	}
	if event.Type < nodis.EventDocumentCreated || event.Type > nodis.EventReset {
		return nodis.Event{}, ErrInvalidFrame
	}
	return event, nil
}
