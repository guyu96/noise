package broadcast

import (
	"github.com/guyu96/noise"
	"github.com/guyu96/noise/payload"
	kad "github.com/guyu96/noise/skademlia"
	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"
)

// Message is a message to be broadcasted with custom data.
type Message struct {
	From      kad.ID
	PrefixLen uint16
	Data      []byte
	Hash      [32]byte
}

func (m Message) Write() []byte {
	writer := payload.NewWriter(nil)
	writer.Write(m.From.Write())
	writer.WriteUint16(m.PrefixLen)
	writer.WriteUint32(uint32(len(m.Data))) // Note: need to write data byte length for reader.ReadBytes method to work
	writer.Write(m.Data)
	writer.Write(m.Hash[:])
	return writer.Bytes()
}

func (m Message) Read(reader payload.Reader) (noise.Message, error) {
	From, err := kad.ID{}.Read(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode From ID")
	}
	m.From = From.(kad.ID)

	prefixlen, err := reader.ReadUint16()
	if err != nil {
		return nil, err
	}
	m.PrefixLen = prefixlen

	data, err := reader.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.Data = data

	var hash [32]byte
	for i := 0; i < 32; i++ {
		h, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		hash[i] = h
	}
	m.Hash = hash

	return m, err
}

// Note: only hash the broadcaster and the data, not prefixLen because the same broadcast msg can be rebroadcasted by multiple parties.
// QUESTION: can add a timestamp, but spamming?
func (m *Message) generateHash() {
	bytes := append(m.From.Hash(), m.Data...)
	m.Hash = blake2b.Sum256(bytes)
}

// NewMessage creates a new Message instance.
func NewMessage(from kad.ID, prefixlen uint16, data []byte) *Message {
	msg := &Message{
		From:      from,
		PrefixLen: prefixlen,
		Data:      data,
	}
	msg.generateHash()
	return msg
}
