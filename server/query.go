package server

import (
	"context"

	"github.com/diiyw/nodis"
	"github.com/diiyw/nodis/protocol"
)

func (wire *serverWire) find(db *nodis.DB, principal nodis.Principal, payload []byte) (protocol.Status, []byte) {
	query, err := protocol.DecodeQueryInto(payload, &wire.queryDecoder)
	if err != nil {
		return protocol.StatusInvalid, nil
	}
	page, err := db.FindInto(context.Background(), principal, query, &wire.queryBuffer)
	if err != nil {
		return errorStatus(err), nil
	}
	encoded, err := protocol.AppendQueryPage(wire.queryResponse[:0], page)
	if err != nil {
		return protocol.StatusInternal, nil
	}
	wire.queryResponse = encoded
	return protocol.StatusOK, encoded
}
