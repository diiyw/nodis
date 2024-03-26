package pb

import (
	"encoding/binary"
	"hash/crc32"

	"google.golang.org/protobuf/proto"
)

func (e *Entry) Marshal() ([]byte, error) {
	data, err := proto.Marshal(e)
	if err != nil {
		return nil, err
	}
	c32 := crc32.ChecksumIEEE(data)
	var buf = make([]byte, len(data)+4)
	binary.LittleEndian.PutUint32(buf, c32)
	copy(buf[4:], data)
	return buf, nil
}
