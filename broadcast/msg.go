package broadcast

import (
	"github.com/guyu96/noise"
	"github.com/guyu96/noise/payload"
	kad "github.com/guyu96/noise/skademlia"
	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"
)

const (
	hashSize = blake2b.Size256
)

// Message is a message to be broadcasted with custom data.
type Message struct {
	From      kad.ID
	PrefixLen uint16
	Data      []byte
	Hash      [hashSize]byte
	// No SeenPeers (unlike relay message) because the number of seen peers could be much larger.
}

func (msg Message) Write() []byte {
	writer := payload.NewWriter(nil)
	writer.Write(msg.From.Write())
	writer.WriteUint16(msg.PrefixLen)
	// Need to specify number of bytes for reader.ReadBytes method to work
	writer.WriteUint32(uint32(len(msg.Data)))
	writer.Write(msg.Data)
	writer.Write(msg.Hash[:])
	return writer.Bytes()
}

func (msg Message) Read(reader payload.Reader) (noise.Message, error) {
	From, err := kad.ID{}.Read(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode From ID")
	}
	msg.From = From.(kad.ID)

	prefixlen, err := reader.ReadUint16()
	if err != nil {
		return nil, err
	}
	msg.PrefixLen = prefixlen

	data, err := reader.ReadBytes()
	if err != nil {
		return nil, err
	}
	msg.Data = data

	var hash [hashSize]byte
	for i := 0; i < hashSize; i++ {
		h, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		hash[i] = h
	}
	msg.Hash = hash

	return msg, nil
}

// Note: only hash the broadcaster and the data, not prefixLen because the same broadcast msg can be rebroadcasted by multiple parties.
func (msg *Message) generateHash() {
	bytes := append(msg.From.Hash(), msg.Data...)
	msg.Hash = blake2b.Sum256(bytes)
}

// NewMessage creates a new Message instance.
func NewMessage(from kad.ID, prefixlen uint16, data []byte) Message {
	msg := Message{
		From:      from,
		PrefixLen: prefixlen,
		Data:      data,
	}
	msg.generateHash()
	return msg
}
