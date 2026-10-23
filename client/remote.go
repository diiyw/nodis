package client

import (
	"context"
	"encoding/binary"
	"errors"
	"sort"

	"github.com/diiyw/nodis"
	"github.com/diiyw/nodis/protocol"
)

func (r *Remote) Document(ctx context.Context, name string) (nodis.DocumentID, error) {
	response, err := r.request(ctx, protocol.OpGetDocument, []byte(name))
	if err != nil {
		return 0, err
	}
	if len(response) < 8 {
		return 0, protocol.ErrInvalidCommand
	}
	return nodis.DocumentID(binary.LittleEndian.Uint64(response)), nil
}

func (r *Remote) CreateDocument(ctx context.Context, name string) (nodis.DocumentID, error) {
	response, err := r.request(ctx, protocol.OpCreateDocument, []byte(name))
	if err != nil {
		return 0, err
	}
	if len(response) != 8 {
		return 0, protocol.ErrInvalidCommand
	}
	return nodis.DocumentID(binary.LittleEndian.Uint64(response)), nil
}

func (r *Remote) RenameDocument(ctx context.Context, document nodis.DocumentID, name string) error {
	payload := binary.LittleEndian.AppendUint64(nil, uint64(document))
	payload = append(payload, name...)
	_, err := r.request(ctx, protocol.OpRenameDocument, payload)
	return err
}

func (r *Remote) DeleteDocument(ctx context.Context, document nodis.DocumentID) error {
	_, err := r.request(ctx, protocol.OpDeleteDocument, binary.LittleEndian.AppendUint64(nil, uint64(document)))
	return err
}

func (r *Remote) Get(ctx context.Context, document nodis.DocumentID, object nodis.ObjectID) (*nodis.PortableObject, error) {
	payload := binary.LittleEndian.AppendUint64(nil, uint64(document))
	payload = binary.LittleEndian.AppendUint64(payload, uint64(object))
	response, err := r.request(ctx, protocol.OpGet, payload)
	if err != nil {
		return nil, err
	}
	return nodis.DecodePortable(response)
}

func (r *Remote) Create(ctx context.Context, document nodis.DocumentID, fields map[string]any) (nodis.ObjectID, error) {
	payload, err := appendRemoteMutation(nil, document, 0, 0, fields, false)
	if err != nil {
		return 0, err
	}
	response, err := r.request(ctx, protocol.OpCreate, payload)
	if err != nil {
		return 0, err
	}
	if len(response) != 8 {
		return 0, protocol.ErrInvalidCommand
	}
	return nodis.ObjectID(binary.LittleEndian.Uint64(response)), nil
}

func (r *Remote) Update(ctx context.Context, document nodis.DocumentID, object nodis.ObjectID, revision uint64, fields map[string]any) error {
	payload, err := appendRemoteMutation(nil, document, object, revision, fields, true)
	if err != nil {
		return err
	}
	_, err = r.request(ctx, protocol.OpUpdate, payload)
	return err
}

func (r *Remote) Delete(ctx context.Context, document nodis.DocumentID, object nodis.ObjectID, revision uint64) error {
	payload := binary.LittleEndian.AppendUint64(nil, uint64(document))
	payload = binary.LittleEndian.AppendUint64(payload, uint64(object))
	payload = binary.LittleEndian.AppendUint64(payload, revision)
	_, err := r.request(ctx, protocol.OpDelete, payload)
	return err
}

func (r *Remote) GetAsync(ctx context.Context, document nodis.DocumentID, object nodis.ObjectID) Future[*nodis.PortableObject] {
	done := make(chan struct{})
	result := &Result[*nodis.PortableObject]{}
	go func() { result.Value, result.Err = r.Get(ctx, document, object); close(done) }()
	return Future[*nodis.PortableObject]{done: done, result: result}
}

func (r *Remote) CreateAsync(ctx context.Context, document nodis.DocumentID, fields map[string]any) Future[nodis.ObjectID] {
	fields = cloneFields(fields)
	done := make(chan struct{})
	result := &Result[nodis.ObjectID]{}
	go func() { result.Value, result.Err = r.Create(ctx, document, fields); close(done) }()
	return Future[nodis.ObjectID]{done: done, result: result}
}

func (r *Remote) UpdateAsync(ctx context.Context, document nodis.DocumentID, object nodis.ObjectID, revision uint64, fields map[string]any) Future[struct{}] {
	fields = cloneFields(fields)
	done := make(chan struct{})
	result := &Result[struct{}]{}
	go func() { result.Err = r.Update(ctx, document, object, revision, fields); close(done) }()
	return Future[struct{}]{done: done, result: result}
}

func (r *Remote) DeleteAsync(ctx context.Context, document nodis.DocumentID, object nodis.ObjectID, revision uint64) Future[struct{}] {
	done := make(chan struct{})
	result := &Result[struct{}]{}
	go func() { result.Err = r.Delete(ctx, document, object, revision); close(done) }()
	return Future[struct{}]{done: done, result: result}
}

func appendRemoteMutation(dst []byte, document nodis.DocumentID, object nodis.ObjectID, revision uint64, fields map[string]any, update bool) ([]byte, error) {
	dst = binary.LittleEndian.AppendUint64(dst, uint64(document))
	if update {
		dst = binary.LittleEndian.AppendUint64(dst, uint64(object))
		dst = binary.LittleEndian.AppendUint64(dst, revision)
	}
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)
	dst = binary.AppendUvarint(dst, uint64(len(names)))
	for _, name := range names {
		value, err := nodis.ValueOf(fields[name])
		if err != nil {
			return nil, err
		}
		dst = binary.AppendUvarint(dst, uint64(len(name)))
		dst = append(dst, name...)
		dst = nodis.AppendValue(dst, value)
	}
	return dst, nil
}

func statusError(status protocol.Status) error {
	switch status {
	case protocol.StatusUnauthorized, protocol.StatusForbidden:
		return nodis.ErrPermission
	case protocol.StatusNotFound:
		return nodis.ErrNotFound
	case protocol.StatusConflict:
		return nodis.ErrConflict
	case protocol.StatusInvalid:
		return nodis.ErrInvalidValue
	case protocol.StatusTooLarge:
		return nodis.ErrQueryTooLarge
	default:
		return errors.New("nodis: remote server error")
	}
}
