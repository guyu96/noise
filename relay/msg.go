package relay

import (
	"github.com/guyu96/noise"
	"github.com/guyu96/noise/payload"
	kad "github.com/guyu96/noise/skademlia"
	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"
)

// Message is a message to be relayed to To node with custom data.
type Message struct {
	From kad.ID
	To   kad.ID
	Hash [32]byte
	Data []byte
}

func (m Message) Write() []byte {
	writer := payload.NewWriter(nil)
	writer.Write(m.From.Write())
	writer.Write(m.To.Write())
	writer.Write(m.Hash[:])
	writer.WriteUint32(uint32(len(m.Data))) // Note: need to write data byte length for reader.ReadBytes method to work
	writer.Write(m.Data)
	return writer.Bytes()
}

func (m Message) Read(reader payload.Reader) (noise.Message, error) {
	From, err := kad.ID{}.Read(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode From ID")
	}
	m.From = From.(kad.ID)

	To, err := kad.ID{}.Read(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode To ID")
	}
	m.To = To.(kad.ID)

	var hash [32]byte
	for i := 0; i < 32; i++ {
		h, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		hash[i] = h
	}
	m.Hash = hash

	data, err := reader.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.Data = data

	return m, err
}

func (m *Message) generateHash() {
	bytes := append(m.From.Hash(), m.To.Hash()...)
	bytes = append(bytes, m.Data...)
	m.Hash = blake2b.Sum256(bytes)
}

// NewMessage creates a new Message instance.
func NewMessage(from kad.ID, to kad.ID, data []byte) *Message {
	msg := &Message{
		From: from,
		To:   to,
		Data: data,
	}
	msg.generateHash()
	return msg
}
