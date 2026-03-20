// Package wire provides length-prefixed protobuf message framing over
// io.Reader / io.Writer, suitable for raw TCP connections.
package wire

import (
	"encoding/binary"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
)

// MaxMessageSize is the upper bound on a single framed message (16 MB).
const MaxMessageSize = 16 << 20

// WriteMessage serialises msg as a length-prefixed protobuf frame:
// 4-byte big-endian uint32 payload length, then the protobuf bytes.
func WriteMessage(w io.Writer, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("wire: marshal: %w", err)
	}

	if len(data) > MaxMessageSize {
		return fmt.Errorf("wire: message too large (%d bytes, max %d)", len(data), MaxMessageSize)
	}

	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(len(data))) //nolint:gosec // len checked above

	if _, err := w.Write(header); err != nil {
		return fmt.Errorf("wire: write header: %w", err)
	}

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("wire: write payload: %w", err)
	}

	return nil
}

// ReadMessage reads a length-prefixed protobuf frame from r and
// unmarshals the payload into msg.
func ReadMessage(r io.Reader, msg proto.Message) error {
	header := make([]byte, 4)
	if _, err := io.ReadFull(r, header); err != nil {
		return fmt.Errorf("wire: read header: %w", err)
	}

	size := binary.BigEndian.Uint32(header)
	if size > MaxMessageSize {
		return fmt.Errorf("wire: message too large (%d bytes, max %d)", size, MaxMessageSize)
	}

	data := make([]byte, size)
	if _, err := io.ReadFull(r, data); err != nil {
		return fmt.Errorf("wire: read payload: %w", err)
	}

	if err := proto.Unmarshal(data, msg); err != nil {
		return fmt.Errorf("wire: unmarshal: %w", err)
	}

	return nil
}
