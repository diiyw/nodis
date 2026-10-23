package protocol

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
)

const CommandHeaderSize = 22

var (
	ErrInvalidCommand = errors.New("nodis protocol: invalid command frame")
	commandMagic      = [4]byte{'N', 'D', 'C', '1'}
)

type Op uint8

const (
	OpAuthenticate Op = iota + 1
	OpGet
	OpCreate
	OpUpdate
	OpDelete
	OpSubscribe
	OpGetDocument
	OpCreateDocument
	OpRenameDocument
	OpDeleteDocument
	OpFind
	OpEvent Op = 128
)

type Status uint8

const (
	StatusOK Status = iota
	StatusInvalid
	StatusUnauthorized
	StatusForbidden
	StatusNotFound
	StatusConflict
	StatusInternal
	StatusTooLarge
)

type Command struct {
	RequestID uint64
	Op        Op
	Status    Status
	Payload   []byte
}

func AppendCommand(dst []byte, command Command) []byte {
	start := len(dst)
	dst = append(dst, commandMagic[:]...)
	dst = append(dst, 0, 0, 0, 0)
	dst = binary.LittleEndian.AppendUint64(dst, command.RequestID)
	dst = append(dst, byte(command.Op), byte(command.Status))
	dst = binary.LittleEndian.AppendUint32(dst, uint32(len(command.Payload)))
	dst = append(dst, command.Payload...)
	binary.LittleEndian.PutUint32(dst[start+4:start+8], uint32(len(dst)-start+4))
	dst = binary.LittleEndian.AppendUint32(dst, crc32.Checksum(dst[start:], table))
	return dst
}

func DecodeCommand(frame []byte) (Command, error) {
	if len(frame) < CommandHeaderSize+4 || len(frame) > MaxFrameSize || string(frame[:4]) != string(commandMagic[:]) {
		return Command{}, ErrInvalidCommand
	}
	if binary.LittleEndian.Uint32(frame[4:8]) != uint32(len(frame)) {
		return Command{}, ErrInvalidCommand
	}
	if crc32.Checksum(frame[:len(frame)-4], table) != binary.LittleEndian.Uint32(frame[len(frame)-4:]) {
		return Command{}, ErrInvalidCommand
	}
	size := int(binary.LittleEndian.Uint32(frame[18:22]))
	if size != len(frame)-CommandHeaderSize-4 {
		return Command{}, ErrInvalidCommand
	}
	return Command{RequestID: binary.LittleEndian.Uint64(frame[8:16]), Op: Op(frame[16]), Status: Status(frame[17]), Payload: frame[22 : len(frame)-4]}, nil
}
